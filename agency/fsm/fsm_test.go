package fsm

import (
	"os"
	"testing"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

const transientTestMsg = "test-msg"

var (
	machineTransientState = Machine{
		Name: "machine",
		Initial: &Transition{
			Sends: []*Event{{
				Protocol: "basic_message",
				Data:     "Hello!",
			}},
			Target: "OUTPUT_MESSAGE",
		},
		States: map[string]*State{
			"OUTPUT_MESSAGE": {
				Transitions: []*Transition{
					{
						Trigger: &Event{Protocol: "transient"},
						Sends: []*Event{{
							Protocol: "transient",
							Rule:     "TRANSIENT",
							Data:     transientTestMsg,
						}},
						Target: "TERMINATE",
					},
				},
			},
			"TERMINATE": {
				Terminate: true,
			},
		},
	}

	machineTerminates = Machine{
		Name: "machine",
		Initial: &Transition{
			Sends: []*Event{{
				Protocol: "basic_message",
				Data:     "Hello!",
			}},
			Target: "IDLE",
		},
		States: map[string]*State{
			"IDLE": {
				Transitions: []*Transition{
					{
						Trigger: &Event{Protocol: "basic_message"},
						Sends: []*Event{{
							Protocol: "basic_message",
							Rule:     "INPUT",
						}},
						Target: "TERMINATE",
					},
				},
			},
			"TERMINATE": {
				Terminate: true,
			},
		},
	}

	machine = Machine{
		Name: "machine",
		Initial: &Transition{
			Sends: []*Event{{
				Protocol: "basic_message",
				Data:     "Hello!",
			}},
			Target: "IDLE",
		},
		States: map[string]*State{
			"IDLE": {
				Transitions: []*Transition{
					{
						Trigger: &Event{Protocol: "basic_message"},
						Sends: []*Event{{
							Protocol: "basic_message",
							Rule:     "INPUT",
						}},
						Target: "WAITING_STATUS",
					},
				},
			},
			"WAITING_STATUS": {
				Transitions: []*Transition{{
					Trigger: &Event{
						Protocol: "basic_message",
						Rule:     "OUR_STATUS",
					},
					Sends: []*Event{{
						Protocol: "basic_message",
						Rule:     "OUR_STATUS",
					}},
					Target: "IDLE",
				}},
			},
		},
	}

	showProofMachine = Machine{
		Name: "proof machine",
		Initial: &Transition{
			Target: "IDLE",
		},
		States: map[string]*State{
			"IDLE": {
				Transitions: []*Transition{
					{
						Trigger: &Event{
							Protocol: "hook",
							TypeID:   "input-hook",
						},
						Sends: []*Event{
							{
								Protocol: "hook",
								TypeID:   "output-hook",
								// Testing if it can handle this!
								Rule: "INPUT",
							},
						},
						Target: "IDLE",
					},
					{
						Trigger: &Event{
							Protocol: "connection",
							TypeID:   "STATUS_UPDATE",
						},
						Sends: []*Event{
							{
								Protocol: "basic_message",
								Rule:     "INPUT",
							},
						},
						Target: "WAITING_STATUS",
					},
				},
			},
			"WAITING_STATUS": {
				Transitions: []*Transition{
					{
						Trigger: &Event{
							Protocol: "basic_message",
							TypeID:   "STATUS_UPDATE",
							Rule:     "OUR_STATUS",
						},
						Sends: []*Event{
							{
								Protocol: "basic_message",
								Rule:     "OUR_STATUS",
							},
						},
						Target: "IDLE",
					},
				},
			},
		},
	}
)

func TestMachine_Initialize(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.ThatNot(machine.Initialized)
	assert.That(machine.States["IDLE"].Transitions[0].Trigger.ProtocolType == 0)
	assert.That(machine.States["IDLE"].Transitions[0].Sends[0].ProtocolType == 0)
	assert.NoError(machine.Initialize())
	assert.That(machine.Initialized)
	assert.DeepEqual(agency.Protocol_BASIC_MESSAGE,
		machine.States["IDLE"].Transitions[0].Trigger.ProtocolType)
	assert.DeepEqual(agency.Protocol_BASIC_MESSAGE,
		machine.States["IDLE"].Transitions[0].Sends[0].ProtocolType)
}

func TestMachine_InitializeTerminate(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.ThatNot(machineTerminates.Initialized)
	assert.That(machineTerminates.States["IDLE"].Transitions[0].Trigger.ProtocolType == 0)
	assert.That(machineTerminates.States["IDLE"].Transitions[0].Sends[0].ProtocolType == 0)
	assert.NoError(machineTerminates.Initialize())
	assert.That(machineTerminates.Initialized)
	assert.DeepEqual(agency.Protocol_BASIC_MESSAGE,
		machineTerminates.States["IDLE"].Transitions[0].Trigger.ProtocolType)
	assert.DeepEqual(agency.Protocol_BASIC_MESSAGE,
		machineTerminates.States["IDLE"].Transitions[0].Sends[0].ProtocolType)
}

func TestMachine_Triggers(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.NoError(machine.Initialize())
	assert.That(nil == machine.Triggers(
		protocolStatus(agency.Protocol_PRESENT_PROOF)))
	assert.NotNil(t, machine.Triggers(
		protocolStatus(agency.Protocol_BASIC_MESSAGE)))
}

func TestMachine_TriggersTerminate(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.NoError(machine.Initialize())
	assert.That(nil == machine.Triggers(
		protocolStatus(agency.Protocol_PRESENT_PROOF)))
	assert.NotNil(t, machine.Triggers(
		protocolStatus(agency.Protocol_BASIC_MESSAGE)))
}

func protocolStatus(typeID agency.Protocol_Type, a ...string) *agency.ProtocolStatus {
	content := "test string"
	if len(a) > 0 {
		assert.SLen(a, 1, "currently only one (1) is supported")
		content = a[0]
	}
	agencyProof := &agency.ProtocolStatus{
		State: &agency.ProtocolState{ProtocolID: &agency.ProtocolID{TypeID: typeID}},
		Status: &agency.ProtocolStatus_BasicMessage{
			BasicMessage: &agency.ProtocolStatus_BasicMessageStatus{
				Content: content,
			},
		},
	}
	return agencyProof
}

func TestMachine_Step(t *testing.T) {
	defer assert.PushTester(t)()

	try.To(machine.Initialize())
	transition := machine.Triggers(protocolStatus(agency.Protocol_PRESENT_PROOF))
	assert.Nil(transition)
	transition = machine.Triggers(protocolStatus(agency.Protocol_BASIC_MESSAGE))
	assert.NotNil(transition)
	machine.Step(transition)
	assert.Equal("WAITING_STATUS", machine.Current)
}

func TestMachine_StepTerminate(t *testing.T) {
	defer assert.PushTester(t)()

	try.To(machineTerminates.Initialize())
	termChan := make(TerminateChan)
	go func() {
		defer assert.PushTester(t)()

		termSignaled, ok := <-termChan
		assert.That(ok)
		if ok {
			assert.That(termSignaled)
		}
		assert.Equal("TERMINATE", machineTerminates.Current)
	}()
	_ = machine.Start(termChan)
	transition := machineTerminates.Triggers(protocolStatus(agency.Protocol_PRESENT_PROOF))
	assert.That(nil == transition)
	transition = machineTerminates.Triggers(protocolStatus(agency.Protocol_BASIC_MESSAGE))
	assert.NotNil(transition)
	assert.Equal("IDLE", machineTerminates.Current)
	go machineTerminates.Step(transition)
}

func TestMachine_Step2(t *testing.T) {
	defer assert.PushTester(t)()

	try.To(showProofMachine.Initialize())
	status := protocolStatus(agency.Protocol_DIDEXCHANGE)
	transition := showProofMachine.Triggers(status)
	assert.NotNil(transition)
	e := transition.buildInputEvent(status)
	assert.NotNil(e)
	o := transition.BuildSendEvents(status)
	assert.SNotNil(o)
	showProofMachine.Step(transition)
	assert.Equal("WAITING_STATUS", showProofMachine.Current)
}

func TestMachine_StepByTransient(t *testing.T) {
	defer assert.PushTester(t)()

	try.To(machineTransientState.Initialize())
	transition := machineTransientState.TriggersByStep()
	assert.NotNil(transition)
	stepData := "transient_msg"
	e := transition.BuildSendEventsFromStep(stepData)
	assert.SNotNil(e)
	assert.SNotEmpty(e)
	event := e[0]
	assert.Equal(TransientProtocol, int(event.ProtocolType))
	assert.Equal(event.BasicMessage.Content, transientTestMsg)
	machineTransientState.Step(transition)
	assert.Equal("TERMINATE", machineTransientState.Current)
}

func TestMachine_StepByHook(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.NoError(showProofMachine.Initialize())
	transition := showProofMachine.TriggersByHook()
	assert.NotNil(transition)
	hookData := map[string]string{"key": "value"}
	e := transition.BuildSendEventsFromHook(hookData)
	assert.SNotNil(e)
	assert.SNotEmpty(e)
	event := e[0]
	assert.Equal(HookProtocol, int(event.ProtocolType))
	assert.Equal("output-hook", event.Hook.Data["ID"])
	showProofMachine.Step(transition)
	assert.Equal("IDLE", showProofMachine.Current)
}

func TestMachine_Start(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.NoError(machine.Initialize())
	sends := machine.Start(nil)
	assert.SNotNil(sends)
	assert.SLen(sends, 1)
	assert.NotNil(sends[0].Transition)
	assert.NotNil(sends[0].Transition.Machine)
}

func TestMachine_Start_ProofMachine(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.NoError(showProofMachine.Initialize())
	sends := showProofMachine.Start(nil)
	assert.SLen(sends, 0)
}

func Test_filterEnvs(t *testing.T) {
	type args struct {
		i      string
		envKey string
		envVal string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{"simplest",
			args{
				`${CRED_DEF_ID}`,
				"CRED_DEF_ID",
				"2K74emXCd9H8FV54MbVYjD:3:CL:13:TAG",
			},
			`2K74emXCd9H8FV54MbVYjD:3:CL:13:TAG`},
		{"simple",
			args{
				`[{"name":"email","credDefId":"${CRED_DEF_ID}"}]`,
				"CRED_DEF_ID",
				"2K74emXCd9H8FV54MbVYjD:3:CL:13:TAG",
			},
			`[{"name":"email","credDefId":"2K74emXCd9H8FV54MbVYjD:3:CL:13:TAG"}]`},
		{"double",
			args{
				`${CRED_DEF_ID}${CRED_DEF_ID}`,
				"CRED_DEF_ID",
				"2K74emXCd9H8FV54MbVYjD:3:CL:13:TAG",
			},
			`2K74emXCd9H8FV54MbVYjD:3:CL:13:TAG2K74emXCd9H8FV54MbVYjD:3:CL:13:TAG`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.PushTester(t)
			defer assert.PopTester()

			os.Setenv(tt.args.envKey, tt.args.envVal)
			got := filterEnvs(tt.args.i)
			os.Setenv(tt.args.envKey, "")
			assert.Equal(got, tt.want)
		})
	}
}
