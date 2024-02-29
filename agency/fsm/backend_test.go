package fsm

import (
	"testing"

	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
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
		{"mem save and format", args{&saveMemBackendMachine, "TEST"}, "she says, TEST", true},

		{"simple", args{&simpleBackendMachine, "TEST"}, "she says: TEST", true},
		{"simple_not", args{&simpleBackendMachine, "not"}, "not", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer assert.PushTester(t)()

			try.To(tt.args.m.Initialize())
			tt.args.m.InitLua()
			assert.Equal(tt.args.m.Type, MachineTypeBackend)
			beData := newBackend(tt.args.content, tt.args.content)
			transition := tt.args.m.TriggersByBackendData(beData)
			assert.NotNil(transition)
			o := transition.BuildSendEventsFromBackendData(beData)
			assert.SNotNil(o)
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
	type wants struct {
		want bool
		tgt  string
	}
	tests := []struct {
		name    string
		args    args
		wantStr string
		wants
	}{
		{"dynamic target", args{&luaNewTgtBackendMachine, "yes"}, "yes-yes", wants{true, "yes"}},

		{"simple", args{&luaBackendMachine, "TEST"}, "TEST-TEST", wants{true, ""}},
		{"simple_not", args{&luaBackendMachine, "not"}, "", wants{false, ""}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer assert.PushTester(t)()

			try.To(tt.args.m.Initialize())
			tt.args.m.InitLua()
			assert.Equal(tt.args.m.Type, MachineTypeBackend)
			if tt.want {
				beData := newBackend(tt.args.content, tt.args.content)
				transition := tt.args.m.TriggersByBackendData(beData)
				assert.NotNil(transition)
				o := transition.BuildSendEventsFromBackendData(beData)
				assert.SLen(o, 1)
				assert.NotNil(o[0].EventData.Backend)
				assert.Equal(o[0].EventData.Backend.Content, tt.wantStr)
				assert.SNotNil(o)
			} else {
				assert.Nil(tt.args.m.Triggers(
					protocolStatus(agency.Protocol_BASIC_MESSAGE, tt.args.content)))
			}
		})
	}
}

func newBackend(c, s string) *BackendData {
	return &BackendData{
		ConnID:   "TEST_CONN_ID_SET_IN_UNIT_TEST",
		Protocol: MessageBackend,
		Subject:  s,
		Content:  c,
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
	luaBackendScriptNewTgt = `
local i=getRegValue("MEM", "INPUT")
if i == "yes" then
	setRegValue("MEM", "TARGET", "YES")
	setRegValue("MEM", "OUTPUT", "OK")
else
	setRegValue("MEM", "TARGET", "TERMINATE")
	setRegValue("MEM", "OUTPUT", "NO")
end
 `

	luaBackendScript2 = `
local i=getRegValue("MEM", "INPUT")
local retval=i .. "-" .. i
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

	luaNewTgtBackendMachine = Machine{
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
							Data:     luaBackendScriptNewTgt,
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

	saveMemBackendMachine = Machine{
		Name: "save mem backend machine",
		Type: MachineTypeBackend,
		Initial: &Transition{
			Target: "IDLE",
		},
		States: map[string]*State{
			"IDLE": {
				Transitions: []*Transition{
					{
						Trigger: &Event{
							Data:     "LINE",
							Protocol: "backend",
							Rule:     "INPUT_SAVE",
						},
						Sends: []*Event{{
							Protocol: "backend",
							Rule:     "FORMAT_MEM",
							Data:     "she says, {{.LINE}}",
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
)
