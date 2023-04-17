package fsm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

const (
	// monitors how our proof/issue protocol goes
	TriggerTypeOurMessage = "OUR_STATUS"

	// used just for echo/forward
	TriggerTypeUseInput = "INPUT"

	// saves input data to event that we can use it, data tells the name of
	// memory slot
	TriggerTypeUseInputSave = "INPUT_SAVE"

	// formates input data with then format string which is in send data
	TriggerTypeFormat = "FORMAT"

	// formates send event where data is themplate and every memory map value
	// are available. See exmaples for more information.
	TriggerTypeFormatFromMem = "FORMAT_MEM"

	// helps to generate a PIN code to send e.g. email (endpoint not yet
	// supported).
	TriggerTypePIN = "GEN_PIN"

	// quides to use send events `data` as is.
	TriggerTypeData = ""

	// these three validate 'operations' compare input data to send data
	TriggerTypeValidateInputEqual    = "INPUT_VALIDATE_EQUAL"
	TriggerTypeValidateInputNotEqual = "INPUT_VALIDATE_NOT_EQUAL"
	TriggerTypeInputEqual            = "INPUT_EQUAL"

	// these two need other states to help them (in production). The previous
	// states decide to which of these the FSM transits.
	// accept and stores present proof values and stores them to FSM memory map
	TriggerTypeAcceptAndInputValues = "ACCEPT_AND_INPUT_VALUES"
	// not accept present proof protocol
	TriggerTypeNotAcceptValues = "NOT_ACCEPT_VALUES"
)

const (
	// these are Aries DIDComm protocols
	MessageNone         = ""
	MessageBasicMessage = "basic_message"
	MessageIssueCred    = "issue_cred"
	MessageTrustPing    = "trust_ping"
	MessagePresentProof = "present_proof"
	MessageConnection   = "connection"

	MessageAnswer = "answer"

	MessageEmail = "email" // not supported yet
	MessageHook  = "hook"  // internal program call back
)

const (
	EmailProtocol = 100
	QAProtocol    = 101
	HookProtocol  = 102
)

const digitsInPIN = 6

var seed = time.Now().UnixNano()

func init() {
	rand.NewSource(seed)
}

// NewBasicMessage creates a new message which can be send to machine
func _(content string) *agency.ProtocolStatus {
	agencyProof := &agency.ProtocolStatus{
		State: &agency.ProtocolState{ProtocolID: &agency.ProtocolID{
			TypeID: agency.Protocol_BASIC_MESSAGE}},
		Status: &agency.ProtocolStatus_BasicMessage{
			BasicMessage: &agency.ProtocolStatus_BasicMessageStatus{
				Content: content,
			},
		},
	}
	return agencyProof
}

type MachineData struct {
	FType string
	Data  []byte
}

func NewMachine(data MachineData) *Machine {
	var machine Machine
	if filepath.Ext(data.FType) == ".json" {
		try.To(json.Unmarshal(data.Data, &machine))
	} else {
		try.To(yaml.Unmarshal(data.Data, &machine))
	}
	return &machine
}

type Machine struct {
	Name string `json:"name,omitempty"`

	// marks the start state: there can be only one for the Machine, but there
	// can be 0..n termination states. See State.Terminate field.
	Initial *Transition `json:"initial"`

	States map[string]*State `json:"states"`

	Current     string `json:"-"`
	Initialized bool   `json:"-"`

	Memory map[string]string `json:"-"`

	termChan TerminateOutChan `json:"-"`
}

type State struct {
	Transitions []*Transition `json:"transitions"`

	Terminate bool `json:"terminate,omitempty"`

	// TODO: add KeepMemory, see Step()
	// TODO: transient state (empedding Lua is tested) + new rules
	// - we should find proper use case to develop these

	// we could have onEntry and OnExit ? If that would help, we shall see
}

type Transition struct {
	Trigger *Event `json:"trigger,omitempty"`

	Sends []*Event `json:"sends,omitempty"`

	Target string `json:"target"`

	// Script, or something to execute in future?? idea we could have LUA
	// script which communicates our Memory map, that would be a simple data
	// model

	Machine *Machine `json:"-"`
}

