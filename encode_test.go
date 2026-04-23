//go:build !ignore

package aac_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gen2brain/aac-go"
	"github.com/youpy/go-wav"
)

func TestEncode(t *testing.T) {
	file, err := os.Open(filepath.Join("testdata", "test.wav"))
	if err != nil {
		t.Fatal(err)
	}

	wr := wav.NewReader(file)
	f, err := wr.Format()
	if err != nil {
		t.Fatal(err)
	}

	buf := bytes.NewBuffer(make([]byte, 0))

	opts := &aac.Options{}
	opts.SampleRate = int(f.SampleRate)
	opts.NumChannels = int(f.NumChannels)

	enc, err := aac.NewEncoderV2(opts)
	if err != nil {
		t.Fatal(err)
	}

	data, err := enc.EncodeOneFrame(buf.Bytes())
	if err != nil {
		t.Error(err)
	}
	for _, frame := range data {
		buf.Write(frame)
	}

	err = enc.Close()
	if err != nil {
		t.Error(err)
	}

	err = os.WriteFile(filepath.Join(os.TempDir(), "test.aac"), buf.Bytes(), 0640)
	if err != nil {
		t.Error(err)
	}
}

func TestEncoderCloseIdempotent(t *testing.T) {
	opts := &aac.Options{SampleRate: 44100, NumChannels: 2}
	enc, err := aac.NewEncoderV2(opts)
	if err != nil {
		t.Fatal(err)
	}

	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}

	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestEncodeAfterClose(t *testing.T) {
	opts := &aac.Options{SampleRate: 44100, NumChannels: 2}
	enc, err := aac.NewEncoderV2(opts)
	if err != nil {
		t.Fatal(err)
	}

	if err := enc.Close(); err != nil {
		t.Fatal(err)
	}

	_, err = enc.EncodeOneFrame(nil)
	if err == nil {
		t.Fatal("expected error after close")
	}
	if !strings.Contains(err.Error(), "closed") {
		t.Fatalf("unexpected error: %v", err)
	}
}
