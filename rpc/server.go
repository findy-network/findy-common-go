package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/findy-network/findy-grpc/jwt"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/test/bufconn"
)

// ServerCfg is gRPC server configuration struct for service init
type ServerCfg struct {
	PKI
	Port     int
	TLS      bool
	TestLis  *bufconn.Listener
	Register func(s *grpc.Server) error
}

// Server creates a gRPC server with TLS and JWT token authorization.
func Server(cfg ServerCfg) (s *grpc.Server, err error) {
	defer err2.Return(&err)

	opts := make([]grpc.ServerOption, 0, 4)
	if cfg.TLS {
		creds, err := loadTLSCredentials(cfg.PKI)
		err2.Check(err)

		opts = append(opts,
			grpc.Creds(creds),
			grpc.UnaryInterceptor(jwt.EnsureValidToken),
			grpc.StreamInterceptor(jwt.EnsureValidTokenStream),
		)
	}
	return grpc.NewServer(opts...), nil
}

// Serve builds up the gRPC test server and starts to serve. This function
// blocks. In most cases you should start it as goroutine. TODO: graceful stop!
func Serve(cfg ServerCfg) {
	defer err2.Catch(func(err error) {
		glog.Error(err)
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	var lis net.Listener
	if cfg.TestLis != nil {
		lis = cfg.TestLis
		glog.V(0).Infoln("listen to TEST BUFFER:", addr)
	} else {
		var err error
		lis, err = net.Listen("tcp", addr)
		err2.Check(err)
		glog.V(5).Infoln("listen to:", addr)
	}
	s, err := Server(cfg)
	err2.Check(err)
	err2.Check(cfg.Register(s))
	glog.V(5).Infoln("start to serve..")
	err2.Check(s.Serve(lis))
}

func loadTLSCredentials(pw PKI) (creds credentials.TransportCredentials, err error) {
	defer err2.Return(&err)

	caCert := err2.Bytes.Try(ioutil.ReadFile(pw.Client.CertFile))
	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(caCert)

	// Load server's certificate and private key
	serverCert, err := tls.LoadX509KeyPair(pw.Server.CertFile, pw.Server.KeyFile)
	err2.Check(err)

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    rootCAs,
	}

	glog.V(0).Infoln("cert files loaded")
	return credentials.NewTLS(config), nil
}
