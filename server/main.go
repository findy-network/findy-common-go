package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"

	agency "github.com/findy-network/findy-agent-api/grpc/new"
	"github.com/findy-network/findy-grpc/rpc"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"google.golang.org/grpc"
)

var useTls = true

func main() {
	flag.Parse()

	Serve()
}


func Serve() {
	goPath := os.Getenv("GOPATH")
	tlsPath := path.Join(goPath, "src/github.com/findy-network/findy-grpc/tls")
	certFile := path.Join(tlsPath, "server.crt")
	keyFile := path.Join(tlsPath, "server.pem")

	glog.V(1).Infoln("starting gRPC server with tls path:", tlsPath)

	rpc.Serve(rpc.ServerCfg{
		Port:     50051,
		TLS:      useTls,
		CertFile: certFile,
		KeyFile:  keyFile,
		Register: func(s *grpc.Server) error {
			agency.RegisterDevOpsServer(s, &devOpsServer{Root:"findy-root"})
			glog.Infoln("GRPC registration IIIIIII OK")
			return nil
		},
	})
}


type devOpsServer struct {
	agency.UnimplementedDevOpsServer
	Root string
}

func (d devOpsServer) Enter(ctx context.Context, cmd *agency.Cmd) (cr *agency.CmdReturn, err error) {
	defer err2.Return(&err)

//	user := jwt.User(ctx)

	//if user != d.Root {
	//	return &agency.CmdReturn{Type: cmd.Type}, errors.New("access right")
	//}

	glog.V(3).Infoln("dev ops cmd", cmd.Type)
	cmdReturn := &agency.CmdReturn{Type: cmd.Type}

	switch cmd.Type {
	case agency.Cmd_PING:
		response := fmt.Sprintf("%s, ping ok", "TEST")
		cmdReturn.Response = &agency.CmdReturn_Ping{Ping: response}
	case agency.Cmd_LOGGING:
		//agencyCmd.ParseLoggingArgs(cmd.GetLogging())
		//response = fmt.Sprintf("logging = %s", cmd.GetLogging())
	case agency.Cmd_COUNT:
		response := fmt.Sprintf("%d/%d cloud agents",
			100, 1000)
		cmdReturn.Response = &agency.CmdReturn_Ping{Ping: response}
	}
	return cmdReturn, nil
}
