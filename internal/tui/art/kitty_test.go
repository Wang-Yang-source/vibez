package art

import (
	"bytes"
	"compress/zlib"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/x/ansi"
)

func TestKittyPreservesResolutionAndCellWidths(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 512, 512))
	size := Size{32, 16}
	data, err := KittyUpload(src, 0x560001, size)
	if err != nil {
		t.Fatal(err)
	}
	var encoded strings.Builder
	for _, chunk := range strings.Split(data, "\x1b_G")[1:] {
		head, body, ok := strings.Cut(chunk, ";")
		if !ok {
			t.Fatal("missing graphics payload")
		}
		_ = head
		payload, _, _ := strings.Cut(body, "\x1b\\")
		encoded.WriteString(payload)
	}
	raw, err := base64.StdEncoding.DecodeString(encoded.String())
	if err != nil {
		t.Fatal(err)
	}
	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil || img.Bounds() != src.Bounds() {
		t.Fatalf("image resolution changed: %v", err)
	}
	for _, line := range KittyLines(0x560001, size) {
		if ansi.StringWidth(line) != size.Width {
			t.Fatalf("incorrect placeholder width: %d", ansi.StringWidth(line))
		}
	}
	if !strings.Contains(data, "U=1") || !strings.Contains(data, "q=2") {
		t.Fatal("must use silent virtual placement")
	}
}

func TestAnimationTransferPreservesStraightAlphaAndChunking(t *testing.T) {
	src := image.NewNRGBA(image.Rect(0, 0, 160, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 160; x++ {
			src.SetNRGBA(x, y, color.NRGBA{R: uint8(x*17 + y), G: uint8(x + y*7), B: uint8(x*3 + y*13), A: uint8(x + y)})
		}
	}
	data, err := KittyUploadAnimation(src, 0x570001, Size{40, 20})
	if err != nil {
		t.Fatal(err)
	}
	var encoded strings.Builder
	chunks := strings.Split(data, "\x1b_G")[1:]
	for _, chunk := range chunks {
		_, body, ok := strings.Cut(chunk, ";")
		if !ok {
			t.Fatal("missing payload")
		}
		payload, _, _ := strings.Cut(body, "\x1b\\")
		if len(payload) > 4096 {
			t.Fatal("oversized graphics chunk")
		}
		encoded.WriteString(payload)
	}
	payload, err := base64.StdEncoding.DecodeString(encoded.String())
	if err != nil {
		t.Fatal(err)
	}
	reader, err := zlib.NewReader(bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, src.Pix) {
		t.Fatal("lossy color or alpha transfer")
	}
	if len(chunks) < 2 || strings.Contains(data, "f=100") || strings.Contains(data, "f=24") || !strings.Contains(data, "o=z") || !strings.Contains(data, "s=160") || !strings.Contains(data, "v=100") || !strings.Contains(chunks[len(chunks)-1], "m=0") {
		t.Fatal("incorrect chunked RGBA protocol")
	}
}

func TestAnimationTransferUnpremultipliesTransparentRGBA(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 1))
	src.SetRGBA(0, 0, color.RGBA{R: 128, A: 128})
	data, err := KittyUploadAnimation(src, 1, Size{2, 1})
	if err != nil {
		t.Fatal(err)
	}
	var encoded strings.Builder
	for _, chunk := range strings.Split(data, "\x1b_G")[1:] {
		_, body, _ := strings.Cut(chunk, ";")
		payload, _, _ := strings.Cut(body, "\x1b\\")
		encoded.WriteString(payload)
	}
	compressed, err := base64.StdEncoding.DecodeString(encoded.String())
	if err != nil {
		t.Fatal(err)
	}
	reader, err := zlib.NewReader(bytes.NewReader(compressed))
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	raw, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, []byte{255, 0, 0, 128, 0, 0, 0, 0}) {
		t.Fatalf("incorrect straight alpha pixels: %v", raw)
	}
}
