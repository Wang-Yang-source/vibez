package lyrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEnhancedSourcePreservesDuetLabels(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/search/get/web" {
			w.Write([]byte(`{"code":200,"result":{"songs":[{"id":1,"name":"测试","duration":60000,"artists":[{"name":"歌手"}]}]}}`))
			return
		}
		if r.URL.Path != "/api/song/lyric/v1" || r.Method != "POST" {
			t.Errorf("unexpected endpoint %s", r.URL.Path)
		}
		w.Write([]byte(`{"code":200,"lrc":{"lyric":"[00:01.00]甲：测试甲\n[00:03.00]乙：测试乙\n[00:05.00]合：合唱"}}`))
	}))
	defer s.Close()
	c := &Client{http: s.Client(), Enhanced: true, enhancedURL: s.URL}
	res, err := c.Fetch(context.Background(), "歌手", "测试", "", time.Minute)
	if err != nil || res.Lines[0].Speaker != "甲" || res.Lines[1].Speaker != "乙" || !IsChorus(res.Lines[2].Speaker) {
		t.Fatalf("roles lost: %+v %v", res, err)
	}
}
func TestWordTimingParsers(t *testing.T) {
	l, _ := parseLRC("[00:01.00]甲：<00:01.00>测<00:02.00>试<00:03.00>")
	if len(l) != 1 || l[0].Text != "测试" || len(l[0].Words) != 2 || l[0].Words[0].End != 2*time.Second {
		t.Fatalf("enhanced LRC lost timing: %+v", l)
	}
	y := parseYRC("[1000,2000](1000,500,0)测(1500,1500,0)试")
	if len(y) != 1 || y[0].Text != "测试" || y[0].Words[1].Start != 1500*time.Millisecond || y[0].End != 3*time.Second {
		t.Fatalf("YRC lost timing: %+v", y)
	}
}
