package jwt

import (
	"context"
	"errors"
	"fmt"

	"github.com/form3tech-oss/jwt-go"
	"github.com/golang/glog"
)

const defaultLabel = "n/a"

type Token struct {
	Label   string
	AgentID string
	Raw     string
}

// TokenFromContext receives the user data stored to context
// NOTE: token is validated by middleware
// before storing to context, actual verification is done currently there
func TokenFromContext(ctx context.Context, contextKey interface{}) (*Token, error) {
	user := ctx.Value(contextKey)
	if user == nil {
		return nil, errors.New("no authenticated user found")
	}

	jwtToken, ok := user.(*jwt.Token)
	if !ok {
		return nil, errors.New("no authenticated user found")
	}

	if !jwtToken.Valid {
		return nil, errors.New("token is not valid")
	}

	if jwtToken.Raw == "" {
		return nil, fmt.Errorf("no raw token found")
	}

	// TODO: middleware does not support custom claims currently
	// so token is parsed twice, refactor!
	tokenWithClaims, err := jwt.ParseWithClaims(jwtToken.Raw, &customClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})
	if err != nil {
		glog.Errorf("custom claims parsing failed, is token generated properly? (%s)", err)
		return nil, errors.New("custom claims missing")
	}

	claims, ok := tokenWithClaims.Claims.(*customClaims)
	if !ok {
		return nil, errors.New("no claims found for token")
	}

	if claims.Username == "" {
		return nil, errors.New("no cloud agent DID found for token")
	}

	label := defaultLabel
	if claims.Label != "" {
		label = claims.Label
	}

	return &Token{
		AgentID: claims.Username,
		Label:   label,
		Raw:     jwtToken.Raw,
	}, nil
}

// TokenToContext stores user data from raw token to context
// Used with tests
func TokenToContext(ctx context.Context, contextKey interface{}, token *Token) context.Context {
	jwtToken, err := jwt.ParseWithClaims(token.Raw, &customClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return key, nil
		})

	if err != nil {
		glog.Error(err)
		return nil
	}
	return context.WithValue(ctx, contextKey, jwtToken)
}
