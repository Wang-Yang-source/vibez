package tui

import (
	"context"
	"math"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

type displayMode struct {
	ID            string
	Width, Height int32
	Hz, Scale     float64
	Scales        []float64
	Properties    map[string]dbus.Variant
}
type displayMonitor struct {
	Spec       struct{ Connector, Vendor, Product, Serial string }
	Modes      []displayMode
	Properties map[string]dbus.Variant
}
type refreshRateMsg int

// Reuse Mutter's display API and the existing D-Bus dependency. Other desktops
// retain a conservative 60 Hz fallback; cap the terminal animation at 120 Hz.
func detectRefreshRate() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		conn, err := dbus.ConnectSessionBus()
		if err != nil {
			return refreshRateMsg(60)
		}
		defer conn.Close()
		call := conn.Object("org.gnome.Mutter.DisplayConfig", "/org/gnome/Mutter/DisplayConfig").CallWithContext(ctx, "org.gnome.Mutter.DisplayConfig.GetCurrentState", 0)
		if call.Err != nil || len(call.Body) < 2 {
			return refreshRateMsg(60)
		}
		var monitors []displayMonitor
		if err := dbus.Store([]interface{}{call.Body[1]}, &monitors); err != nil {
			return refreshRateMsg(60)
		}
		return refreshRateMsg(currentRefreshRate(monitors))
	}
}
func currentRefreshRate(monitors []displayMonitor) int {
	hz := 0.0
	for _, monitor := range monitors {
		for _, mode := range monitor.Modes {
			current, ok := mode.Properties["is-current"]
			if ok && current.Value() == true && !math.IsNaN(mode.Hz) && !math.IsInf(mode.Hz, 0) {
				hz = max(hz, mode.Hz)
			}
		}
	}
	if hz <= 0 {
		return 60
	}
	return min(120, max(30, int(math.Round(hz))))
}
func (m *Model) frameInterval() time.Duration {
	hz := m.frameRate
	if hz <= 0 {
		hz = 60
	}
	return time.Second / time.Duration(hz)
}
func (m *Model) animationTick() tea.Cmd {
	return tea.Tick(m.frameInterval(), func(t time.Time) tea.Msg { return glowTickMsg(t) })
}
