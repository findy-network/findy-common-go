package fsm

import (
	"testing"

	"github.com/findy-network/findy-agent-api/grpc/agency"
	"github.com/stretchr/testify/assert"
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
	assert.False(t, machine.Initialized)
	assert.Zero(t, machine.States["IDLE"].Transitions[0].Trigger.ProtocolType)
	assert.Zero(t, machine.States["IDLE"].Transitions[0].Sends[0].ProtocolType)
	assert.NoError(t, machine.Initialize())
	assert.True(t, machine.Initialized)
	assert.Equal(t, agency.Protocol_BASIC_MESSAGE,
		machine.States["IDLE"].Transitions[0].Trigger.ProtocolType)
	assert.Equal(t, agency.Protocol_BASIC_MESSAGE,
		machine.States["IDLE"].Transitions[0].Sends[0].ProtocolType)
}

func TestMachine_Triggers(t *testing.T) {
	assert.NoError(t, machine.Initialize())
	assert.Nil(t, machine.Triggers(
		protocolStatus(agency.Protocol_PROOF)))
	assert.NotNil(t, machine.Triggers(
		protocolStatus(agency.Protocol_BASIC_MESSAGE)))
}

func protocolStatus(typeID agency.Protocol_Type) *agency.ProtocolStatus {
	agencyProof := &agency.ProtocolStatus{
		State: &agency.ProtocolState{ProtocolId: &agency.ProtocolID{TypeId: typeID}},
		Status: &agency.ProtocolStatus_BasicMessage_{BasicMessage: &agency.ProtocolStatus_BasicMessage{
			Content: "test string",
		}},
	}
	return agencyProof
}

func TestMachine_Step(t *testing.T) {
	assert.NoError(t, machine.Initialize())
	transition := machine.Triggers(protocolStatus(agency.Protocol_PROOF))
	assert.Nil(t, transition)
	transition = machine.Triggers(protocolStatus(agency.Protocol_BASIC_MESSAGE))
	assert.NotNil(t, transition)
	machine.Step(transition)
	assert.Equal(t, "WAITING_STATUS", machine.Current)
}

func TestMachine_Step2(t *testing.T) {
	assert.NoError(t, showProofMachine.Initialize())
	status := protocolStatus(agency.Protocol_CONNECT)
	transition := showProofMachine.Triggers(status)
	assert.NotNil(t, transition)
	e := transition.buildInputEvent(status)
	assert.NotNil(t, e)
	o := transition.BuildSendEvents(status)
	assert.NotNil(t, o)
	showProofMachine.Step(transition)
	assert.Equal(t, "WAITING_STATUS", showProofMachine.Current)
}

func TestMachine_StepByHook(t *testing.T) {
	assert.NoError(t, showProofMachine.Initialize())
	transition := showProofMachine.TriggersByHook()
	assert.NotNil(t, transition)
	hookData := map[string]string{"key": "value"}
	e := transition.BuildSendEventsFromHook(hookData)
	assert.NotNil(t, e)
	assert.NotEmpty(t, e)
	event := e[0]
	assert.Equal(t, HookProtocol, int(event.ProtocolType))
	assert.Equal(t, "output-hook", event.Hook.Data["ID"])
	showProofMachine.Step(transition)
	assert.Equal(t, "IDLE", showProofMachine.Current)
}

func TestMachine_Start(t *testing.T) {
	assert.NoError(t, machine.Initialize())
	sends := machine.Start()
	assert.NotNil(t, sends)
	assert.Len(t, sends, 1)
	assert.NotNil(t, sends[0].Transition)
	assert.NotNil(t, sends[0].Transition.Machine)
}

func TestMachine_Start_ProofMachine(t *testing.T) {
	assert.NoError(t, showProofMachine.Initialize())
	sends := showProofMachine.Start()
	assert.Len(t, sends, 0)
}
