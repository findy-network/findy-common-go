package rpc

import (
	"fmt"
	"net"

	"github.com/findy-network/findy-grpc/jwt"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

// ServerCfg is configuration to setup a gRPC server.
type ServerCfg struct {
	Port     int
	TLS      bool
	CertFile string
	KeyFile  string
	Register func(s *grpc.Server) error
}

// Server creates a gRPC server with TLS and JWT token authorization.
func Server(cfg ServerCfg) (s *grpc.Server, err error) {
	defer err2.Return(&err)

	opts := make([]grpc.ServerOption, 0, 4)
	if cfg.TLS {
		creds, err := credentials.NewServerTLSFromFile(
			cfg.CertFile, cfg.KeyFile)
		err2.Check(err)

		opts = append(opts,
			grpc.Creds(creds),
			grpc.UnaryInterceptor(jwt.EnsureValidToken),
			grpc.StreamInterceptor(jwt.EnsureValidTokenStream),
		)
	}
	return grpc.NewServer(opts...), nil
}

// Serve builds up the gRPC server and starts to serve. This function blocks.
// In most cases you should start it as goroutine. TODO: graceful stop!
func Serve(cfg ServerCfg) {
	defer err2.Catch(func(err error) {
		glog.Error(err)
	})

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.Port))
	err2.Check(err)
	s, err := Server(cfg)
	err2.Check(err)
	err2.Check(cfg.Register(s))
	err2.Check(s.Serve(lis))
}
