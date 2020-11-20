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

type ConnStatus *agency.AgentStatus

type StatusChan chan ConnStatus

type Conversation struct {
	client.Conn
	lastProtocol *agency.ProtocolID
	StatusChan
	fsm.Machine
}

// should be move these to Bot? soon, baby soon, ..
var (
	Status = make(StatusChan)

	conversations = make(map[string]*Conversation)

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

		if transition := c.Machine.Triggers(t.Notification.ProtocolType); transition != nil {
			status := c.getStatus(t)

			if status.GetState().State != agency.ProtocolState_OK {
				glog.Warningln("current FSM steps only completed protocol steps", status.GetState().State)
				continue
			}
			glog.V(1).Infoln("role:", status.GetState().ProtocolId.Role)

			glog.V(1).Infoln("TRiGGERiNG", transition.Trigger.ProtocolType)
			if transition.Trigger.Rule != "OUR_STATUS" {
				event := c.buildMessage(status, transition)
				c.sendBasicMessage(t, event[0].BasicMessage)
			} else {
				glog.V(1).Infoln("our message, just getting status")
			}
			c.Machine.Step(transition)
		}
	}
}

func (c *Conversation) getStatus(status ConnStatus) *agency.ProtocolStatus {
	ctx := context.Background()
	didComm := agency.NewDIDCommClient(c.Conn)
	statusResult, err := didComm.Status(ctx, &agency.ProtocolID{
		TypeId: status.Notification.ProtocolType,

		// todo: we should test if this is really needed, if it isn't then this getter is totally general
		//Role: agency.Protocol_ADDRESSEE,

		Id:               status.Notification.ProtocolId,
		NotificationTime: status.Notification.Timestamp,
	})
	err2.Check(err)
	return statusResult
}

func (c *Conversation) buildMessage(statusResult *agency.ProtocolStatus, transition *fsm.Transition) []fsm.Event {
	event := fsm.Event{
		BasicMessage: &fsm.BasicMessage{Content: statusResult.GetBasicMessage().Content},
		ProtocolType: statusResult.GetState().ProtocolId.TypeId,
	}

	sends := make([]fsm.Event, len(transition.Sends))
	for i, send := range transition.Sends {
		sends[i] = send
		if send.Rule == "INPUT" {
			sends[i].BasicMessage = event.BasicMessage
		}
	}
	return sends
}

func (c *Conversation) sendBasicMessage(status ConnStatus, message *fsm.BasicMessage) {
	r, err := async.NewPairwise(
		c.Conn,
		status.Notification.ConnectionId,
	).BasicMessage(context.Background(),
		message.Content)
	err2.Check(err)
	glog.V(1).Infoln("protocol id:", r.Id)
	c.SetLastProtocolID(r)
}

func (c *Conversation) SetLastProtocolID(pid *agency.ProtocolID) {
	c.lastProtocol = pid
}
