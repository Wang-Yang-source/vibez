package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
)

func TestSearchCursorAnchorsIMEToInput(t *testing.T) {
	m := newModel(nil)
	m.hideHints = true
	m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	m.introStep = introDone
	m.mode = modeSearch
	m.searchQuery = "周杰伦 album"
	m.searchCursor = 3
	v := m.View()
	if v.Cursor == nil {
		t.Fatal("real cursor missing; IME attaches to stale terminal position")
	}
	if v.Cursor.X != 11 || v.Cursor.Y != m.nowPlayingHeight()+4 {
		t.Fatalf("cursor at %v; want search input column 11 row %d", v.Cursor.Position, m.nowPlayingHeight()+4)
	}
	m.mode = modeNormal
	if m.View().Cursor != nil {
		t.Fatal("normal mode must hide editing cursor")
	}
}

func TestVibeCursorAnchorsIMEToChineseInput(t *testing.T) {
	for _, width := range []int{60, 100, 140} {
		for _, hidden := range []bool{false, true} {
			m := newModel(nil)
			m.hideHints = hidden
			m.Update(tea.WindowSizeMsg{Width: width, Height: 40})
			m.introStep = introDone
			m.mode = modeNormal
			m.vibe.Focus()
			m.vibe.Update(tea.KeyPressMsg{Code: '周', Text: "周杰伦"})
			v := m.View()
			split := (width - 2) / 2
			if v.Cursor == nil || v.Cursor.X != split+11 || v.Cursor.Y != m.nowPlayingHeight()+7 {
				t.Fatalf("width %d: cursor %v", width, v.Cursor)
			}
			m.vibe.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
			v = m.View()
			if v.Cursor.X != split+9 {
				t.Fatalf("mid-text Chinese cursor: %v", v.Cursor)
			}
			m.vibe.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
			if m.View().Cursor != nil {
				t.Fatal("cursor remained after leaving Vibe input")
			}
		}
	}
}
