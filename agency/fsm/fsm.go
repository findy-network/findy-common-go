package fsm

import (
	"errors"
	"fmt"

	"github.com/findy-network/findy-agent-api/grpc/agency"
)

const (
	TriggerTypeOurMessage = "OUR_STATUS"
	TriggerTypeUseInput   = "INPUT"
	TriggerTypeFormat     = "FORMAT"
	TriggerTypeData       = ""
)

type Machine struct {
	Initial string           `json:"initial"`
	States  map[string]State `json:"states"`

	Current     string `json:"-"`
	Initialized bool   `json:"-"`
}

type State struct {
	ID string `json:"id"` // todo: no need, will removed, id is map key

	Transitions []Transition `json:"transitions"`

	// we could have onEntry and OnExit ? If that would help, we shall see
}

type Transition struct {
	Trigger Event `json:"trigger"`

	// todo: we will allow only one at the time because we get status back!?
	//  or we do nothing, we send nothing when we get our own message status
	//  still, even we don't send nothing, it doesn't mean that we don't do
	//  transition
	//  maybe we could send many but wait only one, what it would mean when we
	//  get many status messages about our own sending? would it be much better
	//  to have one clear step at the time, and then that many receiver states
	//  amount of them is not the issue?! Keep it like it is for now
	Sends []Event `json:"sends,omitempty"`

	Target string `json:"target"`
	// Script, or something to execute
}

type Event struct {
	TypeID string `json:"type_id"`
	Rule   string `json:"rule"`
	Data   string `json:"data,omitempty"`

	*EventData `json:"event_data,omitempty"`

	ProtocolType           agency.Protocol_Type `json:"-"`
	*agency.ProtocolStatus `json:"-"`
}

type EventData struct {
	BasicMessage *BasicMessage `json:"basic_message"`
}

type BasicMessage struct {
	Content string
}

// Initialize initializes and optimizes the state machine because the JSON is
// meant for humans to write and machines to read. Initialize also moves machine
// to the initial state. It returns error if machine has them.
func (m *Machine) Initialize() (err error) {
	initSet := false
	for id := range m.States {
		for j := range m.States[id].Transitions {
			m.States[id].Transitions[j].Trigger.ProtocolType =
				protocolType[m.States[id].Transitions[j].Trigger.TypeID]
			for k := range m.States[id].Transitions[j].Sends {
				m.States[id].Transitions[j].Sends[k].ProtocolType =
					protocolType[m.States[id].Transitions[j].Sends[k].TypeID]
			}
		}
		if id == m.Initial {
			if initSet {
				return errors.New("machine has two initial states")
			}
			m.Current = m.Initial
			initSet = true
		}
	}
	m.Initialized = true
	return nil
}

func (m *Machine) CurrentState() State {
	return m.States[m.Current]
}

// Triggers returns a transition if machine has it in its current state. If not
// it returns nil.
func (m *Machine) Triggers(ptype agency.Protocol_Type) *Transition {
	for _, transition := range m.CurrentState().Transitions {
		if transition.Trigger.ProtocolType == ptype {
			return &transition
		}
	}
	return nil
}

func (m *Machine) Step(t *Transition) {
	m.Current = t.Target
}

func (t *Transition) BuildSendEvents(status *agency.ProtocolStatus) []Event {
	input := t.buildInputEvent(status)

	sends := make([]Event, len(t.Sends))
	for i, send := range t.Sends {
		sends[i] = send
		switch send.Rule {
		case TriggerTypeUseInput:
			sends[i].EventData = input.EventData
		case TriggerTypeData:
			sends[i].EventData = &EventData{BasicMessage: &BasicMessage{
				Content: send.Data,
			}}
		case TriggerTypeFormat:
			sends[i].EventData = &EventData{BasicMessage: &BasicMessage{
				Content: fmt.Sprintf(send.Data, input.Data),
			}}
		}
	}
	return sends
}

func (t *Transition) buildInputEvent(status *agency.ProtocolStatus) Event {
	e := Event{
		ProtocolType:   status.GetState().ProtocolId.TypeId,
		ProtocolStatus: status,
	}
	switch status.GetState().ProtocolId.TypeId {
	case agency.Protocol_CONNECT:
		return e
	case agency.Protocol_BASIC_MESSAGE:
		switch t.Trigger.Rule {
		case TriggerTypeUseInput:
			e.Data = status.GetBasicMessage().Content
			e.EventData = &EventData{BasicMessage: &BasicMessage{
				Content: status.GetBasicMessage().Content,
			}}
		case TriggerTypeData:
			e.EventData = &EventData{BasicMessage: &BasicMessage{
				Content: t.Trigger.Data,
			}}
		}
	}
	return e
}

var protocolType = map[string]agency.Protocol_Type{
	"connection":    agency.Protocol_CONNECT,
	"issue_cred":    agency.Protocol_ISSUE,
	"present_proof": agency.Protocol_PROOF,
	"trust_ping":    agency.Protocol_TRUST_PING,
	"basic_message": agency.Protocol_BASIC_MESSAGE,
}
