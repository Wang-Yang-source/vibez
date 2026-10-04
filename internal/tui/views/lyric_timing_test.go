package views

import (
	"github.com/simone-vibes/vibez/internal/lyrics"
	"testing"
	"time"
)

func TestOverlappingTimedVoicesStayActiveIndependently(t *testing.T) {
	l := NewLyrics()
	l.SetLyrics(&lyrics.Result{Synced: true, Lines: []lyrics.Line{
		{Start: time.Second, End: 5 * time.Second, Speaker: "甲", Text: "第一句"},
		{Start: 3 * time.Second, End: 6 * time.Second, Speaker: "乙", Text: "第二句"},
	}}, nil)
	l.SetPosition(4 * time.Second)
	if !l.lineActive(0) || !l.lineActive(1) {
		t.Fatal("second voice prematurely ended first voice highlight")
	}
	l.SetPosition(5500 * time.Millisecond)
	if l.lineActive(0) || !l.lineActive(1) {
		t.Fatal("completed voice still active")
	}
}
