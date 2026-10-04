package lyrics

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestEnhancedSelectsFullDuetAcrossChineseScripts(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/search/get/web" {
			w.Write([]byte(`{"result":{"songs":[{"id":2,"name":"没有一杯茶解决不了的事","duration":223448,"artists":[{"name":"周传雄"}]},{"id":1,"name":"沒有一杯茶解決不了的事 (feat. 杨和苏KeyNG)","duration":223448,"artists":[{"name":"周传雄"},{"name":"杨和苏KeyNG"}]}]}}`))
			return
		}
		r.ParseForm()
		if r.Form.Get("id") != "1" {
			t.Errorf("selected solo edit: %s", r.Form.Get("id"))
		}
		w.Write([]byte(`{"code":200,"lrc":{"lyric":"[00:01.00]周传雄：\n[00:02.00]测试主唱\n[00:23.40]杨和苏KeyNG：\n[00:24.00]测试说唱"}}`))
	}))
	defer s.Close()
	c := &Client{http: s.Client(), enhancedURL: s.URL}
	res, err := c.fetchEnhanced(context.Background(), "周传雄", "没有一杯茶解决不了的事 (feat. 杨和苏KeyNG)", 223*time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Lines) != 4 || res.Lines[3].Speaker != "杨和苏KeyNG" || res.Lines[3].Text != "测试说唱" {
		t.Fatalf("rap section lost: %+v", res.Lines)
	}
}
func TestTitleScriptMatchingDoesNotChangeRecordingIdentity(t *testing.T) {
	if !matchingTitle("沒有一杯茶解決不了的事 (feat. 杨和苏KeyNG)", "没有一杯茶解决不了的事") {
		t.Fatal("traditional title missed")
	}
	if matchingTitle("沒有一杯茶解決不了的事 (Live)", "没有一杯茶解决不了的事") {
		t.Fatal("live recording incorrectly matched")
	}
}
