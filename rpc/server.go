package rpc

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"

	"github.com/findy-network/findy-grpc/jwt"
	"github.com/golang/glog"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/lainio/err2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/test/bufconn"
)

// ServerCfg is gRPC server configuration struct for service init
type ServerCfg struct {
	*PKI
	Port      int
	TestLis   *bufconn.Listener
	Register  func(s *grpc.Server) error
	JWTSecret string
}

// Server creates a gRPC server with TLS and JWT token authorization.
func Server(cfg *ServerCfg) (s *grpc.Server, err error) {
	defer err2.Return(&err)

	// TODO: require always a custom secret in production mode
	if cfg.JWTSecret != "" {
		jwt.SetJWTSecret(cfg.JWTSecret)
	}

	opts := make([]grpc.ServerOption, 0, 4)
	if cfg.PKI != nil {
		creds, err := loadTLSCredentials(cfg.PKI)
		err2.Check(err)

		opts = append(opts,
			grpc.Creds(creds),
			//grpc.UnaryInterceptor(jwt.EnsureValidToken),
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				grpc_auth.UnaryServerInterceptor(jwt.CheckTokenValidity),
				grpc_recovery.UnaryServerInterceptor(),
			)),
			//grpc.StreamInterceptor(jwt.EnsureValidTokenStream),
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
				grpc_auth.StreamServerInterceptor(jwt.CheckTokenValidity),
				grpc_recovery.StreamServerInterceptor(),
			)),
		)
	}
	return grpc.NewServer(opts...), nil
}

// Serve builds up the gRPC server and starts to serve. Note that the function
// blocks. In most cases you should start it as goroutine. To be able to
// gracefully stop the gRPC server you should call PrepareServe which builds
// everything ready but leaves calling the grpcServer.Serve for you.
func Serve(cfg *ServerCfg) {
	defer err2.Catch(func(err error) {
		glog.Error(err)
	})

	s, lis, err := PrepareServe(cfg)
	err2.Check(err)

	glog.V(5).Infoln("start to serve..")
	err2.Check(s.Serve(lis))
}

func PrepareServe(cfg *ServerCfg) (s *grpc.Server, lis net.Listener, err error) {
	defer err2.Return(&err)

	addr := fmt.Sprintf(":%d", cfg.Port)
	if cfg.TestLis != nil {
		lis = cfg.TestLis
	} else {
		lis, err = net.Listen("tcp", addr)
		err2.Check(err)
		glog.V(5).Infoln("listen to:", addr)
	}
	s, err = Server(cfg)
	err2.Check(err)
	err2.Check(cfg.Register(s))

	return s, lis, nil
}

func loadTLSCredentials(pw *PKI) (creds credentials.TransportCredentials, err error) {
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

	glog.V(3).Infoln("cert files loaded")
	return credentials.NewTLS(config), nil
}
