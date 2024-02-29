package chat

import (
	"context"
	"os"
	"syscall"

	"github.com/findy-network/findy-common-go/agency/client"
	"github.com/findy-network/findy-common-go/agency/client/async"
	"github.com/findy-network/findy-common-go/agency/fsm"
	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/golang/glog"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

type HookFn func(data map[string]string)

type QuestionStatus *agency.Question
type ConnStatus *agency.AgentStatus

type StatusChan chan ConnStatus
type QuestionChan chan QuestionStatus

type HookChan chan map[string]string

// Backend is optional state-machine for service level. We use backend name
// because service is too generic and we want to underline that Conversations
// are in the front and backend has our back!
type Backend struct {
	fsm.TerminateChan // FSM tells us if machine has reached the end.
	fsm.BackendChan

	// machine can be ptr because multiplexer creates a new for each one
	machine *fsm.Machine
}

type Conversation struct {
	StatusChan
	QuestionChan
	HookChan
	fsm.TerminateChan // FSM tells us if machine has reached the end.
	fsm.BackendChan
	TransientChan fsm.TransientChan

	id string
	client.Conn
	lastProtocolID map[string]struct{} // *agency.ProtocolID

	// machine can be ptr because multiplexer creates a new for each one
	machine *fsm.Machine
}

// These are class level variables for this chat bot which means that every
// conversation of this bot will share these variables
var (
	ConversationBackendChan = make(fsm.BackendChan, 1)

	// Status is input channel for multiplexing this chat bot i.e. CA sends
	// every message to this channel from here they are transported to
	// according conversations
	Status = make(StatusChan)

	// Question is input channel for multiplexing this chat bot i.e. CA sends
	// every question to this channel.
	Question = make(QuestionChan)

	// conversations is a map of all instances indexed by connection id aka
	// pairwise id.
	// thread-safe because the multiplexer gorountine is the only one who
	// processes the map.
	conversations = make(map[string]*Conversation)

	// backendMachine is the only instance needed, if needed, i.e. 0..1
	// created by function RunBackendService.
	backendMachine *Backend

	// SharedMem is memory register to be shared between all conversations.
	// It can be used e.g. gather information.. not sure if this is needed.
	SharedMem = make(map[string]string)

	// Hook is function to be set from user of the chat bot. It will be called
	// when state machine has triggered transition where Hook should be called.
	// Hook allows process who is running FSM extend it with a callback.
	Hook HookFn
)

func newBackendService() *Backend {
	backendMachine = &Backend{
		TerminateChan: make(chan bool),
		BackendChan:   make(fsm.BackendChan, 1),
	}
	return backendMachine
}

type MultiplexerInfo struct {
	Conn                client.Conn
	InterruptCh         chan<- os.Signal
	ConversationMachine fsm.MachineData
	BackendMachine      *fsm.MachineData
}

// Multiplexer is a goroutine function to started multiplex all the
// conversations an agent is currently having. It takes a gRPC connection handle
// and a signaling channel as an arguments. The second argument, the interrupt
// channel, is needed to tell when the multiplexer has reached to the end.
// NOTE. Most of the state machines never do that.
func Multiplexer(info MultiplexerInfo) {
	glog.V(3).Infoln("starting multiplexer", info.ConversationMachine.FType)
	termChan := make(fsm.TerminateChan, 1)

	var backendChan fsm.BackendInChan
	if info.BackendMachine.IsValid() {
		b := newBackendService()
		backendChan = b.BackendChan
		b.machine = fsm.NewBackendMachine(*info.BackendMachine)
		try.To(b.machine.Initialize())
		b.machine.InitLua()

		glog.V(2).Infoln("starting and send first step:", info.BackendMachine.FType)
		b.send(b.machine.Start(fsm.TerminateOutChan(b.TerminateChan)))
		glog.V(2).Infoln("going to for loop:", info.BackendMachine.FType)
	}

	for {
		select {
		// we serve backendMachine here if we have a REAL channel, see above
		// NOTE. It's OK to listen nil channel (especially) in select.
		case bd := <-backendChan:
			backendMachine.backendReceived(bd)

		case d := <-ConversationBackendChan:
			c, ok := conversations[d.ConnID]
			assert.That(ok, "backend msgs to existing conversations only")
			c.BackendChan <- d
		case t := <-Status:
			connID := t.Notification.ConnectionID
			c, ok := conversations[connID]
			if !ok {
				c = newConversation(info, connID, termChan)
			}
			c.StatusChan <- t
		case question := <-Question:
			connID := question.Status.Notification.ConnectionID
			c, ok := conversations[connID]
			if !ok {
				c = newConversation(info, connID, termChan)
			}
			c.QuestionChan <- question
		case <-termChan:
			// One machine has reached its terminate state. Let's signal
			// outside that the whole system is ready to stop.
			//
			// We have two (2) different types of mahines:
			// 1. Multiplexer (class lvl) machine that includes rules what to
			// do at service/process lvl.
			// 2. The pairwise lvl, aka conversation FSMs
			//
			// In future we could offer an API to what to send to
			// the intCh. SIGTERM is very good compromise in K8s, etc.
			glog.V(1).Infoln("<- FSM Terminate signal received")
			info.InterruptCh <- syscall.SIGTERM
			glog.V(1).Infoln("-> signaled SIGTERM")
		}
		// TODO HookChan handler isn't implemented yet!
	}
}

func newConversation(
	info MultiplexerInfo,
	connID string,
	termChan fsm.TerminateChan,
) *Conversation {
	glog.V(5).Infoln("Starting new conversation",
		info.ConversationMachine.FType)
	c := &Conversation{
		id:            connID,
		Conn:          info.Conn,
		StatusChan:    make(StatusChan),
		QuestionChan:  make(QuestionChan),
		HookChan:      make(HookChan),
		BackendChan:   make(fsm.BackendChan, 1),
		TransientChan: make(fsm.TransientChan, 1),
		TerminateChan: termChan,
	}
	conversations[connID] = c
	go c.Run(info.ConversationMachine)
	return c
}

func (b *Backend) backendReceived(data *fsm.BackendData) {
	glog.V(2).Infof("+++ b-fsm data(%v):%v", data, b.machine.Type)
	assert.Equal(b.machine.Type, fsm.MachineTypeBackend)
	if transition := b.machine.TriggersByBackendData(data); transition != nil {
		b.send(transition.BuildSendEventsFromBackendData(data))
		b.machine.Step(transition)
	}
}

func (b *Backend) send(outputs []*fsm.Event) {
	if outputs == nil {
		return
	}
	for _, output := range outputs {
		switch output.ProtocolType {
		case fsm.BackendProtocol:
			b.sendBackendData(output.EventData.Backend, false)
		}
	}
}

func (b *Backend) sendBackendData(data *fsm.BackendData, _ bool) {
	for _, conversation := range conversations {
		glog.V(2).Infof("b-fsm-> BackendData:%v", data)
		conversation.BackendChan <- data
	}
}

func (c *Conversation) Run(data fsm.MachineData) {
	c.machine = fsm.NewMachine(data)
	try.To(c.machine.Initialize())
	c.machine.ConnID = c.id // conversation machines need ConnectionID
	c.machine.InitLua()
	c.send(c.machine.Start(fsm.TerminateOutChan(c.TerminateChan)), nil)

	for {
		select {
		case t := <-c.StatusChan:
			c.statusReceived(t)
		case q := <-c.QuestionChan:
			c.questionReceived(q)
		case hookData := <-c.HookChan:
			c.hookReceived(hookData)
		case backendData := <-c.BackendChan:
			c.backendReceived(backendData)
		case stepData := <-c.TransientChan:
			c.stepReceived(stepData)
		}
	}
}

func (c *Conversation) stepReceived(data string) {
	glog.V(3).Infoln("conversation: step w/ str:", data)
	if transition := c.machine.TriggersByStep(); transition != nil {
		c.send(transition.BuildSendEventsFromStep(data), nil)
		c.machine.Step(transition)
	}
}

func (c *Conversation) backendReceived(data *fsm.BackendData) {
	glog.V(2).Infof("+++ b-fsm data(%v):%v", data, c.machine.Type)
	if data.ConnID == "" {
		// TODO: maybe this is only place to se it? First opportunity? It
		// seems that sender cannot get ConnID at b-fsm because the
		// `keep_memory` isn't `true` in all machines. Maybe rm Warning.
		glog.Warningln("!!! ConnID is empty, fixing !!!")
		data.ConnID = c.id
	}
	sessionID, weHaveSessionID := c.machine.Memory[fsm.LUA_SESSION_ID]
	glog.V(3).Infof("conversation: backend w/ content: %v, type:%v, SID:%v",
		data.Content, c.machine.Type, sessionID)
	if weHaveSessionID && sessionID != data.SessionID {
		glog.V(1).Infof("--- (f/my: %v != b-fsm: %v)",
			sessionID, data.SessionID)
		return
	} else if data.NoEcho && c.id == data.ConnID {
		glog.V(1).Infoln("--- no echo")
		return
	}

	assert.Equal(c.machine.Type, fsm.MachineTypeConversation)
	if transition := c.machine.TriggersByBackendData(data); transition != nil {
		c.send(transition.BuildSendEventsFromBackendData(data), nil)
		c.machine.Step(transition)
	}
}

func (c *Conversation) hookReceived(hookData map[string]string) {
	glog.V(4).Infoln("hook data arriwed:", hookData)
	if transition := c.machine.TriggersByHook(); transition != nil {
		c.send(transition.BuildSendEventsFromHook(hookData), nil)
		c.machine.Step(transition)
	}
}

func (c *Conversation) questionReceived(q QuestionStatus) {
	glog.V(10).Infoln("conversation:", q.Status.Notification.ConnectionID)

	switch q.TypeID {
	case agency.Question_PROOF_VERIFY_WAITS:
		glog.V(1).Infof("- %s: proof QA (%p)", c.machine.Name,
			c.machine)
		if transition := c.machine.Answers(q); transition != nil {
			c.send(transition.BuildSendAnswers(q.Status), q.Status)
			c.machine.Step(transition)
		}
	}
}

func (c *Conversation) statusReceived(as *agency.AgentStatus) {
	glog.V(10).Infoln("conversation:", as.Notification.ConnectionID)

	switch as.Notification.TypeID {
	case agency.Notification_STATUS_UPDATE:
		glog.V(3).Infoln("status update:", as.Notification.GetProtocolType())
		if c.isOursAndRm(as.Notification.ProtocolID) {
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
				glog.Infof("machine: %s, ptr(%p)", c.id[:8], c.machine)
				glog.Infoln("role:", status.GetState().ProtocolID.Role)
				glog.Infoln("TRiGGERiNG", transition.Trigger.ProtocolType)
			}

			c.send(transition.BuildSendEvents(status), as)
			c.machine.Step(transition)
		} else {
			glog.V(1).Infoln("machine doesn't have transition for:",
				as.Notification.ProtocolType)
		}
	}
}

