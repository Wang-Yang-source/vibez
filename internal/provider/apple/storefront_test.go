package apple_test

import (
	"context"
	"github.com/simone-vibes/vibez/internal/config"
	"github.com/simone-vibes/vibez/internal/provider/apple"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAccountStorefrontIgnoresStaleUSConfig(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/me/storefront" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		w.Write([]byte(`{"data":[{"id":"cn"}]}`))
	}))
	defer server.Close()
	cfg := &config.Config{StoreFront: "us", AppleDeveloperToken: "fixture", AppleUserToken: "fixture"}
	p := apple.New(cfg)
	p.SetBaseURL(server.URL)
	sf, err := p.AccountStorefront(context.Background())
	if err != nil || sf != "cn" {
		t.Fatalf("stale region accepted: %s %v", sf, err)
	}
}
