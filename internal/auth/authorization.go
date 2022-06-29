package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

	lara "github.com/guvense/lara/internal"
)

func RetrieveAuthBearerToken(tokenServerDetail lara.TokenServerDetail) (string, error) {

	if tokenServerDetail.Type == "" {
		return "", nil
	}

	data := url.Values{}

	if tokenServerDetail.Type == "Password-Credential" {

		data.Set("password", tokenServerDetail.Password)
		data.Set("username", tokenServerDetail.Username)
		data.Set("scope", tokenServerDetail.Scope)
		data.Set("grant_type", "password")

		
	} else if tokenServerDetail.Type == "Client-Credential" {

		data.Set("scope", tokenServerDetail.Scope)
		data.Set("grant_type", "client_credentials")

	}

	req, err := http.NewRequest("POST", tokenServerDetail.TokenUrl, strings.NewReader(data.Encode()))

	req.SetBasicAuth(tokenServerDetail.ClientID, tokenServerDetail.ClientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		fmt.Print("Error occurred while sending request server: %w", err)
		return "", err
	}

	defer resp.Body.Close()

	resBytes, err := io.ReadAll(resp.Body)

	var jsonRes map[string]interface{}
	err = json.Unmarshal(resBytes, &jsonRes)

	if err != nil {
		log.Fatalln(err)
	}

	if w, ok := jsonRes["access_token"].(string); ok {
        return w, nil
    }

	return "", nil

}
