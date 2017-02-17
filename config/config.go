package config

import (
	"os"

	"github.com/golang/glog"
)

var (
	// PrimaryDomain is the main DNS name used to access this service.
	PrimaryDomain string
	// Port is the TCP port on which this service will accept connections.
	Port string
	// TLSCert is the PEM encoded value of the TLS certificate
	TLSCert string
	// TLSKey is the PEM encoded value of the TLS private key used to generate the certificate
	TLSKey string
)

// Read loads the configuration values.
func Read() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "9999"
	}
	Port = port

	PrimaryDomain = os.Getenv("PRIMARY_DOMAIN")
	if PrimaryDomain == "" {
		PrimaryDomain = "localhost:" + Port
	}

	// For development purposes, use the following to regenerate key:
	// openssl ecparam -genkey -name secp384r1 -out cert-key.pem
	TLSKey = os.Getenv("TLS_KEY")
	if TLSKey == "" {
		glog.Fatal("TLS_KEY config variable must be set")
	}

	// For development purposes, use the following command to regenerate cert:
	// openssl req -new -x509 -key cert-key.pem -out cert.pem -days 1920
	TLSCert = os.Getenv("TLS_CERT")
	if TLSCert == "" {
		glog.Fatal("TLS_CERT config variable must be set")
	}
}
