package views

import (
	"slices"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/simone-vibes/vibez/internal/lyrics"
	"github.com/simone-vibes/vibez/internal/tui/styles"
)

// InlineLines centers the active timestamp group and uses separate alignment
// for explicitly labelled vocalists. It does not infer singers from audio.
func (l *LyricsModel) InlineLines(w, h int) []string {
	rows := make([]string, max(0, h))
	if h <= 0 || w <= 0 {
		return rows
	}
	if l.loading || l.errMsg != "" || len(l.lines) == 0 {
		label := "Waiting for playback"
		if l.loading {
			label = "Loading lyrics…"
		} else if l.errMsg != "" {
			label = "Lyrics source has no matching result"
			if !l.notFound {
				label = "Lyrics request failed; retry with y"
			}
		}
		rows[h/2] = centerLyric(styles.QueueItemMuted.Render(l.Locale.Text(label)), w)
		return rows
	}

	l.SetSize(w, h)
	speakers := []string{}
	for _, line := range l.lines {
		if line.Speaker != "" && !lyrics.IsChorus(line.Speaker) && !slices.Contains(speakers, line.Speaker) {
			speakers = append(speakers, line.Speaker)
		}
	}
	type visualRow struct {
		text   string
		source int
		offset int
	}
	var visual []visualRow
	anchor := max(0, l.currentIdx)
	for anchor > 0 && l.synced && l.currentIdx >= 0 && l.lines[anchor-1].Start == l.lines[l.currentIdx].Start {
		anchor--
	}
	anchorRow := 0
	blockW := min(max(1, w-6), 44)
	blockX := max(0, (w-blockW)/2)
	for i, line := range l.lines {
		if i == anchor {
			anchorRow = len(visual)
		}
		text := l.speakerPrefix(i) + line.Text
		offset := 0
		for _, part := range strings.Split(ansi.Wrap(text, blockW, ""), "\n") {
			partStart := offset
			if index := strings.Index(text[offset:], part); index >= 0 {
				partStart += index
			}
			visual = append(visual, visualRow{part, i, partStart})
			offset = partStart + len(part)
		}
		// Separate phrases, while keeping simultaneous singers together.
		if i+1 == len(l.lines) || !l.synced || line.Start != l.lines[i+1].Start {
			visual = append(visual, visualRow{"", -1, 0})
		}
	}
	target := anchorRow - h/2
	if !l.synced {
		target = l.scroll * 2
	}
	if !l.viewportReady || l.viewportWidth != w || l.viewportHeight != h || absInt(target-l.viewportTarget) > h {
		l.viewportOffset = target
		l.viewportReady = true
	}
	l.viewportTarget = target
	l.viewportWidth = w
	l.viewportHeight = h
	start := l.viewportOffset
	for row := range h {
		idx := start + row
		if idx < 0 || idx >= len(visual) || visual[idx].source < 0 {
			continue
		}
		item := visual[idx]
		line := l.lines[item.source]
		active := l.synced && l.currentIdx >= 0 && line.Start == l.lines[l.currentIdx].Start
		distance := absInt(idx - anchorRow)
		color := styles.ColorMuted
		if distance > h/3 {
			color = styles.ColorSurface
		}
		style := lipgloss.NewStyle().Foreground(color)
		if active {
			style = lipgloss.NewStyle().Foreground(styles.ColorFg).Bold(true)
		}
		if !l.synced {
			style = lipgloss.NewStyle().Foreground(styles.ColorFg)
		}
		x := blockX
		if len(speakers) >= 2 && line.Speaker != "" && !lyrics.IsChorus(line.Speaker) {
			if line.Speaker == speakers[0] {
				x = 2
			} else if line.Speaker == speakers[1] {
				x = max(0, w-2-lipgloss.Width(item.text))
			}
		} else if lyrics.IsChorus(line.Speaker) {
			x = max(0, (w-lipgloss.Width(item.text))/2)
		}
		rendered := style.Render(item.text)
		if active {
			completed, _ := l.lineProgress(item.source)
			completed += len(l.speakerPrefix(item.source))
			split := min(max(0, completed-item.offset), len(item.text))
			rendered = style.Render(item.text[:split]) + lipgloss.NewStyle().Foreground(styles.ColorMuted).Bold(true).Render(item.text[split:])
		}
		rows[row] = strings.Repeat(" ", x) + rendered
	}

	return rows
}
func centerLyric(text string, w int) string {
	return strings.Repeat(" ", max(0, (w-lipgloss.Width(text))/2)) + text
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// AdvanceFrame moves one terminal row per animation tick. Seeking across a
// whole screen snaps immediately; normal lyric transitions glide into place.
func (l *LyricsModel) AdvanceFrame() {
	l.canvasOffset += (l.canvasTarget - l.canvasOffset) * 0.28
	if absFloat(l.canvasTarget-l.canvasOffset) < 0.3 {
		l.canvasOffset = l.canvasTarget
	}
	if l.viewportOffset < l.viewportTarget {
		l.viewportOffset++
	} else if l.viewportOffset > l.viewportTarget {
		l.viewportOffset--
	}
}
