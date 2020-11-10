package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"github.com/findy-network/findy-agent-api/grpc/ops"
	"github.com/findy-network/findy-grpc/jwt"
	"github.com/findy-network/findy-grpc/rpc"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"google.golang.org/grpc"
)

var (
	user = flag.String("user", "findy-root", "test user name")
)

func main() {
	defer err2.Catch(func(err error) {
		glog.Error(err)
	})
	flag.Parse()

	// whe want this for glog, this is just a tester, not a real world service
	err2.Check(flag.Set("logtostderr", "true"))

	conn, err := newClient(*user, "localhost:50051")
	err2.Check(err)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	c := ops.NewDevOpsClient(conn)
	r, err := c.Enter(ctx, &ops.Cmd{
		Type: ops.Cmd_PING,
	})
	err2.Check(err)
	fmt.Println("result:", r.GetPing())
	defer cancel()

}

func newClient(user, addr string) (conn *grpc.ClientConn, err error) {
	defer err2.Return(&err)

	pki := rpc.LoadPKI("../cert")
	glog.V(5).Infoln("client with user:", user)
	conn, err = rpc.ClientConn(rpc.ClientCfg{
		PKI:  pki,
		JWT:  jwt.BuildJWT(user),
		Addr: addr,
	})
	err2.Check(err)
	return
}
