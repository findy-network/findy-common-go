package invitation

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
)

func TestBuild(t *testing.T) {
	var inv Invitation
	var err error

	inv, err = Translate(invJSON)
	assert.NoError(t, err)
	assert.NotNil(t, inv)

	s, err := Build(inv)
	assert.NoError(t, err)
	assert.NotEmpty(t, s)

	inv, err = Translate(s)
	assert.NoError(t, err)
	assert.NotNil(t, inv)
}

func TestTranslate(t *testing.T) {
	tests := []struct {
		name   string
		readConnID string
		input  string
	}{
		{"plain json", "e8f9dfc9-bcf3-4f47-a7f9-5b898efbf885", invJSON},
		{"url from bc cov",  "36a884f7-a105-4494-a395-c6cc7f578e28", "https://email-verification-agent.vonx.io?c_i=eyJAdHlwZSI6ICJkaWQ6c292OkJ6Q2JzTlloTXJqSGlxWkRUVUFTSGc7c3BlYy9jb25uZWN0aW9ucy8xLjAvaW52aXRhdGlvbiIsICJAaWQiOiAiMzZhODg0ZjctYTEwNS00NDk0LWEzOTUtYzZjYzdmNTc4ZTI4IiwgInNlcnZpY2VFbmRwb2ludCI6ICJodHRwczovL2VtYWlsLXZlcmlmaWNhdGlvbi1hZ2VudC52b254LmlvIiwgImxhYmVsIjogIkVtYWlsIFZlcmlmaWNhdGlvbiBTZXJ2aWNlIiwgInJlY2lwaWVudEtleXMiOiBbIkdvQ01MeHZjcTlBbzZqaEhlUkJZN0Vuam4zc042WThCd2hjMTJCNWs4NEJhIl19"},
		{"url from lissi", "a1eb3595-00fb-4e58-89b4-89b1fe164441", "didcomm://aries_connection_invitation?c_i=eyJAdHlwZSI6ImRpZDpzb3Y6QnpDYnNOWWhNcmpIaXFaRFRVQVNIZztzcGVjL2Nvbm5lY3Rpb25zLzEuMC9pbnZpdGF0aW9uIiwiQGlkIjoiYTFlYjM1OTUtMDBmYi00ZTU4LTg5YjQtODliMWZlMTY0NDQxIiwibGFiZWwiOiJTdGFkdCBNdXN0ZXJoYXVzZW4iLCJzZXJ2aWNlRW5kcG9pbnQiOiJodHRwczovL2RlbW8tYWdlbnQuaW5zdGl0dXRpb25hbC1hZ2VudC5saXNzaS5pZC9kaWRjb21tLyIsImltYWdlVXJsIjoiaHR0cHM6Ly9yb3V0aW5nLmxpc3NpLmlvL2FwaS9JbWFnZS9kZW1vTXVzdGVyaGF1c2VuIiwicmVjaXBpZW50S2V5cyI6WyJCNm5hQ1I5TlRENHdKbmZrNXRxZVdhOWdYNWNIWVpEeWgyaUxOcHF3cldBbyJdfQ"},
		{"url from us", "d6903f21-1c92-45d2-82ad-34e894bfb01b", "didcomm://aries_connection_invitation?c_i=eyJzZXJ2aWNlRW5kcG9pbnQiOiJodHRwOi8vbG9jYWxob3N0OjgwODAvYTJhLzNKY3NUYW9tR2NQdlRBMmlpdDdSZ2YvM0pjc1Rhb21HY1B2VEEyaWl0N1JnZi9MVkRZa1ZMZXpyREhUa0VGSG1vQ216L2Q2OTAzZjIxLTFjOTItNDVkMi04MmFkLTM0ZTg5NGJmYjAxYiIsInJlY2lwaWVudEtleXMiOlsiQmQxTkZnSDlQeW5MakJzSmlqNFFIc2U5WXlxbjJTcFB1RHNSaVBVZFdveXUiXSwiQGlkIjoiZDY5MDNmMjEtMWM5Mi00NWQyLTgyYWQtMzRlODk0YmZiMDFiIiwibGFiZWwiOiJlbXB0eS1sYWJlbCIsIkB0eXBlIjoiZGlkOnNvdjpCekNic05ZaE1yakhpcVpEVFVBU0hnO3NwZWMvY29ubmVjdGlvbnMvMS4wL2ludml0YXRpb24ifQ"},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			inv, err := Translate(tc.input)
			if err != nil {
				t.Errorf("%s = err (%v)", tc.name, err)
			}
			assert.Equal(t, tc.readConnID, inv.ID)
		})
	}

}
