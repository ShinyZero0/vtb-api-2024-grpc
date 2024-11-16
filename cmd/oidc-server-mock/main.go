package main

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"time"

	"codeberg.org/shinyzero0/vtb-api-2024-grpc/utils"
	"github.com/oauth2-proxy/mockoidc"
)

func main() {
	if err := fTLS(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func fTLS() error {
	time.Sleep(10 * time.Second)
	certfile, err1 := utils.GetEnv("CERTFILE")
	keyfile, err2 := utils.GetEnv("KEYFILE")
	cafile, err3 := utils.GetEnv("CAFILE")
	srvAddr, err4 := utils.GetEnv("LISTEN_ADDR")
	if err := errors.Join(err1, err2, err3, err4); err1 != nil {
		return err
	}
	return serveTLS(certfile, keyfile, cafile, srvAddr)
}
func f() error {
	srvAddr, err := utils.GetEnv("LISTEN_ADDR")
	if err != nil {
		return err
	}
	return serve(srvAddr)
}
func serve(srvAddr string) error {
	return withServer(srvAddr, nil, queueUsers)
}
func withServer(srvAddr string, tlsc *tls.Config, cb func(m *mockoidc.MockOIDC) error) error {
	m, err := mockoidc.NewServer(nil)
	if err != nil {
		return err
	}
	ln, err := net.Listen("tcp", srvAddr)
	if err != nil {
		return err
	}
	if err := m.Start(ln, tlsc); err != nil {
		return err
	}
	if err := cb(m); err != nil {
		return err
	}
	fmt.Println(m.DiscoveryEndpoint())
	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt)
	<-sigch
	return m.Shutdown()

}
func serveTLS(certFile, keyFile, caFile, srvAddr string) error {
	tlsc, err := utils.LoadTLSConfig(certFile, keyFile, caFile)
	if err != nil {
		return err
	}

	return withServer(srvAddr, tlsc, queueUsers)
}
func queueUsers(m *mockoidc.MockOIDC) error {
	for i := 0; i < 500; i++ {
		m.QueueUser(&mockoidc.MockUser{
			Subject:           fmt.Sprintf("user-%d", i),
			Email:             fmt.Sprintf("user-%d@example.com", i),
			EmailVerified:     true,
			PreferredUsername: fmt.Sprintf("user-%d", i),
		})
	}
	return nil
}
