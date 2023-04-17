package fsm

import (
	"testing"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/lainio/err2/assert"
)

func TestLuaTestLua(t *testing.T) {
	type args struct {
		m *Machine
		c string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"simple", args{&luaMachine, "TEST"}, true},
		{"simple_not", args{&luaMachine, "not"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.PushTester(t)
			defer assert.PopTester()
			assert.NoError(tt.args.m.Initialize())
			tt.args.m.InitLua()
			if tt.want {
				assert.NotNil(tt.args.m.Triggers(
					protocolStatus(agency.Protocol_BASIC_MESSAGE, tt.args.c)))
			} else {
				assert.Nil(tt.args.m.Triggers(
					protocolStatus(agency.Protocol_BASIC_MESSAGE, tt.args.c)))
			}
		})
	}
}

var (
	luaScript1 = `
local i=getValue("INPUT")
if i == "TEST" then
	setValue("OUTPUT", "OK")
else
	setValue("OUTPUT", "NO")
end
 `

	luaMachine = Machine{
		Name: "lua_machine",
		Initial: &Transition{
			Sends: []*Event{{
				Protocol: "basic_message",
				Data:     "Hello!",
			}},
			Target: "IDLE",
		},
		States: map[string]*State{
			"IDLE": {
				Transitions: []*Transition{
					{
						Trigger: &Event{
							Protocol: "basic_message",
							Rule:     "LUA",
							Data:     luaScript1,
						},
						Sends: []*Event{{
							Protocol: "basic_message",
							Rule:     "LUA",
							Data:     luaScript1,
						}},
						Target: "TERMINATE",
					},
				},
			},
			"TERMINATE": {
				Terminate: true,
			},
		},
	}
)
