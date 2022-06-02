package callback

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"time"

	model "github.com/guvense/lara/internal/model"
	parser "github.com/guvense/lara/internal/parser"

	lara "github.com/guvense/lara/internal"
	auth "github.com/guvense/lara/internal/auth"
)

type CallBack struct {
	TokenServers lara.TokenServers
}

func (s *CallBack) Call(callbacks []model.Callback, parser *parser.Parser) {
	for _, callback := range callbacks {
		if callback.CallbackRest.Request.Endpoint != "" {

			var token string
		
			if callback.CallbackRest.AuthorizationKey != "" {
				tokenServers := s.TokenServers
				token, err := auth.RetrieveAuthBearerToken(tokenServers[callback.CallbackRest.AuthorizationKey])
				log.Printf("access token: %s", token);
				if err != nil {
					fmt.Print("Error occurred preparing token: %w", err)
				}
			}

			sendRestCallback(callback.CallbackRest, token, parser)
		}
	}
}

func sendRestCallback(restCallback model.CallbackRest, token string, parser *parser.Parser) {

	if restCallback.Delay.Delay() > 0 {
		time.Sleep(restCallback.Delay.Delay())
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{Transport: tr}

	var jsonBody []byte

	if len(restCallback.Request.Body) > 0 {
		stringJson := fmt.Sprintf("%v", string(restCallback.Request.Body))
		body := parser.Parse(stringJson)
		jsonBody = []byte(body)
		log.Printf("Callback body : %s", body)
	}

	endPoint := parser.Parse(restCallback.Request.Endpoint)

	req, err := http.NewRequest(restCallback.Request.Method, endPoint, bytes.NewBuffer(jsonBody))

	if err != nil {
		fmt.Print("Error occurred while preparing request: %w", err)
	}

	req.Header = http.Header{}

	for key, val := range *restCallback.Request.Headers {
		req.Header.Add(key, val)
	}

	if token != "" {
		req.Header.Add("Authorization", "Bearer "+token)
	}

	_, err = client.Do(req)

	if err != nil {
		fmt.Print("Error occurred while sending request server: %w", err)
	}

}
