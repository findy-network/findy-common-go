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

	TriggerTypeAcceptAndInputValues = "ACCEPT_AND_INPUT_VALUES"
	TriggerTypeNotAcceptValues      = "NOT_ACCEPT_VALUES"
)

const (
	MessageNone         = ""
	MessageBasicMessage = "basic_message"
	MessageIssueCred    = "issue_cred"
	MessageTrustPing    = "trust_ping"
	MessagePresentProof = "present_proof"
	MessageConnection   = "connection"

	MessageAnswer = "answer"

	MessageEmail = "email"
)

const (
	EmailProtocol = 100
	QAProtocol    = 101
)

//	*ProtocolStatus_Connection_
//	*ProtocolStatus_Issue_
//	*ProtocolStatus_Proof
//	*ProtocolStatus_TrustPing_
//	*ProtocolStatus_BasicMessage_

// NewBasicMessage creates a new message which can be send to machine
func NewBasicMessage(content string) *agency.ProtocolStatus {
	agencyProof := &agency.ProtocolStatus{
		State: &agency.ProtocolState{ProtocolId: &agency.ProtocolID{
			TypeId: agency.Protocol_BASIC_MESSAGE}},
		Status: &agency.ProtocolStatus_BasicMessage_{BasicMessage: &agency.ProtocolStatus_BasicMessage{Content: content}},
	}
	return agencyProof
}

type Machine struct { // todo: should these have a name, for humans at least?
	Initial string            `json:"initial"`
	States  map[string]*State `json:"states"`

	Current     string `json:"-"`
	Initialized bool   `json:"-"`

	Memory map[string]string `json:"-"`
}

type State struct {
	Transitions []*Transition `json:"transitions"`

	// we could have onEntry and OnExit ? If that would help, we shall see
}

type Transition struct {
	Trigger *Event `json:"trigger"`

	Sends []*Event `json:"sends,omitempty"`

	Target string `json:"target"`

	// Script, or something to execute in future?? idea we could have LUA
	// script which communicates our Memory map, that would be a simple data
	// model

	Machine *Machine `json:"-"`
}

type EventType string

type Event struct {
	// todo: questions could be protocols here, then TypeID would not be needed?
	Protocol string `json:"protocol"`
	TypeID   string `json:"type_id"`

	Rule     string `json:"rule"`
	Data     string `json:"data,omitempty"`
	NoStatus bool   `json:"no_status,omitempty"`

	*EventData `json:"event_data,omitempty"`

	ProtocolType     agency.Protocol_Type     `json:"-"`
	NotificationType agency.Notification_Type `json:"-"`

	*agency.ProtocolStatus `json:"-"`
	*Transition            `json:"-"`
}

