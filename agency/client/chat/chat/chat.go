package chat

import (
	"context"

	agency "github.com/findy-network/findy-agent-api/grpc/agency/v1"
	"github.com/findy-network/findy-common-go/agency/client"
	"github.com/findy-network/findy-common-go/agency/client/async"
	"github.com/findy-network/findy-common-go/agency/fsm"
	"github.com/golang/glog"
	"github.com/lainio/err2"
)

type HookFn func(data map[string]string)

type ConnStatus *agency.AgentStatus

type StatusChan chan ConnStatus

type HookChan chan map[string]string

type Conversation struct {
	StatusChan
	HookChan

	id string
	client.Conn
	lastProtocolID map[string]struct{} //*agency.ProtocolID

	// machine can be ptr because multiplexer creates a new for each one
	machine *fsm.Machine
}

// These are class level variables for this chat bot which means that every
// conversation of this bot will share these variables
var (
	// Status is input channel for multiplexing this chat bot i.e. CA sends
	// every message t this channel from here they are transported to according
	// conversations
	Status = make(StatusChan)

	// conversations is a map of all instances indexed by connection id aka
	// pairwise id
	conversations = make(map[string]*Conversation)

	// Machine is the initial finite-state machine from where every
	// conversations will be started
	Machine fsm.MachineData

	// Hook is function to be set from user of the chat bot. It will be called
	// when state machine has triggered transition where Hook should be called.
	Hook HookFn
)

func Multiplexer(conn client.Conn) {
	glog.V(4).Infoln("starting multiplexer", Machine.FType)
	for {
		t := <-Status
		c, ok := conversations[t.Notification.ConnectionID]
		if !ok {
			glog.V(5).Infoln("Starting new conversation",
				Machine.FType)
			c = &Conversation{
				id:         t.Notification.ConnectionID,
				Conn:       conn,
				StatusChan: make(StatusChan),
			}
			go c.RunConversation(Machine)
			conversations[t.Notification.ConnectionID] = c
		}
		c.StatusChan <- t
	}
}

func (c *Conversation) RunConversation(data fsm.MachineData) {
	c.machine = fsm.NewMachine(data)
	err2.Check(c.machine.Initialize())

	c.send(c.machine.Start(), nil)

	for {
		select {
		case t := <-c.StatusChan:
			c.statusReceived(t)
		case hookData := <-c.HookChan:
			c.hookReceived(hookData)
		}
	}
}

func (c *Conversation) hookReceived(hookData map[string]string) {
	glog.V(4).Infoln("hook data arriwed:", hookData)
	if transition := c.machine.TriggersByHook(); transition != nil {
		c.send(transition.BuildSendEventsFromHook(hookData), nil)
		c.machine.Step(transition)
	}
}

func (c *Conversation) statusReceived(as *agency.AgentStatus) {
	glog.V(10).Infoln("conversation:", as.Notification.ConnectionID)

	switch as.Notification.TypeID {
	case agency.Notification_STATUS_UPDATE:
		glog.V(3).Infoln("status update")
		if c.IsOursAndRm(as.Notification.ProtocolID) {
			glog.V(10).Infoln("discarding event")
			return
		}

		// translate notification to actual status data
		status := c.getStatus(as)

		if transition := c.machine.Triggers(status); transition != nil {
			// TODO: different transitions to FSM, move error handling to it!
			if status.GetState().State != agency.ProtocolState_OK {
				glog.Warningln("FSM steps only completed protocol steps",
					status.GetState().State)
				return
			}
			if glog.V(3) {
				glog.Infof("machine: %s (%p)", c.machine.Name, c.machine)
				glog.Infoln("role:", status.GetState().ProtocolID.Role)
				glog.Infoln("TRiGGERiNG", transition.Trigger.ProtocolType)
			}

			c.send(transition.BuildSendEvents(status), as)
			c.machine.Step(transition)
		} else {
			glog.V(1).Infoln("machine don't have transition for:",
				as.Notification.ProtocolType)
		}
	case agency.Notification_ANSWER_NEEDED_PROOF_VERIFY:
		glog.V(1).Infof("- %s: proof QA (%p)", c.machine.Name,
			c.machine)
		if transition := c.machine.Answers(as); transition != nil {
			c.send(transition.BuildSendAnswers(as), as)
			c.machine.Step(transition)
		}
	}
}

