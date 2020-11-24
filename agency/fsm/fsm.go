package fsm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"text/template"

	"github.com/findy-network/findy-agent-api/grpc/agency"
	"github.com/golang/glog"
	"github.com/lainio/err2"
)

const (
	TriggerTypeOurMessage    = "OUR_STATUS"
	TriggerTypeUseInput      = "INPUT"
	TriggerTypeUseInputSave  = "INPUT_SAVE"
	TriggerTypeFormat        = "FORMAT"
	TriggerTypeFormatFromMem = "FORMAT_MEM"
	TriggerTypePIN           = "GEN_PIN"
	TriggerTypeData          = ""

	TriggerTypeValidateInputEqual    = "INPUT_VALIDATE_EQUAL"
	TriggerTypeValidateInputNotEqual = "INPUT_VALIDATE_NOT_EQUAL"
	TriggerTypeInputEqual            = "INPUT_EQUAL"
)

const (
	MessageEmail        = "email"
	MessageBasicMessage = "basic_message"
	MessageIssueCred    = "issue_cred"
	MessageTrustPing    = "trust_ping"
	MessagePresentProof = "present_proof"
	MessageConnection   = "connection"
	MessageNone         = ""
)

const (
	EmailProtocol = 100
)

type Machine struct {
	Initial string           `json:"initial"`
	States  map[string]State `json:"states"`

	Current     string `json:"-"`
	Initialized bool   `json:"-"`

	Memory map[string]string `json:"-"`
}

type State struct {
	Transitions []Transition `json:"transitions"`

	// we could have onEntry and OnExit ? If that would help, we shall see
}

type Transition struct {
	Trigger Event `json:"trigger"`

	Sends []Event `json:"sends,omitempty"`

	Target string `json:"target"`

	// Script, or something to execute in future?? idea we could have LUA
	// script which communicates our Memory map, that would be a simple data
	// model

	Machine *Machine `json:"-"`
}

type Event struct {
	TypeID   string `json:"type_id"`
	Rule     string `json:"rule"`
	Data     string `json:"data,omitempty"`
	NoStatus bool   `json:"no_status,omitempty"`

	*EventData `json:"event_data,omitempty"`

	ProtocolType           agency.Protocol_Type `json:"-"`
	*agency.ProtocolStatus `json:"-"`

	FailTarget string `json:"fail_target,omitempty"`
	FailEvent  *Event `json:"fail_event,omitempty"`

	*Transition `json:"-"`
}

func (e Event) Triggers(status *agency.ProtocolStatus) bool {
	switch status.GetState().ProtocolId.TypeId {
	case agency.Protocol_ISSUE:
		return true
	case agency.Protocol_BASIC_MESSAGE:
		content := status.GetBasicMessage().Content
		switch e.Rule {
		case TriggerTypeValidateInputNotEqual:
			return e.Machine.Memory[e.Data] != content
		case TriggerTypeValidateInputEqual:
			return e.Machine.Memory[e.Data] == content
		case TriggerTypeInputEqual:
			return content == e.Data
		case TriggerTypeData, TriggerTypeUseInput, TriggerTypeUseInputSave:
			return true
		}
	}
	return false
}

type EventData struct {
	BasicMessage *BasicMessage `json:"basic_message,omitempty"`
	Issuing      *Issuing      `json:"issuing,omitempty"`
	Email        *Email        `json:"email,omitempty"`
}

type Email struct {
	To      string `json:"to,omitempty"`
	From    string `json:"from,omitempty"`
	Subject string `json:"subject,omitempty"`
	Body    string `json:"body,omitempty"`
}

type Issuing struct {
	CredDefID string
	AttrsJSON string
}

type BasicMessage struct {
	Content string
}

