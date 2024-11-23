package main

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"codeberg.org/shinyzero0/vtb-api-2024-grpc/utils"

	"fmt"
	"net"

	proto "codeberg.org/shinyzero0/vtb-api-2024-grpc/generated-proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/peer"
)

type Storage interface {
	AppendMessage(ctx context.Context, m msg, sid string) error
	// GetHistory(ctx context.Context, until int64) []*models.Message
}

func main() {
	fmt.Println(f())
}

type Chat struct {
	Clients     map[Client]struct{}
	clients_mtx sync.RWMutex
	strg        Storage
}
type Client struct {
	ch (chan *proto.StreamResponse)
	id string
	// srv grpc.BidiStreamingServer[proto.StreamRequest, proto.StreamResponse]
}

func (c *Client) HandleMessages(str grpc.ServerStreamingServer[proto.StreamResponse]) {
	for msg := range c.ch {
		if err := str.Send(msg); err != nil {
			log.Println(err)
		}
		fmt.Printf("sent %s to %s\n", msg, c.id)
	}
}
func (c *Client) SendMessage(req msg, sid string) {
	c.ch <- &proto.StreamResponse{
		Message:  req.Message,
		SenderId: sid,
	}
}
func (c *Chat) SubscribeClient(cid string) Client {
	c.clients_mtx.Lock()
	defer c.clients_mtx.Unlock()
	cli := Client{
		ch: make(chan *proto.StreamResponse),
		id: cid,
	}
	c.Clients[cli] = struct{}{}
	fmt.Printf("client %s connected\n", cid)
	return cli
}
func (c *Chat) UnsubscribeClient(cli Client) {
	c.clients_mtx.Lock()
	defer c.clients_mtx.Unlock()
	delete(c.Clients, cli)
	fmt.Printf("client %s disconnected\n", cli.id)
	close(cli.ch)
}

type msg struct {
	Message   string
	timestamp time.Time
}
type msgxuser struct {
	msg
	CN string
	ID int64
}

func (c *Chat) SendMessage(req msg, sid string) error {
	c.clients_mtx.RLock()
	defer c.clients_mtx.RUnlock()
	for cli := range c.Clients {
		cli.SendMessage(req, sid)
	}
	return c.strg.AppendMessage(context.Background(), req, sid)
}

