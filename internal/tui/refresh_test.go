package tui

import (
	"github.com/godbus/dbus/v5"
	"testing"
	"time"
)

func TestRefreshRateUsesCurrentModeAndCaps120(t *testing.T) {
	modes := []displayMode{{Hz: 60, Properties: map[string]dbus.Variant{"is-current": dbus.MakeVariant(false)}}, {Hz: 120.001, Properties: map[string]dbus.Variant{"is-current": dbus.MakeVariant(true)}}}
	if got := currentRefreshRate([]displayMonitor{{Modes: modes}}); got != 120 {
		t.Fatalf("got %d", got)
	}
	modes[1].Hz = 240
	if got := currentRefreshRate([]displayMonitor{{Modes: modes}}); got != 120 {
		t.Fatalf("got %d", got)
	}
	if currentRefreshRate(nil) != 60 {
		t.Fatal("missing monitor fallback")
	}
	m := Model{frameRate: 120}
	if m.frameInterval() != time.Second/120 {
		t.Fatal("incorrect interval")
	}
}
func TestNativeDisplayRefreshProbe(t *testing.T) {
	if got := detectRefreshRate()().(refreshRateMsg); got < 30 || got > 120 {
		t.Fatalf("invalid detected rate %d", got)
	}
	t.Logf("detected refresh: %d Hz", detectRefreshRate()())
}
