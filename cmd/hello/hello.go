package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/c4milo/handlers/grpcutil"
	"github.com/c4milo/handlers/logger"
	"github.com/c4milo/hello-nyt"
	"github.com/c4milo/hello-nyt/config"
	"github.com/c4milo/hello-nyt/static"
	"github.com/golang/glog"
)

var (
	// AppName is defined during compilation
	AppName string
	// Version is defined during compilation
	Version string
)

func init() {
	flag.Set("logtostderr", "true")
	flag.Parse()
	config.Read()
}

func main() {
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	appName := AppName + "-" + Version
	defer glog.Flush()

	// GRPC services
	services := []grpcutil.ServiceRegisterFn{
		hello.RegisterService,
	}

	tlsKeyPair, err := tls.X509KeyPair([]byte(config.TLSCert), []byte(config.TLSKey))
	if err != nil {
		glog.Fatalf("failed loading TLS certificate and key: %+v", err)
	}

	options := []grpcutil.Option{
		grpcutil.WithTLSCert(&tlsKeyPair),
		grpcutil.WithPort(config.Port),
		grpcutil.WithServices(services),
		// We could extend more the hello gRPC service by adding token verification
		// using gRPC interceptors.
		// grpcutil.WithServerOpts([]grpc.ServerOption{
		// 	grpc.UnaryInterceptor(identity.UnaryInterceptor()),
		// }),

		// We want the OpenAPI spec to be served by our static handler
		grpcutil.WithSkipPath(fmt.Sprintf("/lib/%s.swagger.json", AppName)),
	}

	// These middlewares are invoked bottom up and the order matters.
	// Serves single page app and web static assets such as images, stylecheets and fonts.
	handler := static.Handler(http.DefaultServeMux)
	// Handles gRPC and OpenAPI requests
	handler = grpcutil.Handler(handler, options...)
	// Handles logging for OpenAPI requests as well as static assets requests
	handler = logger.Handler(handler, logger.AppName(appName))

	address := ":" + config.Port
	srv := &http.Server{
		Addr:    address,
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsKeyPair},
			NextProtos:   []string{"h2"}, // Makes sure it upgrades the connection to HTTP2
		},
	}

	go func() {
		glog.Infof("Starting server at %s", address)
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			glog.Errorf("ListenAndServeTLS: %v", err)
		}
	}()
	<-stopChan // wait for SIGINT

	// shut down gracefully, but waits no longer than 5 seconds before shutting down the service.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	glog.Info("Server gracefully stopped")
}
