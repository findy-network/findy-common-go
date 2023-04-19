package chat

import "github.com/findy-network/findy-common-go/agency/fsm"

var EmailIssuerMachine = fsm.Machine{
	Name: "email issuer machine",
	Type: fsm.MachineTypeConversation,
	Initial: &fsm.Transition{
		Sends: []*fsm.Event{{
			Protocol: "basic_message",
			Data:     "Hello!",
		}},
		Target: "IDLE",
	},
	States: map[string]*fsm.State{
		"IDLE": {
			Transitions: []*fsm.Transition{{
				Trigger: &fsm.Event{
					Protocol: "basic_message",
				},
				Sends: []*fsm.Event{{
					Protocol: "basic_message",
					Data: `
Hello! I'm a email issuer.
Please enter your email address.`,
				}},
				Target: "WAITING_EMAIL_ADDRESS",
			}},
		},
		"WAITING_EMAIL_ADDRESS": {
			Transitions: []*fsm.Transition{{
				Trigger: &fsm.Event{
					Protocol: "basic_message",
					Rule:     "INPUT_SAVE",
					Data:     "EMAIL",
				},
				Sends: []*fsm.Event{
					{
						Protocol: "basic_message",
						Rule:     "FORMAT_MEM",
						Data: `Thank you! I sent your pin code to {{.EMAIL}}.
Please enter it here and I'll send your email credential.
Say "reset" if you want to start over.`,
					},
					{
						Protocol: "email",
						Rule:     "GEN_PIN",
						Data: `{"from":"chatbot@our.address.net",
"subject":"Your PIN for email issuer chat bot",
"to":"{{.EMAIL}}",
"body":"Thank you! This is your pin code:\n{{.PIN}}\nPlease enter it back to me, the chat bot, and I'll give your credential."}`,
					},
				},
				Target: "WAITING_EMAIL_PIN",
			}},
		},
		"WAITING_EMAIL_PIN": {
			Transitions: []*fsm.Transition{
				{
					Trigger: &fsm.Event{
						Protocol: "basic_message",
						Rule:     "INPUT_EQUAL",
						Data:     "reset",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Data:     "Please enter your email address.",
						},
					},
					Target: "WAITING_EMAIL_ADDRESS",
				},
				{
					Trigger: &fsm.Event{
						Protocol: "basic_message",
						Rule:     "INPUT_VALIDATE_NOT_EQUAL",
						Data:     "PIN",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Rule:     "FORMAT_MEM",
							Data: `Incorrect PIN code. Please check your emails for:
{{.EMAIL}}`,
						},
					},
					Target: "WAITING_EMAIL_PIN",
				},
				{
					Trigger: &fsm.Event{
						Protocol: "basic_message",
						Rule:     "INPUT_VALIDATE_EQUAL", // validation criterion is will be in??
						Data:     "PIN",                  // this is the name of the memory we are using
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Rule:     "FORMAT_MEM",
							Data: `Thank you! Issuing an email credential for address:
{{.EMAIL}}
Please follow your wallet app's instructions`,
						},
						{
							Protocol: "issue_cred",
							Rule:     "FORMAT_MEM",
							Data:     `[{"name":"email","value":"{{.EMAIL}}"}]`,
							EventData: &fsm.EventData{
								Issuing: &fsm.Issuing{
									CredDefID: "T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1",
									AttrsJSON: `[{"name":"email","value":"{{.EMAIL}}"}]`,
								},
							},
						},
					},
					Target: "WAITING_ISSUING_STATUS",
				},
			},
		},
		"WAITING_ISSUING_STATUS": {
			Transitions: []*fsm.Transition{{
				Trigger: &fsm.Event{
					Protocol: "issue_cred", // there was no questions when it was us who started the issuing
					Rule:     "OUR_STATUS",
				},
				Sends: []*fsm.Event{
					{
						Protocol: "basic_message",
						Rule:     "FORMAT_MEM",
						Data: `Thank you {{.EMAIL}}!
We are ready now. Bye bye!`,
					},
				},
				Target: "IDLE",
			}},
		},
	},
}

