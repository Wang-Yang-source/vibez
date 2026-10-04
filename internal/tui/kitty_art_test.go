package tui

import (
	"github.com/charmbracelet/x/ansi"
	"image"
	"strings"
	"testing"
	"time"

	"github.com/simone-vibes/vibez/internal/provider"
)

func TestGhosttyArtworkUsesImagePlaceholders(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	t.Setenv("TMUX", "")
	m := newModel(nil)
	m.artMode = true
	m.supportsArtGraphics = func() bool { return true }
	m.supportsArtColor = func() bool { return true }
	m.artwork.url = "https://example.invalid/cover.png"
	m.artwork.img = image.NewNRGBA(image.Rect(0, 0, 512, 512))
	m.playerState.Track = &provider.Track{ArtworkURL: m.artwork.url}
	lines := strings.Join(m.nowPlayingArtLines(80, 20), "\n")
	if !strings.ContainsRune(lines, '\U0010EEEE') || strings.ContainsRune(lines, '▀') {
		t.Fatal("Ghostty cover still uses pixelated character blocks")
	}
}

func TestArtworkGraphicsLifecycle(t *testing.T) {
	m := newModel(nil)
	m.supportsArtGraphics = func() bool { return true }
	m.supportsArtColor = func() bool { return true }
	m.artMode = true
	m.width, m.height = 100, 40
	m.artwork.img = image.NewNRGBA(image.Rect(0, 0, 512, 512))
	m.artwork.url = "cover"
	m.playerState.Track = &provider.Track{ArtworkURL: "cover"}
	if m.syncArtworkGraphics() == nil {
		t.Fatal("first cover must upload")
	}
	if m.syncArtworkGraphics() != nil {
		t.Fatal("unchanged cover uploaded again")
	}
	m.width = 20
	if m.syncArtworkGraphics() == nil {
		t.Fatal("resize must update placement")
	}
	m.artworkGen++
	if m.syncArtworkGraphics() == nil {
		t.Fatal("new cover must replace old image")
	}
	m.artMode = false
	if m.syncArtworkGraphics() == nil || m.artGraphics.id != 0 {
		t.Fatal("hidden cover resources not released")
	}
}

func TestSplitCoverMetadataUsesSeparateLeftAlignedRows(t *testing.T) {
	m := newModel(nil)
	m.width = 160
	m.inlineLyrics = true
	m.playerState.Track = &provider.Track{Title: "Test title", Artist: "Test artist", Duration: 3 * time.Minute}
	rows := m.nowPlayingArtLines(60, 25)
	if len(rows) != 25 {
		t.Fatalf("metadata changes panel height: %d", len(rows))
	}
	title, artist := ansi.Strip(rows[21]), ansi.Strip(rows[22])
	if strings.TrimSpace(title) != "Test title" || strings.TrimSpace(artist) != "Test artist" {
		t.Fatal("title and artist must use separate rows")
	}
	if len(title)-len(strings.TrimLeft(title, " ")) != len(artist)-len(strings.TrimLeft(artist, " ")) {
		t.Fatal("metadata left edges differ")
	}
}
