package chat

import "github.com/findy-network/findy-grpc/agency/fsm"

var EmailIssuerMachine = fsm.Machine{
	Initial: "IDLE",
	States: map[string]fsm.State{
		"IDLE": {
			Transitions: []fsm.Transition{{
				Trigger: &fsm.Event{
					TypeID: "basic_message",
				},
				Sends: []*fsm.Event{{
					TypeID: "basic_message",
					Data: `
Hello! I'm a email issuer.
Please enter your email address.`,
					NoStatus: true,
				}},
				Target: "WAITING_EMAIL_ADDRESS",
			}},
		},
		"WAITING_EMAIL_ADDRESS": {
			Transitions: []fsm.Transition{{
				Trigger: &fsm.Event{
					TypeID: "basic_message",
					Rule:   "INPUT_SAVE",
					Data:   "EMAIL",
				},
				Sends: []*fsm.Event{
					{
						TypeID: "basic_message",
						Rule:   "FORMAT_MEM",
						Data: `Thank you! I sent your pin code to {{.EMAIL}}.
Please enter it here and I'll send your email credential.
Say "reset" if you want to start over.`,
						NoStatus: true,
					},
					{
						TypeID: "email",
						Rule:   "GEN_PIN",
						Data: `{"from":"chatbot@our.address.net",
"subject":"Your PIN for email issuer chat bot",
"to":"{{.EMAIL}}",
"body":"Thank you! This is your pin code:\n{{.PIN}}\nPlease enter it back to me, the chat bot, and I'll give your credential."}`,
						NoStatus: true,
					},
				},
				Target: "WAITING_EMAIL_PIN",
			}},
		},
		"WAITING_EMAIL_PIN": {
			Transitions: []fsm.Transition{
				{
					Trigger: &fsm.Event{
						TypeID: "basic_message",
						Rule:   "INPUT_EQUAL",
						Data:   "reset",
					},
					Sends: []*fsm.Event{
						{
							TypeID:   "basic_message",
							Data:     "Please enter your email address.",
							NoStatus: true,
						},
					},
					Target: "WAITING_EMAIL_ADDRESS",
				},
				{
					Trigger: &fsm.Event{
						TypeID: "basic_message",
						Rule:   "INPUT_VALIDATE_NOT_EQUAL",
						Data:   "PIN",
					},
					Sends: []*fsm.Event{
						{
							TypeID: "basic_message",
							Rule:   "FORMAT_MEM",
							Data: `Incorrect PIN code. Please check your emails for:
{{.EMAIL}}`,
							NoStatus: true,
						},
					},
					Target: "WAITING_EMAIL_PIN",
				},
				{
					Trigger: &fsm.Event{
						TypeID: "basic_message",
						Rule:   "INPUT_VALIDATE_EQUAL", // validation criterion is will be in??
						Data:   "PIN",                  // this is the name of the memory we are using
					},
					Sends: []*fsm.Event{
						{
							TypeID:   "basic_message",
							NoStatus: true,
							Rule:     "FORMAT_MEM",
							Data: `Thank you! Issuing an email credential for address:
{{.EMAIL}}
Please follow your wallet app's instructions`,
						},
						{
							TypeID: "issue_cred",
							Rule:   "FORMAT_MEM",
							Data:   `[{"name":"email","value":"{{.EMAIL}}"}]`,
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
			Transitions: []fsm.Transition{{
				Trigger: &fsm.Event{
					TypeID: "issue_cred", // there was no questions when it was us who started the issuing
					Rule:   "OUR_STATUS",
				},
				Sends: []*fsm.Event{
					{
						TypeID:   "basic_message",
						NoStatus: true,
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

var EchoMachine = fsm.Machine{
	Initial: "INITIAL",
	States: map[string]fsm.State{
		"INITIAL": {
			Transitions: []fsm.Transition{
				{
					Trigger: &fsm.Event{
						TypeID: "basic_message",
						Rule:   "INPUT_EQUAL",
						Data:   "run",
					},
					Sends: []*fsm.Event{
						{
							TypeID:   "basic_message",
							Data:     "Let's go!",
							NoStatus: true,
						},
					},
					Target: "IDLE",
				},
				{
					Trigger: &fsm.Event{
						TypeID: "basic_message",
						Rule:   "INPUT",
					},
					Sends: []*fsm.Event{
						{
							TypeID: "basic_message",
							//Rule: "",
							Data:     "Hello! I'm echo bot.\nSay: run, and I'start.\nSay: reset, and I'll go beginning.",
							NoStatus: true,
						},
					},
					Target: "INITIAL",
				},
			},
		},
		"IDLE": {
			Transitions: []fsm.Transition{
				{
					Trigger: &fsm.Event{
						TypeID: "basic_message",
						Rule:   "INPUT_EQUAL",
						Data:   "reset",
					},
					Sends: []*fsm.Event{
						{
							TypeID:   "basic_message",
							Data:     "Going to beginning.",
							NoStatus: true,
						},
					},
					Target: "INITIAL",
				},
				{
					Trigger: &fsm.Event{
						TypeID: "basic_message",
						Rule:   "INPUT",
					},
					Sends: []*fsm.Event{
						{
							TypeID:   "basic_message",
							Rule:     "INPUT",
							NoStatus: true,
						},
					},
					Target: "IDLE",
				},
			},
		},
	},
}
