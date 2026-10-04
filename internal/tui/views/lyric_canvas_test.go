package views

import (
	"github.com/simone-vibes/vibez/internal/lyrics"
	"golang.org/x/image/font/gofont/goregular"
	"image"
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