func f() error {
	certfile, err1 := utils.GetEnv("CERTFILE")
	keyfile, err2 := utils.GetEnv("KEYFILE")
	cafile, err3 := utils.GetEnv("CAFILE")
	dsn, err4 := utils.GetEnv("STORAGE_DSN")
	if err := errors.Join(err1, err2, err3, err4); err != nil {
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
	strg, err := NewStorage(dsn)
	if err != nil {
		return err
	}
	s := &server{
		chat: Chat{
			Clients:     make(map[Client]struct{}),
			clients_mtx: sync.RWMutex{},
			strg:        strg,
		},
		jwtSecret: jwtSecret,
	}
	srv := grpc.NewServer(
		grpc.StreamInterceptor(s.MiddlewareHandler),
		grpc.UnaryInterceptor(s.UnaryMiddlewareHandler),
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

// History implements generated_proto.ChatServer.
func (s *server) History(req *proto.HistoryRequest, ss grpc.ServerStreamingServer[proto.StreamResponse]) error {
	msgs, err := s.chat.strg.GetHistory(ss.Context(), req.UntilTimestamp, req.Amount)
	if err != nil {
		return err
	}
	for _, m := range msgs {
		if err := ss.Send(&proto.StreamResponse{
			Message:   m.Message,
			MessageId: m.ID,
			SenderId:  m.CN,
			Timestamp: m.timestamp.Unix(),
		}); err != nil {
			fmt.Println(err)
		}
	}
	return nil
}

// SendSingle implements generated_proto.ChatServer.
func (s *server) SendSingle(ctx context.Context, req *proto.SendRequest) (*proto.SendResponse, error) {
	// fmt.Printf("req: %#v\n", req)
	cn, ok := ctx.Value("cn").(string)
	if !ok {
		return nil, fmt.Errorf("wtf? no CN in context?")
	}
	// fmt.Println("before")
	if err := s.chat.SendMessage(msg{Message: req.GetMessage(), timestamp: time.Now()}, cn); err != nil {
		return nil, err
	}
	// fmt.Println("after")
	return &proto.SendResponse{}, nil
}

// mustEmbedUnimplementedChatServer implements generated_proto.ChatServer.

// Stream implements generated_proto.ChatServer.
func (s *server) Stream(req *proto.StreamRequest, str grpc.ServerStreamingServer[proto.StreamResponse]) error {
	cli, ok := str.Context().Value("cli").(Client)
	if !ok {
		return fmt.Errorf("wtf? no client in context?")
	}
	cli.HandleMessages(str)
	// LOOP:
	// 	for {
	// 		select {
	// 		case <-str.Context().Done():
	// 			fmt.Println("done")
	// 			break LOOP
	// 		default:
	// 		}
	// 	}
	return str.Context().Err()
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Override the Context method to return the custom context
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
func CNFromContext(ctx context.Context) (string, error) {
	var cn string
	if p, ok := peer.FromContext(ctx); ok {
		if mtls, ok := p.AuthInfo.(credentials.TLSInfo); ok {
			certs := mtls.State.PeerCertificates
			if len(certs) > 0 {
				cn = certs[0].Subject.CommonName
			} else {
				return cn, fmt.Errorf("ugh")
			}

			// for _, item := range mtls.State.PeerCertificates {
			// 	clis = item.Subject.CommonName
			// 	fmt.Println(clis)
			// }
		} else {
			return cn, fmt.Errorf("crap")
		}
	} else {
		return cn, fmt.Errorf("fuck")
	}
	return cn, nil
}
func (s *server) MiddlewareHandler(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {

	cn, err := CNFromContext(ss.Context()) // ctx, cancel := context.WithCancel(ss.Context())
	if err != nil {
		return err
	}
	// go func(c <-chan time.Time) {
	// 	<-c
	// 	cancel()
	// 	return
	// }(time.NewTimer(exp.Time.Sub(time.Now())).C)
	cli := s.chat.SubscribeClient(cn)
	defer s.chat.UnsubscribeClient(cli)
	ctx := context.WithValue(ss.Context(), "cli", cli)
	newss := &wrappedServerStream{
		ServerStream: ss,
		ctx:          ctx,
	}

	return logerr(handler(srv, newss))
}
func (s *server) UnaryMiddlewareHandler(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	// you can write your own code here to check client tls certificate
	cn, err := CNFromContext(ctx)
	if err != nil {
		return nil, err
	}
	// cli := s.chat.ConnectClient(cn)
	// defer s.chat.DisconnectClient(cli)
	return handler(context.WithValue(ctx, "cn", cn), req)
}

func logerr(erri error) error {
	// fmt.Println(erri)
	return erri
}

// func parseInt64(s string) (int64, error) {
// 	return strconv.ParseInt(s, 10, 64)
// }

// func mapTlsSubject(sub string) func(x509.Certificate) string {
// 	return map[string]func(x509.Certificate) string{
// 		"tls_client_auth_subject_dn": func(c x509.Certificate) string { return c.Subject.String() },
// 		"tls_client_auth_san_dns":    func(c x509.Certificate) string { return c.DNSNames[0] },
// 		"tls_client_auth_san_uri":    func(c x509.Certificate) string { return c.URIs[0].String() },
// 		"tls_client_auth_san_ip":     func(c x509.Certificate) string { return string(c.IPAddresses[0]) },
// 		"tls_client_auth_san_email":  func(c x509.Certificate) string { return c.EmailAddresses[0] },
// 	}[sub]
// }
