package tui

import (
	"time"

	"github.com/simone-vibes/vibez/internal/player"
	"github.com/simone-vibes/vibez/internal/tui/views"
)

// A monotonic local clock fills the gaps between coarse engine observations.
// Small timing corrections are eased in; seeks and track changes reset it.
type lyricClock struct {
	anchor                                   time.Time
	position, reported, correction, duration time.Duration
	track                                    string
	running                                  bool
}

func (c *lyricClock) at(now time.Time) time.Duration {
	position := c.position
	if c.running && !c.anchor.IsZero() {
		elapsed := max(0, now.Sub(c.anchor))
		// Stop extrapolating after a missing engine heartbeat, never after 250 ms.
		elapsed = min(elapsed, 2*time.Second)
		t := min(1, elapsed.Seconds())
		position += elapsed + time.Duration(float64(c.correction)*t*t*(3-2*t))
	}
	position = max(0, position)
	if c.duration > 0 {
		position = min(position, c.duration)
	}
	return position
}
func (c *lyricClock) observe(s player.State, now time.Time) {
	track := ""
	duration := time.Duration(0)
	if s.Track != nil {
		track = views.PlaybackID(*s.Track)
		duration = s.Track.Duration
	}
	running := s.Playing && !s.Loading
	if c.anchor.IsZero() || track != c.track {
		*c = lyricClock{anchor: now, position: s.Position, reported: s.Position, track: track, duration: duration, running: running}
		return
	}
	current := c.at(now)
	if running != c.running {
		position := current
		if delta := s.Position - current; s.Position != c.reported && (delta > 500*time.Millisecond || delta < -500*time.Millisecond) {
			position = s.Position
		}
		c.anchor, c.position, c.correction, c.running = now, position, 0, running
	} else if s.Position != c.reported {
		delta := s.Position - current
		c.anchor = now
		if !running || delta > 500*time.Millisecond || delta < -500*time.Millisecond {
			c.position, c.correction = s.Position, 0
		} else {
			c.position, c.correction = current, delta
		}
	}
	// Duplicate volume/log/status reports must not restart interpolation.
	c.reported, c.duration = s.Position, duration
}
func (m *Model) lyricPosition(now time.Time) time.Duration {
	if m.lyricClock.anchor.IsZero() {
		return m.playerState.Position
	}
	return m.lyricClock.at(now)
}
