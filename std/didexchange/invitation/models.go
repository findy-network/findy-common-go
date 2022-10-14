package invitation

import (
	"encoding/base64"
	"encoding/json"

	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

type DIDExchangeVersion int

const (
	DIDExchangeVersionV0 DIDExchangeVersion = iota
	DIDExchangeVersionV1
	DIDExchangeVersionV2
)

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
	ServiceEndpoint() []ServiceEndpoint
	ImageURL() string
	Accept() []string
	HandshakeProtocols() []string
}

type invitationDIDExchangeV0 struct {
	// the Image URL of the connection invitation
	InvImageURL string `json:"imageUrl,omitempty"`

	// the Service endpoint of the connection invitation
	InvServiceEndpoint string `json:"serviceEndpoint,omitempty"`

	// the RecipientKeys for the connection invitation
	InvRecipientKeys []string `json:"recipientKeys,omitempty"`

	// the ID of the connection invitation
	InvID string `json:"@id,omitempty"`

	// the Label of the connection invitation
	InvLabel string `json:"label,omitempty"`

	// the RoutingKeys of the connection invitation
	InvRoutingKeys []string `json:"routingKeys,omitempty"`

	// the Type of the connection invitation
	InvType string `json:"@type,omitempty"`
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
	return inv.InvType
}

func (inv *invitationDIDExchangeV0) ID() string {
	return inv.InvID
}

func (inv *invitationDIDExchangeV0) Label() string {
	return inv.InvLabel
}

func (inv *invitationDIDExchangeV0) ServiceEndpoint() []ServiceEndpoint {
	return []ServiceEndpoint{{
		ServiceEndpoint: inv.InvServiceEndpoint,
		RecipientKeys:   inv.InvRecipientKeys,
		RoutingKeys:     inv.InvRoutingKeys,
	}}
}

func (inv *invitationDIDExchangeV0) ImageURL() string {
	return inv.InvImageURL
}

func (inv *invitationDIDExchangeV0) Accept() []string {
	panic("not implemented")
}

func (inv *invitationDIDExchangeV0) HandshakeProtocols() []string {
	panic("not implemented")
}

type invitationDIDExchangeV1 struct {
	// the Type of the connection invitation
	InvType string `json:"@type,omitempty"`

	// the ID of the connection invitation
	InvID string `json:"@id,omitempty"`

	// the Label of the connection invitation
	InvLabel string `json:"label,omitempty"`

	InvAccept []string `json:"accept,omitempty"`

	InvHandshakeProtocols []string `json:"handshake_protocols,omitempty"`

	InvServices []struct {
		ID              string   `json:"id,omitempty"`
		ServiceEndpoint string   `json:"serviceEndpoint,omitempty"`
		Type            string   `json:"type,omitempty"`
		RecipientKeys   []string `json:"recipientKeys,omitempty"`
		RoutingKeys     []string `json:"routingKeys,omitempty"`
	} `json:"services,omitempty"`

	// the Image URL of the connection invitation
	InvImageURL string `json:"imageUrl,omitempty"`
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
	return inv.InvType
}

func (inv *invitationDIDExchangeV1) ID() string {
	return inv.InvID
}

func (inv *invitationDIDExchangeV1) Label() string {
	return inv.InvLabel
}

func (inv *invitationDIDExchangeV1) ServiceEndpoint() []ServiceEndpoint {
	endpoints := make([]ServiceEndpoint, 0)
	for _, ep := range inv.InvServices {
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
	return inv.InvImageURL
}

func (inv *invitationDIDExchangeV1) Accept() []string {
	return inv.InvAccept
}

func (inv *invitationDIDExchangeV1) HandshakeProtocols() []string {
	return inv.InvHandshakeProtocols
}
