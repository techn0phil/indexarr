package services

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTVClientSearchTV_LoginAndQuery(t *testing.T) {
	loginCalls := 0
	searchCalls := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v4/login":
			loginCalls++
			fmt.Fprint(w, `{"status":"success","data":{"token":"test-token"}}`)
		case "/v4/search":
			searchCalls++
			if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
				t.Fatalf("unexpected auth header: %q", auth)
			}
			if got := r.URL.Query().Get("type"); got != "series" {
				t.Fatalf("unexpected type query: %q", got)
			}
			fmt.Fprint(w, `{"status":"success","data":[{"tvdb_id":"123","name":"Dark"}]}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := NewTVClient("tvdb-key", nil)
	client.httpClient = newRewrittenClient(srv)

	result, err := client.SearchTV("Dark")
	if err != nil {
		t.Fatalf("SearchTV() error = %v", err)
	}
	if loginCalls != 1 || searchCalls != 1 {
		t.Fatalf("unexpected calls: login=%d search=%d", loginCalls, searchCalls)
	}
	if len(result.Data) != 1 || result.Data[0].TVDBId != "123" {
		t.Fatalf("unexpected result payload: %+v", result)
	}
}

func TestTVClientSearchTV_401RefreshAndRetry(t *testing.T) {
	searchCalls := 0
	loginCalls := 0

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v4/login":
			loginCalls++
			fmt.Fprint(w, `{"status":"success","data":{"token":"new-token"}}`)
		case "/v4/search":
			searchCalls++
			if searchCalls == 1 {
				w.WriteHeader(http.StatusUnauthorized)
				fmt.Fprint(w, `{"status":"failure"}`)
				return
			}
			if auth := r.Header.Get("Authorization"); auth != "Bearer new-token" {
				t.Fatalf("expected refreshed token, got %q", auth)
			}
			fmt.Fprint(w, `{"status":"success","data":[]}`)
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	client := NewTVClient("tvdb-key", nil)
	client.httpClient = newRewrittenClient(srv)
	client.token = "stale-token"
	client.tokenExpiry = time.Now().Add(48 * time.Hour)

	if _, err := client.SearchTV("Severance"); err != nil {
		t.Fatalf("SearchTV() error = %v", err)
	}
	if loginCalls != 1 {
		t.Fatalf("expected one login refresh call, got %d", loginCalls)
	}
	if searchCalls != 2 {
		t.Fatalf("expected request retry on 401, got %d search calls", searchCalls)
	}
}

func TestTVClientLogin_MissingToken(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"status":"success","data":{"token":""}}`)
	}))
	defer srv.Close()

	client := NewTVClient("tvdb-key", nil)
	client.httpClient = newRewrittenClient(srv)

	err := client.login()
	if err == nil || !strings.Contains(err.Error(), "missing token") {
		t.Fatalf("expected missing token error, got %v", err)
	}
}
