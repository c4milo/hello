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

	_ "expvar"
)

var (
	// AppName is defined during compilation
	AppName string
	// Version is defined during compilation
	Version string
)

func init() {
	// Sets glog flag to log messages to stderr as well.
	flag.Set("logtostderr", "true")

	// Parses any flags passed to the command line
	flag.Parse()

	// Reads environment variables configuring the service.
	config.Read()
}

func main() {
	// Listens for SIGINT signals in order to gracefully shutdown the service.
	stopChan := make(chan os.Signal)
	signal.Notify(stopChan, os.Interrupt)

	appName := AppName + "-" + Version

	// Makes sure to flush any pending IO before shutting down service.
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
		grpcutil.WithPort(config.TLSPort),
		grpcutil.WithServices(services),
		// We could extend more the hello gRPC service by adding token verification
		// using gRPC interceptors. We could also add more vars to expvar in order
		// to gather stats about the gRPC calls.
		// grpcutil.WithServerOpts([]grpc.ServerOption{
		// 	grpc.UnaryInterceptor(identity.UnaryInterceptor()),
		// }),

		// We want the OpenAPI spec to be served by our static handler
		grpcutil.WithSkipPath(fmt.Sprintf("/lib/%s.swagger.json", AppName)),
	}

	// The following middlewares are invoked bottom up and the order matters.

	// Serves single page app and web static assets such as images, stylecheets and fonts.
	handler := static.Handler(http.DefaultServeMux)
	// Handles gRPC and OpenAPI requests
	handler = grpcutil.Handler(handler, options...)
	// Handles health requests
	handler = func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/health" {
				http.Redirect(w, r, "/debug/vars", http.StatusTemporaryRedirect)
				return
			}
			h.ServeHTTP(w, r)
		})
	}(handler)
	// Handles logging for OpenAPI requests as well as static assets requests
	handler = logger.Handler(handler, logger.AppName(appName))

	tlsAddress := ":" + config.TLSPort
	address := ":" + config.Port
	srv := &http.Server{
		Addr:    tlsAddress,
		Handler: handler,
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{tlsKeyPair},
			NextProtos:   []string{"h2"}, // Makes sure it upgrades the connection to HTTP2
		},
	}

	go func() {
		glog.Infof("Starting HTTPS server at %s", tlsAddress)
		if err := srv.ListenAndServeTLS("", ""); err != nil {
			glog.Errorf("ListenAndServeTLS: %v", err)
		}
	}()

	go func() {
		glog.Infof("Starting HTTP server at %s", address)
		if err := http.ListenAndServe(address, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			url := "https://" + config.PrimaryDomain + r.RequestURI
			glog.Infof("redirecting to %s", url)
			http.Redirect(w, r, url, http.StatusPermanentRedirect)
		})); err != nil {
			glog.Errorf("ListenAndServe: %v", err)
		}
	}()
	<-stopChan // wait for SIGINT signal

	// shut down gracefully, waiting 5 seconds for any established connection to finish before shutting down the service.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	srv.Shutdown(ctx)

	glog.Info("Server gracefully stopped")
}
