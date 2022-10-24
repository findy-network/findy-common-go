package invitation_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/findy-network/findy-common-go/std/didexchange/invitation"
	"github.com/lainio/err2/assert"
)

var (
	invJSON = `{
  "serviceEndpoint": "http://agency.example.com/a2a/VxM84ioz2ct15vN4XhT9ek/VxM84ioz2ct15vN4XhT9ek/UqFQQbuNbJDE3RNEXPW3y5",
  "recipientKeys": [
    "GnJU64Lf3BwfAD2mTnEWVG1H32FcRMhZzm24dC4FbWt3"
  ],
  "@id": "e8f9dfc9-bcf3-4f47-a7f9-5b898efbf885",
  "label": "test_client_agent1",
  "@type": "did:sov:BzCbsNYhMrjHiqZDTUASHg;spec/connections/1.0/invitation"
}
`

	invOOBJSON = `{
	"@type": "https://didcomm.org/out-of-band/1.0/invitation",
	"@id": "4c3d1a7e-ac1f-41ec-a8a0-6c0bf1b952a1",
	"services": [{
		"id": "#inline",
		"type": "did-communication",
		"recipientKeys": ["did:key:z6MknpNFYTqkhT29tc84WPqhBJkFNJnvm9NdepEUWFvfa8hN"],
		"routingKeys": ["did:key:z6MknpNFYTqkhT29tc84WPqhBJkFNJnvm9NdepEUWFvfa8hN"],
		"serviceEndpoint": "https://agent.example.com"
	}],
	"label": "faber.agent",
	"handshake_protocols": ["https://didcomm.org/didexchange/1.0"]
}
`

	invInvalid = `{
	"@type": "https://didcomm.org/invalid/1.0/invitation",
	"@id": "huu-haa",
	"services": [],
	"label": "invalid",
	"handshake_protocols": []
}
`
)

func TestCreate(t *testing.T) {
	tests := []struct {
		name    string
		version invitation.DIDExchangeVersion
		params  invitation.AgentInfo
		fail    bool
	}{
		{
			"V0",
			invitation.DIDExchangeVersionV0,
			invitation.AgentInfo{
				"did:sov:BzCbsNYhMrjHiqZDTUASHg;spec/connections/1.0/invitation",
				"e8f9dfc9-bcf3-4f47-a7f9-5b898efbf885",
				"http://agency.example.com/a2a/VxM84ioz2ct15vN4XhT9ek/VxM84ioz2ct15vN4XhT9ek/UqFQQbuNbJDE3RNEXPW3y5",
				"GnJU64Lf3BwfAD2mTnEWVG1H32FcRMhZzm24dC4FbWt3",
				"test_client_agent1",
			},
			false,
		},
		{"V1",
			invitation.DIDExchangeVersionV1,
			invitation.AgentInfo{
				"https://didcomm.org/out-of-band/1.0/invitation",
				"4c3d1a7e-ac1f-41ec-a8a0-6c0bf1b952a1",
				"https://agent.example.com",
				"did:key:z6MknpNFYTqkhT29tc84WPqhBJkFNJnvm9NdepEUWFvfa8hN",
				"faber.agent",
			},
			false,
		},
		{"invalid", -1, invitation.AgentInfo{}, true},
	}
	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assert.PushTester(t)
			defer assert.PopTester()

			inv, err := invitation.Create(tc.version, tc.params)
			if tc.fail {
				assert.Error(err)
				return
			}
			assert.NoError(err)

			// build URL
			urlStr, err := invitation.Build(inv)
			assert.NoError(err)
			assert.NotEmpty(urlStr)

			invFromURL, err := invitation.Translate(urlStr)
			assert.NoError(err)
			assert.INotNil(invFromURL)
			assert.Equal(invFromURL.ID(), inv.ID())

			// convert to JSON
			jsonBytes, err := json.Marshal(inv)
			assert.NoError(err)
			jsonStr := string(jsonBytes)
			assert.NotEmpty(jsonStr)

			invFromJSON, err := invitation.Translate(jsonStr)
			assert.NoError(err)
			assert.INotNil(invFromJSON)
			assert.Equal(invFromJSON.ID(), inv.ID())
		})
	}
}

