package views

import (
	"fmt"
	"github.com/simone-vibes/vibez/internal/tui/locale"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/simone-vibes/vibez/internal/tui/styles"
	"github.com/simone-vibes/vibez/internal/version"
)

// AboutModel renders application and author information.
type AboutModel struct {
	Locale locale.Locale
	width  int
	height int
	status string
}

func NewAbout() *AboutModel {
	return &AboutModel{}
}

func (a *AboutModel) SetSize(w, h int) {
	a.width = w
	a.height = h
}

func (a *AboutModel) Update(_ tea.KeyPressMsg) tea.Cmd {
	return nil
}

func (a *AboutModel) View() string {
	muted := styles.QueueItemMuted
	normal := lipgloss.NewStyle().Foreground(styles.ColorFg)
	header := styles.TabActive
	primary := lipgloss.NewStyle().Foreground(styles.ColorPrimary).Bold(true)
	secondary := lipgloss.NewStyle().Foreground(styles.ColorSecondary)

	var sb strings.Builder
	sb.WriteString(header.Render(a.Locale.Text("About")) + "\n")
	sb.WriteString(muted.Render(strings.Repeat("─", 5)) + "\n\n")

	contentLines := []string{
		primary.Render("vibez ♪"),
		muted.Render(fmt.Sprintf(a.Locale.Text("version %s"), version.Version)),
		"",
		normal.Render(a.Locale.Text("Apple Music in your terminal.")),
		normal.Render(a.Locale.Text("Vibe-driven. Keyboard-first.")),
		"",
		secondary.Render(a.Locale.Text("made with ❤️ by simonepelosi")),
		"",
	}

	if a.status != "" {
		contentLines = append(contentLines, "", styles.ControlActive.Render(a.status))
	}

	neededHeight := len(contentLines)
	topPad := max(0, (a.height-3-neededHeight)/2)

	for range topPad {
		sb.WriteByte('\n')
	}

	for _, line := range contentLines {
		sb.WriteString(centerStrAbout(line, a.width) + "\n")
	}

	return sb.String()
}

func centerStrAbout(s string, width int) string {
	w := lipgloss.Width(s)
	pad := max(0, (width-w)/2)
	return strings.Repeat(" ", pad) + s
}
