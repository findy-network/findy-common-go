package fsm

import (
	"testing"

	"github.com/findy-network/findy-agent-api/grpc/agency"
	"github.com/stretchr/testify/assert"
)

var (
	machine = Machine{
		Initial: "IDLE",
		States: map[string]State{
			"IDLE": {
				ID: "IDLE",
				Transitions: []Transition{{
					Trigger: Event{TypeID: "basic_message"},
					Sends: []Event{{
						TypeID: "basic_message",
						Rule:   "INPUT",
					}},
					Target: "WAITING_STATUS",
				}},
			},
			"WAITING_STATUS": {
				ID: "WAITING_STATUS",
				Transitions: []Transition{{
					Trigger: Event{
						TypeID: "basic_message",
						Rule:   "OUR_STATUS",
					},
					Sends: []Event{{
						TypeID: "basic_message",
						Rule:   "OUR_STATUS",
					}},
					Target: "IDLE",
				}},
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
	assert.Nil(t, machine.Triggers(agency.Protocol_PROOF))
	assert.NotNil(t, machine.Triggers(agency.Protocol_BASIC_MESSAGE))
}

func TestMachine_Step(t *testing.T) {
	assert.NoError(t, machine.Initialize())
	transition := machine.Triggers(agency.Protocol_PROOF)
	assert.Nil(t, transition)
	transition = machine.Triggers(agency.Protocol_BASIC_MESSAGE)
	assert.NotNil(t, transition)
	machine.Step(transition)
	assert.Equal(t, "WAITING_STATUS", machine.Current)
}
