package acator

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"
	"testing"

	"github.com/duo-labs/webauthn/protocol/webauthncose"
	"github.com/findy-network/findy-grpc/acator/authenticator"
	"github.com/findy-network/findy-grpc/acator/cose"
	"github.com/stretchr/testify/assert"
)

var data = []byte{
	73,
	150,
	13,
	229,
	136,
	14,
	140,
	104,
	116,
	52,
	23,
	15,
	100,
	118,
	96,
	91,
	143,
	228,
	174,
	185,
	162,
	134,
	50,
	199,
	153,
	92,
	243,
	186,
	131,
	29,
	151,
	99,
	69,
	96,
	2,
	220,
	119,
	173,
	206,
	0,
	2,
	53,
	188,
	198,
	10,
	100,
	139,
	11,
	37,
	241,
	240,
	85,
	3,
	0,
	56,
	1,
	237,
	206,
	166,
	129,
	7,
	101,
	231,
	238,
	111,
	72,
	127,
	231,
	19,
	25,
	155,
	133,
	152,
	214,
	130,
	22,
	241,
	21,
	53,
	127,
	108,
	97,
	227,
	213,
	234,
	137,
	15,
	104,
	231,
	125,
	16,
	62,
	158,
	129,
	211,
	214,
	70,
	94,
	221,
	6,
	200,
	118,
	172,
	223,
	102,
	17,
	173,
	18,
	141,
	38,
	128,
	165,
	1,
	2,
	3,
	38,
	32,
	1,
	33,
	88,
	32,
	135,
	4,
	102,
	211,
	195,
	250,
	222,
	210,
	130,
	176,
	66,
	170,
	143,
	3,
	154,
	76,
	202,
	251,
	152,
	79,
	68,
	46,
	192,
	130,
	191,
	126,
	227,
	13,
	132,
	129,
	205,
	196,
	34,
	88,
	32,
	2,
	185,
	214,
	73,
	248,
	10,
	153,
	52,
	227,
	90,
	35,
	175,
	180,
	219,
	189,
	10,
	78,
	227,
	36,
	50,
	105,
	84,
	155,
	223,
	127,
	176,
	151,
	255,
	20,
	33,
	205,
	185,
}

// Credential creation/Register
var challengeJSON = `{
  "publicKey": {
    "challenge": "7vH6L70QspI4ToHZ6gTJLj74jQ9jj/AzlIQkSlkZX8E=",
    "rp": {
      "name": "Foobar Corp.",
      "id": "localhost"
    },
    "user": {
      "name": "debug1",
      "displayName": "debug1",
      "id": "nsWi6Nr9u8nAAQ=="
    },
    "pubKeyCredParams": [
      {
        "type": "public-key",
        "alg": -7
      },
      {
        "type": "public-key",
        "alg": -35
      },
      {
        "type": "public-key",
        "alg": -36
      },
      {
        "type": "public-key",
        "alg": -257
      },
      {
        "type": "public-key",
        "alg": -258
      },
      {
        "type": "public-key",
        "alg": -259
      },
      {
        "type": "public-key",
        "alg": -37
      },
      {
        "type": "public-key",
        "alg": -38
      },
      {
        "type": "public-key",
        "alg": -39
      },
      {
        "type": "public-key",
        "alg": -8
      }
    ],
    "authenticatorSelection": {
      "requireResidentKey": false,
      "userVerification": "preferred"
    },
    "timeout": 60000
  }
}
`

var challengeResponseJSON = `{
  "id": "Ae3OpoEHZefub0h_5xMZm4WY1oIW8RU1f2xh49XqiQ9o530QPp6B09ZGXt0GyHas32YRrRKNJoA",
  "rawId": "Ae3OpoEHZefub0h_5xMZm4WY1oIW8RU1f2xh49XqiQ9o530QPp6B09ZGXt0GyHas32YRrRKNJoA",
  "type": "public-key",
  "response": {
    "attestationObject": "o2NmbXRkbm9uZWdhdHRTdG10oGhhdXRoRGF0YVi8SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2NFYALcd63OAAI1vMYKZIsLJfHwVQMAOAHtzqaBB2Xn7m9If-cTGZuFmNaCFvEVNX9sYePV6okPaOd9ED6egdPWRl7dBsh2rN9mEa0SjSaApQECAyYgASFYIIcEZtPD-t7SgrBCqo8DmkzK-5hPRC7Agr9-4w2Egc3EIlggArnWSfgKmTTjWiOvtNu9Ck7jJDJpVJvff7CX_xQhzbk",
    "clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uY3JlYXRlIiwiY2hhbGxlbmdlIjoiN3ZINkw3MFFzcEk0VG9IWjZnVEpMajc0alE5ampfQXpsSVFrU2xrWlg4RSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0"
  }
}
`

