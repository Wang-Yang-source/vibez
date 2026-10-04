//go:build !linux && !darwin

package tui

// cellAspect returns the conventional cell height/width ratio on platforms
// without a TIOCGWINSZ query. See the unix build for details.
func cellAspect() float64 { return 2.0 }

func terminalCellPixels() (float64, float64) { return 24, 48 }
