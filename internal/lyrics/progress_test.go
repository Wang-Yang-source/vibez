package lyrics

import (
	"testing"
	"time"
	"unicode/utf8"
)

func TestProgressPrefersRealTimesAndRespectsSeek(t *testing.T) {
	line := Line{Start: time.Second, Text: "测试", Words: []Word{{Start: time.Second, End: 2 * time.Second, Text: "测"}, {Start: 4 * time.Second, End: 5 * time.Second, Text: "试"}}}
	n, f := Progress(line, 10*time.Second, 1500*time.Millisecond)
	if n != 0 || f != .5 {
		t.Fatalf("word timing ignored: %d %f", n, f)
	}
	n, f = Progress(line, 10*time.Second, 3*time.Second)
	if n != len("测") || f != 0 {
		t.Fatal("sings through explicit gap")
	}
	n, _ = Progress(line, 10*time.Second, 500*time.Millisecond)
	if n != 0 {
		t.Fatal("seek backward retains highlight")
	}
}
func TestEstimatedProgressKeepsGraphemesIntact(t *testing.T) {
	text := "中👨‍👩‍👧‍👦é文"
	for _, pos := range []time.Duration{0, time.Second, 2 * time.Second, 3 * time.Second, 4 * time.Second} {
		n, _ := Progress(Line{Text: text}, 4*time.Second, pos)
		if !utf8.ValidString(text[:n]) {
			t.Fatal("broken Unicode")
		}
	}
	n, _ := Progress(Line{Text: text}, 4*time.Second, 2*time.Second)
	if text[:n] != "中👨‍👩‍👧‍👦" {
		t.Fatal("family emoji split")
	}
}
