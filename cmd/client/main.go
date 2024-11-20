package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	proto "codeberg.org/shinyzero0/vtb-api-2024-grpc/generated-proto"
	"codeberg.org/shinyzero0/vtb-api-2024-grpc/utils"
	"google.golang.org/grpc"

	"github.com/dgraph-io/badger/v4"
	"github.com/go-loremipsum/loremipsum"
	"golang.org/x/oauth2/dcrp"
)

func main() {
	if err := f(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func f() error {
	srvAddr, err0 := utils.GetEnv("SERVER_ADDR")
	certfile, err1 := utils.GetEnv("CERTFILE")
	keyfile, err2 := utils.GetEnv("KEYFILE")
	cafile, err3 := utils.GetEnv("CAFILE")
	if err := errors.Join(err0, err1, err2, err3); err != nil {
		return err
	}
	tlsc, err := utils.LoadTlSTransport(certfile, keyfile, cafile)
	if err != nil {
		return err
	}
	// authAddr, err := utils.GetEnv("AUTH_ADDR")
	// if err != nil {
	// 	return err
	// }
	// dbPath, err := utils.GetEnv("DATABASE")
	// if err != nil {
	// 	return err
	// }
	// db, err := badger.Open(badger.DefaultOptions(dbPath))
	// if err != nil {
	// 	return err
	// }
	ctx := context.Background()
	// const host = "localhost"
	// certfile, err1 := utils.GetEnv("CERTFILE")
	// keyfile, err2 := utils.GetEnv("KEYFILE")
	// cafile, err3 := utils.GetEnv("CAFILE")
	// tlsc, err := utils.LoadTlSTransport(certfile, keyfile, cafile)
	// if err != nil {
	// 	return err
	// }

	conn, err := grpc.NewClient(srvAddr,
		grpc.WithTransportCredentials(tlsc),
		// grpc.WithPerRPCCredentials(tsrcgrpc),
	)
	if err != nil {
		fmt.Println(5)
		return err
	}
	ccli := proto.NewChatClient(conn)
	str, err := ccli.Stream(ctx)
	if err != nil {
		return err
	}
	go send(str)
	recv(str)

	return nil
}
func send(bidi grpc.BidiStreamingClient[proto.StreamRequest, proto.StreamResponse]) {
	loremIpsumGenerator := loremipsum.New()
	// tick := time.Tick(300 * time.Millisecond)
	for {
		text := loremIpsumGenerator.Paragraph()
		if err := bidi.Send(&proto.StreamRequest{Message: text}); err != nil {
			fmt.Println(err)
			break
		}
		time.Sleep(300 * time.Millisecond)
	}
}
func recv(bidi grpc.BidiStreamingClient[proto.StreamRequest, proto.StreamResponse]) {
	for {
		msg, err := bidi.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			fmt.Println(err)
			break
		}
		fmt.Printf("%s says: %s\n", msg.GetSenderId(), msg.Message)
	}
}

type HandleAllShitAndGetClient func() (clientId string, clientSecret string, err error)

func MakeHandleAllShitAndGetClient(reg Register, gc GetClient, sd SaveClientData) HandleAllShitAndGetClient {
	return func() (clientId string, clientSecret string, err error) {
		var cid, cs string
		cid, cs, err = gc()
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				cid, cs, err = reg()
				if err != nil {
					return
				}
				if err = sd(cid, cs); err != nil {
					return
				}
			}
			return
		}
		return
	}
}

type GetClient func() (clientId string, clientSecret string, err error)
type Register func() (clientId string, clientSecret string, err error)
type SaveClientData func(clientId string, clientSecret string) (err error)

func MakeSaveClientData(db *badger.DB) SaveClientData {
	return func(clientId, clientSecret string) (err error) {
		return db.Update(func(txn *badger.Txn) error {
			return errors.Join(
				txn.Set([]byte(cidkey), []byte(clientId)),
				txn.Set([]byte(cskey), []byte(clientSecret)))
		})
	}
}

const cidkey = "client_id"
const cskey = "client_secret"

func MakeGetClient(db *badger.DB) GetClient {
	return func() (clientId string, clientSecret string, err error) {
		var cid, cs string
		if err := db.View(func(txn *badger.Txn) error {
			var cid_, cs_ *badger.Item
			var err1, err2 error
			cid_, err1 = txn.Get([]byte(cidkey))
			cs_, err2 = txn.Get([]byte(cskey))
			if err := errors.Join(err1, err2); err != nil {
				return err
			}
			err1 = cid_.Value(func(val []byte) error { cid = string(val); return nil })
			err2 = cs_.Value(func(val []byte) error { cs = string(val); return nil })
			if err := errors.Join(err1, err2); err != nil {
				return err
			}
			return nil
		}); err != nil {
			return "", "", err
		}
		return cid, cs, nil
	}
}
func makeReg(endpoint string) Register {
	return func() (clientId string, clientSecret string, err error) {
		dcrpconf := dcrp.Config{
			Metadata: dcrp.Metadata{
				GrantTypes: []string{"client_credentials"}},
			ClientRegistrationEndpointURL: endpoint,
		}
		resp, err := dcrpconf.Register()
		if err != nil {
			return "", "", err
		}
		return resp.ClientID, resp.ClientSecret, err
	}
}
