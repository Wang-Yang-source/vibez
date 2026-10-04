package views

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/rivo/uniseg"
	"github.com/simone-vibes/vibez/internal/lyrics"
	"github.com/simone-vibes/vibez/internal/tui/styles"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

// LyricRenderer reuses x/image's font parser/rasterizer and the terminal's
// existing graphics transport. System fonts are read, never redistributed.
type LyricRenderer struct {
	face                 font.Face
	parsed               *opentype.Font
	baseSize, pixelScale float64
	layoutKey            string
	phrases              []canvasPhrase
	lineY                []int
	speakers             []string
	animationCanvas      *image.RGBA
}

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
	return &LyricRenderer{face: face, parsed: parsed, baseSize: size, pixelScale: 1}, nil
}
func (l *LyricsModel) HasLyrics() bool { return !l.loading && l.errMsg == "" && len(l.lines) > 0 }
func (l *LyricsModel) CanvasKey() string {
	return strconv.Itoa(l.revision) + ":" + strconv.FormatInt(int64(l.position)/int64(8e6), 10) + ":" + strconv.Itoa(int(l.canvasOffset))
}

type canvasPhrase struct {
	text              string
	source, offset, y int
	width, x          int
	mask              *image.Alpha
}

// RenderCanvas lays out large text in pixels instead of fixed terminal cells.
// Pixel clipping of the lit foreground makes the current grapheme sweep smooth.
func (r *LyricRenderer) RenderCanvas(l *LyricsModel, width, height int) *image.RGBA {
	return r.RenderCanvasColors(l, width, height, styles.ColorFg, styles.ColorMuted, styles.ColorBg)
}

func (r *LyricRenderer) RenderCanvasColors(l *LyricsModel, width, height int, fg, muted, bg color.Color) *image.RGBA {
	return r.renderCanvasInto(image.NewRGBA(image.Rect(0, 0, max(1, width), max(1, height))), l, width, height, fg, muted, bg)
}

// RenderAnimationCanvas returns worker-owned storage, valid until the next
// call. The serialized graphics worker encodes it before allowing another frame.
// RenderCanvasColors keeps returning independent snapshots for other callers.
func (r *LyricRenderer) RenderAnimationCanvas(l *LyricsModel, width, height int, fg, muted, bg color.Color) *image.RGBA {
	bounds := image.Rect(0, 0, max(1, width), max(1, height))
	if r.animationCanvas == nil || r.animationCanvas.Bounds() != bounds {
		r.animationCanvas = image.NewRGBA(bounds)
	}
	return r.renderCanvasInto(r.animationCanvas, l, width, height, fg, muted, bg)
}