type NotificationType int32

type Event struct {
	// TODO: questions could be protocols here, then TypeID would not be needed?
	// we will continue with this when other protocol QAs will be implemented
	// New! Hook now uses TypeID for hook name/ID

	// These both are string versions to make writing the yaml fsm easier.
	// There parser methdod, Initialize() that must be call to make the machine
	// to work. It also make other syntax checks.
	Protocol string `json:"protocol"` // Note! See ProtocolType below
	TypeID   string `json:"type_id"`  // Note! See NotificationType below

	Rule string `json:"rule"`
	Data string `json:"data,omitempty"`
	// Deprecated: replaced by WantStatus, left to keep file format
	NoStatus bool `json:"no_status,omitempty"`
	// Tells that we want status updates about our sending, this is calculated
	// automatically
	WantStatus bool `json:"want_status,omitempty"`

	*EventData `json:"event_data,omitempty"`

	ProtocolType     agency.Protocol_Type `json:"-"`
	NotificationType NotificationType     `json:"-"`
	//NotificationType agency.Notification_Type `json:"-"`

	*agency.ProtocolStatus `json:"-"`
	*Transition            `json:"-"`
}

func (e *Event) filterEnvs() {
	if e == nil {
		glog.V(7).Infoln("no event data, in filter:", e.Protocol)
		return
	}
	glog.V(10).Infoln("in filter:", e.Protocol)

	switch e.ProtocolType {
	case agency.Protocol_ISSUE_CREDENTIAL:
		if e.EventData != nil && e.EventData.Issuing != nil {
			e.EventData.Issuing.CredDefID = filterEnvs(e.EventData.Issuing.CredDefID)
		}
		e.Data = filterEnvs(e.Data)
	case agency.Protocol_PRESENT_PROOF:
		if e.EventData != nil && e.EventData.Proof != nil {
			e.EventData.Proof.ProofJSON = filterEnvs(e.EventData.Proof.ProofJSON)
		}
		e.Data = filterEnvs(e.Data)
	default:
		glog.V(7).Infoln("wrong type, in filter:", e.ProtocolType)
	}
}

func (e Event) TriggersByHook() bool {
	return true
}

