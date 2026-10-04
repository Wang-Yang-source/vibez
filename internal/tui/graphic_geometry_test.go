package tui

import (
	"github.com/simone-vibes/vibez/internal/lyrics"
	"github.com/simone-vibes/vibez/internal/tui/views"
	"golang.org/x/image/font/gofont/goregular"
	"testing"
)

func TestGraphicsGeometryMatchesVisibleLayout(t *testing.T) {
	m := newModel(nil)
	for _, w := range []int{40, 69, 70, 140} {
		for _, h := range []int{15, 30, 50} {
			m.width, m.height = w+4, h
			m.inlineLyrics = true
			m.nowPlayingLines(w, m.nowPlayingHeight())
			want := m.lyricViewport
			m.syncNowPlayingViewports()
			if m.lyricViewport != want {
				t.Fatalf("%dx%d: got %+v want %+v", w, h, m.lyricViewport, want)
			}
			cw, ch := m.coverPanelSize(w, m.nowPlayingHeight())
			if m.artworkViewport.Width != cw || m.artworkViewport.Height != ch {
				t.Fatal("cover viewport differs")
			}
		}
	}
	m.inlineLyrics = false
	m.syncNowPlayingViewports()
	if m.lyricViewport.Width != 0 || m.artworkViewport.Width != m.width-4 {
		t.Fatal("single panel geometry incorrect")
	}
}
func TestLyricPlacementCacheTracksResize(t *testing.T) {
	m := newModel(nil)
	m.cfg.LyricsFontScale = 1.8
	m.supportsArtGraphics = func() bool { return true }
	m.lyricRenderer, _ = views.NewLyricRenderer(goregular.TTF, 0, 32)
	m.lyricsP.m.SetLyrics(&lyrics.Result{Synced: true, Lines: []lyrics.Line{{Text: "Test phrase"}}}, nil)
	first := m.inlineLyricLines(80, 20)
	next := m.inlineLyricLines(80, 20)
	if &first[0] != &next[0] {
		t.Fatal("static placement regenerated")
	}
	resized := m.inlineLyricLines(60, 12)
	if len(resized) != 12 || &first[0] == &resized[0] {
		t.Fatal("placement resize missed")
	}
}
func BenchmarkGraphicsViewport(b *testing.B) {
	m := newModel(nil)
	m.width, m.height = 150, 45
	m.inlineLyrics = true
	b.Run("render_layout", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			m.nowPlayingLines(m.width-4, m.nowPlayingHeight())
		}
	})
	b.Run("geometry_only", func(b *testing.B) {
		b.ReportAllocs()
		for range b.N {
			m.syncNowPlayingViewports()
		}
	})
}
