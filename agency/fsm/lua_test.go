package fsm

import (
	"testing"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

func TestLuaTestLuaTrigger(t *testing.T) {
	type args struct {
		m *Machine
		c string
	}
	type wants struct {
		want bool
		tgt  string
	}
	tests := []struct {
		name string
		args args
		wants
	}{
		{"simple_not", args{&luaMachineDynamicTargetState, "yes"}, wants{true, "YES"}},

		{"simple", args{&luaMachine, "TEST"}, wants{true, ""}},
		{"simple_not", args{&luaMachine, "not"}, wants{false, ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer assert.PushTester(t)()

			try.To(tt.args.m.Initialize())
			tt.args.m.InitLua()
			if tt.want {
				transition := tt.args.m.Triggers(
					protocolStatus(agency.Protocol_BASIC_MESSAGE, tt.args.c))
				assert.NotNil(transition)
				if tt.tgt != "" {
					assert.Equal(transition.Target, tt.tgt)
				}
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
			defer assert.PushTester(t)()

			assert.NoError(tt.args.m.Initialize())
			tt.args.m.InitLua()
			if tt.want {
				status := protocolStatus(agency.Protocol_BASIC_MESSAGE, tt.args.c)
				transition := tt.args.m.Triggers(status)
				assert.NotNil(transition)
				o := transition.BuildSendEvents(status)
				assert.SLen(o, 1)
				assert.Equal(o[0].EventData.BasicMessage.Content, tt.wantStr)
				assert.SNotNil(o)
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
if i == "yes" then
	setRegValue("MEM", "TARGET", "YES")
	setRegValue("MEM", "OUTPUT", "OK")
else
	setRegValue("MEM", "TARGET", "TERMINATE")
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
		Type: MachineTypeConversation,
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
							Data:     `${script1.lua}`,
						},
						Sends: []*Event{{
							Protocol: "basic_message",
							Rule:     "LUA",
							Data:     `${script2.lua}`,
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

	luaMachineDynamicTargetState = Machine{
		Name: "lua_machine",
		Type: MachineTypeConversation,
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
							Data:     `@{script2.lua}`,
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
