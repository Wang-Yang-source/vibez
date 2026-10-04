package views

import (
	"github.com/simone-vibes/vibez/internal/lyrics"
	"github.com/simone-vibes/vibez/internal/tui/art"
	"github.com/simone-vibes/vibez/internal/tui/styles"
	"golang.org/x/image/font/gofont/goregular"
	"testing"
	"time"
)

func BenchmarkLyricFrame(b *testing.B) {
	r, _ := NewLyricRenderer(goregular.TTF, 0, 34)
	l := NewLyrics()
	var lines []lyrics.Line
	for i := 0; i < 50; i++ {
		lines = append(lines, lyrics.Line{Start: time.Duration(i) * 5 * time.Second, Text: "This is a test lyric line with several words"})
	}
	l.SetLyrics(&lyrics.Result{Synced: true, Lines: lines}, nil)
	l.SetDuration(250 * time.Second)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		l.SetPosition(120*time.Second + time.Duration(i)*time.Millisecond)
		img := r.RenderAnimationCanvas(l, 900, 600, styles.ColorFg, styles.ColorMuted, styles.ColorBg)
		if _, e := art.KittyUploadAnimation(img, 0x570001, art.Size{Width: 75, Height: 25}); e != nil {
			b.Fatal(e)
		}
	}
}