// Initialize initializes and optimizes the state machine because the JSON is
// meant for humans to write and machines to read. Initialize also moves machine
// to the initial state. It returns error if machine has them.
// todo: switch to pointers in everything, it's better for state machine, no value errors
func (m *Machine) Initialize() (err error) {
	initSet := false
	for id := range m.States {
		for j := range m.States[id].Transitions {
			m.States[id].Transitions[j].Trigger.Transition = &m.States[id].Transitions[j]
			m.States[id].Transitions[j].Trigger.ProtocolType =
				ProtocolType[m.States[id].Transitions[j].Trigger.TypeID]
			if m.States[id].Transitions[j].Trigger.FailEvent != nil {
				m.States[id].Transitions[j].Trigger.FailEvent.ProtocolType =
					ProtocolType[m.States[id].Transitions[j].Trigger.FailEvent.TypeID]
			}
			for k := range m.States[id].Transitions[j].Sends {
				m.States[id].Transitions[j].Sends[k].Transition = &m.States[id].Transitions[j]
				m.States[id].Transitions[j].Sends[k].ProtocolType =
					ProtocolType[m.States[id].Transitions[j].Sends[k].TypeID]
				if m.States[id].Transitions[j].Sends[k].TypeID == MessageIssueCred &&
					m.States[id].Transitions[j].Sends[k].EventData.Issuing == nil {
					return fmt.Errorf("bad format in (%s) missing Issuing data",
						m.States[id].Transitions[j].Sends[k].Data)
				}
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
	m.Memory = make(map[string]string)
	m.Initialized = true
	return nil
}

func (m *Machine) CurrentState() State {
	return m.States[m.Current]
}

// Triggers returns a transition if machine has it in its current state. If not
// it returns nil.
func (m *Machine) Triggers(status *agency.ProtocolStatus) *Transition {
	for _, transition := range m.CurrentState().Transitions {
		transition.Machine = m // todo: not needed
		transition.Trigger.Machine = m
		if transition.Trigger.ProtocolType == status.State.ProtocolId.TypeId &&
			transition.Trigger.Triggers(status) {
			return &transition
		}
	}
	return nil
}

func (m *Machine) Step(t *Transition) {
	glog.V(1).Infoln("--- Transition from", m.Current, "to", t.Target)
	m.Current = t.Target
}

func (t *Transition) BuildSendEvents(status *agency.ProtocolStatus) []Event {
	input, tgtChanged := t.buildInputEvent(status)
	events := t.Sends
	if tgtChanged {
		events = []Event{input}
	}

	sends := make([]Event, len(events))
	for i, send := range events {
		sends[i] = send
		switch send.TypeID {
		case MessageIssueCred:
			switch send.Rule {
			case TriggerTypeFormatFromMem:
				sends[i].EventData = &EventData{Issuing: &Issuing{
					CredDefID: send.EventData.Issuing.CredDefID,
					AttrsJSON: t.FmtFromMem(&send),
				}}
			}
		case MessageEmail:
			switch send.Rule {
			case TriggerTypePIN:
				t.GenPIN(&send)
				emailJSON := t.FmtFromMem(&send)
				var email Email
				err := json.Unmarshal([]byte(emailJSON), &email)
				if err != nil {
					glog.Errorf("json error %v", err)
				}
				glog.Infoln("email:", emailJSON)
				sends[i].EventData = &EventData{Email: &email}
			}
		case MessageBasicMessage:
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
			case TriggerTypeFormatFromMem:
				sends[i].EventData = &EventData{BasicMessage: &BasicMessage{
					Content: t.FmtFromMem(&send),
				}}
			}
		}
	}
	return sends
}

func (t *Transition) buildInputEvent(status *agency.ProtocolStatus) (e Event, tgtSwitch bool) {
	e = Event{
		ProtocolType:   status.GetState().ProtocolId.TypeId,
		ProtocolStatus: status,
	}
	switch status.GetState().ProtocolId.TypeId {
	case agency.Protocol_ISSUE:
		switch t.Trigger.Rule {
		case TriggerTypeOurMessage:
			return e, false
		}
	case agency.Protocol_CONNECT:
		return e, false
	case agency.Protocol_BASIC_MESSAGE:
		content := status.GetBasicMessage().Content
		switch t.Trigger.Rule {
		case TriggerTypeValidateInputNotEqual:
			e.Data = content
			e.EventData = &EventData{BasicMessage: &BasicMessage{
				Content: content,
			}}
		case TriggerTypeValidateInputEqual:
			if t.Machine.Memory[t.Trigger.Data] != content {
				glog.V(1).Infof("want: %s got: %s",
					t.Machine.Memory[t.Trigger.Data], content)
				t.Target = t.Trigger.FailTarget
				return *t.Trigger.FailEvent, true
			}
			e.Data = content
			e.EventData = &EventData{BasicMessage: &BasicMessage{
				Content: content,
			}}
		case TriggerTypeUseInputSave:
			t.Machine.Memory[t.Trigger.Data] = content
			e.Data = content
			e.EventData = &EventData{BasicMessage: &BasicMessage{
				Content: content,
			}}
		case TriggerTypeUseInput:
			e.Data = content
			e.EventData = &EventData{BasicMessage: &BasicMessage{
				Content: content,
			}}
		case TriggerTypeData, TriggerTypeInputEqual:
			e.EventData = &EventData{BasicMessage: &BasicMessage{
				Content: t.Trigger.Data,
			}}
		}
	}
	return e, false
}

func (t *Transition) FmtFromMem(send *Event) string {
	defer err2.Catch(func(err error) {
		glog.Error(err)
	})
	templ := template.Must(template.New("templ").Parse(send.Data))
	var buf bytes.Buffer
	err2.Check(templ.Execute(&buf, t.Machine.Memory))
	return buf.String()
}

func (t *Transition) GenPIN(_ *Event) {
	t.Machine.Memory["PIN"] = "12234" // todo: real generator
	glog.Infoln("pin code:", t.Machine.Memory["PIN"])
}

var ProtocolType = map[string]agency.Protocol_Type{
	MessageNone:         0, // todo: we need the constant here!
	MessageConnection:   agency.Protocol_CONNECT,
	MessageIssueCred:    agency.Protocol_ISSUE,
	MessagePresentProof: agency.Protocol_PROOF,
	MessageTrustPing:    agency.Protocol_TRUST_PING,
	MessageBasicMessage: agency.Protocol_BASIC_MESSAGE,
	MessageEmail:        EmailProtocol,
}