func (c *Conversation) getStatus(status ConnStatus) *agency.ProtocolStatus {
	ctx := context.Background()
	didComm := agency.NewProtocolServiceClient(c.Conn)
	statusResult := try.To1(didComm.Status(ctx, &agency.ProtocolID{
		TypeID:           status.Notification.ProtocolType,
		ID:               status.Notification.ProtocolID,
		NotificationTime: status.Notification.Timestamp,
	}))
	return statusResult
}

func (c *Conversation) reply(status *agency.AgentStatus, ack bool) {
	ctx := context.Background()
	agentClient := agency.NewAgentServiceClient(c.Conn)
	cid := try.To1(agentClient.Give(ctx, &agency.Answer{
		ID:       status.Notification.ID,
		ClientID: status.ClientID,
		Ack:      ack,
		Info:     "testing says hello!",
	}))
	glog.V(3).Infof("+++ Sending the answer (%s) send to client:%s\n",
		status.Notification.ID, cid.ID)
}

func (c *Conversation) sendBasicMessage(message *fsm.BasicMessage, wantStatus bool) {
	glog.V(10).Infoln("start sendBasicMessage")
	r := try.To1(async.NewPairwise(
		c.Conn,
		c.id,
	).BasicMessage(context.Background(),
		message.Content))
	glog.V(10).Infoln("protocol id:", r.ID)
	if !wantStatus {
		c.SetLastProtocolID(r)
	}
}

