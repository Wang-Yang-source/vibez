package views

import (
	"github.com/simone-vibes/vibez/internal/lyrics"
	"golang.org/x/image/font/gofont/goregular"
	"image"
	"image/color"
	"testing"
	"time"
)

func TestLyricCanvasProgressAndDuet(t *testing.T) {
	r, e := NewLyricRenderer(goregular.TTF, 0, 32)
	if e != nil {
		t.Fatal(e)
	}
	l := NewLyrics()
	l.SetLyrics(&lyrics.Result{Lines: []lyrics.Line{{Start: 0, Text: "a long phrase", Speaker: "A"}, {Start: 10 * time.Second, Text: "another phrase", Speaker: "B"}}, Synced: true}, nil)
	l.SetDuration(20 * time.Second)
	l.SetPosition(time.Second)
	first := r.RenderCanvas(l, 800, 400)
	l.SetPosition(8 * time.Second)
	second := r.RenderCanvas(l, 800, 400)
	if first.Bounds() != image.Rect(0, 0, 800, 400) {
		t.Fatal("incorrect canvas bounds")
	}
	different := false
	for y := 0; y < 400; y++ {
		for x := 0; x < 800; x++ {
			if first.At(x, y) != second.At(x, y) {
				different = true
			}
		}
	}
	if !different {
		t.Fatal("timed highlight did not advance")
	}
	l.SetPosition(time.Second)
	again := r.RenderCanvas(l, 800, 400)
	for y := 0; y < 400; y++ {
		for x := 0; x < 800; x++ {
			if first.At(x, y) != again.At(x, y) {
				t.Fatal("seeking back must restore earlier highlight")
			}
		}
	}
}

func TestLyricCanvasUsesTerminalBackgroundOrTransparency(t *testing.T) {
	r, _ := NewLyricRenderer(goregular.TTF, 0, 32)
	l := NewLyrics()
	l.SetLyrics(&lyrics.Result{Lines: []lyrics.Line{{Text: "Test lyric"}}, Synced: true}, nil)
	bg := color.RGBA{R: 30, G: 30, B: 46, A: 255}
	img := r.RenderCanvasColors(l, 800, 400, color.White, color.Gray{Y: 120}, bg)
	if img.RGBAAt(0, 0) != bg {
		t.Fatal("canvas does not match the supplied terminal background")
	}
	transparent := r.RenderCanvasColors(l, 800, 400, color.White, color.Gray{Y: 120}, nil)
	if transparent.RGBAAt(0, 0).A != 0 {
		t.Fatal("unknown terminal background must remain transparent")
	}
}
