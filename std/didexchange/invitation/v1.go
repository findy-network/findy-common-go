package invitation

import (
	"encoding/base64"
	"encoding/json"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type V1Service struct {
	ID              string   `json:"id,omitempty"`
	ServiceEndpoint string   `json:"serviceEndpoint,omitempty"`
	Type            string   `json:"type,omitempty"`
	RecipientKeys   []string `json:"recipientKeys,omitempty"`
	RoutingKeys     []string `json:"routingKeys,omitempty"`
}

type V1 struct {
	// the Type of the connection invitation
	Type string `json:"@type,omitempty"`

	// the ID of the connection invitation
	ID string `json:"@id,omitempty"`

	// the Label of the connection invitation
	Label string `json:"label,omitempty"`

	Accept []string `json:"accept,omitempty"`

	HandshakeProtocols []string `json:"handshake_protocols,omitempty"`

	Services []V1Service `json:"services,omitempty"`

	// the Image URL of the connection invitation
	ImageURL string `json:"imageUrl,omitempty"`
}

type invitationDIDExchangeV1 struct {
	V1
}

func CreateInvitationV1(info *AgentInfo) Invitation {
	return &invitationDIDExchangeV1{
		V1: V1{
			Type:  info.InvitationType,
			ID:    info.InvitationID,
			Label: info.AgentLabel,
			Services: []V1Service{{
				ServiceEndpoint: info.EndpointURL,
				RecipientKeys:   []string{info.RecipientKey},
			}},
		},
	}
}

func (inv *invitationDIDExchangeV1) Build() (s string, err error) {
	defer err2.Returnf(&err, "build invitation V1")

	const prefix = "didcomm://aries_connection_invitation?oob="
	b := try.To1(json.Marshal(inv))
	return prefix + base64.RawURLEncoding.EncodeToString(b), nil
}

func (inv *invitationDIDExchangeV1) Version() DIDExchangeVersion {
	return DIDExchangeVersionV1
}

func (inv *invitationDIDExchangeV1) Type() string {
	return inv.V1.Type
}

func (inv *invitationDIDExchangeV1) ID() string {
	return inv.V1.ID
}

func (inv *invitationDIDExchangeV1) Label() string {
	return inv.V1.Label
}

func (inv *invitationDIDExchangeV1) Services() []ServiceEndpoint {
	endpoints := make([]ServiceEndpoint, 0)
	for _, ep := range inv.V1.Services {
		endpoints = append(endpoints, ServiceEndpoint{
			ServiceEndpoint: ep.ServiceEndpoint,
			RecipientKeys:   ep.RecipientKeys,
			RoutingKeys:     ep.RoutingKeys,
			Type:            ep.Type,
			ID:              ep.ID,
		})
	}
	return endpoints
}
func (inv *invitationDIDExchangeV1) ImageURL() string {
	return inv.V1.ImageURL
}

func (inv *invitationDIDExchangeV1) Accept() []string {
	return inv.V1.Accept
}

func (inv *invitationDIDExchangeV1) HandshakeProtocols() []string {
	return inv.V1.HandshakeProtocols
}
