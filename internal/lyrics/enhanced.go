package lyrics

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"net/url"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"time"
)

func durationsFirst(starts []time.Duration) time.Duration {
	if len(starts) > 0 {
		return starts[0]
	}
	return 0
}

var inlineTime = regexp.MustCompile(`<([0-9]+:[0-9]+(?:\.[0-9]+)?)>`)

func parseEnhancedLRC(text string, offset time.Duration) (string, []Word) {
	matches := inlineTime.FindAllStringSubmatchIndex(text, -1)
	if len(matches) == 0 {
		return text, nil
	}
	var words []Word
	var out strings.Builder
	out.WriteString(text[:matches[0][0]])
	for i, m := range matches {
		start, err := parseLRCTimestamp(text[m[2]:m[3]])
		if err != nil {
			continue
		}
		end := len(text)
		if i+1 < len(matches) {
			end = matches[i+1][0]
		}
		word := text[m[1]:end]
		if len(words) > 0 {
			words[len(words)-1].End = start + offset
		}
		out.WriteString(word)
		if word != "" {
			words = append(words, Word{Start: start + offset, Text: word})
		}
	}
	return out.String(), words
}

var yrcLine = regexp.MustCompile(`^\[(\d+),(\d+)\]`)
var yrcWord = regexp.MustCompile(`\((\d+),(\d+),\d+\)`)

func parseYRC(text string) []Line {
	var lines []Line
	speaker := ""
	for raw := range strings.SplitSeq(text, "\n") {
		header := yrcLine.FindStringSubmatch(raw)
		if header == nil {
			continue
		}
		start, _ := strconv.ParseInt(header[1], 10, 64)
		duration, _ := strconv.ParseInt(header[2], 10, 64)
		if start < 0 || duration <= 0 || start > 24*60*60*1000 || duration > 24*60*60*1000 {
			continue
		}
		line := Line{Start: time.Duration(start) * time.Millisecond, End: time.Duration(start+duration) * time.Millisecond}
		matches := yrcWord.FindAllStringSubmatchIndex(raw, -1)
		var plain strings.Builder
		for i, m := range matches {
			begin, _ := strconv.ParseInt(raw[m[2]:m[3]], 10, 64)
			length, _ := strconv.ParseInt(raw[m[4]:m[5]], 10, 64)
			if begin < start || length < 0 || begin > start+duration {
				continue
			}
			end := len(raw)
			if i+1 < len(matches) {
				end = matches[i+1][0]
			}
			word := raw[m[1]:end]
			plain.WriteString(word)
			line.Words = append(line.Words, Word{Start: time.Duration(begin) * time.Millisecond, End: time.Duration(begin+length) * time.Millisecond, Text: word})
		}
		label, body := splitSpeaker(plain.String())
		if label != "" {
			speaker = label
			line.Words = nil
		} // never retain label offsets as sung text
		line.Text, line.Speaker = body, speaker
		if line.Text != "" {
			lines = append(lines, line)
		}
	}
	slices.SortStableFunc(lines, func(a, b Line) int {
		if a.Start < b.Start {
			return -1
		}
		if a.Start > b.Start {
			return 1
		}
		return 0
	})
	return lines
}

// The public lyric endpoint needs no login, playback URLs, or encrypted client.
// Reuse the existing HTTP client and standard JSON/form decoding. Only accept
// a title, credited artist and duration match before fetching the lyric record.
func (c *Client) fetchEnhanced(ctx context.Context, artist, title string, duration time.Duration) (*Result, error) {
	if duration <= 0 {
		return nil, ErrNotFound
	}
	base := c.enhancedURL
	if base == "" {
		base = "https://music.163.com"
	}
	title = baseTitle(title)
	var search struct {
		Result struct {
			Songs []struct {
				ID       int64
				Name     string
				Duration int64
				Artists  []struct{ Name string }
			}
		}
		Code int
	}
	q := url.Values{"s": {artist + " " + title}, "type": {"1"}, "limit": {"10"}}
	if err := c.enhancedRequest(ctx, http.MethodGet, base+"/api/search/get/web?"+q.Encode(), nil, &search); err != nil {
		return nil, err
	}
	var id int64
	for _, song := range search.Result.Songs {
		if !matchingTitle(song.Name, title) || math.Abs(float64(song.Duration)/1000-duration.Seconds()) > 3 {
			continue
		}
		credited := false
		for _, a := range song.Artists {
			if strings.EqualFold(strings.TrimSpace(a.Name), strings.TrimSpace(artist)) {
				credited = true
			}
		}
		if credited {
			id = song.ID
			break
		}
	}
	if id == 0 {
		return nil, ErrNotFound
	}
	var response struct {
		Code int
		LRC  struct{ Lyric string }
		YRC  struct{ Lyric string }
	}
	body := url.Values{"id": {strconv.FormatInt(id, 10)}, "lv": {"-1"}, "kv": {"-1"}, "tv": {"-1"}, "rv": {"-1"}, "yv": {"-1"}}.Encode()
	if err := c.enhancedRequest(ctx, http.MethodPost, base+"/api/song/lyric/v1", strings.NewReader(body), &response); err != nil {
		return nil, err
	}
	if response.Code != 200 {
		return nil, ErrNotFound
	}
	if lines := parseYRC(response.YRC.Lyric); len(lines) > 0 {
		return &Result{Lines: lines, Synced: true}, nil
	}
	return (lyricRecord{SyncedLyrics: response.LRC.Lyric}).result()
}
func (c *Client) enhancedRequest(ctx context.Context, method, address string, body io.Reader, target any) error {
	req, err := http.NewRequestWithContext(ctx, method, address, body)
	if err != nil {
		return err
	}
	req.Header.Set("Referer", "https://music.163.com/")
	req.Header.Set("User-Agent", "vibez")
	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("enhanced lyrics: status %d", resp.StatusCode)
	}
	return json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(target)
}
