package invitation

import (
	"testing"

	"github.com/lainio/err2/assert"
)

var (
	invJSON = `{
  "serviceEndpoint": "http://findy-agent.op-ai.fi/a2a/VxM84ioz2ct15vN4XhT9ek/VxM84ioz2ct15vN4XhT9ek/UqFQQbuNbJDE3RNEXPW3y5",
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
		"serviceEndpoint": "https://a0cd-91-153-22-154.ngrok.io"
	}],
	"label": "faber.agent",
	"handshake_protocols": ["https://didcomm.org/didexchange/1.0"]
}
`
)

func TestBuild(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()

	inv, err := Translate(invJSON)
	assert.NoError(err)
	assert.INotNil(inv)

	s, err := Build(inv)
	assert.NoError(err)
	assert.NotEmpty(s)

	inv, err = Translate(s)
	assert.NoError(err)
	assert.INotNil(inv)
}

func TestTranslate(t *testing.T) {
	tests := []struct {
		name       string
		readConnID string
		input      string
	}{
		{"plain json", "e8f9dfc9-bcf3-4f47-a7f9-5b898efbf885", invJSON},
		{"plain oob json", "4c3d1a7e-ac1f-41ec-a8a0-6c0bf1b952a1", invOOBJSON},
		{"url from bc cov", "36a884f7-a105-4494-a395-c6cc7f578e28", "https://email-verification-agent.vonx.io?c_i=eyJAdHlwZSI6ICJkaWQ6c292OkJ6Q2JzTlloTXJqSGlxWkRUVUFTSGc7c3BlYy9jb25uZWN0aW9ucy8xLjAvaW52aXRhdGlvbiIsICJAaWQiOiAiMzZhODg0ZjctYTEwNS00NDk0LWEzOTUtYzZjYzdmNTc4ZTI4IiwgInNlcnZpY2VFbmRwb2ludCI6ICJodHRwczovL2VtYWlsLXZlcmlmaWNhdGlvbi1hZ2VudC52b254LmlvIiwgImxhYmVsIjogIkVtYWlsIFZlcmlmaWNhdGlvbiBTZXJ2aWNlIiwgInJlY2lwaWVudEtleXMiOiBbIkdvQ01MeHZjcTlBbzZqaEhlUkJZN0Vuam4zc042WThCd2hjMTJCNWs4NEJhIl19"},
		{"url from aca-py", "b6a5d0c9-f43b-4c41-8757-1e69e46f29e6", "https://a29e-91-153-20-18.ngrok.io?c_i=eyJAdHlwZSI6ICJkaWQ6c292OkJ6Q2JzTlloTXJqSGlxWkRUVUFTSGc7c3BlYy9jb25uZWN0aW9ucy8xLjAvaW52aXRhdGlvbiIsICJAaWQiOiAiYjZhNWQwYzktZjQzYi00YzQxLTg3NTctMWU2OWU0NmYyOWU2IiwgInJlY2lwaWVudEtleXMiOiBbIjg0akRvUGlxbkxpemdKNUdSeE0xZHh5QUZlS1JXbUp4TFhkWkhxeUpUU3VRIl0sICJzZXJ2aWNlRW5kcG9pbnQiOiAiaHR0cHM6Ly9hMjllLTkxLTE1My0yMC0xOC5uZ3Jvay5pbyIsICJsYWJlbCI6ICJhY2EtcHkuRmFiZXIifQ=="},
		{"url from lissi", "a1eb3595-00fb-4e58-89b4-89b1fe164441", "didcomm://aries_connection_invitation?c_i=eyJAdHlwZSI6ImRpZDpzb3Y6QnpDYnNOWWhNcmpIaXFaRFRVQVNIZztzcGVjL2Nvbm5lY3Rpb25zLzEuMC9pbnZpdGF0aW9uIiwiQGlkIjoiYTFlYjM1OTUtMDBmYi00ZTU4LTg5YjQtODliMWZlMTY0NDQxIiwibGFiZWwiOiJTdGFkdCBNdXN0ZXJoYXVzZW4iLCJzZXJ2aWNlRW5kcG9pbnQiOiJodHRwczovL2RlbW8tYWdlbnQuaW5zdGl0dXRpb25hbC1hZ2VudC5saXNzaS5pZC9kaWRjb21tLyIsImltYWdlVXJsIjoiaHR0cHM6Ly9yb3V0aW5nLmxpc3NpLmlvL2FwaS9JbWFnZS9kZW1vTXVzdGVyaGF1c2VuIiwicmVjaXBpZW50S2V5cyI6WyJCNm5hQ1I5TlRENHdKbmZrNXRxZVdhOWdYNWNIWVpEeWgyaUxOcHF3cldBbyJdfQ"},
		{"url from us", "d6903f21-1c92-45d2-82ad-34e894bfb01b", "didcomm://aries_connection_invitation?c_i=eyJzZXJ2aWNlRW5kcG9pbnQiOiJodHRwOi8vbG9jYWxob3N0OjgwODAvYTJhLzNKY3NUYW9tR2NQdlRBMmlpdDdSZ2YvM0pjc1Rhb21HY1B2VEEyaWl0N1JnZi9MVkRZa1ZMZXpyREhUa0VGSG1vQ216L2Q2OTAzZjIxLTFjOTItNDVkMi04MmFkLTM0ZTg5NGJmYjAxYiIsInJlY2lwaWVudEtleXMiOlsiQmQxTkZnSDlQeW5MakJzSmlqNFFIc2U5WXlxbjJTcFB1RHNSaVBVZFdveXUiXSwiQGlkIjoiZDY5MDNmMjEtMWM5Mi00NWQyLTgyYWQtMzRlODk0YmZiMDFiIiwibGFiZWwiOiJlbXB0eS1sYWJlbCIsIkB0eXBlIjoiZGlkOnNvdjpCekNic05ZaE1yakhpcVpEVFVBU0hnO3NwZWMvY29ubmVjdGlvbnMvMS4wL2ludml0YXRpb24ifQ"},
		{"url with prefix space", "d6903f21-1c92-45d2-82ad-34e894bfb01b", " didcomm://aries_connection_invitation?c_i=eyJzZXJ2aWNlRW5kcG9pbnQiOiJodHRwOi8vbG9jYWxob3N0OjgwODAvYTJhLzNKY3NUYW9tR2NQdlRBMmlpdDdSZ2YvM0pjc1Rhb21HY1B2VEEyaWl0N1JnZi9MVkRZa1ZMZXpyREhUa0VGSG1vQ216L2Q2OTAzZjIxLTFjOTItNDVkMi04MmFkLTM0ZTg5NGJmYjAxYiIsInJlY2lwaWVudEtleXMiOlsiQmQxTkZnSDlQeW5MakJzSmlqNFFIc2U5WXlxbjJTcFB1RHNSaVBVZFdveXUiXSwiQGlkIjoiZDY5MDNmMjEtMWM5Mi00NWQyLTgyYWQtMzRlODk0YmZiMDFiIiwibGFiZWwiOiJlbXB0eS1sYWJlbCIsIkB0eXBlIjoiZGlkOnNvdjpCekNic05ZaE1yakhpcVpEVFVBU0hnO3NwZWMvY29ubmVjdGlvbnMvMS4wL2ludml0YXRpb24ifQ"},
		{"url with suffix space", "d6903f21-1c92-45d2-82ad-34e894bfb01b", "didcomm://aries_connection_invitation?c_i=eyJzZXJ2aWNlRW5kcG9pbnQiOiJodHRwOi8vbG9jYWxob3N0OjgwODAvYTJhLzNKY3NUYW9tR2NQdlRBMmlpdDdSZ2YvM0pjc1Rhb21HY1B2VEEyaWl0N1JnZi9MVkRZa1ZMZXpyREhUa0VGSG1vQ216L2Q2OTAzZjIxLTFjOTItNDVkMi04MmFkLTM0ZTg5NGJmYjAxYiIsInJlY2lwaWVudEtleXMiOlsiQmQxTkZnSDlQeW5MakJzSmlqNFFIc2U5WXlxbjJTcFB1RHNSaVBVZFdveXUiXSwiQGlkIjoiZDY5MDNmMjEtMWM5Mi00NWQyLTgyYWQtMzRlODk0YmZiMDFiIiwibGFiZWwiOiJlbXB0eS1sYWJlbCIsIkB0eXBlIjoiZGlkOnNvdjpCekNic05ZaE1yakhpcVpEVFVBU0hnO3NwZWMvY29ubmVjdGlvbnMvMS4wL2ludml0YXRpb24ifQ\n"},
		{"OOB url from Animo", "f4155f35-090b-4bdd-80e5-f2415309fa30", "https://didcomm.demo.animo.id?oob=eyJAdHlwZSI6Imh0dHBzOi8vZGlkY29tbS5vcmcvb3V0LW9mLWJhbmQvMS4xL2ludml0YXRpb24iLCJAaWQiOiJmNDE1NWYzNS0wOTBiLTRiZGQtODBlNS1mMjQxNTMwOWZhMzAiLCJsYWJlbCI6IkFuaW1vIiwiYWNjZXB0IjpbImRpZGNvbW0vYWlwMSIsImRpZGNvbW0vYWlwMjtlbnY9cmZjMTkiXSwiaGFuZHNoYWtlX3Byb3RvY29scyI6WyJodHRwczovL2RpZGNvbW0ub3JnL2RpZGV4Y2hhbmdlLzEuMCIsImh0dHBzOi8vZGlkY29tbS5vcmcvY29ubmVjdGlvbnMvMS4wIl0sInNlcnZpY2VzIjpbeyJpZCI6IiNpbmxpbmUtMCIsInNlcnZpY2VFbmRwb2ludCI6Imh0dHBzOi8vZGlkY29tbS5kZW1vLmFuaW1vLmlkIiwidHlwZSI6ImRpZC1jb21tdW5pY2F0aW9uIiwicmVjaXBpZW50S2V5cyI6WyJkaWQ6a2V5Ono2TWtmZzNoWDhOUVBVTkZOeTh6ZVN2QlE4Z1A3VjVKREd1aTN3YWdvb3o3Z2tkYSJdLCJyb3V0aW5nS2V5cyI6W119XSwiaW1hZ2VVcmwiOiJodHRwczovL2kuaW1ndXIuY29tL2czYWJjQ08ucG5nIn0"},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			assert.PushTester(t)
			defer assert.PopTester()

			inv, err := Translate(tc.input)
			if err != nil {
				t.Errorf("%s = err (%v)", tc.name, err)
			}
			assert.Equal(tc.readConnID, inv.ID())
			assert.SNotEmpty(inv.ServiceEndpoint())
			assert.SNotEmpty(inv.ServiceEndpoint()[0].RecipientKeys)
		})
	}

}
