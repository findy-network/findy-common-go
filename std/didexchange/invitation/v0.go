package invitation

import (
	"encoding/base64"
	"encoding/json"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type V0 struct {
	// the Image URL of the connection invitation
	ImageURL string `json:"imageUrl,omitempty"`

	// the Service endpoint of the connection invitation
	ServiceEndpoint string `json:"serviceEndpoint,omitempty"`

	// the RecipientKeys for the connection invitation
	RecipientKeys []string `json:"recipientKeys,omitempty"`

	// the ID of the connection invitation
	ID string `json:"@id,omitempty"`

	// the Label of the connection invitation
	Label string `json:"label,omitempty"`

	// the RoutingKeys of the connection invitation
	RoutingKeys []string `json:"routingKeys,omitempty"`

	// the Type of the connection invitation
	Type string `json:"@type,omitempty"`
}

type invitationDIDExchangeV0 struct {
	V0
}

func CreateInvitationV0(info *AgentInfo) Invitation {
	return &invitationDIDExchangeV0{
		V0: V0{
			Type:            info.InvitationType,
			ID:              info.InvitationID,
			Label:           info.AgentLabel,
			ServiceEndpoint: info.EndpointURL,
			RecipientKeys:   []string{info.RecipientKey},
		},
	}
}

func (inv *invitationDIDExchangeV0) Build() (s string, err error) {
	defer err2.Returnf(&err, "build invitation V0")

	const prefix = "didcomm://aries_connection_invitation?c_i="
	b := try.To1(json.Marshal(inv))
	return prefix + base64.RawURLEncoding.EncodeToString(b), nil

}

func (inv *invitationDIDExchangeV0) Version() DIDExchangeVersion {
	return DIDExchangeVersionV0
}

func (inv *invitationDIDExchangeV0) Type() string {
	return inv.V0.Type
}

func (inv *invitationDIDExchangeV0) ID() string {
	return inv.V0.ID
}

func (inv *invitationDIDExchangeV0) Label() string {
	return inv.V0.Label
}

func (inv *invitationDIDExchangeV0) Services() []ServiceEndpoint {
	return []ServiceEndpoint{{
		ServiceEndpoint: inv.V0.ServiceEndpoint,
		RecipientKeys:   inv.V0.RecipientKeys,
		RoutingKeys:     inv.V0.RoutingKeys,
	}}
}

func (inv *invitationDIDExchangeV0) ImageURL() string {
	return inv.V0.ImageURL
}

func (inv *invitationDIDExchangeV0) Accept() []string {
	panic("not implemented")
}

func (inv *invitationDIDExchangeV0) HandshakeProtocols() []string {
	panic("not implemented")
}
