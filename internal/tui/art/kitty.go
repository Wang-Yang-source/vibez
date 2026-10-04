package art

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"image"
	"os"
	"strings"
	"sync"

	"github.com/charmbracelet/x/ansi"
	"github.com/charmbracelet/x/ansi/kitty"
)

// SupportsKitty limits automatic graphics to terminals with known support.
func SupportsKitty() bool {
	if os.Getenv("TMUX") != "" {
		return false
	}
	return os.Getenv("TERM") == "xterm-ghostty" || os.Getenv("TERM") == "xterm-kitty" || os.Getenv("TERM_PROGRAM") == "ghostty"
}

// KittyUpload preserves the decoded image resolution; the terminal scales it.
func KittyUpload(img image.Image, id int, size Size) (string, error) {
	var buf bytes.Buffer
	err := kitty.EncodeGraphics(&buf, img, &kitty.Options{
		Action: kitty.TransmitAndPut, Format: kitty.PNG, Transmission: kitty.Direct,
		ID: id, Quite: 2, Chunk: true, VirtualPlacement: true,
		Columns: size.Width, Rows: size.Height, DoNotMoveCursor: true,
	})
	return buf.String(), err
}

func KittyPlacement(id int, size Size) string {
	return ansi.KittyGraphics(nil, (&kitty.Options{
		Action: kitty.Put, ID: id, Quite: 2, VirtualPlacement: true,
		Columns: size.Width, Rows: size.Height, DoNotMoveCursor: true,
	}).Options()...)
}

func KittyDelete(id int) string {
	return ansi.KittyGraphics(nil, (&kitty.Options{
		Action: kitty.Delete, ID: id, Delete: kitty.DeleteID, DeleteResources: true, Quite: 2,
	}).Options()...)
}

// KittyLines are normal text cells, so the TUI can center, clip and move them.
func KittyLines(id int, size Size) []string {
	lines := make([]string, size.Height)
	color := fmt.Sprintf("\x1b[38;2;%d;%d;%dm", (id>>16)&255, (id>>8)&255, id&255)
	for row := range size.Height {
		var s strings.Builder
		s.WriteString(color)
		for col := range size.Width {
			s.WriteRune(kitty.Placeholder)
			s.WriteRune(kitty.Diacritic(row))
			s.WriteRune(kitty.Diacritic(col))
		}
		s.WriteString("\x1b[0m")
		lines[row] = s.String()
	}
	return lines
}

var animationCompressors = sync.Pool{New: func() any { w, _ := zlib.NewWriterLevel(&bytes.Buffer{}, zlib.BestSpeed); return w }}

// KittyUploadAnimation uses the standard library's fast lossless compressor.
// x/ansi's image encoder offers neither a compression level nor a packed-pixel
// fast path; retain its protocol options and framing instead of encoding PNG.
func KittyUploadAnimation(img image.Image, id int, size Size) (string, error) {
	var compressed bytes.Buffer
	zw := animationCompressors.Get().(*zlib.Writer)
	zw.Reset(&compressed)
	defer animationCompressors.Put(zw)
	var err error
	bounds := img.Bounds()
	var pixels []byte
	var stride int
	switch packed := img.(type) {
	case *image.NRGBA:
		pixels, stride = packed.Pix, packed.Stride
	case *image.RGBA:
		if !packed.Opaque() {
			return "", fmt.Errorf("animation RGBA canvas must be opaque")
		}
		pixels, stride = packed.Pix, packed.Stride
	default:
		return "", fmt.Errorf("animation requires packed pixels")
	}
	if stride == bounds.Dx()*4 {
		_, err = zw.Write(pixels[:stride*bounds.Dy()])
	} else {
		for y := 0; y < bounds.Dy(); y++ {
			if _, err = zw.Write(pixels[y*stride : y*stride+bounds.Dx()*4]); err != nil {
				break
			}
		}
	}
	if err != nil {
		return "", err
	}

	if err = zw.Close(); err != nil {
		return "", err
	}
	payload := base64.StdEncoding.EncodeToString(compressed.Bytes())
	opts := (&kitty.Options{Action: kitty.TransmitAndPut, Format: kitty.RGBA,
		Compression: kitty.Zlib, Transmission: kitty.Direct, ID: id, Quite: 2,
		VirtualPlacement: true, Columns: size.Width, Rows: size.Height,
		ImageWidth: bounds.Dx(), ImageHeight: bounds.Dy(), DoNotMoveCursor: true}).Options()
	var out strings.Builder
	for start := 0; start < len(payload); start += kitty.MaxChunkSize {
		end := min(start+kitty.MaxChunkSize, len(payload))
		chunkOpts := []string{"q=2"}
		if start == 0 {
			chunkOpts = append([]string(nil), opts...)
		}
		if end < len(payload) {
			chunkOpts = append(chunkOpts, "m=1")
		} else {
			chunkOpts = append(chunkOpts, "m=0")
		}
		out.WriteString(ansi.KittyGraphics([]byte(payload[start:end]), chunkOpts...))
	}
	return out.String(), nil
}
