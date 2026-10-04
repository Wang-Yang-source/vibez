package views

import (
	"github.com/simone-vibes/vibez/internal/lyrics"
	"golang.org/x/image/font/gofont/goregular"
	"image/color"
	"testing"
	"time"
)

func TestCanvasFocusBaselineAndReadableHighlight(t *testing.T) {
	r, _ := NewLyricRenderer(goregular.TTF, 0, 34)
	l := NewLyrics()
	l.SetDuration(10 * time.Second)
	l.SetLyrics(&lyrics.Result{Synced: true, Lines: []lyrics.Line{{Start: 0, Text: "A readable lyric"}, {Start: 5 * time.Second, Text: "Another phrase"}}}, nil)
	l.SetPosition(4 * time.Second)
	img := r.RenderCanvasColors(l, 900, 600, color.RGBA{R: 180, A: 255}, color.Gray{Y: 90}, color.RGBA{R: 30, G: 30, B: 46, A: 255})
	p := r.phrases[0]
	baseline := p.y - int(l.canvasOffset) + r.face.Metrics().Ascent.Ceil()
	if baseline != 252 {
		t.Fatalf("focus baseline %d; want 42%% of viewport", baseline)
	}
	if p.x != 63 || p.width > 738 {
		t.Fatal("lyric column does not retain side padding")
	}
	white := false
	for y := baseline - 60; y < baseline; y++ {
		for x := p.x; x < p.x+p.width; x++ {
			pixel := img.RGBAAt(x, y)
			if pixel.R > 230 && pixel.G > 230 && pixel.B > 230 {
				white = true
			}
		}
	}
	if !white {
		t.Fatal("current lyric did not use a bright neutral highlight")
	}
}
