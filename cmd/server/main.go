package main

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"sync"

	"codeberg.org/shinyzero0/vtb-api-2024-grpc/utils"
	"github.com/golang-jwt/jwt/v5"

	"fmt"
	"net"

	proto "codeberg.org/shinyzero0/vtb-api-2024-grpc/generated-proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func main() {
	fmt.Println(f())
}

type Chat struct {
	Clients     map[Client]struct{}
	clients_mtx sync.RWMutex
}
type Client struct {
	ch  (chan *proto.StreamResponse)
	id  int64
	srv grpc.BidiStreamingServer[proto.StreamRequest, proto.StreamResponse]
}

func (c *Client) HandleMessages() {
	for msg := range c.ch {
		if err := c.srv.Send(msg); err != nil {
			log.Println(err)
		}
	}
}
func (c *Client) SendMessage(req *proto.StreamRequest) {
	c.ch <- &proto.StreamResponse{
		Message:  req.GetMessage(),
		SenderId: 0,
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
func (c *Chat) SendMessage(req *proto.StreamRequest) {
	c.clients_mtx.RLock()
	for cli := range c.Clients {
		cli.SendMessage(req)
	}
	defer c.clients_mtx.RUnlock()
}

func f() error {
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
	tlsconf, err := utils.LoadTlSTransport("server.pem", "server-key.pem", "root.pem")
	if err != nil {
		return err
	}
	srv := grpc.NewServer(
		grpc.StreamInterceptor(MiddlewareHandler),
		grpc.Creds(tlsconf),
		// grpc.Creds(tlsconf),
	)

	proto.RegisterChatServer(srv, &server{
		chat:      Chat{},
		jwtSecret: jwtSecret,
	})
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
	go cli.HandleMessages()
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
			s.chat.SendMessage(req)
		}
	}
	return nil
}
func MiddlewareHandler(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
	// you can write your own code here to check client tls certificate
	// if p, ok := peer.FromContext(ss.Context()); ok {
	// 	if mtls, ok := p.AuthInfo.(credentials.TLSInfo); ok {
	// 		for _, item := range mtls.State.PeerCertificates {
	// 			log.Println("client certificate subject:", item.Subject.String())
	// 		}
	// 	}
	// }
	md, ok := metadata.FromIncomingContext(ss.Context())
	if !ok {
		panic("fuck")
	}
	token := strings.TrimPrefix(md.Get("authorization")[0], "Bearer ")
	s := srv.(*server)
	tok, err := jwt.Parse(
		token,
		func(t *jwt.Token) (interface{}, error) { return []byte(s.jwtSecret), nil },
	)
	exp, err := tok.Claims.GetExpirationTime()
	// ctx, cancel := context.WithCancel(ss.Context())
	// go func(c <-chan time.Time) {
	// 	<-c
	// 	cancel()
	// 	return
	// }(time.NewTimer(exp.Time.Sub(time.Now())).C)
	ctx, cancel := context.WithDeadline(ss.Context(), exp.Time)
	defer cancel()
	sub_, err := tok.Claims.GetSubject()
	if err != nil {
		return err
	}
	sub, err := parseInt64(sub_)
	cli := s.chat.ConnectClient(sub)
	defer s.chat.DisconnectClient(cli)
	ctx = context.WithValue(ctx, "cli", cli)

	return handler(ctx, ss)
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
