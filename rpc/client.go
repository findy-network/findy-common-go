package rpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/findy-network/findy-grpc/jwt"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/oauth"
)

// ClientCfg is gRPC client initialization and configuration struct.
type ClientCfg struct {
	*PKI
	JWT  string
	Addr string
	Opts []grpc.DialOption
}

// ClientConn opens client connection with given configuration.
func ClientConn(cfg ClientCfg) (conn *grpc.ClientConn, err error) {
	defer err2.Return(&err)

	// for now we use only server side TLS, if we go mTLS use NewTLS()
	creds, err := loadClientTLSFromFile(cfg.PKI)
	err2.Check(err)

	glog.V(5).Infoln("new tls client ready")

	opts := []grpc.DialOption{grpc.WithBlock(), grpc.WithInsecure()}
	if cfg.PKI != nil {
		// we wrap our JWT token to Oauth token
		perRPC := oauth.NewOauthAccess(jwt.OauthToken(cfg.JWT))
		glog.V(10).Infoln("grpc oauth wrap for JWT done")

		opts = []grpc.DialOption{
			grpc.WithPerRPCCredentials(perRPC),
			grpc.WithTransportCredentials(creds),
			//grpc.WithBlock(), // dont use!! you don't get immediate error messages
		}
	}
	if cfg.Opts != nil {
		opts = append(opts, cfg.Opts...)
	}
	glog.V(5).Infoln("going to dial:", cfg.Addr)
	return grpc.DialContext(context.Background(), cfg.Addr, opts...)
}

func loadClientTLSFromFile(pw *PKI) (creds credentials.TransportCredentials, err error) {
	defer err2.Return(&err)

	caCert := err2.Bytes.Try(ioutil.ReadFile(pw.Server.CertFile))
	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(caCert)

	clientCert, err := tls.LoadX509KeyPair(pw.Client.CertFile, pw.Client.KeyFile)
	tlsConf := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            rootCAs,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS13,
		ServerName:         "localhost",
	}

	return credentials.NewTLS(tlsConf), nil
}
