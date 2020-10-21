package rpc

import (
	. "github.com/lainio/err2"
	"github.com/optechlab/findy-grpc/jwt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

type ClientCfg struct {
	CertFile string
	JWT      string
	Addr     string
	TLS      bool
}

func ClientConn(cfg ClientCfg) (conn *grpc.ClientConn, err error) {
	defer Return(&err)

	// for now we use only server side TLS, if we go mTLS use NewTLS()
	creds, err := credentials.NewClientTLSFromFile(cfg.CertFile, "")
	Check(err)

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
