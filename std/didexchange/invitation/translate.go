package invitation

import (
	"encoding/base64"
	"encoding/json"
	"net/url"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

const prefix = "didcomm://aries_connection_invitation?c_i="

func decodeB64(str string) ([]byte, error) {
	data, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		data, err = base64.RawURLEncoding.DecodeString(str)
	}
	return data, err
}

func Translate(s string) (i Invitation, err error) {
	defer err2.Annotate("invitation translate", &err)

	u, err := url.Parse(s)

	// this is not URL formated invitation, it must be JSON then
	if err != nil {
		try.To(json.Unmarshal([]byte(s), &i))
		return i, nil
	}

	m := try.To1(url.ParseQuery(u.RawQuery))

	raw := m["c_i"][0]
	d := err2.Bytes.Try(decodeB64(raw))
	try.To(json.Unmarshal(d, &i))

	return i, nil
}

func Build(inv Invitation) (s string, err error) {
	defer err2.Annotate("invitation build", &err)

	b := err2.Bytes.Try(json.Marshal(inv))
	return prefix + base64.RawURLEncoding.EncodeToString(b), nil
}
