package fsm

import (
	"os"
	"testing"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/lainio/err2/assert"
)

var (
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

func protocolStatus(typeID agency.Protocol_Type) *agency.ProtocolStatus {
	agencyProof := &agency.ProtocolStatus{
		State: &agency.ProtocolState{ProtocolID: &agency.ProtocolID{TypeID: typeID}},
		Status: &agency.ProtocolStatus_BasicMessage{
			BasicMessage: &agency.ProtocolStatus_BasicMessageStatus{
				Content: "test string",
			},
		},
	}
	return agencyProof
}

func TestMachine_Step(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.NoError(machine.Initialize())
	transition := machine.Triggers(protocolStatus(agency.Protocol_PRESENT_PROOF))
	assert.That(nil == transition)
	transition = machine.Triggers(protocolStatus(agency.Protocol_BASIC_MESSAGE))
	assert.INotNil(transition)
	machine.Step(transition)
	assert.Equal("WAITING_STATUS", machine.Current)
}

func TestMachine_StepTerminate(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.NoError(machineTerminates.Initialize())
	termChan := make(TerminateChan)
	go func() {
		assert.PushTester(t)
		defer assert.PopTester()
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
	assert.INotNil(transition)
	assert.Equal("IDLE", machineTerminates.Current)
	go machineTerminates.Step(transition)
}

func TestMachine_Step2(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.NoError(showProofMachine.Initialize())
	status := protocolStatus(agency.Protocol_DIDEXCHANGE)
	transition := showProofMachine.Triggers(status)
	assert.INotNil(transition)
	e := transition.buildInputEvent(status)
	assert.INotNil(e)
	o := transition.BuildSendEvents(status)
	assert.INotNil(o)
	showProofMachine.Step(transition)
	assert.Equal("WAITING_STATUS", showProofMachine.Current)
}

func TestMachine_StepByHook(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.NoError(showProofMachine.Initialize())
	transition := showProofMachine.TriggersByHook()
	assert.INotNil(transition)
	hookData := map[string]string{"key": "value"}
	e := transition.BuildSendEventsFromHook(hookData)
	assert.INotNil(e)
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
	assert.INotNil(sends)
	assert.SLen(sends, 1)
	assert.INotNil(sends[0].Transition)
	assert.INotNil(sends[0].Transition.Machine)
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
