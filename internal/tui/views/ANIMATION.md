Animation design reference:
https://github.com/Steve-xmh/applemusic-like-lyrics/tree/main/packages/core/src/lyric-player
AMLL separates lyric time bounds from scroll motion and supports overlapping
vocal lines. Its DOM/Web Animation implementation is AGPL-3.0-only; no source
code is copied or bundled here. Vibez keeps its existing Go/Kitty renderer.

Scrolling reuses github.com/charmbracelet/harmonica v0.2.0 (MIT, with its retained
spring implementation notice), pinned with hashes in go.mod/go.sum. Go's
standard library has no spring animator. Harmonica is pure Go, requires Go 1.16,
has no runtime dependencies, and accepts the actual elapsed frame time. The
critically damped spring retains velocity when targets change without overshoot.
Word highlights remain driven by playback time, never spring animation time;
line-only lyric data continues to use an explicitly approximate word sweep.
