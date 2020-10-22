package rpc

import (
	"github.com/findy-network/findy-grpc/jwt"
	"github.com/lainio/err2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// ClientCfg is configuration struct for making a gRPC client connection.
type ClientCfg struct {
	CertFile string
	JWT      string
	Addr     string
	TLS      bool
}

// ClientConn creates a client connection according the configuration to gRPC
// server.
func ClientConn(cfg ClientCfg) (conn *grpc.ClientConn, err error) {
	defer err2.Return(&err)

	// for now we use only server side TLS, if we go mTLS use NewTLS()
	creds, err := credentials.NewClientTLSFromFile(cfg.CertFile, "")
	err2.Check(err)

	opts := []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure()}
	if cfg.TLS {
		// we wrap our JWT token to Oauth token
		perRPC := oauth.NewOauthAccess(jwt.OauthToken(cfg.JWT))

		opts = []grpc.DialOption{
			grpc.WithPerRPCCredentials(perRPC),
			grpc.WithTransportCredentials(creds),
			grpc.WithBlock(),
		}
	}
	return grpc.Dial(cfg.Addr, opts...)
}
