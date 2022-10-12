package invitation

import (
	"encoding/base64"
	"encoding/json"
	"net/url"
	"strings"

	"github.com/hyperledger/aries-framework-go/pkg/vdr/fingerprint"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
	"github.com/mr-tron/base58"
)

const prefix = "didcomm://aries_connection_invitation?c_i="

func decodeB64(str string) ([]byte, error) {
	data, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		data, err = base64.RawURLEncoding.DecodeString(str)
	}
	return data, err
}

func didKeysToB58(keys []string) []string {
	for index, key := range keys {
		assert.That(strings.HasPrefix(key, "did:key"))

		keyBytes := try.To1(fingerprint.PubKeyFromDIDKey(key))
		keys[index] = base58.Encode(keyBytes)
	}
	return keys
}

func convertFromOOB(invBytes []byte) (i Invitation, err error) {
	defer err2.Returnf(&err, "oob conversion")

	var oobInv OOBInvitation
	try.To(json.Unmarshal(invBytes, &oobInv))

	assert.D.True(len(oobInv.Services) > 0)

	service := oobInv.Services[0]

	i.ID = oobInv.ID
	i.Type = oobInv.Type
	i.Label = oobInv.Label
	i.ServiceEndpoint = service.ServiceEndpoint
	i.RecipientKeys = didKeysToB58(service.RecipientKeys)
	i.RoutingKeys = didKeysToB58(service.RoutingKeys)
	i.ImageURL = oobInv.ImageURL

	return i, nil
}

// TODO: finalize and cleanup OOB parsing
func Translate(s string) (i Invitation, err error) {
	defer err2.Returnf(&err, "invitation translate")

	u, err := url.Parse(strings.TrimSpace(s))

	// this is not URL formated invitation, it must be JSON then
	if err != nil {
		invBytes := []byte(s)
		if strings.Contains(s, "https://didcomm.org/out-of-band/1.0/invitation") {
			i = try.To1(convertFromOOB(invBytes))
		} else {
			try.To(json.Unmarshal(invBytes, &i))
		}
		return i, nil
	}

	m := try.To1(url.ParseQuery(u.RawQuery))

	if param, ok := m["c_i"]; ok {
		d := try.To1(decodeB64(param[0]))
		try.To(json.Unmarshal(d, &i))
		return i, nil
	}

	param := m["oob"]
	assert.D.True(param != nil)

	d := try.To1(decodeB64(param[0]))
	i = try.To1(convertFromOOB(d))

	return i, nil

}

func Build(inv Invitation) (s string, err error) {
	defer err2.Returnf(&err, "invitation build")

	b := try.To1(json.Marshal(inv))
	return prefix + base64.RawURLEncoding.EncodeToString(b), nil
}
