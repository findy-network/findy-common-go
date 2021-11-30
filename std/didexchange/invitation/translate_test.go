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

	oldInvURL = "https://email-verification-agent.vonx.io?c_i=eyJAdHlwZSI6ICJkaWQ6c292OkJ6Q2JzTlloTXJqSGlxWkRUVUFTSGc7c3BlYy9jb25uZWN0aW9ucy8xLjAvaW52aXRhdGlvbiIsICJAaWQiOiAiMzZhODg0ZjctYTEwNS00NDk0LWEzOTUtYzZjYzdmNTc4ZTI4IiwgInNlcnZpY2VFbmRwb2ludCI6ICJodHRwczovL2VtYWlsLXZlcmlmaWNhdGlvbi1hZ2VudC52b254LmlvIiwgImxhYmVsIjogIkVtYWlsIFZlcmlmaWNhdGlvbiBTZXJ2aWNlIiwgInJlY2lwaWVudEtleXMiOiBbIkdvQ01MeHZjcTlBbzZqaEhlUkJZN0Vuam4zc042WThCd2hjMTJCNWs4NEJhIl19"
	newInvURL = "didcomm://aries_connection_invitation?c_i=eyJAdHlwZSI6ImRpZDpzb3Y6QnpDYnNOWWhNcmpIaXFaRFRVQVNIZztzcGVjL2Nvbm5lY3Rpb25zLzEuMC9pbnZpdGF0aW9uIiwiQGlkIjoiYTFlYjM1OTUtMDBmYi00ZTU4LTg5YjQtODliMWZlMTY0NDQxIiwibGFiZWwiOiJTdGFkdCBNdXN0ZXJoYXVzZW4iLCJzZXJ2aWNlRW5kcG9pbnQiOiJodHRwczovL2RlbW8tYWdlbnQuaW5zdGl0dXRpb25hbC1hZ2VudC5saXNzaS5pZC9kaWRjb21tLyIsImltYWdlVXJsIjoiaHR0cHM6Ly9yb3V0aW5nLmxpc3NpLmlvL2FwaS9JbWFnZS9kZW1vTXVzdGVyaGF1c2VuIiwicmVjaXBpZW50S2V5cyI6WyJCNm5hQ1I5TlRENHdKbmZrNXRxZVdhOWdYNWNIWVpEeWgyaUxOcHF3cldBbyJdfQ"
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
	inv, err := Translate(invJSON)
	assert.NoError(t, err)
	assert.NotNil(t, inv)

	inv, err = Translate(oldInvURL)
	assert.NoError(t, err)
	assert.NotNil(t, inv)

	inv, err = Translate(newInvURL)
	assert.NoError(t, err)
	assert.NotNil(t, inv)
}
