package gocoin

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
	"time"
)

var client *Client

type Client struct {
	Signer Signer
	Sender Dispatcher
}

type Signer interface {
	GenerateHMAC(data string, secret string) string
	GenerateSignature(message Message) string
}

type Dispatcher interface {
	Sign(m Message, apiSecret string) string
	GetEpochTime() int64
}

type Sig string
type Req string

func NewClient() *Client {

	if client == nil {
		client = &Client{
			Signer: Sig("Signer"),
			Sender: Req("Request Handler"),
		}
	}
	return client
}

// Hash-based message authentication code
func (s Sig) GenerateHMAC(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (s Sig) GenerateSignature(m Message) string {
	sig := s.GenerateHMAC(fmt.Sprintf("%v%s%s%s", m.Epoch, m.Method, m.Path, m.Body), m.Secret)
	return sig
}

func (r Req) Sign(m Message, apiSecret string) string {
	m.Secret = apiSecret
	m.Epoch = r.GetEpochTime()
	return NewClient().Signer.GenerateSignature(m)
}

func (r Req) GetEpochTime() int64 {
	return time.Now().Unix()
}

func (c *Client) NewRequest(ctx context.Context, m Message, apikey, apiSecret string) ([]byte, error) { //signature string, epoch int64, url string) ([]byte, error) {

	httpclient := &http.Client{}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, m.URL, bytes.NewReader([]byte("")))

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("CB-ACCESS-SIGN", c.Sender.Sign(m, apiSecret))
	req.Header.Set("CB-ACCESS-TIMESTAMP", strconv.Itoa(int(c.Sender.GetEpochTime())))
	req.Header.Set("CB-ACCESS-KEY", apikey)
	req.Header.Set("CB-VERSION", "2015-07-22")

	res, err := httpclient.Do(req)
	if err != nil {
		logger.Error("request.NewRequest failure..", zap.Error(err))
		return nil, err
	}
	defer res.Body.Close()

	b, err := io.ReadAll(res.Body)
	if err != nil {
		logger.Error("request.NewRequest failed reading byte stream..", zap.Error(err))
		return nil, err
	}

	return b, err
}