// ==== assertion/Login
var credentialRequestOptions = `{
  "publicKey": {
    "challenge": "yifGGzsupyIW3xxZoL09vEbJQYBrQaarZf4CN8GUvWE=",
    "timeout": 60000,
    "rpId": "localhost",
    "allowCredentials": [
      {
        "type": "public-key",
        "id": "Ae3OpoEHZefub0h/5xMZm4WY1oIW8RU1f2xh49XqiQ9o530QPp6B09ZGXt0GyHas32YRrRKNJoA="
      }
    ]
  }
}
`

var credentialRequestOptionsFmt = `{
  "publicKey": {
    "challenge": "yifGGzsupyIW3xxZoL09vEbJQYBrQaarZf4CN8GUvWE=",
    "timeout": 60000,
    "rpId": "localhost",
    "allowCredentials": [
      {
        "type": "public-key",
        "id": "%s"
      }
    ]
  }
}
`

var authenticatorAssertionResponse = `{
  "id": "Ae3OpoEHZefub0h_5xMZm4WY1oIW8RU1f2xh49XqiQ9o530QPp6B09ZGXt0GyHas32YRrRKNJoA",
  "rawId": "Ae3OpoEHZefub0h_5xMZm4WY1oIW8RU1f2xh49XqiQ9o530QPp6B09ZGXt0GyHas32YRrRKNJoA",
  "type": "public-key",
  "response": {
    "authenticatorData": "SZYN5YgOjGh0NBcPZHZgW4_krrmihjLHmVzzuoMdl2MFYAU-Hg",
    "clientDataJSON": "eyJ0eXBlIjoid2ViYXV0aG4uZ2V0IiwiY2hhbGxlbmdlIjoieWlmR0d6c3VweUlXM3h4Wm9MMDl2RWJKUVlCclFhYXJaZjRDTjhHVXZXRSIsIm9yaWdpbiI6Imh0dHA6Ly9sb2NhbGhvc3Q6ODA4MCIsImNyb3NzT3JpZ2luIjpmYWxzZX0",
    "signature": "MEQCIEXR3OAmve4jJoER6-HR1qVQrCkuj9xFapZqdj13Ms6qAiB8Vj1QuZog67zpzt8ywVwK8iqViaH34y9AXPKTh1JImA",
    "userHandle": "nsWi6Nr9u8nAAQ"
  }
}
`

func TestRegister(t *testing.T) {
	r := strings.NewReader(challengeJSON)
	out, err := Register(r)
	assert.NoError(t, err)
	assert.NotNil(t, out)

	js, err := json.Marshal(out)
	assert.NoError(t, err)

	ccd, err := ParseResponse(string(js))
	assert.NoError(t, err)
	assert.NotNil(t, ccd)
	println(ccd.ID)
}

func TestLogin(t *testing.T) {
	credID := "QABRwuCGuynqf0lf35FK-CG-PY_WXai1oCzIZdIbY4S-81SMwZg1hD_V75cWyPwrGmFS4NpVegzMg8c-XnIBYPvmsl0hmkoxMCPDe7tKgV0kcSBC2Fy-BN8B22Ftt78CrZQUbYMJruutTWEp818XaVH9KDlRuV4s9k0G-T23lMUjqJHzUn-gfMbuP1uuVILV6rQu6kw"
	data, err := base64.RawURLEncoding.DecodeString(credID)
	assert.NoError(t, err)

	newStr := base64.StdEncoding.EncodeToString(data)
	assert.NoError(t, err)
	credID = newStr
	credReq := fmt.Sprintf(credentialRequestOptionsFmt, credID)
	car, err := Login(strings.NewReader(credReq))
	assert.NoError(t, err)
	assert.NotNil(t, car)

	js, err := json.Marshal(car)
	assert.NoError(t, err)

	pcad, err := ParseAssertionResponse(string(js))
	assert.NoError(t, err)
	assert.NotNil(t, pcad)

	credentialBytes := pcad.Response.AuthenticatorData.AttData.CredentialPublicKey
	err = pcad.Verify("yifGGzsupyIW3xxZoL09vEbJQYBrQaarZf4CN8GUvWE",
		"localhost", "http://localhost:8080", false,
		credentialBytes)
	assert.NoError(t, err)

}