func (c *Conversation) getStatus(status ConnStatus) *agency.ProtocolStatus {
	ctx := context.Background()
	didComm := agency.NewProtocolServiceClient(c.Conn)
	statusResult, err := didComm.Status(ctx, &agency.ProtocolID{
		TypeID:           status.Notification.ProtocolType,
		ID:               status.Notification.ProtocolID,
		NotificationTime: status.Notification.Timestamp,
	})
	err2.Check(err)
	return statusResult
}

func (c *Conversation) reply(status *agency.AgentStatus, ack bool) {
	ctx := context.Background()
	agentClient := agency.NewAgentServiceClient(c.Conn)
	cid, err := agentClient.Give(ctx, &agency.Answer{
		ID:       status.Notification.ID,
		ClientID: status.ClientID,
		Ack:      ack,
		Info:     "testing says hello!",
	})
	err2.Check(err)
	glog.V(3).Infof("Sending the answer (%s) send to client:%s\n",
		status.Notification.ID, cid.ID)
}

func (c *Conversation) sendBasicMessage(message *fsm.BasicMessage, noAck bool) {
	r, err := async.NewPairwise(
		c.Conn,
		c.id,
	).BasicMessage(context.Background(),
		message.Content)
	err2.Check(err)
	glog.V(10).Infoln("protocol id:", r.ID)
	if noAck {
		c.SetLastProtocolID(r)
	}
}

func (c *Conversation) sendIssuing(message *fsm.Issuing, noAck bool) {
	r, err := async.NewPairwise(
		c.Conn,
		c.id,
	).Issue(context.Background(),
		message.CredDefID, message.AttrsJSON)
	err2.Check(err)
	glog.V(10).Infoln("protocol id:", r.ID)
	if noAck {
		c.SetLastProtocolID(r)
	}
}

func (c *Conversation) sendReqProof(message *fsm.Proof, noAck bool) {
	r, err := async.NewPairwise(
		c.Conn,
		c.id,
	).ReqProof(context.Background(),
		message.ProofJSON)
	err2.Check(err)
	glog.V(10).Infoln("protocol id:", r.ID)
	if noAck {
		c.SetLastProtocolID(r)
	}
}

func (c *Conversation) sendHook(hookData *fsm.Hook, _ bool) {
	// TODO: implement call real hook implementation here
	glog.V(0).Infoln("call hook implementation")
	callHook(hookData.Data)
}

func callHook(hookData map[string]string) {
	if Hook != nil {
		glog.V(3).Infoln("calling hook")
		Hook(hookData)
	} else {
		glog.Warningln("call hook but hook not set")
	}
}

func (c *Conversation) sendEmail(message *fsm.Email, _ bool) {
	// TODO: implement send email here
	glog.V(0).Infoln("sending email to", message.To, message.Body)
}

func (c *Conversation) SetLastProtocolID(pid *agency.ProtocolID) {
	if c.lastProtocolID == nil {
		c.lastProtocolID = make(map[string]struct{})
	}
	c.lastProtocolID[pid.ID] = struct{}{}
}

func (c *Conversation) send(outputs []*fsm.Event, status ConnStatus) {
	if outputs == nil {
		return
	}
	for _, output := range outputs {
		switch output.ProtocolType {
		case agency.Protocol_DIDEXCHANGE:
			glog.Warningf("we should not be here!!")
		case agency.Protocol_BASIC_MESSAGE:
			c.sendBasicMessage(output.BasicMessage, output.NoStatus)
		case agency.Protocol_ISSUE_CREDENTIAL:
			c.sendIssuing(output.Issuing, output.NoStatus)
		case agency.Protocol_PRESENT_PROOF:
			c.sendReqProof(output.Proof, output.NoStatus)
		case fsm.EmailProtocol:
			c.sendEmail(output.Email, output.NoStatus)
		case fsm.QAProtocol:
			if status == nil {
				panic("FSM syntax error")
			}
			ack := false
			if output.Data == "ACK" {
				ack = true
			}
			c.reply(status, ack)
		case fsm.HookProtocol:
			c.sendHook(output.Hook, false)
		}
	}
}

func (c *Conversation) IsOursAndRm(id string) bool {
	if _, ok := c.lastProtocolID[id]; ok {
		c.lastProtocolID[id] = struct{}{}
		return true
	}
	return false
}
