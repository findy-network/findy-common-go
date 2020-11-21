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
				}},
				Target: "WAITING_MY_STATUS",
			}},
		},
		"WAITING_MY_STATUS": {
			Transitions: []fsm.Transition{{
				Trigger: fsm.Event{
					TypeID: "basic_message",
					Rule:   "OUR_STATUS",
				},
				Sends: []fsm.Event{{
					TypeID: "basic_message",
					Rule:   "OUR_STATUS",
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
				}},
				Target: "WAITING_MY_EMAIL_PIN_STATUS",
			}},
		},
		"WAITING_MY_EMAIL_PIN_STATUS": {
			Transitions: []fsm.Transition{{
				Trigger: fsm.Event{
					TypeID: "basic_message",
					Rule:   "OUR_STATUS",
				},
				Sends: []fsm.Event{{
					TypeID: "basic_message",
					Rule:   "OUR_STATUS",
				}},
				Target: "WAITING_EMAIL_PIN",
			}},
		},
		"WAITING_EMAIL_PIN": {
			Transitions: []fsm.Transition{{
				Trigger: fsm.Event{
					TypeID: "basic_message",
					Rule:   "INPUT",
				},
				Sends: []fsm.Event{{
					TypeID: "basic_message",
					Rule: `Hello! I'm a email issuer.
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
			Transitions: []fsm.Transition{{
				Trigger: fsm.Event{
					TypeID: "basic_message",
					Rule:   "INPUT",
				},
				Sends: []fsm.Event{{
					TypeID: "basic_message",
					Rule:   "INPUT", // todo: not here, or both?
				}},
				Target: "WAITING_STATUS",
			}},
		},
		"WAITING_STATUS": {
			Transitions: []fsm.Transition{{
				Trigger: fsm.Event{
					TypeID: "basic_message",
					Rule:   "OUR_STATUS",
				},
				Sends: []fsm.Event{{
					TypeID: "basic_message",
					Rule:   "OUR_STATUS",
				}},
				Target: "IDLE",
			}},
		},
	},
}
