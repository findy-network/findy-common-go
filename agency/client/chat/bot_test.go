package chat

import (
	"reflect"
	"testing"

	"github.com/findy-network/findy-grpc/agency/fsm"
)

var issuerFSMYaml = `initial:
  sends:
  - data: Hello!
    no_status: true
    protocol: basic_message
    rule: ""
    type_id: ""
  target: IDLE
name: email issuer machine
states:
  IDLE:
    transitions:
    - sends:
      - data: |2-

          Hello! I'm a email issuer.
          Please enter your email address.
        no_status: true
        protocol: basic_message
        rule: ""
        type_id: ""
      target: WAITING_EMAIL_ADDRESS
      trigger:
        protocol: basic_message
        rule: ""
        type_id: ""
  WAITING_EMAIL_ADDRESS:
    transitions:
    - sends:
      - data: |-
          Thank you! I sent your pin code to {{.EMAIL}}.
          Please enter it here and I'll send your email credential.
          Say "reset" if you want to start over.
        no_status: true
        protocol: basic_message
        rule: FORMAT_MEM
        type_id: ""
      - data: |-
          {"from":"chatbot@our.address.net",
          "subject":"Your PIN for email issuer chat bot",
          "to":"{{.EMAIL}}",
          "body":"Thank you! This is your pin code:\n{{.PIN}}\nPlease enter it back to me, the chat bot, and I'll give your credential."}
        no_status: true
        protocol: email
        rule: GEN_PIN
        type_id: ""
      target: WAITING_EMAIL_PIN
      trigger:
        data: EMAIL
        protocol: basic_message
        rule: INPUT_SAVE
        type_id: ""
  WAITING_EMAIL_PIN:
    transitions:
    - sends:
      - data: Please enter your email address.
        no_status: true
        protocol: basic_message
        rule: ""
        type_id: ""
      target: WAITING_EMAIL_ADDRESS
      trigger:
        data: reset
        protocol: basic_message
        rule: INPUT_EQUAL
        type_id: ""
    - sends:
      - data: |-
          Incorrect PIN code. Please check your emails for:
          {{.EMAIL}}
        no_status: true
        protocol: basic_message
        rule: FORMAT_MEM
        type_id: ""
      target: WAITING_EMAIL_PIN
      trigger:
        data: PIN
        protocol: basic_message
        rule: INPUT_VALIDATE_NOT_EQUAL
        type_id: ""
    - sends:
      - data: |-
          Thank you! Issuing an email credential for address:
          {{.EMAIL}}
          Please follow your wallet app's instructions
        no_status: true
        protocol: basic_message
        rule: FORMAT_MEM
        type_id: ""
      - data: '[{"name":"email","value":"{{.EMAIL}}"}]'
        event_data:
          issuing:
            AttrsJSON: '[{"name":"email","value":"{{.EMAIL}}"}]'
            CredDefID: T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1
        protocol: issue_cred
        rule: FORMAT_MEM
        type_id: ""
      target: WAITING_ISSUING_STATUS
      trigger:
        data: PIN
        protocol: basic_message
        rule: INPUT_VALIDATE_EQUAL
        type_id: ""
  WAITING_ISSUING_STATUS:
    transitions:
    - sends:
      - data: |-
          Thank you {{.EMAIL}}!
          We are ready now. Bye bye!
        no_status: true
        protocol: basic_message
        rule: FORMAT_MEM
        type_id: ""
      target: IDLE
      trigger:
        protocol: issue_cred
        rule: OUR_STATUS
        type_id: ""
`

var proofFSMYaml = `initial:
  sends:
  - data: Hello!
    no_status: true
    protocol: basic_message
    rule: ""
    type_id: ""
  target: INITIAL
name: machine
states:
  INITIAL:
    transitions:
    - sends:
      - data: |-
          Hello! I'm echo bot.
          First I need your verified email.
          I'm now sending you a proof request.
          Please accept it and we can continue.
        no_status: true
        protocol: basic_message
        rule: ""
        type_id: ""
      target: INITIAL
      trigger:
        protocol: connection
        rule: ""
        type_id: ""
    - sends:
      - data: '[{"name":"email","credDefId":"T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1"}]'
        protocol: present_proof
        rule: ""
        type_id: ""
      target: WAITING_EMAIL_PROOF_QA
      trigger:
        protocol: basic_message
        rule: ""
        type_id: ""
  WAITING_EMAIL_PROOF_QA:
    transitions:
    - sends:
      - data: Going to beginning...
        no_status: true
        protocol: basic_message
        rule: ""
        type_id: ""
      target: INITIAL
      trigger:
        data: reset
        protocol: basic_message
        rule: INPUT_EQUAL
        type_id: ""
    - sends:
      - data: NACK
        protocol: answer
        rule: ""
        type_id: ""
      target: INITIAL
      trigger:
        data: '[{"name":"email","credDefId":"T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1"}]'
        protocol: present_proof
        rule: NOT_ACCEPT_VALUES
        type_id: ANSWER_NEEDED_PROOF_VERIFY
    - sends:
      - data: ACK
        protocol: answer
        rule: ""
        type_id: ""
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
        no_status: true
        protocol: basic_message
        rule: ""
        type_id: ""
      target: INITIAL
      trigger:
        data: reset
        protocol: basic_message
        rule: INPUT_EQUAL
        type_id: ""
    - sends:
      - data: '{{.email}} says: {{.LINE}}'
        no_status: true
        protocol: basic_message
        rule: FORMAT_MEM
        type_id: ""
      target: WAITING_NEXT_CMD
      trigger:
        data: LINE
        protocol: basic_message
        rule: INPUT_SAVE
        type_id: ""
  WAITING2_EMAIL_PROOF:
    transitions:
    - sends:
      - data: |-
          Hello {{.email}}! I'm stupid bot who knows you have verified email address!!!
          I can trust you.
        no_status: true
        protocol: basic_message
        rule: FORMAT_MEM
        type_id: ""
      target: WAITING_NEXT_CMD
      trigger:
        protocol: present_proof
        rule: ""
        type_id: ""`

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
		{name: "proof machine",
			args:    args{fName: "file.yaml", data: []byte(proofFSMYaml)},
			wantFsm: &ReqProofMachine,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFsm := loadFSMData(tt.args.fName, tt.args.data)
			if !reflect.DeepEqual(gotFsm, tt.wantFsm) {
				t.Errorf("loadFSMData() = %v, want %v", gotFsm, tt.wantFsm)
			}
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
		want string
	}{
		{name: "proof machine", args: args{fName: "file.yaml", fsm: &ReqProofMachine},
			want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := string(marshalFSM(tt.args.fName, tt.args.fsm))
			if reflect.DeepEqual(got, tt.want) {
				t.Errorf("marshalFSM() = %v, want %v", got, tt.want)
			}
		})
	}
}
