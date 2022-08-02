package rpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io/ioutil"

	"github.com/findy-network/findy-common-go/jwt"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/credentials/oauth"
)

// ClientCfg is gRPC client initialization and configuration struct.
type ClientCfg struct {
	*PKI
	JWT  string
	Addr string
	Opts []grpc.DialOption
	// Client should set the insecure flag if token should be sent
	// over insecure connection - intended for internal, secure networks only
	Insecure bool
}

func (c ClientCfg) GetRequestMetadata(ctx context.Context, uri ...string) (map[string]string, error) {
	res := make(map[string]string)
	// bypass oauth transport security restrictions if in insecure mode
	if c.Insecure && c.PKI == nil {
		res["Authorization"] = "Bearer " + c.JWT
	}
	return res, nil
}

func (c ClientCfg) RequireTransportSecurity() bool {
	return !c.Insecure || c.PKI != nil
}

// ClientConn opens client connection with given configuration.
func ClientConn(cfg ClientCfg) (conn *grpc.ClientConn, err error) {
	defer err2.Return(&err)

	opts := []grpc.DialOption{
		grpc.WithBlock(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}
	if cfg.PKI != nil {
		glog.V(5).Infoln("new tls client ready")
		opts = make([]grpc.DialOption, 0)
		if cfg.JWT != "" {
			// we wrap our JWT token to Oauth token
			perRPC := oauth.NewOauthAccess(jwt.OauthToken(cfg.JWT))
			glog.V(10).Infoln("grpc oauth wrap for JWT done")
			opts = append(opts, grpc.WithPerRPCCredentials(perRPC))
		}
		// for now, we use only server side TLS, if we go mTLS use NewTLS()
		creds := try.To1(loadClientTLSFromFile(cfg.PKI))
		opts = append(opts, grpc.WithTransportCredentials(creds))
		// dont use grpc.WithBlock()!! you don't get immediate error messages
	} else if cfg.Insecure && cfg.JWT != "" {
		glog.V(10).Infoln("sending token over insecure transport")
		opts = append(opts, grpc.WithPerRPCCredentials(cfg))
	}

	if cfg.Opts != nil {
		opts = append(opts, cfg.Opts...)
	}
	glog.V(5).Infoln("going to dial:", cfg.Addr)
	return grpc.DialContext(context.Background(), cfg.Addr, opts...)
}

func loadClientTLSFromFile(pw *PKI) (creds credentials.TransportCredentials, err error) {
	defer err2.Return(&err)

	caCert := try.To1(ioutil.ReadFile(pw.Server.CertFile))
	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(caCert)

	clientCert, err := tls.LoadX509KeyPair(pw.Client.CertFile, pw.Client.KeyFile)
	tlsConf := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            rootCAs,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
		ServerName:         pw.ServerName,
	}

	return credentials.NewTLS(tlsConf), nil
}
