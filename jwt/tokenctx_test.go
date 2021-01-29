package jwt

import (
	"context"
	"reflect"
	"testing"

	"github.com/form3tech-oss/jwt-go"
)

func TestTokenFromContext(t *testing.T) {
	ctx := context.Background()
	key := "user"
	userName := "cloudDID"
	label := "label"
	raw := BuildJWTWithLabel(userName, label)
	rawNoLabel := BuildJWT(userName)

	var createCtx = func(val interface{}) context.Context {
		return context.WithValue(ctx, key, val)
	}

	tests := []struct {
		name string
		ctx  context.Context
		err  bool
		data *Token
	}{
		{"non-existing user", context.Background(), true, nil},
		{"invalid struct", createCtx(&customClaims{}), true, nil},
		{"invalid token", createCtx(&jwt.Token{}), true, nil},
		{"no claims", createCtx(&jwt.Token{Valid: true}), true, nil},
		{"no id", createCtx(&jwt.Token{Valid: true, Claims: &customClaims{}}), true, nil},
		{"no raw token", createCtx(
			&jwt.Token{Valid: true, Claims: &customClaims{Username: userName}},
		), true, nil},
		{"no raw token", createCtx(
			&jwt.Token{Valid: true, Claims: &customClaims{Username: userName}},
		), true, nil},
		{"no label", createCtx(
			&jwt.Token{Valid: true, Claims: &customClaims{Username: userName}, Raw: rawNoLabel},
		), false, &Token{Label: defaultLabel, AgentID: userName, Raw: rawNoLabel}},
		{"full data", createCtx(
			&jwt.Token{Valid: true, Claims: &customClaims{Username: userName, Label: label}, Raw: raw},
		), false, &Token{Label: label, AgentID: userName, Raw: raw}},
	}

	for _, testCase := range tests {
		tc := testCase
		t.Run(tc.name, func(t *testing.T) {
			data, err := TokenFromContext(tc.ctx, key)
			if tc.err {
				if err == nil {
					t.Errorf("%s = err (%v)\n want (%v)", tc.name, err, tc.err)
				}
				if data != nil {
					t.Errorf("%s = err (%v)\n want (%v)", tc.name, data, nil)
				}
			} else if !reflect.DeepEqual(data, tc.data) {
				t.Errorf("%s = err (%v)\n want (%v)", tc.name, data, tc.data)
			}
		})
	}

}
