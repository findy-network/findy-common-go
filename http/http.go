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

// Run starts a http(s) server. It returns a channel which is configured to
// listen system process termination signals: INTR, TERM. If server object is
// given as a first argument, it is used. If also certFile and keyFile names are
// given (in this order) the https server is started.
func Run(s *nethttp.Server, a ...string) <-chan os.Signal {
	server = s
	if s == nil {
		server = &nethttp.Server{
			ReadTimeout:       20 * time.Second,
			ReadHeaderTimeout: 10 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
		}
	}
	startHTTPS := len(a) == 2
	glog.V(3).Infof("startHTTPS: %v, length a: %v", startHTTPS, len(a))
	go func() {
		defer err2.Catch(func(err error) {
			glog.Error(err)
		})

		if startHTTPS {
			glog.V(3).Infof("starting https server w/ cert: %s, key: %s",
				a[0], a[1])
			if try.Is(server.ListenAndServeTLS(a[0], a[1]), nethttp.ErrServerClosed) {
				glog.Infoln("Stopped serving new connections.")
			}
		} else {
			if try.Is(server.ListenAndServe(), nethttp.ErrServerClosed) {
				glog.Infoln("Stopped serving new connections.")
			}
		}
	}()

	shutdownCh := make(chan os.Signal, 1)
	signal.Notify(shutdownCh, os.Interrupt, syscall.SIGTERM)

	return shutdownCh
}

// GracefulStop stops the http(s) server gracefully.
func GracefulStop() {
	defer err2.Catch()

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	try.To(server.Shutdown(ctx))
}
