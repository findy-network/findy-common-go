package chat

import (
	"context"

	"github.com/findy-network/findy-agent-api/grpc/agency"
	"github.com/findy-network/findy-grpc/agency/client"
	"github.com/findy-network/findy-grpc/agency/client/async"
	"github.com/findy-network/findy-grpc/agency/fsm"
	"github.com/golang/glog"
	"github.com/lainio/err2"
)

// todo: check what of these types need to be exported.

type ConnStatus *agency.AgentStatus

type StatusChan chan ConnStatus

type Conversation struct {
	id string
	client.Conn
	lastProtocolID map[string]struct{} //*agency.ProtocolID
	StatusChan
	fsm.Machine
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
	Machine *fsm.Machine
)

func Multiplexer(conn client.Conn) {
	glog.V(3).Infoln("starting multiplexer")
	for {
		t := <-Status
		c, ok := conversations[t.Notification.ConnectionId]
		if !ok {
			glog.V(1).Infoln("Starting new conversation")
			c = &Conversation{
				id:         t.Notification.ConnectionId,
				Conn:       conn,
				StatusChan: make(StatusChan),
				Machine:    *Machine,
			}
			go c.RunConversation()
			conversations[t.Notification.ConnectionId] = c
		}
		c.StatusChan <- t
	}
}

func (c *Conversation) RunConversation() {
	for {
		t := <-c.StatusChan

		glog.V(3).Infoln("conversation:", t.Notification.ConnectionId)

		if c.IsOursAndRm(t.Notification.ProtocolId) {
			glog.V(1).Infoln("discarding event")
			continue
		}

		if transition := c.Machine.Triggers(t.Notification.ProtocolType); transition != nil {
			status := c.getStatus(t)

			if status.GetState().State != agency.ProtocolState_OK {
				glog.Warningln("current FSM steps only completed protocol steps", status.GetState().State)
				continue
			}
			glog.V(1).Infoln("role:", status.GetState().ProtocolId.Role)

			glog.V(1).Infoln("TRiGGERiNG", transition.Trigger.ProtocolType)
			if transition.Trigger.Rule != fsm.TriggerTypeOurMessage { // todo: not used any more
				outputs := transition.BuildSendEvents(status)
				if outputs != nil {
					c.send(outputs)
				}
			} else {
				glog.V(1).Infoln("our message, just getting status")
			}
			c.Machine.Step(transition)
		} else {
			glog.V(1).Infoln("machine don't have transition for:",
				t.Notification.ProtocolType)
		}
	}
}

func (c *Conversation) getStatus(status ConnStatus) *agency.ProtocolStatus {
	ctx := context.Background()
	didComm := agency.NewDIDCommClient(c.Conn)
	statusResult, err := didComm.Status(ctx, &agency.ProtocolID{
		TypeId:           status.Notification.ProtocolType,
		Id:               status.Notification.ProtocolId,
		NotificationTime: status.Notification.Timestamp,
	})
	err2.Check(err)
	return statusResult
}

func (c *Conversation) sendBasicMessage(message *fsm.BasicMessage, noAck bool) {
	r, err := async.NewPairwise(
		c.Conn,
		c.id,
	).BasicMessage(context.Background(),
		message.Content)
	err2.Check(err)
	glog.V(1).Infoln("protocol id:", r.Id)
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
	glog.V(1).Infoln("protocol id:", r.Id)
	if noAck {
		c.SetLastProtocolID(r)
	}
}

func (c *Conversation) sendEmail(message *fsm.Email, noAck bool) {
	// todo: implement send email here
	glog.V(1).Infoln("sending email to", message.To, message.Body)
}

func (c *Conversation) SetLastProtocolID(pid *agency.ProtocolID) {
	if c.lastProtocolID == nil {
		c.lastProtocolID = make(map[string]struct{})
	}
	c.lastProtocolID[pid.Id] = struct{}{}
}

func (c *Conversation) send(outputs []fsm.Event) {
	for _, output := range outputs {
		switch output.ProtocolType {
		case agency.Protocol_CONNECT:
			glog.Warningf("we should not be here!!")
		case agency.Protocol_BASIC_MESSAGE:
			c.sendBasicMessage(output.BasicMessage, output.NoStatus)
		case agency.Protocol_ISSUE:
			c.sendIssuing(output.Issuing, output.NoStatus)
		case fsm.EmailProtocol:
			c.sendEmail(output.Email, output.NoStatus)
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
