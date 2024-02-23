package fsm

import (
	"encoding/json"

	"github.com/Shopify/go-lua"
	agency "github.com/findy-network/findy-common-go/grpc/agency/v1"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

type NotificationType int32

type Event struct {
	// TODO: questions could be protocols here, then TypeID would not be needed?
	// we will continue with this when other protocol QAs will be implemented
	// New! Hook now uses TypeID for hook name/ID

	// These both are string versions to make writing the yaml fsm easier.
	// There parser methdod, Initialize() that must be call to make the machine
	// to work. It also make other syntax checks.
	// NOTE. Don't use this here at code level, use ProtocolType!
	Protocol string `json:"protocol"` // Note! See ProtocolType below
	TypeID   string `json:"type_id"`  // Note! See NotificationType below

	Rule string `json:"rule"`
	Data string `json:"data,omitempty"`
	// Deprecated: replaced by WantStatus, left to keep file format
	NoStatus bool `json:"no_status,omitempty"`
	// Tells that we want status updates about our sending, this is calculated
	// automatically
	WantStatus bool `json:"want_status,omitempty"`

	*EventData `json:"event_data,omitempty"`

	ProtocolType     agency.Protocol_Type `json:"-"`
	NotificationType NotificationType     `json:"-"`
	// NotificationType agency.Notification_Type `json:"-"`

	*agency.ProtocolStatus `json:"-"`
	*Transition            `json:"-"`
	*BackendData           `json:"-"`
}

func (e *Event) filterEnvs() {
	if e == nil {
		glog.V(7).Infoln("no event data, in filter:", e.Protocol)
		return
	}
	glog.V(10).Infoln("in filter:", e.Protocol)

	switch e.ProtocolType {
	case agency.Protocol_ISSUE_CREDENTIAL:
		if e.EventData != nil && e.EventData.Issuing != nil {
			e.EventData.Issuing.CredDefID = filterEnvs(e.EventData.Issuing.CredDefID)
		}
		e.Data = filterEnvs(e.Data)
	case agency.Protocol_PRESENT_PROOF:
		if e.EventData != nil && e.EventData.Proof != nil {
			e.EventData.Proof.ProofJSON = filterEnvs(e.EventData.Proof.ProofJSON)
		}
		e.Data = filterEnvs(e.Data)
	default:
		glog.V(7).Infoln("wrong type, in filter:", e.ProtocolType)
	}
}

func (e Event) TriggersByBackendData(data *BackendData) (ok bool, tgt string) {
	if data == nil {
		return true, ""
	}
	e.Machine.Memory[TriggerTypeUseInput] = data.Subject
	content := data.Content
	switch e.Rule {
	case TriggerTypeValidateInputNotEqual:
		return e.Machine.Memory[e.Data] != content, ""
	case TriggerTypeValidateInputEqual:
		return e.Machine.Memory[e.Data] == content, ""
	case TriggerTypeInputEqual:
		return content == e.Data, ""
	case TriggerTypeData, TriggerTypeUseInput, TriggerTypeUseInputSave:
		return true, ""
	case TriggerTypeLua:
		_, target, ok := e.ExecLua(content)
		return ok, target
	}
	return false, ""
}

func (e Event) TriggersByHook() bool {
	return true
}

func (e Event) Triggers(status *agency.ProtocolStatus) (ok bool, tgt string) {
	if status == nil {
		return true, ""
	}
	switch status.GetState().ProtocolID.TypeID {
	case agency.Protocol_ISSUE_CREDENTIAL, agency.Protocol_DIDEXCHANGE, agency.Protocol_PRESENT_PROOF:
		return true, ""
	case agency.Protocol_BASIC_MESSAGE:
		content := status.GetBasicMessage().Content
		switch e.Rule {
		case TriggerTypeValidateInputNotEqual:
			return e.Machine.Memory[e.Data] != content, ""
		case TriggerTypeValidateInputEqual:
			return e.Machine.Memory[e.Data] == content, ""
		case TriggerTypeInputEqual:
			return content == e.Data, ""
		case TriggerTypeData, TriggerTypeUseInput,
			TriggerTypeUseInputSave, TriggerTypeTransient:
			return true, ""
		case TriggerTypeLua:
			_, target, ok := e.ExecLua(content)
			return ok, target
		}
	}
	return false, ""
}

func (e Event) ExecLua(content string, a ...string) (out, tgt string, ok bool) {
	defer err2.Catch(err2.Err(func(err error) {
		ok = false
	}))

	okStr := LUA_OK
	if len(a) > 0 {
		okStr = a[0]
	}
	e.Machine.Memory[LUA_INPUT] = content
	luaScript := filterFilelink(e.Data)
	try.To(lua.DoString(e.Machine.luaState, luaScript))
	out, ok = e.Machine.Memory[LUA_OUTPUT]
	if !ok {
		glog.Warning("lua script: no output. Trying to get error")
		errMsg := assert.MKeyExists(e.Machine.Memory, LUA_ERROR)
		glog.Errorln("lua error:", errMsg)
	}
	tgt = e.Machine.Memory[LUA_TARGET]
	if okStr == LUA_ALL_OK {
		return out, tgt, true
	}
	ok = ok && out == okStr
	return out, tgt, ok
}

func (e Event) Answers(status *agency.Question) bool {
	switch status.TypeID {
	case agency.Question_PING_WAITS:
	case agency.Question_ISSUE_PROPOSE_WAITS:
	case agency.Question_PROOF_PROPOSE_WAITS:
	case agency.Question_PROOF_VERIFY_WAITS:
		assert.Equal(e.ProtocolType, agency.Protocol_PRESENT_PROOF)

		var attrValues []ProofAttr
		try.To(json.Unmarshal([]byte(e.Data), &attrValues))

		switch e.Rule {
		case TriggerTypeNotAcceptValues:
			if len(attrValues) != len(status.GetProofVerify().Attributes) {
				return true
			}
			for _, attr := range status.GetProofVerify().Attributes {
				for i, value := range attrValues {
					if value.Name == attr.Name && value.CredDefID == attr.CredDefID {
						attrValues[i].found = true
					}
				}
			}
			for _, value := range attrValues {
				if !value.found {
					return true
				}
			}
		case TriggerTypeAcceptAndInputValues:
			count := 0
			for _, attr := range status.GetProofVerify().Attributes {
				for _, value := range attrValues {
					if value.Name == attr.Name {
						e.Machine.Memory[value.Name] = attr.Value
						count++
					}
				}
			}
			return count == len(status.GetProofVerify().Attributes)
		}
	}
	return false
}
