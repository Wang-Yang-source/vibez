package tui

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/simone-vibes/vibez/internal/tui/art"
	"github.com/simone-vibes/vibez/internal/tui/styles"
	"github.com/simone-vibes/vibez/internal/tui/views"
)

const lyricImageID = 0x570001

// Separate IDs and dimensions keep the lyric canvas independent of cover art.
type lyricGraphics struct {
	key     string
	size    art.Size
	visible bool
	failed  bool
}
type lyricFontLoadedMsg struct {
	renderer *views.LyricRenderer
	err      error
}

func (m *Model) loadLyricFontCmd() tea.Cmd {
	path, scale := m.cfg.LyricsFont, m.cfg.LyricsFontScale
	return func() tea.Msg {
		index := 0
		if path == "" {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()
			out, err := exec.CommandContext(ctx, "fc-match", "-f", "%{file}\n%{index}\n", "sans:lang=zh-cn").Output()
			if err != nil {
				return lyricFontLoadedMsg{err: err}
			}
			parts := strings.Split(strings.TrimSpace(string(out)), "\n")
			path = parts[0]
			if len(parts) > 1 {
				index, _ = strconv.Atoi(parts[1])
			}
		}
		renderer, err := views.LoadLyricRenderer(path, index, 24*min(max(scale, 1), 3)*.78)
		return lyricFontLoadedMsg{renderer: renderer, err: err}
	}
}
func (m *Model) largeLyricsAvailable() bool {
	return m.cfg.LyricsFontScale > 1 && m.lyricRenderer != nil && !m.lyricGraphics.failed && m.supportsArtGraphics != nil && m.supportsArtGraphics()
}
func (m *Model) inlineLyricLines(w, h int) []string {
	m.lyricViewport = art.Size{Width: w, Height: h}
	if m.largeLyricsAvailable() && m.lyricsP.m.HasLyrics() && w >= 20 && h >= 6 {
		return art.KittyLines(lyricImageID, m.lyricViewport)
	}
	return m.lyricsP.m.InlineLines(w, h)
}
func (m *Model) syncLyricGraphics() tea.Cmd {
	if !m.inlineLyrics || !m.largeLyricsAvailable() || !m.lyricsP.m.HasLyrics() {
		if m.lyricGraphics.visible {
			m.lyricGraphics = lyricGraphics{}
			return tea.Raw(art.KittyDelete(lyricImageID))
		}
		return nil
	}
	m.nowPlayingLines(m.width-4, m.nowPlayingHeight())
	size := m.lyricViewport
	if size.Width < 20 || size.Height < 6 {
		return nil
	}
	key := fmt.Sprintf("%s:%dx%d:%v:%v", m.lyricsP.m.CanvasKey(), size.Width, size.Height, styles.ColorFg, styles.ColorMuted)
	if m.lyricGraphics.visible && m.lyricGraphics.key == key {
		return nil
	}
	aspect := m.artCellAsp
	if aspect <= 0 {
		aspect = 2
	}
	img := m.lyricRenderer.RenderCanvas(m.lyricsP.m, size.Width*12, int(float64(size.Height*12)*aspect))
	data, err := art.KittyUpload(img, lyricImageID, size)
	m.lyricGraphics = lyricGraphics{key: key, size: size, visible: err == nil, failed: err != nil}
	if err != nil {
		return nil
	}
	return tea.Raw(data)
}
