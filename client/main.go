package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	ops "github.com/findy-network/findy-common-go/grpc/ops/v1"
	"github.com/findy-network/findy-common-go/jwt"
	"github.com/findy-network/findy-common-go/rpc"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	_ "github.com/lainio/err2/assert" // we want an --asserter flag
	"github.com/lainio/err2/try"
	"google.golang.org/grpc"
)

var (
	cert       = flag.String("cert", "../cert", "pki cert path")
	user       = flag.String("user", "", "test user name")
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

	conn := try.To1(newClient(*user, fmt.Sprintf("%s:%d", *serverAddr, *port)))
	defer conn.Close()

	glog.V(1).Info("connection created")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	c := ops.NewDevOpsServiceClient(conn)
	r := try.To1(c.Enter(ctx, &ops.Cmd{
		Type: ops.Cmd_PING,
	}))
	fmt.Println("result:", r.GetPing())
	defer cancel()

}

func newClient(user, addr string) (conn *grpc.ClientConn, err error) {
	defer err2.Handle(&err)

	var pki *rpc.PKI
	jwtStr := ""
	if !*noTLS {
		pki = rpc.LoadPKIWithServerName(*cert, addr)
		jwtStr = jwt.BuildJWT(user)
	}
	glog.V(5).Infoln("client with user:", user)
	conn = try.To1(rpc.ClientConn(rpc.ClientCfg{
		PKI:  pki,
		JWT:  jwtStr,
		Addr: addr,

		Insecure: *noTLS,
	}))
	return
}
