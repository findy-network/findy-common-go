package invitation

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
)

type invitationHeader struct {
	// the Type of the connection invitation
	Type string `json:"@type,omitempty"`
}

func decodeB64(str string) ([]byte, error) {
	data, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		data, err = base64.RawURLEncoding.DecodeString(str)
	}
	return data, err
}

func parseInvitationJSON(jsonBytes []byte) (i Invitation, err error) {
	defer err2.Return(&err)

	var header invitationHeader
	err = json.Unmarshal(jsonBytes, &header)

	switch {
	case strings.Contains(header.Type, "connections"):
		{
			var invV0 invitationDIDExchangeV0
			try.To(json.Unmarshal(jsonBytes, &invV0))
			i = &invV0
		}
	case strings.Contains(header.Type, "out-of-band"): // TODO: check versions
		{
			var invV1 invitationDIDExchangeV1
			try.To(json.Unmarshal(jsonBytes, &invV1))
			i = &invV1
		}
	default:
		return nil, errors.New("unknown invitation type")
	}

	return i, nil
}

func Translate(s string) (i Invitation, err error) {
	defer err2.Returnf(&err, "invitation translate")

	u, err := url.Parse(strings.TrimSpace(s))

	// this is not URL formated invitation, it must be JSON then
	if err != nil {
		return parseInvitationJSON([]byte(s))
	}

	m := try.To1(url.ParseQuery(u.RawQuery))

	var (
		invB64Str []string
		ok        bool
	)
	if invB64Str, ok = m["c_i"]; !ok {
		invB64Str = m["oob"]
	}
	assert.SNotEmpty(invB64Str, "invalid invitation url format")

	bytes := try.To1(decodeB64(invB64Str[0]))
	return parseInvitationJSON(bytes)
}

func Build(inv Invitation) (s string, err error) {
	return inv.Build()
}

func Create(version DIDExchangeVersion, info AgentInfo) (inv Invitation, err error) {
	switch version {
	case DIDExchangeVersionV0:
		inv = CreateInvitationV0(&info)
	case DIDExchangeVersionV1:
		inv = CreateInvitationV1(&info)
	default:
		return nil, fmt.Errorf("unknown DIDExhange version %d", version)
	}
	return
}