func TestBuild(t *testing.T) {
	tests := []struct {
		name string
		json string
		fail bool
	}{
		{"V0", invJSON, false},
		{"V1", invOOBJSON, false},
		{"invalid", invInvalid, true},
	}
	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assert.PushTester(t)
			defer assert.PopTester()

			inv, err := invitation.Translate(tc.json)

			if tc.fail {
				assert.Error(err)
				return
			}
			assert.NoError(err)
			assert.INotNil(inv)

			s, err := invitation.Build(inv)
			assert.NoError(err)
			assert.NotEmpty(s)

			inv, err = invitation.Translate(s)
			assert.NoError(err)
			assert.INotNil(inv)
		})
	}
}

func TestTranslate(t *testing.T) {
	tests := []struct {
		name       string
		readConnID string
		input      string
		version    invitation.DIDExchangeVersion
	}{
		{
			"plain json",
			"e8f9dfc9-bcf3-4f47-a7f9-5b898efbf885",
			invJSON,
			invitation.DIDExchangeVersionV0,
		},
		{
			"plain oob json",
			"4c3d1a7e-ac1f-41ec-a8a0-6c0bf1b952a1",
			invOOBJSON,
			invitation.DIDExchangeVersionV1,
		},
		{
			"url from bc cov",
			"36a884f7-a105-4494-a395-c6cc7f578e28",
			"https://email-verification-agent.vonx.io?c_i=eyJAdHlwZSI6ICJkaWQ6c292OkJ6Q2JzTlloTXJqSGlxWkRUVUFTSGc7c3BlYy9jb25uZWN0aW9ucy8xLjAvaW52aXRhdGlvbiIsICJAaWQiOiAiMzZhODg0ZjctYTEwNS00NDk0LWEzOTUtYzZjYzdmNTc4ZTI4IiwgInNlcnZpY2VFbmRwb2ludCI6ICJodHRwczovL2VtYWlsLXZlcmlmaWNhdGlvbi1hZ2VudC52b254LmlvIiwgImxhYmVsIjogIkVtYWlsIFZlcmlmaWNhdGlvbiBTZXJ2aWNlIiwgInJlY2lwaWVudEtleXMiOiBbIkdvQ01MeHZjcTlBbzZqaEhlUkJZN0Vuam4zc042WThCd2hjMTJCNWs4NEJhIl19",
			invitation.DIDExchangeVersionV0,
		},
		{
			"url from aca-py",
			"b6a5d0c9-f43b-4c41-8757-1e69e46f29e6",
			"https://a29e-91-153-20-18.ngrok.io?c_i=eyJAdHlwZSI6ICJkaWQ6c292OkJ6Q2JzTlloTXJqSGlxWkRUVUFTSGc7c3BlYy9jb25uZWN0aW9ucy8xLjAvaW52aXRhdGlvbiIsICJAaWQiOiAiYjZhNWQwYzktZjQzYi00YzQxLTg3NTctMWU2OWU0NmYyOWU2IiwgInJlY2lwaWVudEtleXMiOiBbIjg0akRvUGlxbkxpemdKNUdSeE0xZHh5QUZlS1JXbUp4TFhkWkhxeUpUU3VRIl0sICJzZXJ2aWNlRW5kcG9pbnQiOiAiaHR0cHM6Ly9hMjllLTkxLTE1My0yMC0xOC5uZ3Jvay5pbyIsICJsYWJlbCI6ICJhY2EtcHkuRmFiZXIifQ==",
			invitation.DIDExchangeVersionV0,
		},
		{
			"url from lissi",
			"a1eb3595-00fb-4e58-89b4-89b1fe164441",
			"didcomm://aries_connection_invitation?c_i=eyJAdHlwZSI6ImRpZDpzb3Y6QnpDYnNOWWhNcmpIaXFaRFRVQVNIZztzcGVjL2Nvbm5lY3Rpb25zLzEuMC9pbnZpdGF0aW9uIiwiQGlkIjoiYTFlYjM1OTUtMDBmYi00ZTU4LTg5YjQtODliMWZlMTY0NDQxIiwibGFiZWwiOiJTdGFkdCBNdXN0ZXJoYXVzZW4iLCJzZXJ2aWNlRW5kcG9pbnQiOiJodHRwczovL2RlbW8tYWdlbnQuaW5zdGl0dXRpb25hbC1hZ2VudC5saXNzaS5pZC9kaWRjb21tLyIsImltYWdlVXJsIjoiaHR0cHM6Ly9yb3V0aW5nLmxpc3NpLmlvL2FwaS9JbWFnZS9kZW1vTXVzdGVyaGF1c2VuIiwicmVjaXBpZW50S2V5cyI6WyJCNm5hQ1I5TlRENHdKbmZrNXRxZVdhOWdYNWNIWVpEeWgyaUxOcHF3cldBbyJdfQ",
			invitation.DIDExchangeVersionV0,
		},
		{
			"url from us",
			"d6903f21-1c92-45d2-82ad-34e894bfb01b",
			"didcomm://aries_connection_invitation?c_i=eyJzZXJ2aWNlRW5kcG9pbnQiOiJodHRwOi8vbG9jYWxob3N0OjgwODAvYTJhLzNKY3NUYW9tR2NQdlRBMmlpdDdSZ2YvM0pjc1Rhb21HY1B2VEEyaWl0N1JnZi9MVkRZa1ZMZXpyREhUa0VGSG1vQ216L2Q2OTAzZjIxLTFjOTItNDVkMi04MmFkLTM0ZTg5NGJmYjAxYiIsInJlY2lwaWVudEtleXMiOlsiQmQxTkZnSDlQeW5MakJzSmlqNFFIc2U5WXlxbjJTcFB1RHNSaVBVZFdveXUiXSwiQGlkIjoiZDY5MDNmMjEtMWM5Mi00NWQyLTgyYWQtMzRlODk0YmZiMDFiIiwibGFiZWwiOiJlbXB0eS1sYWJlbCIsIkB0eXBlIjoiZGlkOnNvdjpCekNic05ZaE1yakhpcVpEVFVBU0hnO3NwZWMvY29ubmVjdGlvbnMvMS4wL2ludml0YXRpb24ifQ", invitation.DIDExchangeVersionV0,
		},
		{
			"url with prefix space",
			"d6903f21-1c92-45d2-82ad-34e894bfb01b",
			" didcomm://aries_connection_invitation?c_i=eyJzZXJ2aWNlRW5kcG9pbnQiOiJodHRwOi8vbG9jYWxob3N0OjgwODAvYTJhLzNKY3NUYW9tR2NQdlRBMmlpdDdSZ2YvM0pjc1Rhb21HY1B2VEEyaWl0N1JnZi9MVkRZa1ZMZXpyREhUa0VGSG1vQ216L2Q2OTAzZjIxLTFjOTItNDVkMi04MmFkLTM0ZTg5NGJmYjAxYiIsInJlY2lwaWVudEtleXMiOlsiQmQxTkZnSDlQeW5MakJzSmlqNFFIc2U5WXlxbjJTcFB1RHNSaVBVZFdveXUiXSwiQGlkIjoiZDY5MDNmMjEtMWM5Mi00NWQyLTgyYWQtMzRlODk0YmZiMDFiIiwibGFiZWwiOiJlbXB0eS1sYWJlbCIsIkB0eXBlIjoiZGlkOnNvdjpCekNic05ZaE1yakhpcVpEVFVBU0hnO3NwZWMvY29ubmVjdGlvbnMvMS4wL2ludml0YXRpb24ifQ",
			invitation.DIDExchangeVersionV0,
		},
		{
			"url with suffix space",
			"d6903f21-1c92-45d2-82ad-34e894bfb01b",
			"didcomm://aries_connection_invitation?c_i=eyJzZXJ2aWNlRW5kcG9pbnQiOiJodHRwOi8vbG9jYWxob3N0OjgwODAvYTJhLzNKY3NUYW9tR2NQdlRBMmlpdDdSZ2YvM0pjc1Rhb21HY1B2VEEyaWl0N1JnZi9MVkRZa1ZMZXpyREhUa0VGSG1vQ216L2Q2OTAzZjIxLTFjOTItNDVkMi04MmFkLTM0ZTg5NGJmYjAxYiIsInJlY2lwaWVudEtleXMiOlsiQmQxTkZnSDlQeW5MakJzSmlqNFFIc2U5WXlxbjJTcFB1RHNSaVBVZFdveXUiXSwiQGlkIjoiZDY5MDNmMjEtMWM5Mi00NWQyLTgyYWQtMzRlODk0YmZiMDFiIiwibGFiZWwiOiJlbXB0eS1sYWJlbCIsIkB0eXBlIjoiZGlkOnNvdjpCekNic05ZaE1yakhpcVpEVFVBU0hnO3NwZWMvY29ubmVjdGlvbnMvMS4wL2ludml0YXRpb24ifQ\n",
			invitation.DIDExchangeVersionV0,
		},
		{
			"OOB url from Animo",
			"f4155f35-090b-4bdd-80e5-f2415309fa30",
			"https://didcomm.demo.animo.id?oob=eyJAdHlwZSI6Imh0dHBzOi8vZGlkY29tbS5vcmcvb3V0LW9mLWJhbmQvMS4xL2ludml0YXRpb24iLCJAaWQiOiJmNDE1NWYzNS0wOTBiLTRiZGQtODBlNS1mMjQxNTMwOWZhMzAiLCJsYWJlbCI6IkFuaW1vIiwiYWNjZXB0IjpbImRpZGNvbW0vYWlwMSIsImRpZGNvbW0vYWlwMjtlbnY9cmZjMTkiXSwiaGFuZHNoYWtlX3Byb3RvY29scyI6WyJodHRwczovL2RpZGNvbW0ub3JnL2RpZGV4Y2hhbmdlLzEuMCIsImh0dHBzOi8vZGlkY29tbS5vcmcvY29ubmVjdGlvbnMvMS4wIl0sInNlcnZpY2VzIjpbeyJpZCI6IiNpbmxpbmUtMCIsInNlcnZpY2VFbmRwb2ludCI6Imh0dHBzOi8vZGlkY29tbS5kZW1vLmFuaW1vLmlkIiwidHlwZSI6ImRpZC1jb21tdW5pY2F0aW9uIiwicmVjaXBpZW50S2V5cyI6WyJkaWQ6a2V5Ono2TWtmZzNoWDhOUVBVTkZOeTh6ZVN2QlE4Z1A3VjVKREd1aTN3YWdvb3o3Z2tkYSJdLCJyb3V0aW5nS2V5cyI6W119XSwiaW1hZ2VVcmwiOiJodHRwczovL2kuaW1ndXIuY29tL2czYWJjQ08ucG5nIn0",
			invitation.DIDExchangeVersionV1,
		},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assert.PushTester(t)
			defer assert.PopTester()

			inv, err := invitation.Translate(tc.input)
			if err != nil {
				t.Errorf("%s = err (%v)", tc.name, err)
			}
			assert.NotEmpty(inv.Label())
			assert.NotEmpty(inv.Type())
			assert.Equal(tc.readConnID, inv.ID())
			assert.Equal(tc.version, inv.Version())
			assert.SNotEmpty(inv.Services())
			assert.SNotEmpty(inv.Services()[0].RecipientKeys)

			if len(inv.Services()[0].RecipientKeys) > 0 {
				assert.ThatNot(strings.HasPrefix(inv.Services()[0].RecipientKeysAsB58()[0], invitation.DIDKeyPrefix))
			}
			if len(inv.Services()[0].RoutingKeys) > 0 {
				assert.ThatNot(strings.HasPrefix(inv.Services()[0].RoutingKeysAsB58()[0], invitation.DIDKeyPrefix))
			}
		})
	}

}