func (c *Conversation) sendIssuing(message *fsm.Issuing, wantStatus bool) {
	r := try.To1(async.NewPairwise(
		c.Conn,
		c.id,
	).Issue(context.Background(),
		message.CredDefID, message.AttrsJSON))
	glog.V(10).Infoln("protocol id:", r.ID)
	if !wantStatus {
		c.SetLastProtocolID(r)
	}
}

func (c *Conversation) sendReqProof(message *fsm.Proof, wantStatus bool) {
	glog.V(5).Infoln("+++ message.ProofJSON", message.ProofJSON)
	r := try.To1(async.NewPairwise(
		c.Conn,
		c.id,
	).ReqProof(context.Background(),
		message.ProofJSON))
	glog.V(10).Infoln("protocol id:", r.ID)
	if !wantStatus {
		c.SetLastProtocolID(r)
	}
}

func (c *Conversation) sendHook(hookData *fsm.Hook, _ bool) {
	// TODO: implement call real hook implementation here
	glog.V(0).Infoln("call hook implementation")
	callHook(hookData.Data)
}

func (c *Conversation) sendTransient(msg string, _ bool) {
	c.TransientChan <- msg
}

func callHook(hookData map[string]string) {
	if Hook != nil {
		glog.V(3).Infoln("calling hook")
		Hook(hookData)
	} else {
		glog.Warningln("call hook but hook not set")
	}
}