func (e Event) Triggers(status *agency.ProtocolStatus) bool {
	switch status.GetState().ProtocolId.TypeId {
	case agency.Protocol_ISSUE, agency.Protocol_CONNECT, agency.Protocol_PROOF:
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

func (e Event) Answers(status *agency.AgentStatus) bool {
	switch status.Notification.TypeId {
	case agency.Notification_ANSWER_NEEDED_PING:
	case agency.Notification_ANSWER_NEEDED_ISSUE_PROPOSE:
	case agency.Notification_ANSWER_NEEDED_PROOF_PROPOSE:
	case agency.Notification_ANSWER_NEEDED_PROOF_VERIFY:
		if e.ProtocolType != agency.Protocol_PROOF {
			panic("programming error")
		}
		var attrValues []ProofAttr
		err2.Check(json.Unmarshal([]byte(e.Data), &attrValues))

		switch e.Rule {
		case TriggerTypeNotAcceptValues:
			if len(attrValues) != len(status.Notification.GetProofVerify().Attrs) {
				return true
			}
			for _, attr := range status.Notification.GetProofVerify().Attrs {
				for i, value := range attrValues {
					if value.Name == attr.Name && value.CredDefID == attr.CredDefId {
						attrValues[i].found = true
					}
				}
			}
			for _, value := range attrValues {
				if !value.found {
					return true
				}
			}
		case TriggerTypeAcceptAndInputValues:
			count := 0
			for _, attr := range status.Notification.GetProofVerify().Attrs {
				for _, value := range attrValues {
					if value.Name == attr.Name {
						e.Machine.Memory[value.Name] = attr.Value
						count++
					}
				}
			}
			return count == len(status.Notification.GetProofVerify().Attrs)
		}
	}
	return false
}

type EventData struct {
	BasicMessage *BasicMessage `json:"basic_message,omitempty"`
	Issuing      *Issuing      `json:"issuing,omitempty"`
	Email        *Email        `json:"email,omitempty"`
	Proof        *Proof        `json:"proof,omitempty"`
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

type Proof struct {
	ProofJSON string `json:"proof_json"`
}

type ProofAttr struct {
	ID        string `json:"-"`
	Name      string `json:"name,omitempty"`
	CredDefID string `json:"credDefId,omitempty"`
	Predicate string `json:"predicate,omitempty"`

	found bool
}

type BasicMessage struct {
	Content string
}

// Initialize initializes and optimizes the state machine because the JSON is
// meant for humans to write and machines to read. Initialize also moves machine
// to the initial state. It returns error if machine has them.
func (m *Machine) Initialize() (err error) {
	m.Memory = make(map[string]string)
	initSet := false
	for id := range m.States {
		for j := range m.States[id].Transitions {
			m.States[id].Transitions[j].Machine = m
			m.States[id].Transitions[j].Trigger.Transition = m.States[id].Transitions[j]
			m.States[id].Transitions[j].Trigger.ProtocolType =
				ProtocolType[m.States[id].Transitions[j].Trigger.Protocol]
			m.States[id].Transitions[j].Trigger.NotificationType =
				NotificationTypeID[m.States[id].Transitions[j].Trigger.TypeID]
			for k := range m.States[id].Transitions[j].Sends {
				m.States[id].Transitions[j].Sends[k].Transition = m.States[id].Transitions[j]
				m.States[id].Transitions[j].Sends[k].ProtocolType =
					ProtocolType[m.States[id].Transitions[j].Sends[k].Protocol]
				m.States[id].Transitions[j].Sends[k].NotificationType =
					NotificationTypeID[m.States[id].Transitions[j].Sends[k].TypeID]
				if m.States[id].Transitions[j].Sends[k].Protocol == MessageIssueCred &&
					m.States[id].Transitions[j].Sends[k].EventData.Issuing == nil {
					return fmt.Errorf("bad format in (%s) missing Issuing data",
						m.States[id].Transitions[j].Sends[k].Data)
				}
			}
		}
		if id == m.Initial {
			if initSet {
				return errors.New("machine has multiple initial states")
			}
			m.Current = m.Initial
			initSet = true
		}
	}
	m.Initialized = true
	return nil
}

func (m *Machine) CurrentState() *State {
	return m.States[m.Current]
}

// Triggers returns a transition if machine has it in its current state. If not
// it returns nil.
func (m *Machine) Triggers(status *agency.ProtocolStatus) *Transition {
	for _, transition := range m.CurrentState().Transitions {
		if transition.Trigger.ProtocolType == status.State.ProtocolId.TypeId &&
			transition.Trigger.Triggers(status) {
			return transition
		}
	}
	return nil
}

func (m *Machine) Step(t *Transition) {
	glog.V(1).Infoln("--- Transition from", m.Current, "to", t.Target)
	m.Current = t.Target
}

func (m *Machine) Answers(status *agency.AgentStatus) *Transition {
	for _, transition := range m.CurrentState().Transitions {
		if transition.Trigger.ProtocolType == status.Notification.ProtocolType &&
			transition.Trigger.Answers(status) {
			return transition
		}
	}
	return nil
}

func (t *Transition) BuildSendEvents(status *agency.ProtocolStatus) []*Event {
	input := t.buildInputEvent(status)
	return t.doBuildSendEvents(input)
}

func (t *Transition) doBuildSendEvents(input *Event) []*Event {
	events := t.Sends
	sends := make([]*Event, len(events))
	for i, send := range events {
		sends[i] = send
		switch send.Protocol {
		case MessageIssueCred:
			switch send.Rule {
			case TriggerTypeFormatFromMem:
				sends[i].EventData = &EventData{Issuing: &Issuing{
					CredDefID: send.EventData.Issuing.CredDefID,
					AttrsJSON: t.FmtFromMem(send),
				}}
			}
		case MessagePresentProof:
			switch send.Rule {
			case TriggerTypeData:
				sends[i].EventData = &EventData{Proof: &Proof{
					ProofJSON: send.Data,
				}}
			}
		case MessageAnswer:
			glog.V(1).Infoln("building answer") // it's so easy
		case MessageEmail:
			switch send.Rule {
			case TriggerTypePIN:
				t.GenPIN(send)
				emailJSON := t.FmtFromMem(send)
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
					Content: t.FmtFromMem(send),
				}}
			}
		}
	}
	return sends
}

func (t *Transition) buildInputEvent(status *agency.ProtocolStatus) (e *Event) {
	e = &Event{
		ProtocolType:   status.GetState().ProtocolId.TypeId,
		ProtocolStatus: status,
	}
	switch status.GetState().ProtocolId.TypeId {
	case agency.Protocol_ISSUE, agency.Protocol_PROOF:
		switch t.Trigger.Rule {
		case TriggerTypeOurMessage:
			return e
		}
	case agency.Protocol_CONNECT:
		return e
	case agency.Protocol_BASIC_MESSAGE:
		content := status.GetBasicMessage().Content
		switch t.Trigger.Rule {
		case TriggerTypeValidateInputNotEqual, TriggerTypeValidateInputEqual, TriggerTypeUseInput:
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
		case TriggerTypeData, TriggerTypeInputEqual:
			e.EventData = &EventData{BasicMessage: &BasicMessage{
				Content: t.Trigger.Data,
			}}
		}
	}
	return e
}

func (t *Transition) buildInputAnswers(status *agency.AgentStatus) (e *Event) {
	e = &Event{
		ProtocolType: status.Notification.ProtocolType,
	}
	return e
}

func (t *Transition) FmtFromMem(send *Event) string {
	defer err2.Catch(func(err error) {
		glog.Error(err)
	})
	tmpl := template.Must(template.New("template").Parse(send.Data))
	var buf bytes.Buffer
	err2.Check(tmpl.Execute(&buf, t.Machine.Memory))
	return buf.String()
}

func (t *Transition) GenPIN(_ *Event) {
	t.Machine.Memory["PIN"] = "12234" // todo: real generator
	glog.Infoln("pin code:", t.Machine.Memory["PIN"])
}

func (t *Transition) BuildSendAnswers(status *agency.AgentStatus) []*Event {
	inputs := t.buildInputAnswers(status)
	return t.doBuildSendEvents(inputs)
}

var ProtocolType = map[string]agency.Protocol_Type{
	MessageNone:         agency.Protocol_NONE,
	MessageConnection:   agency.Protocol_CONNECT,
	MessageIssueCred:    agency.Protocol_ISSUE,
	MessagePresentProof: agency.Protocol_PROOF,
	MessageTrustPing:    agency.Protocol_TRUST_PING,
	MessageBasicMessage: agency.Protocol_BASIC_MESSAGE,
	MessageEmail:        EmailProtocol,
	MessageAnswer:       QAProtocol,
}

var NotificationTypeID = map[string]agency.Notification_Type{
	"STATUS_UPDATE":               agency.Notification_STATUS_UPDATE,
	"ACTION_NEEDED":               agency.Notification_ACTION_NEEDED,
	"ANSWER_NEEDED_PING":          agency.Notification_ANSWER_NEEDED_PING,
	"ANSWER_NEEDED_ISSUE_PROPOSE": agency.Notification_ANSWER_NEEDED_ISSUE_PROPOSE,
	"ANSWER_NEEDED_PROOF_PROPOSE": agency.Notification_ANSWER_NEEDED_PROOF_PROPOSE,
	"ANSWER_NEEDED_PROOF_VERIFY":  agency.Notification_ANSWER_NEEDED_PROOF_VERIFY,
}