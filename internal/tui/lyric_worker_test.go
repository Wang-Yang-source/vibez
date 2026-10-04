package tui

import (
	"github.com/simone-vibes/vibez/internal/lyrics"
	"github.com/simone-vibes/vibez/internal/tui/views"
	"golang.org/x/image/font/gofont/goregular"
	"testing"
	"time"
)

func TestLyricWorkerDoesNotBlockOrOverwriteNewSong(t *testing.T) {
	m := newModel(nil)
	m.width, m.height = 100, 40
	m.inlineLyrics = true
	m.cfg.LyricsFontScale = 1.8
	m.supportsArtGraphics = func() bool { return true }
	m.lyricRenderer, _ = views.NewLyricRenderer(goregular.TTF, 0, 32)
	m.lyricsP.m.SetLyrics(&lyrics.Result{Synced: true, Lines: []lyrics.Line{{Text: "first song", Start: 0}, {Text: "next line", Start: time.Second}}}, nil)
	cmd := m.syncLyricGraphics()
	if cmd == nil || !m.lyricGraphics.busy {
		t.Fatal("frame was not scheduled")
	}
	if m.syncLyricGraphics() != nil {
		t.Fatal("parallel frame jobs must be coalesced")
	}
	frame := cmd().(lyricFrameMsg)
	if frame.err != nil || frame.data == "" {
		t.Fatalf("frame failed: %v", frame.err)
	}
	m.lyricsP.m.SetLyrics(&lyrics.Result{Synced: true, Lines: []lyrics.Line{{Text: "new song"}}}, nil)
	_, _ = m.Update(frame)
	if m.lyricGraphics.visible {
		t.Fatal("old song frame became visible")
	}
}
