package views

import (
	"image"
	"image/color"
	"image/draw"
	"os"
	"strconv"

	"github.com/rivo/uniseg"
	"github.com/simone-vibes/vibez/internal/lyrics"
	"github.com/simone-vibes/vibez/internal/tui/styles"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// LyricRenderer reuses x/image's font parser/rasterizer and the terminal's
// existing graphics transport. System fonts are read, never redistributed.
type LyricRenderer struct{ face font.Face }

func LoadLyricRenderer(path string, index int, size float64) (*LyricRenderer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return NewLyricRenderer(data, index, size)
}
func NewLyricRenderer(data []byte, index int, size float64) (*LyricRenderer, error) {
	collection, err := opentype.ParseCollection(data)
	if err != nil {
		return nil, err
	}
	parsed, err := collection.Font(index)
	if err != nil {
		return nil, err
	}
	face, err := opentype.NewFace(parsed, &opentype.FaceOptions{Size: size, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil, err
	}
	return &LyricRenderer{face: face}, nil
}
func (l *LyricsModel) HasLyrics() bool { return !l.loading && l.errMsg == "" && len(l.lines) > 0 }
func (l *LyricsModel) CanvasKey() string {
	return strconv.Itoa(l.revision) + ":" + strconv.FormatInt(int64(l.position)/int64(33e6), 10) + ":" + strconv.Itoa(int(l.canvasOffset))
}

type canvasPhrase struct {
	text              string
	source, offset, y int
}

// RenderCanvas lays out large text in pixels instead of fixed terminal cells.
// Pixel clipping of the lit foreground makes the current grapheme sweep smooth.
func (r *LyricRenderer) RenderCanvas(l *LyricsModel, width, height int) image.Image {
	img := image.NewNRGBA(image.Rect(0, 0, max(1, width), max(1, height)))
	face := r.face
	fontH := face.Metrics().Height.Ceil()
	lineH := fontH + max(8, fontH/3)
	blockW := min(max(1, width-48), 780)
	inset := (width - blockW) / 2
	anchor := max(0, l.currentIdx)
	for anchor > 0 && l.synced && l.currentIdx >= 0 && l.lines[anchor-1].Start == l.lines[l.currentIdx].Start {
		anchor--
	}
	var phrases []canvasPhrase
	cursor, anchorY := 0, 0
	speakers := []string{}
	for i, line := range l.lines {
		if line.Speaker != "" && !lyrics.IsChorus(line.Speaker) {
			found := false
			for _, s := range speakers {
				found = found || s == line.Speaker
			}
			if !found {
				speakers = append(speakers, line.Speaker)
			}
		}
		if i == anchor {
			anchorY = cursor
		}
		text := line.Text
		if line.Speaker != "" {
			text = line.Speaker + " · " + text
		}
		offset, start := 0, 0
		g := uniseg.NewGraphemes(text)
		for g.Next() {
			from, to := g.Positions()
			if from > start && font.MeasureString(face, text[start:to]).Ceil() > blockW {
				phrases = append(phrases, canvasPhrase{text[start:from], i, offset, cursor})
				cursor += lineH
				start = from
				offset = from
			}
		}
		phrases = append(phrases, canvasPhrase{text[start:], i, offset, cursor})
		cursor += lineH
		if i+1 == len(l.lines) || !l.synced || line.Start != l.lines[i+1].Start {
			cursor += lineH / 2
		}
	}
	target := float64(anchorY - height/2)
	if !l.synced {
		target = float64(l.scroll * lineH)
	}
	if !l.canvasReady || l.canvasWidth != width || l.canvasHeight != height || absFloat(target-l.canvasTarget) > float64(height) {
		l.canvasOffset = target
		l.canvasReady = true
	}
	l.canvasTarget = target
	l.canvasWidth = width
	l.canvasHeight = height
	for _, phrase := range phrases {
		y := phrase.y - int(l.canvasOffset) + face.Metrics().Ascent.Ceil()
		if y+fontH < 0 || y-fontH > height {
			continue
		}
		line := l.lines[phrase.source]
		active := l.synced && l.currentIdx >= 0 && line.Start == l.lines[l.currentIdx].Start
		x := inset
		textW := font.MeasureString(face, phrase.text).Ceil()
		if len(speakers) >= 2 && line.Speaker != "" && !lyrics.IsChorus(line.Speaker) {
			if line.Speaker == speakers[0] {
				x = 20
			} else if line.Speaker == speakers[1] {
				x = max(0, width-20-textW)
			}
		} else if lyrics.IsChorus(line.Speaker) {
			x = (width - textW) / 2
		}
		dim := color.NRGBAModel.Convert(styles.ColorMuted).(color.NRGBA)
		distance := absInt(y - height/2)
		if distance > height/3 {
			dim.A = 115
		} else if distance > height/5 {
			dim.A = 190
		}
		drawer := font.Drawer{Dst: img, Src: image.NewUniform(dim), Face: face, Dot: fixed.P(x, y)}
		drawer.DrawString(phrase.text)
		if active || !l.synced {
			lit := image.NewNRGBA(img.Bounds())
			d := font.Drawer{Dst: lit, Src: image.NewUniform(color.NRGBAModel.Convert(styles.ColorFg).(color.NRGBA)), Face: face, Dot: fixed.P(x, y)}
			d.DrawString(phrase.text)
			// Overlay at a one-pixel offset gives the active line a stronger weight.
			d.Dot = fixed.P(x+1, y)
			d.DrawString(phrase.text)
			highlight := textW
			if l.synced {
				complete, partial := l.lineProgress(phrase.source)
				if line.Speaker != "" {
					complete += len(line.Speaker + " · ")
				}
				complete = min(max(0, complete-phrase.offset), len(phrase.text))
				highlight = font.MeasureString(face, phrase.text[:complete]).Ceil()
				if complete < len(phrase.text) {
					g := uniseg.NewGraphemes(phrase.text[complete:])
					if g.Next() {
						highlight += int(float64(font.MeasureString(face, g.Str()).Ceil()) * partial)
					}
				}
			}
			area := image.Rect(x, max(0, y-fontH), x+highlight+1, min(height, y+fontH/3)).Intersect(img.Bounds())
			if highlight > 0 {
				draw.Draw(img, area, lit, area.Min, draw.Over)
			}
		}
	}
	return img
}
func absFloat(v float64) float64 {
	if v < 0 {
		return -v
	}
	return v
}
