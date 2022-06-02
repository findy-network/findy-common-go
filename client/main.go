package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	ops "github.com/findy-network/findy-common-go/grpc/ops/v1"
	"github.com/findy-network/findy-common-go/jwt"
	"github.com/findy-network/findy-common-go/rpc"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
	"google.golang.org/grpc"
)

var (
	user       = flag.String("user", "findy-root", "test user name")
	serverAddr = flag.String("addr", "localhost", "agency host gRPC address")
	port       = flag.Int("port", 50051, "agency host gRPC port")
)

func main() {
	defer err2.Catch(func(err error) {
		glog.Error(err)
	})
	flag.Parse()

	// we want this for glog, this is just a tester, not a real world service
	try.To(flag.Set("logtostderr", "true"))

	conn := try.To1(newClient(*user, fmt.Sprintf("%s:%d", *serverAddr, *port)))
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	c := ops.NewDevOpsServiceClient(conn)
	r := try.To1(c.Enter(ctx, &ops.Cmd{
		Type: ops.Cmd_PING,
	}))
	fmt.Println("result:", r.GetPing())
	defer cancel()

}

func newClient(user, addr string) (conn *grpc.ClientConn, err error) {
	defer err2.Return(&err)

	pki := rpc.LoadPKIWithServerName("../cert", addr)
	glog.V(5).Infoln("client with user:", user)
	conn = try.To1(rpc.ClientConn(rpc.ClientCfg{
		PKI:  pki,
		JWT:  jwt.BuildJWT(user),
		Addr: addr,
	}))
	return
}
