package fsm

import (
	"errors"

	"github.com/findy-network/findy-agent-api/grpc/agency"
)

const (
	TriggerTypeOurMessage = "OUR_STATUS"
	TriggerTypeUseInput   = "INPUT"
)

type Machine struct {
	Initial string           `json:"initial"`
	States  map[string]State `json:"states"`

	Current     string `json:"-"`
	Initialized bool   `json:"-"`
}

type State struct {
	ID          string       `json:"id"`
	Transitions []Transition `json:"transitions"`
}

type Transition struct {
	Trigger Event `json:"trigger"`
	// todo: we will allow only one at the time because we get status back!?
	// or we do nothing, we send nothing when we get our own message status
	// still, even we don't send nothing, it doesn't mean that we don't do
	// transition
	Sends  []Event `json:"sends,omitempty"`
	Target string  `json:"target"`
	// Script, or something to execute
}

type Event struct {
	TypeID string `json:"type_id"`
	Rule   string `json:"rule"`

	// todo: all these should be inside own struct that they can handled together
	BasicMessage *BasicMessage `json:"basic_message"`

	ProtocolType agency.Protocol_Type `json:"-"`
}

type BasicMessage struct {
	Content string
}

// Initialize initializes and optimizes the state machine because the JSON is
// meant for humans to write and machines to read. Initialize also moves machine
// to the initial state. It returns error if machine has them.
func (m *Machine) Initialize() (err error) {
	initSet := false
	for i := range m.States {
		for j := range m.States[i].Transitions {
			m.States[i].Transitions[j].Trigger.ProtocolType =
				protocolType[m.States[i].Transitions[j].Trigger.TypeID]
			for k := range m.States[i].Transitions[j].Sends {
				m.States[i].Transitions[j].Sends[k].ProtocolType =
					protocolType[m.States[i].Transitions[j].Sends[k].TypeID]
			}
		}
		if m.States[i].ID == m.Initial {
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

var protocolType = map[string]agency.Protocol_Type{
	"connection":    agency.Protocol_CONNECT,
	"issue_cred":    agency.Protocol_ISSUE,
	"present_proof": agency.Protocol_PROOF,
	"trust_ping":    agency.Protocol_TRUST_PING,
	"basic_message": agency.Protocol_BASIC_MESSAGE,
}