func (r *LyricRenderer) renderCanvasInto(img *image.RGBA, l *LyricsModel, width, height int, fg, muted, bg color.Color) *image.RGBA {
	if bg == nil {
		clear(img.Pix)
	}
	if bg != nil {
		draw.Draw(img, img.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
	}
	face := r.face
	fontH := face.Metrics().Height.Ceil()
	scale := r.pixelScale
	em := r.baseSize * scale
	lineH := int(math.Ceil(em * 1.18))
	blockW := min(max(1, int(float64(width)*.82)), int(620*scale))
	inset := int(float64(width) * .07)
	anchor := max(0, l.currentIdx)
	for anchor > 0 && l.synced && l.currentIdx >= 0 && l.lines[anchor-1].Start == l.lines[l.currentIdx].Start {
		anchor--
	}
	layoutKey := strconv.Itoa(l.revision) + ":" + strconv.Itoa(width) + ":" + strconv.FormatFloat(scale, 'f', 4, 64)
	if r.layoutKey != layoutKey {
		var phrases []canvasPhrase
		cursor := 0
		r.lineY = make([]int, len(l.lines))
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
			r.lineY[i] = cursor
			if line.Speaker != "" && strings.TrimSpace(line.Text) == "" {
				continue
			}
			text := l.speakerPrefix(i) + line.Text
			offset, start := 0, 0
			g := uniseg.NewGraphemes(text)
			for g.Next() {
				from, to := g.Positions()
				if from > start && font.MeasureString(face, text[start:to]).Ceil() > blockW {
					phrases = append(phrases, canvasPhrase{text: text[start:from], source: i, offset: offset, y: cursor})
					cursor += lineH
					start = from
					offset = from
				}
			}
			phrases = append(phrases, canvasPhrase{text: text[start:], source: i, offset: offset, y: cursor})
			cursor += lineH
			if i+1 == len(l.lines) || !l.synced || line.Start != l.lines[i+1].Start {
				cursor += max(3, int(math.Ceil(em*.65)))
			}
		}
		for i := range phrases {
			phrase := &phrases[i]
			phrase.width = font.MeasureString(face, phrase.text).Ceil()
			phrase.x = inset
			line := l.lines[phrase.source]
			if len(speakers) >= 2 && line.Speaker != "" && !lyrics.IsChorus(line.Speaker) {
				if line.Speaker == speakers[0] {
					phrase.x = int(20 * scale)
				} else if line.Speaker == speakers[1] {
					phrase.x = max(0, width-int(20*scale)-phrase.width)
				}
			} else if lyrics.IsChorus(line.Speaker) {
				phrase.x = (width - phrase.width) / 2
			}
			// Cache glyph coverage; each frame only clips and colors these masks.
			phrase.mask = image.NewAlpha(image.Rect(-fontH, -fontH, phrase.width+fontH, fontH))
			d := font.Drawer{Dst: phrase.mask, Src: image.White, Face: face, Dot: fixed.P(0, 0)}
			d.DrawString(phrase.text)
		}
		r.phrases, r.speakers = phrases, speakers
		r.layoutKey = layoutKey
	}
	phrases := r.phrases
	anchorY := r.lineY[anchor]

	focusY := int(float64(height) * .42)
	target := float64(anchorY + face.Metrics().Ascent.Ceil() - focusY)
	if !l.synced {
		target = float64(l.scroll * lineH)
	}
	if !l.canvasReady || l.canvasWidth != width || l.canvasHeight != height || absFloat(target-l.canvasTarget) > float64(height) {
		l.canvasOffset = target
		l.canvasVelocity = 0
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
		active := l.lineActive(phrase.source)
		x, textW := phrase.x, phrase.width
		// Keep the focus bright and readable; fade surrounding lines continuously
		// instead of switching between three abrupt opacity bands.
		dim := color.NRGBA{R: 255, G: 255, B: 255, A: 115}
		distance := absFloat(float64(y-focusY)) / float64(max(1, height))
		dim.A = uint8(max(30, 115*(1-min(1, distance*1.8))))
		if active {
			dim.A = 135
		}
		bounds := phrase.mask.Bounds().Add(image.Pt(x, y)).Intersect(img.Bounds())
		draw.DrawMask(img, bounds, image.NewUniform(dim), image.Point{}, phrase.mask, bounds.Min.Sub(image.Pt(x, y)), draw.Over)
		if active || !l.synced {
			highlight := float64(textW)
			if l.synced {
				complete, partial := l.lineProgress(phrase.source)
				complete += len(l.speakerPrefix(phrase.source))
				complete = min(max(0, complete-phrase.offset), len(phrase.text))
				highlight = float64(font.MeasureString(face, phrase.text[:complete])) / 64
				if complete < len(phrase.text) {
					g := uniseg.NewGraphemes(phrase.text[complete:])
					if g.Next() {
						highlight += float64(font.MeasureString(face, g.Str())) / 64 * partial
					}
				}
			}
			if highlight > 0 {
				// Feather the advancing boundary without blurring the glyphs.
				feather := max(2, int(em*.18))
				right := float64(x) + highlight
				solidRight := max(x, int(right)-feather)
				area := image.Rect(x, max(0, y-fontH), solidRight, min(height, y+fontH/3)).Intersect(img.Bounds())
				draw.DrawMask(img, area, image.White, image.Point{}, phrase.mask, area.Min.Sub(image.Pt(x, y)), draw.Over)
				for column := solidRight; column < int(math.Ceil(right)); column++ {
					alpha := uint8(255 * min(1, max(0, (right-float64(column))/float64(feather))))
					edge := image.Rect(column, max(0, y-fontH), column+1, min(height, y+fontH/3)).Intersect(img.Bounds())
					draw.DrawMask(img, edge, image.NewUniform(color.NRGBA{R: 255, G: 255, B: 255, A: alpha}), image.Point{}, phrase.mask, edge.Min.Sub(image.Pt(x, y)), draw.Over)
				}
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

// SetPixelScale rasterizes at the terminal's native cell pixel density.
func (r *LyricRenderer) SetPixelScale(scale float64) error {
	if scale == r.pixelScale {
		return nil
	}
	face, err := opentype.NewFace(r.parsed, &opentype.FaceOptions{Size: r.baseSize * scale, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return err
	}
	_ = r.face.Close()
	r.face = face
	r.pixelScale = scale
	return nil
}

// ApplyCanvasLayout accepts only layout geometry from a completed worker frame;
// playback position and lyric selection remain owned by the UI event loop.
func (l *LyricsModel) ApplyCanvasLayout(frame *LyricsModel) bool {
	if l.revision != frame.revision {
		return false
	}
	l.canvasTarget = frame.canvasTarget
	if !l.canvasReady || l.canvasWidth != frame.canvasWidth || l.canvasHeight != frame.canvasHeight || absFloat(l.canvasTarget-l.canvasOffset) > float64(frame.canvasHeight) {
		l.canvasOffset = frame.canvasOffset
		l.canvasVelocity = 0
	}
	l.canvasReady = frame.canvasReady
	l.canvasWidth, l.canvasHeight = frame.canvasWidth, frame.canvasHeight
	return true
}
