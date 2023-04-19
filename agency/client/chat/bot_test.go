package chat

import (
	"testing"

	"github.com/findy-network/findy-common-go/agency/fsm"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
)

var issuerFSMYaml = `
type: MachineTypeConversation
name: email issuer machine
initial:
  sends:
  - data: Hello!
    protocol: basic_message
  target: IDLE
states:
  IDLE:
    transitions:
    - sends:
      - data: |2-

          Hello! I'm a email issuer.
          Please enter your email address.
        protocol: basic_message
      target: WAITING_EMAIL_ADDRESS
      trigger:
        protocol: basic_message
  WAITING_EMAIL_ADDRESS:
    transitions:
    - sends:
      - data: |-
          Thank you! I sent your pin code to {{.EMAIL}}.
          Please enter it here and I'll send your email credential.
          Say "reset" if you want to start over.
        protocol: basic_message
        rule: FORMAT_MEM
      - data: |-
          {"from":"chatbot@our.address.net",
          "subject":"Your PIN for email issuer chat bot",
          "to":"{{.EMAIL}}",
          "body":"Thank you! This is your pin code:\n{{.PIN}}\nPlease enter it back to me, the chat bot, and I'll give your credential."}
        protocol: email
        rule: GEN_PIN
      target: WAITING_EMAIL_PIN
      trigger:
        data: EMAIL
        protocol: basic_message
        rule: INPUT_SAVE
  WAITING_EMAIL_PIN:
    transitions:
    - sends:
      - data: Please enter your email address.
        protocol: basic_message
      target: WAITING_EMAIL_ADDRESS
      trigger:
        data: reset
        protocol: basic_message
        rule: INPUT_EQUAL
    - sends:
      - data: |-
          Incorrect PIN code. Please check your emails for:
          {{.EMAIL}}
        protocol: basic_message
        rule: FORMAT_MEM
      target: WAITING_EMAIL_PIN
      trigger:
        data: PIN
        protocol: basic_message
        rule: INPUT_VALIDATE_NOT_EQUAL
    - sends:
      - data: |-
          Thank you! Issuing an email credential for address:
          {{.EMAIL}}
          Please follow your wallet app's instructions
        protocol: basic_message
        rule: FORMAT_MEM
      - data: '[{"name":"email","value":"{{.EMAIL}}"}]'
        event_data:
          issuing:
            AttrsJSON: '[{"name":"email","value":"{{.EMAIL}}"}]'
            CredDefID: T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1
        protocol: issue_cred
        rule: FORMAT_MEM
      target: WAITING_ISSUING_STATUS
      trigger:
        data: PIN
        protocol: basic_message
        rule: INPUT_VALIDATE_EQUAL
  WAITING_ISSUING_STATUS:
    transitions:
    - sends:
      - data: |-
          Thank you {{.EMAIL}}!
          We are ready now. Bye bye!
        protocol: basic_message
        rule: FORMAT_MEM
      target: IDLE
      trigger:
        protocol: issue_cred
        rule: OUR_STATUS
`

