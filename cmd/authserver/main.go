package main

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"os"
	"time"

	models "codeberg.org/shinyzero0/vtb-api-2024-grpc/server-models"
	"codeberg.org/shinyzero0/vtb-api-2024-grpc/utils"
	"github.com/gofiber/fiber/v2"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"golang.org/x/oauth2/dcrp"
	_ "modernc.org/sqlite"
)

func main() {
	if err := f(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
func f() error {
	app := fiber.New(fiber.Config{})
	dsn, err := utils.GetEnv("DSN")
	if err != nil {
		return err
	}
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return err
	}
	laddr, err := utils.GetEnv("LISTEN_ADDR")
	ctx := context.Background()
	reg := MakeRegisterClient(ctx, db)
	app.Post("/register", MakeResgisterHandler(reg))
	app.Post("/token", MakeTokenHandler(ctx, db))
	return app.Listen(laddr)
}

type expirationTime int32

func (e *expirationTime) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || string(b) == "null" {
		return nil
	}
	var n json.Number
	err := json.Unmarshal(b, &n)
	if err != nil {
		return err
	}
	i, err := n.Int64()
	if err != nil {
		return err
	}
	if i > math.MaxInt32 {
		i = math.MaxInt32
	}
	*e = expirationTime(i)
	return nil
}

type tokenJSON struct {
	AccessToken  string         `json:"access_token"`
	TokenType    string         `json:"token_type"`
	RefreshToken string         `json:"refresh_token"`
	ExpiresIn    expirationTime `json:"expires_in"` // at least PayPal returns string, while most return number
	// error fields
	// https://datatracker.ietf.org/doc/html/rfc6749#section-5.2
	ErrorCode        string `json:"error"`
	ErrorDescription string `json:"error_description"`
	ErrorURI         string `json:"error_uri"`
}

const ttl = 15

func MakeTokenHandler(ctx context.Context, db boil.ContextExecutor) fiber.Handler {
	return func(c *fiber.Ctx) error {
		tok, err1 := randBytes()
		refresh, err2 := randBytes()
		if err := errors.Join(err1, err2); err != nil {
			return err
		}
		m := models.Token{
			ExpiresAt: time.Now().Add(time.Second * ttl).Unix(),
			Value:     hex.EncodeToString(tok),
			Refresh:   hex.EncodeToString(refresh),
		}
		if err := m.Insert(ctx, db, boil.Infer()); err != nil {
			return err
		}
		return c.JSON(tokenJSON{
			AccessToken:  hex.EncodeToString(tok),
			TokenType:    "bearer",
			RefreshToken: hex.EncodeToString(refresh),
			ExpiresIn:    ttl,
			// ErrorCode:        "",
			// ErrorDescription: "",
			// ErrorURI:         "",
		})
	}
}
func MakeResgisterHandler(reg RegisterClient) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var input dcrp.Metadata
		if err := c.BodyParser(&input); err != nil {
			return err
		}
		resp, err := reg(input)
		if err != nil {
			return err
		}
		
		return c.Status(201).JSON(resp)
	}
}

type RegisterClient func(dcrp.Metadata) (dcrp.Response, error)

func randBytes() ([]byte, error) {
	buf := make([]byte, 256/8)
	_, err := rand.Read(buf)
	return buf, err
}
func MakeRegisterClient(ctx context.Context, db boil.ContextExecutor) RegisterClient {
	return func(md dcrp.Metadata) (dcrp.Response, error) {
		// for _, v := range md.GrantTypes {
		// 	if v != "client_credentials" {
		// 		return dcrp.Response{}, errors.New("unsupported grant type")
		// 	}
		// }
		buf, err := randBytes()
		if err != nil {
			return dcrp.Response{}, err
		}

		model := models.Client{
			ClientSecret: hex.EncodeToString(buf),
		}
		if err := model.Insert(ctx, db, boil.Infer()); err != nil {
			return dcrp.Response{}, err
		}

		return dcrp.Response{
			ClientID:              fmt.Sprint(model.ClientID),
			ClientSecret:          model.ClientSecret,
			ClientSecretExpiresAt: time.Unix(0, 0),
			ClientIDIssuedAt:      time.Unix(0, 0),
		}, nil
	}
}
