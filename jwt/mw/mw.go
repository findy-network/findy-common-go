package mw

import (
	"net/http"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/form3tech-oss/jwt-go"
)

type Options struct {
	ValidationKey  string
	QueryParameter string
	ErrorHandler   func(w http.ResponseWriter, r *http.Request, err string)
}

type Middleware struct {
	*jwtmiddleware.JWTMiddleware
}

func New(options *Options) *Middleware {
	validationKey := []byte(options.ValidationKey)
	return &Middleware{
		jwtmiddleware.New(jwtmiddleware.Options{
			ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
				return validationKey, nil
			},
			SigningMethod:       jwt.SigningMethodHS256,
			EnableAuthOnOptions: true, // TODO: disable in production?
			Extractor: jwtmiddleware.FromFirst(
				jwtmiddleware.FromAuthHeader,
				jwtmiddleware.FromParameter(options.QueryParameter), // TODO: unsafe but needed for browser websocket auth
			),
			ErrorHandler: options.ErrorHandler,
		}),
	}
}
