package views

import (
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
		label := "等待播放"
		if l.loading {
			label = "正在加载歌词…"
		} else if l.errMsg != "" {
			label = "暂无可用歌词"
		}
		rows[h/2] = centerLyric(styles.QueueItemMuted.Render(label), w)
		return rows
	}
	l.SetSize(w, h)
	speakers := []string{}
	for _, line := range l.lines {
		if line.Speaker == "" || lyrics.IsChorus(line.Speaker) {
			continue
		}
		found := false
		for _, s := range speakers {
			if s == line.Speaker {
				found = true
				break
			}
		}
		if !found {
			speakers = append(speakers, line.Speaker)
		}
	}
	anchor := max(0, l.currentIdx)
	if l.synced && l.currentIdx >= 0 {
		for anchor > 0 && l.lines[anchor-1].Start == l.lines[l.currentIdx].Start {
			anchor--
		}
	}
	start := max(0, anchor-h/2)
	if !l.synced {
		start = l.scroll
	}
	for row := range h {
		idx := start + row
		if idx >= len(l.lines) {
			break
		}
		line := l.lines[idx]
		text := line.Text
		if line.Speaker != "" {
			text = line.Speaker + " · " + text
		}
		active := l.synced && l.currentIdx >= 0 && line.Start == l.lines[l.currentIdx].Start
		style := styles.QueueItemMuted
		if active {
			style = lipgloss.NewStyle().Foreground(styles.ColorPrimary).Bold(true)
		}
		if !l.synced {
			style = lipgloss.NewStyle().Foreground(styles.ColorFg)
		}
		columnW := w
		duet := len(speakers) >= 2 && line.Speaker != "" && !lyrics.IsChorus(line.Speaker)
		if duet {
			columnW = max(1, w*3/4)
		}
		text = style.Render(ansi.Truncate(text, columnW, "…"))
		pad := max(0, w-lipgloss.Width(text))
		if duet && line.Speaker == speakers[0] {
			rows[row] = text
		} else if duet && line.Speaker == speakers[1] {
			rows[row] = strings.Repeat(" ", pad) + text
		} else {
			rows[row] = centerLyric(text, w)
		}
	}
	return rows
}
func centerLyric(text string, w int) string {
	return strings.Repeat(" ", max(0, (w-lipgloss.Width(text))/2)) + text
}
