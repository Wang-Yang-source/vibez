package tui

import (
	"github.com/simone-vibes/vibez/internal/tui/art"
	"strings"
)

func (m *Model) coverPanelSize(w, h int) (int, int) {
	if !m.inlineLyrics {
		return w, h
	}
	if w >= 70 {
		return (w - 3) * 45 / 100, h
	}
	return w, max(8, h/2)
}
func (m *Model) coverAndLyricsLines(w, h int) []string {
	cw, ch := m.coverPanelSize(w, h)
	var cover []string
	if m.artModeActive() {
		cover = m.nowPlayingArtLines(cw, ch)
	} else {
		cover = m.nowPlayingTextLines(cw, ch)
	}
	if w < 70 {
		lines := append(cover, m.inlineLyricLines(w, max(0, h-ch))...)
		return toLines(strings.Join(lines, "\n"), h)
	}
	rw := w - cw - 3
	lyrics := m.inlineLyricLines(rw, h)
	lines := make([]string, h)
	for i := range h {
		lines[i] = padRight(safeIdx(cover, i), cw) + "   " + padRight(safeIdx(lyrics, i), rw)
	}
	return lines
}

// Graphics synchronization needs dimensions, not a second rendering of the
// entire text interface. Keep this arithmetic shared with the visible layout.
func (m *Model) syncNowPlayingViewports() {
	w, h := m.width-4, m.nowPlayingHeight()
	cw, ch := m.coverPanelSize(w, h)
	m.artworkViewport = art.Size{Width: cw, Height: ch}
	m.lyricViewport = art.Size{}
	if !m.inlineLyrics {
		return
	}
	if w < 70 {
		m.lyricViewport = art.Size{Width: w, Height: max(0, h-ch)}
	} else {
		m.lyricViewport = art.Size{Width: w - cw - 3, Height: h}
	}
}