func (c *Conversation) sendBackend(data *fsm.BackendData, _ bool) {
	glog.V(2).Infof("+++ f-fsm-> sending backend:%v", data)
	if data.ConnID == "" {
		// TODO: maybe this is only place to se it? First opportunity?
		glog.Warningln("!!! ConnID is empty, fixing !!!")
		data.ConnID = c.id
	}
	if backendMachine != nil {
		backendMachine.BackendChan <- data
	} else {
		glog.V(0).Infoln("!!! cannot send message to Service FSM")
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
			assert.Equal(output.ProtocolType, agency.Protocol_BASIC_MESSAGE)
			assert.NotNil(output.EventData)
			assert.NotNil(output.EventData.BasicMessage)
			c.sendBasicMessage(output.BasicMessage, output.WantStatus)
		case agency.Protocol_ISSUE_CREDENTIAL:
			c.sendIssuing(output.Issuing, output.WantStatus)
		case agency.Protocol_PRESENT_PROOF:
			c.sendReqProof(output.Proof, output.WantStatus)
		case fsm.BackendProtocol:
			c.sendBackend(output.Backend, output.WantStatus)
		case fsm.EmailProtocol:
			c.sendEmail(output.Email, output.WantStatus)
		case fsm.QAProtocol:
			assert.NotNil(status, "FSM syntax error")

			ack := false
			if output.Data == "ACK" {
				ack = true
			}
			c.reply(status, ack)
		case fsm.HookProtocol:
			c.sendHook(output.Hook, false)
		case fsm.TransientProtocol:
			c.sendTransient(output.BasicMessage.Content, false)
		}
	}
}

func (c *Conversation) isOursAndRm(id string) bool {
	if _, ok := c.lastProtocolID[id]; ok {
		delete(c.lastProtocolID, id)
		return true
	}
	return false
}
