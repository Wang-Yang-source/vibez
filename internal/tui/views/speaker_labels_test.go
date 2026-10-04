package views

import (
	"github.com/charmbracelet/x/ansi"
	"github.com/simone-vibes/vibez/internal/lyrics"
	"strings"
	"testing"
)

func TestSpeakerLabelsOnlyIntroduceChangedVocalSections(t *testing.T) {
	l := NewLyrics()
	l.SetLyrics(&lyrics.Result{Lines: []lyrics.Line{{Speaker: "浩"}, {Speaker: "浩", Text: "first"}, {Speaker: "浩", Text: "second"}, {Speaker: "aMEI", Text: "reply"}, {Speaker: "aMEI", Text: "reply again"}, {Speaker: "合", Text: "together"}, {Speaker: "合", Text: "together again"}, {Speaker: "浩", Text: "returns"}}}, nil)
	want := []string{"", "浩 · ", "", "aMEI · ", "", "合 · ", "", "浩 · "}
	for i, prefix := range want {
		if actual := l.speakerPrefix(i); actual != prefix {
			t.Errorf("line %d prefix %q, want %q", i, actual, prefix)
		}
	}
	text := ansi.Strip(strings.Join(l.InlineLines(80, 30), "\n"))
	if strings.Count(text, "浩 · ") != 2 || strings.Count(text, "aMEI · ") != 1 || strings.Count(text, "合 · ") != 1 {
		t.Fatalf("repeated labels: %s", text)
	}
}
