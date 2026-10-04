package lyrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchRetriesWithoutOverSpecificAlbum(t *testing.T) {
	calls := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.URL.Query().Get("artist_name") != "周杰伦" || r.URL.Query().Get("track_name") != "七里香" {
			t.Error("localized names lost")
		}
		if r.URL.Query().Get("album_name") != "" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"syncedLyrics":"[00:01.00]甲：第一句\n[00:01.00]乙：第二句"}`))
	}))
	defer server.Close()
	client := &Client{http: server.Client(), baseURL: server.URL}
	res, err := client.Fetch(context.Background(), "周杰伦", "七里香", "Deluxe version", time.Minute)
	if err != nil || !res.Synced || len(res.Lines) != 2 || calls != 2 {
		t.Fatalf("fallback failed: %v calls:%d", err, calls)
	}
}
