package fsm

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

// We cannot (atleast yet) to use JSON enum type like MachineType, because we
// have used different naming like snake_case, etc. maybe refactor later, or
// migrate old FSM files?
const (
	// Executes Lua script that can access to machines memory and which must
	// return true/false if trigger can be executed.
	TriggerTypeLua = "LUA"

	// monitors how our proof/issue protocol goes
	TriggerTypeOurMessage = "OUR_STATUS"

	// used just for echo/forward
	TriggerTypeUseInput = "INPUT"

	// saves input data to event that we can use it, data tells the name of
	// memory slot
	TriggerTypeUseInputSave = "INPUT_SAVE"

	// saves input data to event that we can use it, data tells the end of the
	// name of memory slot. Name is callculated with concat:
	// connID+<given-name>
	TriggerTypeUseInputSaveConnID = "INPUT_SAVE_CONN"

	// saves input data to event as an SESSION_ID that is used solutions that
	// need chat room algorithm. The date goes from f-fsm -> b-fsm -> f-fsm
	// where in the last stage it's filtered by the SESSION_ID that must be the
	// same. Filtering is done in last phase to keep b-fsm simple.
	TriggerTypeUseInputSaveSessionID = "INPUT_SAVE_SESSION_ID"

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

	// transient state, just executes without any triggering checks
	TriggerTypeTransient = "TRANSIENT"
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

	// these are internal messages send between Backend (service) FSM and
	// conversation (pairwise connection) FSM
	MessageBackend = "backend"

	MessageTransient = "transient"
)

const (
	EmailProtocol = 100
	QAProtocol    = 101
	HookProtocol  = 102

	BackendProtocol   = 103 // see MessageBackend
	TransientProtocol = 104 // see MessageTransient
)

const (
	digitsInPIN = 6

	// register names for communication thru machine's memory map.
	LUA_CONN_ID = "CONN_ID" // pairwise ID
	LUA_SUBJECT = "SUBJECT" // reserved

	// this is in use generally, not only in lua
	LUA_SESSION_ID = "SESSION_ID"

	LUA_INPUT  = "INPUT"  // current incoming data like basic_message.content
	LUA_OUTPUT = "OUTPUT" // lua scripts output register name
	LUA_TARGET = "TARGET" // lua scripts target register name
	LUA_OK     = "OK"     // lua scripts OK return value
	LUA_ALL_OK = ""       // lua scripts return values are OK
	LUA_ERROR  = "ERR"    // lua scripts key for error message
)

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

type State struct {
	Transitions []*Transition `json:"transitions"`

	Terminate bool `json:"terminate,omitempty"`

	// TODO: transient state (empedding Lua is tested) + new rules
	// - we should find proper use case to develop these
	// - chat room could be a good case
	//   + use credential to get in. Email credential is need. Or callsign
	//     cred.
	//   + login happens by presenting proof of callsign
	//   + PW-chatbot keeps track of callsigns and the room user is connected
	//   + backend-chatbot uses room when forwarding msgs
	//   + end user sends only message. callsign is only added by PW-chatbot
	//   + pairwise FSM sends room and callsign to backend with msg
	//   + backend FSM forwards msgs per room and and keeps: callsign

	// we could have onEntry and OnExit ? If that would help, we shall see
}

var ruleMap = map[string]string{
	TriggerTypeOurMessage: "STATUS",
	TriggerTypeUseInput:   "<-",

	TriggerTypeUseInputSave:          ":=",
	TriggerTypeUseInputSaveConnID:    ":=",
	TriggerTypeUseInputSaveSessionID: ":=",

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

	Backend *BackendData `json:"backend,omitempty"`
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

// ------ lua stuff ------
const (
	REG_MEMORY  = "MEM"
	REG_DB      = "DB"
	REG_PROCESS = "PROC"
)

func filterFilelink(in string) (o string) {
	return filterLink(in, "$", func(k string) string {
		defer err2.Catch()
		d := try.To1(os.ReadFile(k))
		s := string(d)
		return s
	})
}

func filterEnvs(in string) (o string) {
	return filterLink(in, "$", func(k string) string {
		return os.Getenv(k)
	})
}

func filterLink(in, keyword string, getter func(k string) string) (o string) {
	defer func() {
		glog.V(5).Infoln(in, "->", o)
	}()
	s := strings.Split(in, keyword+"{")
	for i, sub := range s {
		if strings.HasPrefix(in, sub) {
			o += sub
		} else {
			s2 := strings.Split(sub, "}")
			e := ""
			if len(s2) > 1 {
				e = getter(s2[0])
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
	MessageBackend:      BackendProtocol,
	MessageTransient:    TransientProtocol,
}

var toFileProtocolType = map[agency.Protocol_Type]string{
	agency.Protocol_NONE:             MessageNone,
	agency.Protocol_DIDEXCHANGE:      MessageConnection,
	agency.Protocol_ISSUE_CREDENTIAL: MessageIssueCred,
	agency.Protocol_PRESENT_PROOF:    MessagePresentProof,
	agency.Protocol_TRUST_PING:       MessageTrustPing,
	agency.Protocol_BASIC_MESSAGE:    MessageBasicMessage,
	EmailProtocol:                    MessageEmail,
	QAProtocol:                       MessageAnswer,
	HookProtocol:                     MessageHook,
	BackendProtocol:                  MessageBackend,
	TransientProtocol:                MessageTransient,
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