var ReqProofMachine = fsm.Machine{
	Type: fsm.MachineTypeConversation,
	Name: "machine",
	Initial: &fsm.Transition{
		Sends: []*fsm.Event{{
			Protocol: "basic_message",
			Data:     "Hello!",
		}},
		Target: "INITIAL",
	},
	States: map[string]*fsm.State{
		"INITIAL": {
			Transitions: []*fsm.Transition{
				{
					Trigger: &fsm.Event{
						Protocol: "connection",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Data:     "Hello! I'm echo bot.\nFirst I need your verified email.\nI'm now sending you a proof request.\nPlease accept it and we can continue.",
						},
					},
					Target: "INITIAL",
				},
				{
					Trigger: &fsm.Event{
						Protocol: "basic_message",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "present_proof",
							Data:     `[{"name":"email","credDefId":"T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1"}]`,
						},
					},
					Target: "WAITING_EMAIL_PROOF_QA",
				},
			},
		},
		"WAITING_EMAIL_PROOF_QA": {
			Transitions: []*fsm.Transition{
				{
					Trigger: &fsm.Event{
						Protocol: "basic_message",
						Rule:     "INPUT_EQUAL",
						Data:     "reset",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Data:     "Going to beginning...",
						},
					},
					Target: "INITIAL",
				}, // reset cmd
				{
					Trigger: &fsm.Event{
						Protocol: "present_proof",
						TypeID:   "ANSWER_NEEDED_PROOF_VERIFY", // this should/could be a protocol?
						Rule:     "NOT_ACCEPT_VALUES",
						Data:     `[{"name":"email","credDefId":"T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1"}]`,
					},
					Sends: []*fsm.Event{
						{
							Protocol: "answer",
							Data:     "NACK",
						},
					},
					Target: "INITIAL",
				},
				{
					Trigger: &fsm.Event{
						Protocol: "present_proof",
						TypeID:   "ANSWER_NEEDED_PROOF_VERIFY",
						Rule:     "ACCEPT_AND_INPUT_VALUES",
						Data:     `[{"name":"email","credDefId":"T2o5osjKcK6oVDPxcLjKnB:3:CL:T2o5osjKcK6oVDPxcLjKnB:2:my-schema:1.0:t1"}]`,
					},
					Sends: []*fsm.Event{
						{
							Protocol: "answer",
							Data:     `ACK`,
						},
					},
					Target: "WAITING2_EMAIL_PROOF",
				},
			},
		},
		// we don't have any error handling here. TODO: this shows that we need timers when we have error handling!
		"WAITING2_EMAIL_PROOF": {
			Transitions: []*fsm.Transition{
				{
					Trigger: &fsm.Event{
						Protocol: "present_proof", // this is the final message that proof has been finalized and ready for pickup
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Rule:     "FORMAT_MEM",
							Data:     "Hello {{.email}}! I'm stupid bot who knows you have verified email address!!!\nI can trust you.",
						},
					},
					Target: "WAITING_NEXT_CMD",
				},
			},
		},
		"WAITING_NEXT_CMD": {
			Transitions: []*fsm.Transition{
				{
					Trigger: &fsm.Event{
						Protocol: "basic_message",
						Rule:     "INPUT_EQUAL",
						Data:     "reset",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Data:     "Going to beginning.",
						},
					},
					Target: "INITIAL",
				},
				{
					Trigger: &fsm.Event{
						Protocol: "basic_message",
						Rule:     "INPUT_SAVE",
						Data:     "LINE",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Rule:     "FORMAT_MEM",
							Data:     "{{.email}} says: {{.LINE}}",
						},
					},
					Target: "WAITING_NEXT_CMD",
				},
			},
		},
	},
}

var EchoMachine = fsm.Machine{
	Name: "echo machine",
	Type: fsm.MachineTypeConversation,
	Initial: &fsm.Transition{
		Sends: []*fsm.Event{{
			Protocol: "basic_message",
			Data:     "Hello!",
		}},
		Target: "INITIAL",
	},
	States: map[string]*fsm.State{
		"INITIAL": {
			Transitions: []*fsm.Transition{
				{
					Trigger: &fsm.Event{
						Protocol: "connection",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Data:     "Hello! I'm echo bot.\nSay: run, and I'start.\nSay: reset, and I'll go beginning.",
						},
					},
					Target: "INITIAL",
				},
				{
					Trigger: &fsm.Event{
						Protocol: "basic_message",
						Rule:     "INPUT_EQUAL",
						Data:     "run",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Data:     "Let's go!",
						},
					},
					Target: "IDLE",
				},
				{
					Trigger: &fsm.Event{
						Protocol: "basic_message",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Data:     "Hello! I'm echo bot.\nSay: run, and I'start.\nSay: reset, and I'll go beginning.",
						},
					},
					Target: "INITIAL",
				},
			},
		},
		"IDLE": {
			Transitions: []*fsm.Transition{
				{
					Trigger: &fsm.Event{
						Protocol: "basic_message",
						Rule:     "INPUT_EQUAL",
						Data:     "reset",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Data:     "Going to beginning.",
						},
					},
					Target: "INITIAL",
				},
				{
					Trigger: &fsm.Event{
						Protocol: "basic_message",
						Rule:     "INPUT",
					},
					Sends: []*fsm.Event{
						{
							Protocol: "basic_message",
							Rule:     "INPUT",
						},
					},
					Target: "IDLE",
				},
			},
		},
	},
}
