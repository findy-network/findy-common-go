package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"

	ops "github.com/findy-network/findy-common-go/grpc/ops/v1"
	"github.com/findy-network/findy-common-go/jwt"
	"github.com/findy-network/findy-common-go/rpc"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	_ "github.com/lainio/err2/assert" // we want an --asserter flag
	"google.golang.org/grpc"
)

var (
	user       = flag.String("user", "findy-root", "test user name")
	serverAddr = flag.String("addr", "localhost", "agency host gRPC address")
	port       = flag.Int("port", 50051, "agency host gRPC port")
	noTLS      = flag.Bool("no-tls", false, "do NOT use TLS and cert files (hard coded)")
)

func main() {
	os.Args = append(os.Args,
		"-logtostderr",
	)
	glog.CopyStandardLogTo("ERROR") // for err2 binging

	defer err2.Catch(err2.Stderr)

	flag.Parse()

	var pki *rpc.PKI
	if !*noTLS {
		pki = rpc.LoadPKI("./cert")
		glog.V(3).Infof("starting gRPC server with\ncrt:\t%s\nkey:\t%s\nclient:\t%s",
			pki.Server.CertFile, pki.Server.KeyFile, pki.Client.CertFile)
	}
	rpc.Serve(&rpc.ServerCfg{
		NoAuthorization: *noTLS,

		Port: *port,
		PKI:  pki,
		Register: func(s *grpc.Server) error {
			ops.RegisterDevOpsServiceServer(s, &devOpsServer{Root: *user})
			glog.V(10).Infoln("GRPC registration all done")
			return nil
		},
	})
}

type devOpsServer struct {
	ops.UnimplementedDevOpsServiceServer
	Root string
}

func (d devOpsServer) Enter(ctx context.Context, cmd *ops.Cmd) (cr *ops.CmdReturn, err error) {
	defer err2.Handle(&err)

	glog.V(1).Info("enter Enter()")
	if !*noTLS {
		user := jwt.User(ctx)

		if user != d.Root {
			return &ops.CmdReturn{Type: cmd.Type}, errors.New("access right")
		}
	}

	glog.V(3).Infoln("dev ops cmd", cmd.Type)
	cmdReturn := &ops.CmdReturn{Type: cmd.Type}

	switch cmd.Type {
	case ops.Cmd_PING:
		response := fmt.Sprintf("%s, ping ok", "TEST")
		cmdReturn.Response = &ops.CmdReturn_Ping{Ping: response}
	case ops.Cmd_LOGGING:
		//agencyCmd.ParseLoggingArgs(cmd.GetLogging())
		//response = fmt.Sprintf("logging = %s", cmd.GetLogging())
	case ops.Cmd_COUNT:
		response := fmt.Sprintf("%d/%d cloud agents",
			100, 1000)
		cmdReturn.Response = &ops.CmdReturn_Ping{Ping: response}
	}
	return cmdReturn, nil
}
