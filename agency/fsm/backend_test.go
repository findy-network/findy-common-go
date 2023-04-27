package fsm

import (
	"testing"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
)

func TestBackendSimple(t *testing.T) {
	type args struct {
		m       *Machine
		content string
	}
	tests := []struct {
		name    string
		args    args
		wantStr string
		want    bool
	}{
		{"simple", args{&simpleBackendMachine, "TEST"}, "she says: TEST", true},
		{"simple_not", args{&simpleBackendMachine, "not"}, "not", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.PushTester(t)
			defer assert.PopTester()
			defer err2.Catch(func(err error) { assert.NoError(err) })

			err := tt.args.m.Initialize()
			assert.NoError(err)
			tt.args.m.InitLua()
			assert.Equal(tt.args.m.Type, MachineTypeBackend)
			beData := newBackend(tt.args.content)
			transition := tt.args.m.TriggersByBackendData(beData)
			assert.NotNil(transition)
			o := transition.BuildSendEventsFromBackendData(beData)
			assert.INotNil(o)
			assert.SLen(o, 1)
			assert.NotNil(o[0].EventData.Backend)
			if tt.want {
				assert.Equal(o[0].EventData.Backend.Content, tt.wantStr)
			} else {
				assert.NotEqual(o[0].EventData.Backend.Content, tt.wantStr)
			}
		})
	}
}

func TestBackendTestLuaTrigger(t *testing.T) {
	type args struct {
		m       *Machine
		content string
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

			err := tt.args.m.Initialize()
			assert.NoError(err)
			tt.args.m.InitLua()
			assert.Equal(tt.args.m.Type, MachineTypeBackend)
			if tt.want {
				beData := newBackend(tt.args.content)
				transition := tt.args.m.TriggersByBackendData(beData)
				assert.NotNil(transition)
				o := transition.BuildSendEventsFromBackendData(beData)
				assert.SLen(o, 1)
				assert.NotNil(o[0].EventData.Backend)
				assert.Equal(o[0].EventData.Backend.Content, tt.wantStr)
				assert.INotNil(o)
			} else {
				assert.Nil(tt.args.m.Triggers(
					protocolStatus(agency.Protocol_BASIC_MESSAGE, tt.args.content)))
			}
		})
	}
}

func newBackend(c string) *BackendData {
	return &BackendData{
		ToConnID:   "",
		Protocol:   MessageBackend,
		FromConnID: "123456", // uuid, this is for testing only
		Subject:    "test",
		Content:    c,
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
	simpleBackendMachine = Machine{
		Name: "simple backend machine",
		Type: MachineTypeBackend,
		Initial: &Transition{
			Target: "IDLE",
		},
		States: map[string]*State{
			"IDLE": {
				Transitions: []*Transition{
					{
						Trigger: &Event{
							Protocol: "backend",
							Rule:     "INPUT",
						},
						Sends: []*Event{{
							Protocol: "backend",
							Rule:     "FORMAT",
							Data:     "she says: %s",
						}},
						Target: "IDLE",
					},
				},
			},
			"TERMINATE": {
				Terminate: true,
			},
		},
	}

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
							Protocol: "backend",
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
