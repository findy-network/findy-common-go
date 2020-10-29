package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path"
	"time"

	agency "github.com/findy-network/findy-agent-api/grpc/new"
	"github.com/findy-network/findy-grpc/jwt"
	"github.com/findy-network/findy-grpc/rpc"
	"github.com/golang/glog"
	"github.com/lainio/err2"
	"google.golang.org/grpc"
)

var useTls = true

func main() {
	defer err2.Catch(func(err error) {
		glog.Error(err)
	})
	flag.Parse()

	conn, err := NewClient("findy-root", "localhost:50051")
	err2.Check(err)
	defer conn.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)

	c := agency.NewDevOpsClient(conn)
	r, err := c.Enter(ctx, &agency.Cmd{
		Type:    agency.Cmd_PING,
	})
	err2.Check(err)
	fmt.Println("result:", r.GetPing())
	defer cancel()

}

var Conn  *grpc.ClientConn

func NewClient(user, addr string) (conn *grpc.ClientConn, err error) {
	defer err2.Return(&err)

	goPath := os.Getenv("GOPATH")
	tlsPath := path.Join(goPath, "src/github.com/findy-network/findy-grpc/tls")
	certFile := path.Join(tlsPath, "ca.crt")

	glog.V(5).Infoln("client with user:", user)
	conn, err = rpc.ClientConn(rpc.ClientCfg{
		CertFile: certFile,
		JWT:      jwt.BuildJWT(user),
		Addr:     addr,
		TLS:      useTls,
	})
	err2.Check(err)
	Conn = conn
	return
}
