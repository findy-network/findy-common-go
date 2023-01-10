package fsm

import (
	"testing"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/lainio/err2/assert"
)

var (
	machine = Machine{
		Name: "machine",
		Initial: &Transition{
			Sends: []*Event{{
				Protocol: "basic_message",
				Data:     "Hello!",
				NoStatus: true,
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

func TestMachine_Triggers(t *testing.T) {
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
	sends := machine.Start()
	assert.INotNil(sends)
	assert.SLen(sends, 1)
	assert.INotNil(sends[0].Transition)
	assert.INotNil(sends[0].Transition.Machine)
}

func TestMachine_Start_ProofMachine(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	assert.NoError(showProofMachine.Initialize())
	sends := showProofMachine.Start()
	assert.SLen(sends, 0)
}
