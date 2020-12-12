package fsm

import (
	"testing"

	"github.com/findy-network/findy-agent-api/grpc/agency"
	"github.com/findy-network/findy-grpc/agency/client/chat"
	"github.com/stretchr/testify/assert"
)

var (
	plantUMLMachine = `
@startuml
title email issuer machine
[*] -> IDLE
state "                  IDLE                  " as IDLE
IDLE --> WAITING_EMAIL_ADDRESS: **basic_message{ ""}**\n{basic_message{ " Hello! I'm "}} ==>\n

state "          WAITING_EMAIL_ADDRESS         " as WAITING_EMAIL_ADDRESS
WAITING_EMAIL_ADDRESS --> WAITING_EMAIL_PIN: **basic_message{:= "EMAIL"}**\n{basic_message{%s "Thank you! I"}} ==>\n{email{new PIN "{"from":"cha"}} ==>\n

state "            WAITING_EMAIL_PIN           " as WAITING_EMAIL_PIN
WAITING_EMAIL_PIN --> WAITING_EMAIL_ADDRESS: **basic_message{== "reset"}**\n{basic_message{ "Please enter"}} ==>\n
WAITING_EMAIL_PIN --> WAITING_EMAIL_PIN: **basic_message{!= "PIN"}**\n{basic_message{%s "Incorrect PI"}} ==>\n
WAITING_EMAIL_PIN --> WAITING_ISSUING_STATUS: **basic_message{== "PIN"}**\n{basic_message{%s "Thank you! I"}} ==>\n{issue_cred{%s "[{"name":"em"}} ==>\n

state "         WAITING_ISSUING_STATUS         " as WAITING_ISSUING_STATUS
WAITING_ISSUING_STATUS --> IDLE: **issue_cred{STATUS ""}**\n{basic_message{%s "Thank you {{"}} ==>\n

@enduml


`
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
	//assert.NotNil(t, sends)
	assert.Len(t, sends, 0)
}

func TestMachine_Print(t *testing.T) {
	assert.NoError(t, chat.EmailIssuerMachine.Initialize())
	plantUml := chat.EmailIssuerMachine.String()
	print(plantUml)
	assert.NotEmpty(t, plantUml)
	assert.NoError(t, chat.ReqProofMachine.Initialize())
	plantUml = chat.ReqProofMachine.String()
	print(plantUml)
	assert.NotEmpty(t, plantUml)
}