var proofFSMJson = `{
        "type": "MachineTypeConversation",
        "name": "machine",
        "initial": {
                "sends": [
                        {
                                "protocol": "basic_message",
                                "type_id": "",
                                "rule": "",
                                "data": "Hello!"
                        }
                ],
                "target": "INITIAL"
        },
        "states": {
                "INITIAL": {
                        "transitions": [
                                {
                                        "trigger": {
                                                "protocol": "connection",
                                                "type_id": "",
                                                "rule": ""
                                        },
                                        "sends": [
                                                {
                                                        "protocol": "basic_message",
                                                        "type_id": "",
                                                        "rule": "",
                                                        "data": "Hello! I'm echo bot.\nFirst I need your verified email.\nI'm now sending you a proof request.\nPlease accept it and we can continue."
                                                }
                                        ],
                                        "target": "INITIAL"
                                },
                                {
                                        "trigger": {
                                                "protocol": "basic_message",
                                                "type_id": "",
                                                "rule": ""
                                        },
                                        "sends": [
                                                {
                                                        "protocol": "present_proof",
                                                        "type_id": "",
                                                        "rule": "",
                                                        "data": "[{\"name\":\"email\",\"credDefId\":\"T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1\"}]",
                                                        "want_status": true
                                                }
                                        ],
                                        "target": "WAITING_EMAIL_PROOF_QA"
                                }
                        ]
                },
                "WAITING2_EMAIL_PROOF": {
                        "transitions": [
                                {
                                        "trigger": {
                                                "protocol": "present_proof",
                                                "type_id": "",
                                                "rule": ""
                                        },
                                        "sends": [
                                                {
                                                        "protocol": "basic_message",
                                                        "type_id": "",
                                                        "rule": "FORMAT_MEM",
                                                        "data": "Hello {{.email}}! I'm stupid bot who knows you have verified email address!!!\nI can trust you."
                                                }
                                        ],
                                        "target": "WAITING_NEXT_CMD"
                                }
                        ]
                },
                "WAITING_EMAIL_PROOF_QA": {
                        "transitions": [
                                {
                                        "trigger": {
                                                "protocol": "basic_message",
                                                "type_id": "",
                                                "rule": "INPUT_EQUAL",
                                                "data": "reset"
                                        },
                                        "sends": [
                                                {
                                                        "protocol": "basic_message",
                                                        "type_id": "",
                                                        "rule": "",
                                                        "data": "Going to beginning..."
                                                }
                                        ],
                                        "target": "INITIAL"
                                },
                                {
                                        "trigger": {
                                                "protocol": "present_proof",
                                                "type_id": "ANSWER_NEEDED_PROOF_VERIFY",
                                                "rule": "NOT_ACCEPT_VALUES",
                                                "data": "[{\"name\":\"email\",\"credDefId\":\"T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1\"}]"
                                        },
                                        "sends": [
                                                {
                                                        "protocol": "answer",
                                                        "type_id": "",
                                                        "rule": "",
                                                        "data": "NACK"
                                                }
                                        ],
                                        "target": "INITIAL"
                                },
                                {
                                        "trigger": {
                                                "protocol": "present_proof",
                                                "type_id": "ANSWER_NEEDED_PROOF_VERIFY",
                                                "rule": "ACCEPT_AND_INPUT_VALUES",
                                                "data": "[{\"name\":\"email\",\"credDefId\":\"T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1\"}]"
                                        },
                                        "sends": [
                                                {
                                                        "protocol": "answer",
                                                        "type_id": "",
                                                        "rule": "",
                                                        "data": "ACK"
                                                }
                                        ],
                                        "target": "WAITING2_EMAIL_PROOF"
                                }
                        ]
                },
                "WAITING_NEXT_CMD": {
                        "transitions": [
                                {
                                        "trigger": {
                                                "protocol": "basic_message",
                                                "type_id": "",
                                                "rule": "INPUT_EQUAL",
                                                "data": "reset"
                                        },
                                        "sends": [
                                                {
                                                        "protocol": "basic_message",
                                                        "type_id": "",
                                                        "rule": "",
                                                        "data": "Going to beginning."
                                                }
                                        ],
                                        "target": "INITIAL"
                                },
                                {
                                        "trigger": {
                                                "protocol": "basic_message",
                                                "type_id": "",
                                                "rule": "INPUT_SAVE",
                                                "data": "LINE"
                                        },
                                        "sends": [
                                                {
                                                        "protocol": "basic_message",
                                                        "type_id": "",
                                                        "rule": "FORMAT_MEM",
                                                        "data": "{{.email}} says: {{.LINE}}"
                                                }
                                        ],
                                        "target": "WAITING_NEXT_CMD"
                                }
                        ]
                }
        }
}`

