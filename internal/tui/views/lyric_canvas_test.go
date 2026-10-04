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

func TestLyricCanvasUsesCompactPhraseSpacing(t *testing.T) {
	r, _ := NewLyricRenderer(goregular.TTF, 0, 32)
	l := NewLyrics()
	l.SetLyrics(&lyrics.Result{Synced: true, Lines: []lyrics.Line{{Start: 0, Text: "A short phrase"}, {Start: time.Second, Text: "The next phrase"}}}, nil)
	r.RenderCanvas(l, 800, 400)
	height := r.face.Metrics().Height.Ceil()
	gap := r.lineY[1] - r.lineY[0]
	if gap > height*2 || gap < height {
		t.Fatalf("paragraph spacing is not readable: %d for font height %d", gap, height)
	}
	if err := r.SetPixelScale(2); err != nil {
		t.Fatal(err)
	}
	r.RenderCanvas(l, 1600, 800)
	height = r.face.Metrics().Height.Ceil()
	gap = r.lineY[1] - r.lineY[0]
	if gap > height*2 || gap < height {
		t.Fatal("spacing changes proportion at native pixel density")
	}
}

func TestLyricCanvasDoesNotReserveRowsForSingerMarkers(t *testing.T) {
	r, _ := NewLyricRenderer(goregular.TTF, 0, 32)
	l := NewLyrics()
	l.SetLyrics(&lyrics.Result{Synced: true, Lines: []lyrics.Line{{Start: 0, Speaker: "A"}, {Start: 0, Speaker: "A", Text: "First phrase"}, {Start: time.Second, Speaker: "B"}, {Start: time.Second, Speaker: "B", Text: "Second phrase"}}}, nil)
	r.RenderCanvas(l, 800, 400)
	if len(r.phrases) != 2 || r.lineY[0] != r.lineY[1] || r.lineY[2] != r.lineY[3] {
		t.Fatal("speaker metadata introduced extra lyric spacing")
	}
}

func TestAnimationCanvasReuseClearsAndResizes(t *testing.T) {
	r, _ := NewLyricRenderer(goregular.TTF, 0, 32)
	l := NewLyrics()
	l.SetLyrics(&lyrics.Result{Synced: true, Lines: []lyrics.Line{{Text: "Test phrase"}}}, nil)
	bg := color.RGBA{R: 30, G: 30, B: 46, A: 255}
	first := r.RenderAnimationCanvas(l, 800, 400, color.White, color.Gray{Y: 120}, bg)
	first.Pix[0] = 255 // an old frame must not leak into the next frame
	next := r.RenderAnimationCanvas(l, 800, 400, color.White, color.Gray{Y: 120}, nil)
	if first != next {
		t.Fatal("same-sized frames reallocated")
	}
	expected := r.RenderCanvasColors(l, 800, 400, color.White, color.Gray{Y: 120}, nil)
	for i, b := range expected.Pix {
		if next.Pix[i] != b {
			t.Fatal("reused transparent canvas retained old pixels")
		}
	}
	resized := r.RenderAnimationCanvas(l, 700, 300, color.White, color.Gray{Y: 120}, bg)
	if resized == next || resized.Bounds() != image.Rect(0, 0, 700, 300) {
		t.Fatal("resize did not replace canvas")
	}
	if resized.RGBAAt(0, 0) != bg {
		t.Fatal("resized background wrong")
	}
}
