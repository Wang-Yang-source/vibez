package lyrics

import (
	"github.com/rivo/uniseg"
	"time"
)

// Progress returns a completed UTF-8 prefix and the fraction of the next
// grapheme. Real word timings take precedence; line-only lyrics interpolate
// to the next phrase as requested, without splitting emoji or combining marks.
func Progress(line Line, until, position time.Duration) (int, float64) {
	if position < line.Start {
		return 0, 0
	}
	if len(line.Words) > 0 {
		offset := 0
		for i, word := range line.Words {
			end := word.End
			if end <= word.Start {
				if i+1 < len(line.Words) {
					end = line.Words[i+1].Start
				} else {
					end = until
				}
			}
			if position < word.Start {
				return offset, 0
			}
			if position >= end {
				offset += len(word.Text)
				continue
			}
			n, f := textProgress(word.Text, word.Start, end, position)
			return offset + n, f
		}
		return len(line.Text), 0
	}
	if line.End > line.Start {
		until = line.End
	}
	return textProgress(line.Text, line.Start, until, position)
}
func textProgress(text string, start, end, position time.Duration) (int, float64) {
	if end <= start || position >= end {
		return len(text), 0
	}
	if position <= start {
		return 0, 0
	}
	total := uniseg.GraphemeClusterCount(text)
	progress := float64(position-start) / float64(end-start) * float64(total)
	complete := int(progress)
	g := uniseg.NewGraphemes(text)
	offset := 0
	for i := 0; i < complete && g.Next(); i++ {
		_, offset = g.Positions()
	}
	return offset, progress - float64(complete)
}