var proofFSMYaml = `
type: MachineTypeConversation
name: machine
initial:
  sends:
  - data: Hello!
    protocol: basic_message
  target: INITIAL
states:
  INITIAL:
    transitions:
    - sends:
      - data: |-
          Hello! I'm echo bot.
          First I need your verified email.
          I'm now sending you a proof request.
          Please accept it and we can continue.
        protocol: basic_message
      target: INITIAL
      trigger:
        protocol: connection
    - sends:
      - data: '[{"name":"email","credDefId":"T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1"}]'
        protocol: present_proof
      target: WAITING_EMAIL_PROOF_QA
      trigger:
        protocol: basic_message
  WAITING_EMAIL_PROOF_QA:
    transitions:
    - sends:
      - data: Going to beginning...
        protocol: basic_message
      target: INITIAL
      trigger:
        data: reset
        protocol: basic_message
        rule: INPUT_EQUAL
    - sends:
      - data: NACK
        protocol: answer
      target: INITIAL
      trigger:
        data: '[{"name":"email","credDefId":"T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1"}]'
        protocol: present_proof
        rule: NOT_ACCEPT_VALUES
        type_id: ANSWER_NEEDED_PROOF_VERIFY
    - sends:
      - data: ACK
        protocol: answer
      target: WAITING2_EMAIL_PROOF
      trigger:
        data: '[{"name":"email","credDefId":"T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1"}]'
        protocol: present_proof
        rule: ACCEPT_AND_INPUT_VALUES
        type_id: ANSWER_NEEDED_PROOF_VERIFY
  WAITING_NEXT_CMD:
    transitions:
    - sends:
      - data: Going to beginning.
        protocol: basic_message
      target: INITIAL
      trigger:
        data: reset
        protocol: basic_message
        rule: INPUT_EQUAL
    - sends:
      - data: '{{.email}} says: {{.LINE}}'
        protocol: basic_message
        rule: FORMAT_MEM
      target: WAITING_NEXT_CMD
      trigger:
        data: LINE
        protocol: basic_message
        rule: INPUT_SAVE
  WAITING2_EMAIL_PROOF:
    transitions:
    - sends:
      - data: |-
          Hello {{.email}}! I'm stupid bot who knows you have verified email address!!!
          I can trust you.
        protocol: basic_message
        rule: FORMAT_MEM
      target: WAITING_NEXT_CMD
      trigger:
        protocol: present_proof
`

func Test_loadFSMData(t *testing.T) {
	type args struct {
		fName string
		data  []byte
	}
	tests := []struct {
		name    string
		args    args
		wantFsm *fsm.Machine
	}{
		{name: "issuer machine",
			args:    args{fName: "file.yaml", data: []byte(issuerFSMYaml)},
			wantFsm: &EmailIssuerMachine,
		},
		{name: "proof machine from json",
			args:    args{fName: "file.json", data: []byte(proofFSMJson)},
			wantFsm: &ReqProofMachine,
		},
		{name: "proof machine from json",
			args:    args{fName: "file.yaml", data: []byte(proofFSMYaml)},
			wantFsm: &ReqProofMachine,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.PushTester(t)
			defer assert.PopTester()

			defer err2.Catch(func(err error) {
				assert.NoError(err)
			})

			gotFsm := loadFSMData(tt.args.fName, tt.args.data)
			assert.Equal(gotFsm.Type, tt.wantFsm.Type)
			assert.Equal(gotFsm.Name, tt.wantFsm.Name)

			// test after Initialize call
			err := gotFsm.Initialize()
			assert.NoError(err)
			err = tt.wantFsm.Initialize()
			assert.NoError(err)
			assert.DeepEqual(gotFsm, tt.wantFsm)
		})
	}
}

func Test_marshalFSM(t *testing.T) {
	type args struct {
		fName string
		fsm   *fsm.Machine
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "proof machine in json", args: args{fName: "file.json", fsm: &ReqProofMachine}},
		{name: "proof machine in yaml", args: args{fName: "file.yaml", fsm: &ReqProofMachine}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.PushTester(t)
			defer assert.PopTester()

			err := tt.args.fsm.Initialize()
			assert.NoError(err)
			gotbytes := marshalFSM(tt.args.fName, tt.args.fsm)
			glog.V(3).Infoln(len(gotbytes))
			glog.V(3).Infoln(string(gotbytes))

			gotMachine := loadFSMData(tt.args.fName, gotbytes)
			// we know that ReqProofMachine is Initialized
			err = gotMachine.Initialize()
			assert.NoError(err)
			assert.DeepEqual(gotMachine, tt.args.fsm,
				"we should get same machine after marshal round")
		})
	}
}