func TestFinnishLogin(t *testing.T) {
	//reqBody := ioutil.NopCloser(bytes.NewReader([]byte(authenticatorAssertionResponse)))
	//httpReq := &http.Request{Body: reqBody}
}

func TestParseAssertionResponse(t *testing.T) {
	ccd, err := ParseResponse(challengeResponseJSON)

	ad, err := ParseAssertionResponse(authenticatorAssertionResponse)
	assert.NoError(t, err)

	// Step 15. Let hash be the result of computing a hash over the cData using SHA-256.
	clientDataHash := sha256.Sum256(ad.Raw.AssertionResponse.ClientDataJSON)

	// Step 16. Using the credential public key looked up in step 3, verify that sig is
	// a valid signature over the binary concatenation of authData and hash.

	sigData := append(ad.Raw.AssertionResponse.AuthenticatorData, clientDataHash[:]...)

	credentialBytes := ccd.Response.AttestationObject.AuthData.AttData.CredentialPublicKey

	coseKey, err := cose.NewFromData(credentialBytes)
	assert.NoError(t, err)
	valid := coseKey.Verify(sigData, ad.Response.Signature)
	assert.True(t, valid)
	keyData, _ := coseKey.Marshal()
	//assert.NoError(t, err)
	assert.Len(t, credentialBytes, len(keyData))
	//assert.Equal(t, credentialBytes, keyData)

	key, err := webauthncose.ParsePublicKey(keyData)
	assert.NoError(t, err)
	k, ok := key.(webauthncose.EC2PublicKeyData)
	assert.True(t, ok)
	pubkey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     big.NewInt(0).SetBytes(k.XCoord),
		Y:     big.NewInt(0).SetBytes(k.YCoord),
	}

	valid = cose.Verify(pubkey, sigData, ad.Response.Signature)
	assert.NoError(t, err)
	assert.True(t, valid)

	valid, err = webauthncose.VerifySignature(key, sigData, ad.Response.Signature)
	assert.NoError(t, err)
	assert.True(t, valid)

	err = ad.Verify("yifGGzsupyIW3xxZoL09vEbJQYBrQaarZf4CN8GUvWE",
		"localhost", "http://localhost:8080", false,
		credentialBytes)
	assert.NoError(t, err)

	json, err := authenticator.MarshalData(&ad.Response.AuthenticatorData)
	assert.NoError(t, err)
	assert.Equal(t, json, []uint8(ad.Raw.AssertionResponse.AuthenticatorData))
}

func TestActorUnmarshal(t *testing.T) {
	err := AcatorUnmarshal(data)
	assert.NoError(t, err)
}

func TestParseResponse(t *testing.T) {
	ccd, err := ParseResponse(challengeResponseJSON)
	assert.NoError(t, err)
	assert.NotNil(t, ccd)

	json, err := authenticator.MarshalData(&ccd.Response.AttestationObject.AuthData)
	assert.NoError(t, err)
	assert.Len(t, json, len(ccd.Response.AttestationObject.RawAuthData))
}

func TestNewCreation(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		//wantCred protocol.CredentialCreation
		wantErr bool
	}{
		{name: "success", args: args{s: challengeJSON}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewCredentialCreation(strings.NewReader(tt.args.s))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewCredentialCreation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestNewAssertion(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		//wantCred protocol.CredentialCreation
		wantErr bool
	}{
		{name: "success", args: args{s: credentialRequestOptions}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewAssertion(strings.NewReader(tt.args.s))
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAssertion() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}