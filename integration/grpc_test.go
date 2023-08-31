package integration

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"testing"
	"time"

	ops "github.com/findy-network/findy-common-go/grpc/ops/v1"
	"github.com/findy-network/findy-common-go/jwt"
	"github.com/findy-network/findy-common-go/rpc"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/assert"
	"github.com/lainio/err2/try"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

const bufSize = 1024 * 1024
const pingReturn = "This is a TEST"

var (
	lis            = bufconn.Listen(bufSize)
	insecureLis    = bufconn.Listen(bufSize)
	server         *grpc.Server
	conn           *grpc.ClientConn
	insecureServer *grpc.Server
	insecureConn   *grpc.ClientConn
	doPanic        = false
	doServer       = &devOpsServer{Root: "findy-root"}
)

func TestMain(m *testing.M) {
	try.To(flag.Set("logtostderr", "true"))
	try.To(flag.Set("v", "0"))
	setUp()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func setUp() {
	err2.SetTracers(os.Stderr)
	defer err2.Catch(err2.Err(func(err error) {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}))

	runServer()
	conn = try.To1(newClient("findy-root", "localhost:50051")) // just dump error info out, we are inside a test

	runInsecureServer()
	insecureConn = try.To1(rpc.ClientConn(rpc.ClientCfg{
		JWT:      jwt.BuildJWT("findy-root"),
		Addr:     "localhost:50052",
		Opts:     []grpc.DialOption{grpc.WithContextDialer(insecureBufDialer)},
		Insecure: true,
	}))

}

func tearDown() {
	err := conn.Close()
	try.To(err) // just dump information out, we are inside a test
	server.GracefulStop()

	err = insecureConn.Close()
	try.To(err) // just dump information out, we are inside a test
	insecureServer.GracefulStop()
}

func TestEnter(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	c := ops.NewDevOpsServiceClient(conn)
	r, err := c.Enter(ctx, &ops.Cmd{
		Type: ops.Cmd_PING,
	})
	assert.NoError(err)
	assert.Equal(pingReturn, r.GetPing())

	doPanic = true
	_, err = c.Enter(ctx, &ops.Cmd{
		Type: ops.Cmd_PING,
	})
	assert.Error(err)
	doPanic = false

	defer cancel()
}

func TestEnterInsecure(t *testing.T) {
	assert.PushTester(t)
	defer assert.PopTester()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	c := ops.NewDevOpsServiceClient(insecureConn)
	r, err := c.Enter(ctx, &ops.Cmd{
		Type: ops.Cmd_PING,
	})
	assert.NoError(err)
	assert.Equal(pingReturn, r.GetPing())

	doPanic = true
	_, err = c.Enter(ctx, &ops.Cmd{
		Type: ops.Cmd_PING,
	})
	assert.Error(err)
	doPanic = false

	defer cancel()
}

func newClient(user, addr string) (conn *grpc.ClientConn, err error) {
	defer err2.Handle(&err)

	pki := rpc.LoadPKI("../cert")

	glog.V(5).Infoln("client with user:", user)
	conn = try.To1(rpc.ClientConn(rpc.ClientCfg{
		PKI:  pki,
		JWT:  jwt.BuildJWT(user),
		Addr: addr,
		Opts: []grpc.DialOption{grpc.WithContextDialer(bufDialer)},
	}))
	return
}

func runServer() {
	pki := rpc.LoadPKI("../cert")
	glog.V(1).Infof("starting gRPC server with\ncrt:\t%s\nkey:\t%s\nclient:\t%s",
		pki.Server.CertFile, pki.Server.KeyFile, pki.Client.CertFile)

	go func() {
		defer err2.Catch(err2.Err(func(err error) {
			log.Fatal(err)
		}))
		s, lis := try.To2(rpc.PrepareServe(&rpc.ServerCfg{
			Port:    50051,
			PKI:     pki,
			TestLis: lis,
			Register: func(s *grpc.Server) error {
				ops.RegisterDevOpsServiceServer(s, doServer)
				glog.V(10).Infoln("GRPC registration all done")
				return nil
			},
		}))
		server = s
		try.To(s.Serve(lis))
	}()
}

func runInsecureServer() {
	go func() {
		defer err2.Catch(err2.Err(func(err error) {
			log.Fatal(err)
		}))
		s, serverLis := try.To2(rpc.PrepareServe(&rpc.ServerCfg{
			Port:    50052,
			TestLis: insecureLis,
			Register: func(s *grpc.Server) error {
				ops.RegisterDevOpsServiceServer(s, doServer)
				glog.V(10).Infoln("GRPC registration all done")
				return nil
			},
		}))
		insecureServer = s
		try.To(s.Serve(serverLis))
	}()
}

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

func insecureBufDialer(context.Context, string) (net.Conn, error) {
	return insecureLis.Dial()
}

type devOpsServer struct {
	ops.UnimplementedDevOpsServiceServer
	Root string
}

func (d devOpsServer) Enter(ctx context.Context, cmd *ops.Cmd) (cr *ops.CmdReturn, err error) {
	defer err2.Handle(&err)

	if doPanic {
		panic("testing panic")
	}

	user := jwt.User(ctx)

	if user != d.Root {
		return &ops.CmdReturn{Type: cmd.Type}, errors.New("access right")
	}

	glog.V(3).Infoln("dev ops cmd", cmd.Type)
	cmdReturn := &ops.CmdReturn{Type: cmd.Type}

	switch cmd.Type {
	case ops.Cmd_PING:
		response := pingReturn
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
