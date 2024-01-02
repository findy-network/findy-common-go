package rpc

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"os"

	"github.com/findy-network/findy-common-go/jwt"
	"github.com/golang/glog"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
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

	NoAuthorization bool
}

// Server creates a gRPC server with TLS and JWT token authorization.
func Server(cfg *ServerCfg) (s *grpc.Server, err error) {
	defer err2.Handle(&err)

	// TODO: require always a custom secret in production mode
	if cfg.JWTSecret != "" {
		jwt.SetJWTSecret(cfg.JWTSecret)
	}

	glog.V(2).Infof("cfg.PKI: %v", cfg.PKI)
	opts := make([]grpc.ServerOption, 0, 4)
	if cfg.PKI != nil {
		creds := try.To1(loadTLSCredentials(cfg.PKI))
		opts = append(opts, grpc.Creds(creds))
	}

	errHandler := func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		glog.V(1).Infoln("-agent gRPC call:", info.FullMethod)
		resp, err := handler(ctx, req)
		if err != nil {
			glog.Errorf("method %q failed: %s", info.FullMethod, err)
		}
		return resp, err
	}

	if cfg.NoAuthorization {
		glog.V(1).Infoln("no authorization")
		opts = append(opts,
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				grpc_recovery.UnaryServerInterceptor(),
				errHandler,
			)),
			grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
				grpc_recovery.StreamServerInterceptor(),
			)),
		)
	} else {
		glog.V(1).Infoln("jwt token validity checked")
		opts = append(opts,
			grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
				grpc_auth.UnaryServerInterceptor(jwt.CheckTokenValidity),
				grpc_recovery.UnaryServerInterceptor(),
				errHandler,
			)),
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
	defer err2.Catch(err2.Err(func(err error) {
		glog.Error(err)
	}))

	s, lis := try.To2(PrepareServe(cfg))

	glog.V(5).Infoln("start to serve..")
	try.To(s.Serve(lis))
}

// PrepareServe builds gRPC server that allows graceful shutdown. You must start
// it by yourself with Serve() that blocks which typically means that you need a
// goroutine for that.
func PrepareServe(cfg *ServerCfg) (s *grpc.Server, lis net.Listener, err error) {
	defer err2.Handle(&err)

	addr := fmt.Sprintf(":%d", cfg.Port)
	if cfg.TestLis != nil {
		lis = cfg.TestLis
	} else {
		lis = try.To1(net.Listen("tcp", addr))
		glog.V(5).Infoln("listen to:", addr)
	}
	s = try.To1(Server(cfg))
	try.To(cfg.Register(s))

	return s, lis, nil
}

func loadTLSCredentials(pw *PKI) (creds credentials.TransportCredentials, err error) {
	defer err2.Handle(&err)

	caCert := try.To1(os.ReadFile(pw.Client.CertFile))
	rootCAs := x509.NewCertPool()
	rootCAs.AppendCertsFromPEM(caCert)

	// Load server's certificate and private key
	serverCert := try.To1(tls.LoadX509KeyPair(pw.Server.CertFile, pw.Server.KeyFile))

	// Create the credentials and return it
	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    rootCAs,
	}

	glog.V(3).Infoln("cert files loaded")
	return credentials.NewTLS(config), nil
}
