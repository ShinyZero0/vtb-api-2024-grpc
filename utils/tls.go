package utils

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"google.golang.org/grpc/credentials"
)

func LoadTlSTransport(certFile, keyFile, caFile string) (credentials.TransportCredentials, error) {
	tlsConfig, err := LoadTLSConfig(certFile, keyFile, caFile)
	if err != nil {
		return nil, err
	}
	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	return credentials.NewTLS(tlsConfig), nil
}
func LoadTLSConfig(certFile, keyFile, caFile string) (*tls.Config, error) {
	certificate, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certification: %w", err)
	}

	data, err := os.ReadFile(caFile)
	if err != nil {
		return nil, fmt.Errorf("faild to read CA certificate: %w", err)
	}

	capool := x509.NewCertPool()
	if !capool.AppendCertsFromPEM(data) {
		return nil, fmt.Errorf("unable to append the CA certificate to CA pool")
	}

	return &tls.Config{
		// ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{certificate},
		RootCAs:      capool,
		ClientCAs:    capool,
		MinVersion:   tls.VersionTLS13,
	}, nil
}
