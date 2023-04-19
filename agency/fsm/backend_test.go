package fsm

import (
	"testing"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
)

func TestBackendTestLuaTrigger(t *testing.T) {
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
		{"simple", args{&luaBackendMachine, "TEST"}, "TEST+TEST", true},
		{"simple_not", args{&luaBackendMachine, "not"}, "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.PushTester(t)
			defer assert.PopTester()
			defer err2.Catch(func(err error) { assert.NoError(err) })

			assert.NoError(tt.args.m.Initialize())
			tt.args.m.InitLua()
			assert.Equal(tt.args.m.Type, MachineTypeBackend)
			if tt.want {
				status := protocolStatus(agency.Protocol_BASIC_MESSAGE, tt.args.c)
				transition := tt.args.m.Triggers(status)
				assert.NotNil(transition)
				o := transition.BuildSendEvents(status)
				assert.SLen(o, 1)
				assert.NotNil(o[0].EventData.Backend)
				assert.Equal(o[0].EventData.Backend.Content, tt.wantStr)
				assert.INotNil(o)

			} else {
				assert.Nil(tt.args.m.Triggers(
					protocolStatus(agency.Protocol_BASIC_MESSAGE, tt.args.c)))
			}
		})
	}
}

var (
	luaBackendScript1 = `
local i=getRegValue("MEM", "INPUT")
if i == "TEST" then
	setRegValue("MEM", "OUTPUT", "OK")
else
	setRegValue("MEM", "OUTPUT", "NO")
end
 `

	luaBackendScript2 = `
local i=getRegValue("MEM", "INPUT")
local retval=i .. "+" .. i
setRegValue("MEM", "OUTPUT", retval)
 `

	luaBackendMachine = Machine{
		Name: "luaBackend_machine",
		Type: MachineTypeBackend,
		Initial: &Transition{
			Target: "IDLE",
		},
		States: map[string]*State{
			"IDLE": {
				Transitions: []*Transition{
					{
						Trigger: &Event{
							Protocol: "basic_message",
							Rule:     "LUA",
							Data:     luaBackendScript1,
						},
						Sends: []*Event{{
							Protocol: "backend",
							Rule:     "LUA",
							Data:     luaBackendScript2,
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
