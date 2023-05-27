package http

import (
	"context"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang/glog"
	"github.com/lainio/err2"
	"github.com/lainio/err2/try"
)

var (
	server *nethttp.Server
)

func Run(s *nethttp.Server) <-chan os.Signal {
	server = s
	if s == nil {
		server = &nethttp.Server{
			ReadTimeout:       20 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
		}
	}
	go func() {
		defer err2.Catch(func(err error) {
			glog.Error(err)
		})

		if try.Is(server.ListenAndServe(), nethttp.ErrServerClosed) {
			glog.Infoln("Stopped serving new connections.")
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	return sigChan
}

func GracefulStop() {
	defer err2.Catch()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	try.To(server.Shutdown(ctx))
}
