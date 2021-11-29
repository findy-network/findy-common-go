package invitation

import (
	"encoding/base64"
	"encoding/json"
	"net/url"

	"github.com/lainio/err2"
)

const prefix = "didcomm://aries_connection_invitation?c_i="

func Translate(s string) (i Invitation, err error) {
	defer err2.Annotate("invitation translate", &err)

	u, err := url.Parse(s)

	// this is not URL formated invitation, it must be JSON then
	if err != nil {
		err2.Check(json.Unmarshal([]byte(s), &i))
		return i, nil
	}

	m, err := url.ParseQuery(u.RawQuery)
	err2.Check(err)

	raw := m["c_i"][0]
	d := err2.Bytes.Try(base64.RawURLEncoding.DecodeString(raw))
	err2.Check(json.Unmarshal(d, &i))

	return i, nil
}

func Build(inv Invitation) (s string, err error) {
	defer err2.Annotate("invitation build", &err)

	b := err2.Bytes.Try(json.Marshal(inv))
	return prefix + base64.RawURLEncoding.EncodeToString(b), nil
}
