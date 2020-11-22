package chat

import "github.com/findy-network/findy-grpc/agency/fsm"

var EmailIssuerMachine = fsm.Machine{
	Initial: "IDLE",
	States: map[string]fsm.State{
		"IDLE": {
			Transitions: []fsm.Transition{{
				Trigger: fsm.Event{
					TypeID: "basic_message",
				},
				Sends: []fsm.Event{{
					TypeID: "basic_message",
					Data: `Hello! I'm a email issuer.
Please enter your email address.`,
					NoStatus: true,
				}},
				Target: "WAITING_EMAIL_ADDRESS",
			}},
		},
		"WAITING_EMAIL_ADDRESS": {
			Transitions: []fsm.Transition{{
				Trigger: fsm.Event{
					TypeID: "basic_message",
					Rule:   "INPUT",
				},
				Sends: []fsm.Event{{
					TypeID: "basic_message",
					Rule:   "FORMAT",
					Data: `Thank you! I sent your pin code to %s.
Please enter it here and I'll send your email credential.`,
					NoStatus: true,
				}},
				Target: "WAITING_EMAIL_PIN",
			}},
		},
		"WAITING_EMAIL_PIN": {
			Transitions: []fsm.Transition{{
				Trigger: fsm.Event{
					TypeID: "basic_message",
					Rule:   "INPUT_VALIDATE", // validation criterion is will be in??
				},
				Sends: []fsm.Event{{
					TypeID: "basic_message",
					Rule:   "FORMAT",
					Data: `.
Please enter your email address.`,
				}},
				Target: "WAITING_STATUS",
			}},
		},
	},
}

var EchoMachine = fsm.Machine{
	Initial: "IDLE",
	States: map[string]fsm.State{
		"INITIAL": {
			Transitions: []fsm.Transition{{
				Trigger: fsm.Event{TypeID: "connection"},
				Sends: []fsm.Event{{
					TypeID: "basic_message",
					Rule:   "Hello! I'm echo bot",
				}},
				Target: "IDLE",
			}},
		},
		"IDLE": {
			Transitions: []fsm.Transition{
				{
					Trigger: fsm.Event{
						TypeID: "basic_message",
						Rule:   "INPUT",
					},
					Sends: []fsm.Event{{
						TypeID:   "basic_message",
						Rule:     "INPUT",
						NoStatus: true,
					}},
					Target: "IDLE",
				},
			},
		},
	},
}
