package fsm

import (
	"testing"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/lainio/err2/assert"
)

func TestLuaTestLuaTrigger(t *testing.T) {
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

func TestLuaTestLuaSend(t *testing.T) {
	type args struct {
		m *Machine
		c string
	}
	tests := []struct {
		name    string
		args    args
		wantStr string
		want    bool
	}{
		{"simple", args{&luaMachine, "TEST"}, "TEST+TEST", true},
		{"simple_not", args{&luaMachine, "not"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.PushTester(t)
			defer assert.PopTester()
			assert.NoError(tt.args.m.Initialize())
			tt.args.m.InitLua()
			if tt.want {
				status := protocolStatus(agency.Protocol_BASIC_MESSAGE, tt.args.c)
				transition := tt.args.m.Triggers(status)
				assert.NotNil(transition)
				o := transition.BuildSendEvents(status)
				assert.SLen(o, 1)
				assert.Equal(o[0].EventData.BasicMessage.Content, tt.wantStr)
				assert.INotNil(o)
			} else {
				assert.Nil(tt.args.m.Triggers(
					protocolStatus(agency.Protocol_BASIC_MESSAGE, tt.args.c)))
			}
		})
	}
}

var (
	luaScript1 = `
local i=getRegValue("MEM", "INPUT")
if i == "TEST" then
	setRegValue("MEM", "OUTPUT", "OK")
else
	setRegValue("MEM", "OUTPUT", "NO")
end
 `

	luaScript2 = `
local i=getRegValue("MEM", "INPUT")
local retval=i .. "+" .. i
setRegValue("MEM", "OUTPUT", retval)
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
							Data:     luaScript2,
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
