package main

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"log"
	"strconv"
	"sync"

	"codeberg.org/shinyzero0/vtb-api-2024-grpc/utils"

	"fmt"
	"net"

	proto "codeberg.org/shinyzero0/vtb-api-2024-grpc/generated-proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

func main() {
	fmt.Println(f())
}

type Chat struct {
	Clients     map[Client]struct{}
	clients_mtx sync.RWMutex
}
type Client struct {
	ch (chan *proto.StreamResponse)
	id int64
	// srv grpc.BidiStreamingServer[proto.StreamRequest, proto.StreamResponse]
}

func (c *Client) HandleMessages(bidi grpc.BidiStreamingServer[proto.StreamRequest, proto.StreamResponse]) {
	for msg := range c.ch {
		if err := bidi.Send(msg); err != nil {
			log.Println(err)
		}
	}
}
func (c *Client) SendMessage(req *proto.StreamRequest, sid int64) {
	c.ch <- &proto.StreamResponse{
		Message:  req.GetMessage(),
		SenderId: sid,
	}
}
func (c *Chat) ConnectClient(cid int64) Client {
	c.clients_mtx.Lock()
	defer c.clients_mtx.Unlock()
	cli := Client{
		ch: make(chan *proto.StreamResponse),
		id: cid,
	}
	c.Clients[cli] = struct{}{}
	return cli
}
func (c *Chat) DisconnectClient(cli Client) {
	c.clients_mtx.Lock()
	defer c.clients_mtx.Unlock()
	delete(c.Clients, cli)
	close(cli.ch)
}
func (c *Chat) SendMessage(req *proto.StreamRequest, sid int64) {
	c.clients_mtx.RLock()
	defer c.clients_mtx.RUnlock()
	for cli := range c.Clients {
		cli.SendMessage(req, sid)
	}
}

func f() error {
	certfile, err1 := utils.GetEnv("CERTFILE")
	keyfile, err2 := utils.GetEnv("KEYFILE")
	cafile, err3 := utils.GetEnv("CAFILE")
	if err := errors.Join(err1, err2, err3); err != nil {
		return err
	}
	lisAddr, err := utils.GetEnv("LISTEN_ADDR")
	if err != nil {
		return err
	}
	jwtSecret, err := utils.GetEnv("JWT_SECRET")
	if err != nil {
		return err
	}
	lis, err := net.Listen("tcp", lisAddr)
	if err != nil {
		return err
	}
	tlsconf, err := utils.LoadTlSTransport(certfile, keyfile, cafile)
	if err != nil {
		return err
	}
	s := &server{
		UnimplementedChatServer: proto.UnimplementedChatServer{},
		chat: Chat{
			Clients:     make(map[Client]struct{}),
			clients_mtx: sync.RWMutex{},
		},
		jwtSecret: jwtSecret,
	}
	srv := grpc.NewServer(
		grpc.StreamInterceptor(s.MiddlewareHandler),
		grpc.Creds(tlsconf),
		// grpc.Creds(tlsconf),
	)

	proto.RegisterChatServer(srv, s)
	return srv.Serve(lis)
}

type server struct {
	proto.UnimplementedChatServer
	chat      Chat
	jwtSecret string
}

// Stream implements generated_proto.ChatServer.
func (s *server) Stream(bidi grpc.BidiStreamingServer[proto.StreamRequest, proto.StreamResponse]) error {
	cli, ok := bidi.Context().Value("cli").(Client)
	if !ok {
		panic("fuck")
	}
	go cli.HandleMessages(bidi)
LOOP:
	for {
		select {
		case <-bidi.Context().Done():
			break LOOP
		default:
			req, err := bidi.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					continue LOOP
				}
				return err
			}
			s.chat.SendMessage(req, cli.id)
		}
	}
	return nil
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Override the Context method to return the custom context
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
func (s *server) MiddlewareHandler(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	var clis string
	if p, ok := peer.FromContext(ss.Context()); ok {
		if mtls, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			for _, item := range mtls.State.PeerCertificates {
				clis = item.Subject.CommonName
			}
		}
	}
	// ctx, cancel := context.WithCancel(ss.Context())
	// go func(c <-chan time.Time) {
	// 	<-c
	// 	cancel()
	// 	return
	// }(time.NewTimer(exp.Time.Sub(time.Now())).C)
	sub, err := parseInt64(clis)
	cli := s.chat.ConnectClient(sub)
	defer s.chat.DisconnectClient(cli)
	ctx := context.WithValue(ss.Context(), "cli", cli)
	newss := &wrappedServerStream{
		ServerStream: ss,
		ctx:          ctx,
	}

	return handler(srv, newss)
}
func parseInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func mapTlsSubject(sub string) func(x509.Certificate) string {
	return map[string]func(x509.Certificate) string{
		"tls_client_auth_subject_dn": func(c x509.Certificate) string { return c.Subject.String() },
		"tls_client_auth_san_dns":    func(c x509.Certificate) string { return c.DNSNames[0] },
		"tls_client_auth_san_uri":    func(c x509.Certificate) string { return c.URIs[0].String() },
		"tls_client_auth_san_ip":     func(c x509.Certificate) string { return string(c.IPAddresses[0]) },
		"tls_client_auth_san_email":  func(c x509.Certificate) string { return c.EmailAddresses[0] },
	}[sub]
}
