package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	lara "github.com/guvense/lara/internal"
)

func TestRetrieveAuthBearerTokenPasswordCredential(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if r.Form.Get("grant_type") != "password" {
			t.Fatalf("expected password grant_type, got %s", r.Form.Get("grant_type"))
		}

		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "token-123"})
	}))
	defer ts.Close()

	token, err := RetrieveAuthBearerToken(lara.TokenServerDetail{
		Type:         "Password-Credential",
		TokenUrl:     ts.URL,
		ClientID:     "client",
		ClientSecret: "secret",
		Username:     "u",
		Password:     "p",
		Scope:        "read",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "token-123" {
		t.Fatalf("expected token-123, got %s", token)
	}
}

func TestRetrieveAuthBearerTokenClientCredential(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Fatalf("parse form failed: %v", err)
		}
		if r.Form.Get("grant_type") != "client_credentials" {
			t.Fatalf("expected client_credentials grant_type, got %s", r.Form.Get("grant_type"))
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"access_token": "token-client"})
	}))
	defer ts.Close()

	token, err := RetrieveAuthBearerToken(lara.TokenServerDetail{
		Type:         "Client-Credential",
		TokenUrl:     ts.URL,
		ClientID:     "client",
		ClientSecret: "secret",
		Scope:        "admin",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "token-client" {
		t.Fatalf("expected token-client, got %s", token)
	}
}

func TestRetrieveAuthBearerTokenEmptyType(t *testing.T) {
	token, err := RetrieveAuthBearerToken(lara.TokenServerDetail{})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if token != "" {
		t.Fatalf("expected empty token, got %s", token)
	}
}
