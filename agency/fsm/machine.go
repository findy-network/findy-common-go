package fsm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/Shopify/go-lua"
	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/ghodss/yaml"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

type MachineData struct {
	FType string
	Data  []byte
}

func (md *MachineData) IsValid() bool {
	return md != nil && md.FType != "" && md.Data != nil
}

func NewBackendMachine(data MachineData) *Machine {
	var machine Machine
	if filepath.Ext(data.FType) == ".json" {
		try.To(json.Unmarshal(data.Data, &machine))
	} else {
		try.To(yaml.Unmarshal(data.Data, &machine))
	}
	machine.Type = MachineTypeBackend
	return &machine
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
	// Tells do we have a Backend (Service) bot or a connection lvl bot
	Type       MachineType `json:"type,omitempty"`
	Name       string      `json:"name,omitempty"`
	KeepMemory bool        `json:"keep_memory,omitempty"`

	// marks the start state: there can be only one for the Machine, but there
	// can be 0..n termination states. See State.Terminate field.
	Initial *Transition `json:"initial"`

	States map[string]*State `json:"states"`

	Current     string `json:"-"`
	Initialized bool   `json:"-"`

	Memory map[string]string `json:"-"`
	// f-fsm uses this, b-fsm gets it from the BackendData
	ConnID string `json:"-"`

	termChan TerminateOutChan `json:"-"`
	luaState *lua.State       `json:"-"`
}

func (m *Machine) register(name string) map[string]string {
	switch name {
	case REG_MEMORY:
		return m.Memory
	default:
		return m.Memory
	}
}

func (m *Machine) registerMemFuncs() {
	m.luaState.Register("getRegValue", func(l *lua.State) (status int) {
		defer err2.Catch(err2.Err(func(err error) {
			status = 0
		}))
		r, ok := l.ToString(1)
		assert.That(ok)
		glog.V(6).Infoln("r:", r)
		k, ok := l.ToString(2)
		assert.That(ok)
		glog.V(6).Infoln("k:", k)
		v := assert.MKeyExists(m.register(r), k)
		glog.V(6).Infoln("v:", v)
		l.PushString(v)
		return 1
	})

	m.luaState.Register("setRegValue", func(l *lua.State) (nResults int) {
		defer err2.Catch(err2.Err(func(err error) {
			nResults = 0
		}))
		r, ok := l.ToString(1)
		assert.That(ok)
		glog.V(6).Infoln("r:", r)
		k, ok := l.ToString(2)
		assert.That(ok)
		v, ok := l.ToString(3)
		assert.That(ok)
		m.register(r)[k] = v
		glog.V(6).Infof("[%s] = '%v'", k, v)
		return 0
	})
}

// Initialize initializes and optimizes the state machine because the JSON is
// meant for humans to write and machines to read. Initialize also moves machine
// to the initial state. It returns error if machine has them.
func (m *Machine) Initialize() (err error) {
	defer err2.Handle(&err)

	if m.Type == MachineTypeNone {
		m.Type = MachineTypeConversation
	}
	m.Memory = make(map[string]string)
	initSet := false
	for id := range m.States {
		for _, transition := range m.States[id].Transitions {
			transition.Machine = m
			transition.Trigger.Transition = transition
			transition.Trigger.ProtocolType =
				ProtocolType[transition.Trigger.Protocol]
			transition.Trigger.NotificationType =
				NotificationTypeID(transition.Trigger.TypeID)
			trEvent := transition.Trigger
			trEvent.filterEnvs()
			for _, send := range transition.Sends {
				send.Transition = transition
				send.ProtocolType =
					ProtocolType[send.Protocol]
				send.NotificationType =
					NotificationTypeID(send.TypeID)
				if send.Protocol == MessageIssueCred && (send.EventData == nil ||
					send.EventData.Issuing == nil) {
					glog.Errorln("missing EventData of issue_cred msg. Target:",
						send.Target)
					return fmt.Errorf("bad format in (%s) missing Issuing data",
						send.Data)
				}
				sEvent := send
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
	for _, initSend := range m.Initial.Sends {
		initSend.Transition = m.Initial
		initSend.ProtocolType = ProtocolType[initSend.Protocol]
		setSendDefs(initSend)
	}

	m.Initialized = true
	return nil
}

func (m *Machine) InitLua() {
	// intitialize lua stuff in own function to help tests
	m.luaState = lua.NewState()
	m.registerMemFuncs()
	lua.OpenLibraries(m.luaState)
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
		if transition.Trigger.ProtocolType == status.State.ProtocolID.TypeID {
			if ok, tgt := transition.Trigger.Triggers(status); ok {
				return transition.withNewTarget(tgt)
			}
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

func (m *Machine) TriggersByStep() *Transition {
	for _, transition := range m.CurrentState().Transitions {
		if transition.Trigger.ProtocolType == TransientProtocol {
			return transition
		}
	}
	return nil
}

func (m *Machine) TriggersByBackendData(data *BackendData) *Transition {
	glog.V(3).Infof("MachineType: %v", m.Type)
	for _, transition := range m.CurrentState().Transitions {
		if transition.Trigger.ProtocolType == BackendProtocol {
			if ok, tgt := transition.Trigger.TriggersByBackendData(data); ok {
				return transition.withNewTarget(tgt)
			}
		}
	}
	return nil
}

func (m *Machine) Step(t *Transition) {
	glog.V(1).Infoln("--- Transition from", m.Current, "to", t.Target)
	m.Current = t.Target

	// coming to Initial state default is to clear the memory map
	if m.Current == m.Initial.Target && !m.KeepMemory {
		m.Memory = make(map[string]string)
		glog.V(1).Infoln("--- clearing memory map")
	} else if m.KeepMemory {
		glog.V(1).Infoln("--- NOT clearing memory map 'cause 'keep_memory'")
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

type TransientChan = chan string
type TransientInChan = <-chan string
type TransientOutChan = chan<- string

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
	if t.Sends != nil {
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
