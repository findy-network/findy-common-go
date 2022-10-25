package invitation

import (
	"encoding/gob"
	"strings"

	"github.com/hyperledger/aries-framework-go/pkg/vdr/fingerprint"
	"github.com/lainio/err2/try"
	"github.com/mr-tron/base58"
)

type DIDExchangeVersion int

const (
	DIDExchangeVersionV0 DIDExchangeVersion = 0
	DIDExchangeVersionV1 DIDExchangeVersion = 10
	DIDExchangeVersionV2 DIDExchangeVersion = 20
)

type AgentInfo struct {
	InvitationType string
	InvitationID   string
	EndpointURL    string
	RecipientKey   string
	AgentLabel     string
}

type ServiceEndpoint struct {
	ID              string
	ServiceEndpoint string
	Type            string
	RecipientKeys   []string
	RoutingKeys     []string
}

type Invitation interface {
	Build() (string, error)
	Version() DIDExchangeVersion
	Type() string
	ID() string
	Label() string
	Services() []ServiceEndpoint
	ImageURL() string
	Accept() []string
	HandshakeProtocols() []string
}

const DIDKeyPrefix = "did:key"

func init() {
	gob.Register(&invitationDIDExchangeV0{})
	gob.Register(&invitationDIDExchangeV1{})
}

func didKeysToB58(keys []string) []string {
	for index, key := range keys {
		if strings.HasPrefix(key, DIDKeyPrefix) {
			keyBytes := try.To1(fingerprint.PubKeyFromDIDKey(key))
			keys[index] = base58.Encode(keyBytes)
		} else {
			keys[index] = key
		}
	}
	return keys
}

func (s ServiceEndpoint) RecipientKeysAsB58() []string {
	return didKeysToB58(s.RecipientKeys)
}

func (s ServiceEndpoint) RoutingKeysAsB58() []string {
	return didKeysToB58(s.RoutingKeys)
}