func (e Event) Triggers(status *agency.ProtocolStatus) bool {
	if status == nil {
		return true
	}
	switch status.GetState().ProtocolID.TypeID {
	case agency.Protocol_ISSUE_CREDENTIAL, agency.Protocol_DIDEXCHANGE, agency.Protocol_PRESENT_PROOF:
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

func (e Event) Answers(status *agency.Question) bool {
	switch status.TypeID {
	case agency.Question_PING_WAITS:
	case agency.Question_ISSUE_PROPOSE_WAITS:
	case agency.Question_PROOF_PROPOSE_WAITS:
	case agency.Question_PROOF_VERIFY_WAITS:
		assert.Equal(e.ProtocolType, agency.Protocol_PRESENT_PROOF)

		var attrValues []ProofAttr
		try.To(json.Unmarshal([]byte(e.Data), &attrValues))

		switch e.Rule {
		case TriggerTypeNotAcceptValues:
			if len(attrValues) != len(status.GetProofVerify().Attributes) {
				return true
			}
			for _, attr := range status.GetProofVerify().Attributes {
				for i, value := range attrValues {
					if value.Name == attr.Name && value.CredDefID == attr.CredDefID {
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
			for _, attr := range status.GetProofVerify().Attributes {
				for _, value := range attrValues {
					if value.Name == attr.Name {
						e.Machine.Memory[value.Name] = attr.Value
						count++
					}
				}
			}
			return count == len(status.GetProofVerify().Attributes)
		}
	}
	return false
}

var ruleMap = map[string]string{
	TriggerTypeOurMessage:    "STATUS",
	TriggerTypeUseInput:      "<-",
	TriggerTypeUseInputSave:  ":=",
	TriggerTypeFormat:        "",
	TriggerTypeFormatFromMem: "%s",
	TriggerTypePIN:           "new PIN",
	TriggerTypeData:          "",

	TriggerTypeValidateInputEqual:    "==",
	TriggerTypeValidateInputNotEqual: "!=",
	TriggerTypeInputEqual:            "==",

	TriggerTypeAcceptAndInputValues: "ACCEPT",
	TriggerTypeNotAcceptValues:      "DECLINE",
}

func removeLF(s string) string {
	return strings.ReplaceAll(s, "\n", " ")
}

func (e Event) String() string {
	w := new(bytes.Buffer)
	fmt.Fprintf(w, "%s{%s \"%.12s\"}", e.Protocol, ruleMap[e.Rule], removeLF(e.Data))
	return w.String()
}

type EventData struct {
	BasicMessage *BasicMessage `json:"basic_message,omitempty"`
	Issuing      *Issuing      `json:"issuing,omitempty"`
	Email        *Email        `json:"email,omitempty"`
	Proof        *Proof        `json:"proof,omitempty"`
	Hook         *Hook         `json:"hook,omitempty"`
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

type Hook struct {
	Data map[string]string
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
				NotificationTypeID(m.States[id].Transitions[j].Trigger.TypeID)
			trEvent := m.States[id].Transitions[j].Trigger
			trEvent.filterEnvs()
			for k := range m.States[id].Transitions[j].Sends {
				m.States[id].Transitions[j].Sends[k].Transition = m.States[id].Transitions[j]
				m.States[id].Transitions[j].Sends[k].ProtocolType =
					ProtocolType[m.States[id].Transitions[j].Sends[k].Protocol]
				m.States[id].Transitions[j].Sends[k].NotificationType =
					NotificationTypeID(m.States[id].Transitions[j].Sends[k].TypeID)
				if m.States[id].Transitions[j].Sends[k].Protocol == MessageIssueCred &&
					m.States[id].Transitions[j].Sends[k].EventData.Issuing == nil {
					return fmt.Errorf("bad format in (%s) missing Issuing data",
						m.States[id].Transitions[j].Sends[k].Data)
				}
				sEvent := m.States[id].Transitions[j].Sends[k]
				sEvent.filterEnvs()

				setSendDefs(sEvent)
			}
		}
		if m.Initial == nil {
			return errors.New("machine doesn't have initial state")
		}
		if id == m.Initial.Target {
			if initSet {
				return errors.New("machine has multiple initial states")
			}
			m.Current = m.Initial.Target
			initSet = true
		}
	}
	m.Initial.Machine = m
	for i := range m.Initial.Sends {
		m.Initial.Sends[i].Transition = m.Initial
		m.Initial.Sends[i].ProtocolType =
			ProtocolType[m.Initial.Sends[i].Protocol]
		setSendDefs(m.Initial.Sends[i])
	}
	m.Initialized = true
	return nil
}

func setSendDefs(e *Event) {
	pType := e.ProtocolType
	switch pType {
	case agency.Protocol_ISSUE_CREDENTIAL, agency.Protocol_PRESENT_PROOF:
		e.WantStatus = true
	default:
		e.WantStatus = false
	}
}
func (m *Machine) CurrentState() *State {
	return m.States[m.Current]
}

// Triggers returns a transition if machine has it in its current state. If not
// it returns nil.
func (m *Machine) Triggers(status *agency.ProtocolStatus) *Transition {
	for _, transition := range m.CurrentState().Transitions {
		if transition.Trigger.ProtocolType == status.State.ProtocolID.TypeID &&
			transition.Trigger.Triggers(status) {
			return transition
		}
	}
	return nil
}

// TriggersByHook returns a transition if machine has it in its current state.
// If not it returns nil.
func (m *Machine) TriggersByHook() *Transition {
	for _, transition := range m.CurrentState().Transitions {
		if transition.Trigger.ProtocolType == HookProtocol &&
			transition.Trigger.TriggersByHook() {
			return transition
		}
	}
	return nil
}

func (m *Machine) Step(t *Transition) {
	glog.V(1).Infoln("--- Transition from", m.Current, "to", t.Target)
	m.Current = t.Target

	// coming to Initial state default is to clear the memory map
	if m.Current == m.Initial.Target { // TODO: KeepMemory field to override
		m.Memory = make(map[string]string)
		glog.V(1).Infoln("--- clearing memory map")
	}
	m.checkTerm()
}

func (m *Machine) Answers(q *agency.Question) *Transition {
	for _, transition := range m.CurrentState().Transitions {
		if transition.Trigger.ProtocolType == q.Status.Notification.ProtocolType &&
			transition.Trigger.Answers(q) {
			return transition
		}
	}
	return nil
}

type TerminateChan = chan bool
type TerminateInChan = <-chan bool
type TerminateOutChan = chan<- bool

func (m *Machine) checkTerm() {
	if m.CurrentState().Terminate {
		if m.termChan != nil {
			glog.V(1).Infoln("--- TERMINATE FSM OK ---")
			m.termChan <- true
		} else {
			glog.Warning("--- Cannot signall TERMINATE FSM ---")
		}

	}
}

// Start starts the FSM. It takes termination channel as an argument to be able
// to signaling outside when machine is stoped. It accept nil as a channel value
// when signaling isn't done.
func (m *Machine) Start(termChan TerminateOutChan) []*Event {
	t := m.Initial
	m.termChan = termChan
	if (t.Trigger == nil || t.Trigger.Triggers(nil)) && t.Sends != nil {
		return t.BuildSendEvents(nil)
	}
	return nil
}

const stateWidthInChar = 100

func padStr(s string) string {
	firstPadWidth := stateWidthInChar / 2
	l := len(s)
	s = fmt.Sprintf("%-*s", firstPadWidth+(l/2), s)
	return fmt.Sprintf("%*s", stateWidthInChar, s)
}

//goland:noinspection ALL
func (m *Machine) String() string {
	w := new(bytes.Buffer)
	fsmName := m.Name
	if fsmName != "" {
		fmt.Fprintf(w, "title %s\n", fsmName)
	}
	fmt.Fprintf(w, "[*] --> %s\n", m.Initial.Target)
	for stateName, state := range m.States {
		fmt.Fprintf(w, "state \"%s\" as %s\n", padStr(stateName), stateName)
		for _, transition := range state.Transitions {
			fmt.Fprintf(w, "%s --> %s: **%s**\\n", stateName,
				transition.Target, transition.Trigger.String())
			for _, send := range transition.Sends {
				fmt.Fprintf(w, "{%s} ==>\\n", send)
			}
			fmt.Fprintln(w)
		}
		glog.V(10).Infof("terminate: %s -> %v", stateName, state.Terminate)
		if state.Terminate {
			fmt.Fprintf(w, "%s --> [*]\n", stateName)
		} else {
			fmt.Fprintln(w)
		}
	}
	return w.String()
}

func (t *Transition) BuildSendEventsFromHook(hookData map[string]string) []*Event {
	input := &Event{
		ProtocolType: HookProtocol,
		EventData:    &EventData{Hook: &Hook{Data: hookData}},
	}
	return t.doBuildSendEvents(input)
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
			glog.V(3).Infoln("building answer") // it's so easy
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
				glog.V(1).Infoln("email:", emailJSON)
				sends[i].EventData = &EventData{Email: &email}
			}
		case MessageBasicMessage:
			assert.That(input != nil ||
				send.Rule == TriggerTypeData ||
				send.Rule == TriggerTypeFormatFromMem,
			)
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
		case MessageHook:
			switch send.Rule {
			case TriggerTypeData:
				sends[i].EventData = &EventData{Hook: &Hook{
					Data: map[string]string{
						"ID":   send.TypeID,
						"data": send.Data,
					},
				}}
			case TriggerTypeUseInput:
				dataStr := ""
				if input.Protocol == MessageBasicMessage {
					dataStr = input.EventData.BasicMessage.Content
				}
				sends[i].EventData = &EventData{Hook: &Hook{
					Data: map[string]string{
						"ID":   send.TypeID,
						"data": dataStr,
					},
				}}
			case TriggerTypeFormat:
				sends[i].EventData = &EventData{Hook: &Hook{
					Data: map[string]string{
						"ID":   send.TypeID,
						"data": fmt.Sprintf(send.Data, input.Data),
					},
				}}
			case TriggerTypeFormatFromMem:
				sends[i].EventData = &EventData{Hook: &Hook{
					Data: map[string]string{
						"ID":   send.TypeID,
						"data": t.FmtFromMem(send),
					},
				}}
			}

		}
	}
	return sends
}

func (t *Transition) buildInputEvent(status *agency.ProtocolStatus) (e *Event) {
	if status == nil {
		return nil
	}
	e = &Event{
		ProtocolType:   status.GetState().ProtocolID.TypeID,
		ProtocolStatus: status,
	}
	switch status.GetState().ProtocolID.TypeID {
	case agency.Protocol_ISSUE_CREDENTIAL, agency.Protocol_PRESENT_PROOF:
		switch t.Trigger.Rule {
		case TriggerTypeOurMessage:
			glog.V(4).Infoln("+++ Our message:", status.GetState().ProtocolID.TypeID)
			return e
		}
	case agency.Protocol_DIDEXCHANGE:
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
	try.To(tmpl.Execute(&buf, t.Machine.Memory))
	return buf.String()
}

func pin(digit int) int {
	min := int(math.Pow(10, float64(digit-1)))
	max := int(math.Pow(10, float64(digit)))
	return min + rand.Intn(max-min)
}

func (t *Transition) GenPIN(_ *Event) {
	t.Machine.Memory["PIN"] = fmt.Sprintf("%v", pin(digitsInPIN))
	glog.V(1).Infoln("pin code:", t.Machine.Memory["PIN"])
}

func (t *Transition) BuildSendAnswers(status *agency.AgentStatus) []*Event {
	input := t.buildInputAnswers(status)
	return t.doBuildSendEvents(input)
}

func filterEnvs(in string) (o string) {
	defer func() {
		glog.V(5).Infoln(in, "->", o)
	}()
	s := strings.Split(in, "${")
	for i, sub := range s {
		if strings.HasPrefix(in, sub) {
			o += sub
		} else {
			s2 := strings.Split(sub, "}")
			e := ""
			if len(s2) > 1 {
				e = os.Getenv(s2[0])
			}
			if e == "" {
				return in
			}
			o += e
			theEnd := i == len(s)-1
			if theEnd {
				for j, sub2 := range s2[i:] {
					if j > 0 {
						o += "}"
					}
					o += sub2
				}
				return o
			}
		}
	}
	return o
}

var ProtocolType = map[string]agency.Protocol_Type{
	MessageNone:         agency.Protocol_NONE,
	MessageConnection:   agency.Protocol_DIDEXCHANGE,
	MessageIssueCred:    agency.Protocol_ISSUE_CREDENTIAL,
	MessagePresentProof: agency.Protocol_PRESENT_PROOF,
	MessageTrustPing:    agency.Protocol_TRUST_PING,
	MessageBasicMessage: agency.Protocol_BASIC_MESSAGE,
	MessageEmail:        EmailProtocol,
	MessageAnswer:       QAProtocol,
	MessageHook:         HookProtocol,
}

func NotificationTypeID(typeName string) NotificationType {
	if _, ok := notificationTypeID[typeName]; ok {
		return NotificationType(notificationTypeID[typeName])
	} else if _, ok := QuestionTypeID[typeName]; ok {
		return NotificationType(10) * NotificationType(QuestionTypeID[typeName])
	}
	glog.V(10).Infof("unknown type: \"%v\" setting zero", typeName)
	return 0
}

var notificationTypeID = map[string]agency.Notification_Type{
	"STATUS_UPDATE": agency.Notification_STATUS_UPDATE,
	"ACTION_NEEDED": agency.Notification_PROTOCOL_PAUSED,
}

var QuestionTypeID = map[string]agency.Question_Type{
	"ANSWER_NEEDED_PING":          agency.Question_PING_WAITS,
	"ANSWER_NEEDED_ISSUE_PROPOSE": agency.Question_ISSUE_PROPOSE_WAITS,
	"ANSWER_NEEDED_PROOF_PROPOSE": agency.Question_PROOF_PROPOSE_WAITS,
	"ANSWER_NEEDED_PROOF_VERIFY":  agency.Question_PROOF_VERIFY_WAITS,
}
