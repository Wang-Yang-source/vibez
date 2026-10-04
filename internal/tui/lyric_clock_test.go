package tui

import (
	"github.com/simone-vibes/vibez/internal/player"
	"github.com/simone-vibes/vibez/internal/provider"
	"testing"
	"time"
)

func clockState() player.State {
	return player.State{Track: &provider.Track{ID: "song", Duration: time.Minute}, Playing: true, Position: 10 * time.Second}
}
func TestLyricClockAdvancesAcrossOneSecondReports(t *testing.T) {
	start := time.Now()
	s := clockState()
	c := lyricClock{}
	c.observe(s, start)
	for i := 1; i < 120; i++ {
		now := start.Add(time.Duration(i) * time.Second / 120)
		// Volume and unrelated snapshots with the same coarse position occur often.
		if i%6 == 0 {
			s.Volume += .01
			c.observe(s, now)
		}
		want := 10*time.Second + now.Sub(start)
		if got := c.at(now); got != want {
			t.Fatalf("frame %d: got %v, want %v", i, got, want)
		}
	}
	s.Position = 11 * time.Second
	c.observe(s, start.Add(time.Second))
	if got := c.at(start.Add(1500 * time.Millisecond)); got != 11500*time.Millisecond {
		t.Fatalf("coarse update interrupted clock: %v", got)
	}
}
func TestLyricClockPauseBufferSeekAndTrackChange(t *testing.T) {
	start := time.Now()
	s := clockState()
	c := lyricClock{}
	c.observe(s, start)
	s.Playing = false
	c.observe(s, start.Add(800*time.Millisecond))
	if got := c.at(start.Add(2 * time.Second)); got != 10800*time.Millisecond {
		t.Fatalf("pause drifted: %v", got)
	}
	s.Playing = true
	c.observe(s, start.Add(3*time.Second))
	s.Position = 30 * time.Second
	c.observe(s, start.Add(4*time.Second))
	if c.at(start.Add(4*time.Second)) != 30*time.Second {
		t.Fatal("forward seek not immediate")
	}
	s.Position = 5 * time.Second
	c.observe(s, start.Add(5*time.Second))
	if c.at(start.Add(5*time.Second)) != 5*time.Second {
		t.Fatal("backward seek not immediate")
	}
	s.Loading = true
	c.observe(s, start.Add(5200*time.Millisecond))
	frozen := c.at(start.Add(5200 * time.Millisecond))
	if c.at(start.Add(7*time.Second)) != frozen {
		t.Fatal("buffering drifted")
	}
	s.Track = &provider.Track{ID: "new", Duration: time.Second}
	s.Loading = false
	s.Position = 0
	c.observe(s, start.Add(8*time.Second))
	if c.at(start.Add(8*time.Second)) != 0 || c.at(start.Add(10*time.Second)) != time.Second {
		t.Fatal("track reset/duration clamp failed")
	}
}
func TestLyricClockCorrectsJitterWithoutJump(t *testing.T) {
	start := time.Now()
	s := clockState()
	c := lyricClock{}
	c.observe(s, start)
	now := start.Add(1200 * time.Millisecond)
	before := c.at(now)
	s.Position = 11 * time.Second
	c.observe(s, now)
	if c.at(now) != before {
		t.Fatal("minor clock correction jumped backwards")
	}
	prev := before
	for i := 1; i <= 120; i++ {
		position := c.at(now.Add(time.Duration(i) * time.Second / 120))
		if position <= prev {
			t.Fatal("correction stalled or reversed")
		}
		prev = position
	}
	if prev != 12*time.Second {
		t.Fatalf("clock failed to settle: %v", prev)
	}
}
