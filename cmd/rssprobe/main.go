//go:build !ignore

package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gen2brain/aac-go"
	"github.com/youpy/go-wav"
)

func main() {
	wavPath := flag.String("wav", "testdata/test.wav", "input wav file")
	duration := flag.Duration("duration", 30*time.Second, "how long to run the encode loop")
	reportEvery := flag.Duration("report-every", time.Second, "stats print interval")
	frameBytes := flag.Int("frame-bytes", 4096, "input bytes per EncodeOneFrame call")
	flag.Parse()

	if *frameBytes <= 0 {
		fmt.Fprintln(os.Stderr, "frame-bytes must be > 0")
		os.Exit(2)
	}

	pcm, opts, err := loadPCM(*wavPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "load wav failed: %v\n", err)
		os.Exit(1)
	}

	enc, err := aac.NewEncoderV2(opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "new encoder failed: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if cerr := enc.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "encoder close failed: %v\n", cerr)
		}
	}()

	fmt.Printf("pid=%d wav=%s sampleRate=%d channels=%d pcmBytes=%d frameBytes=%d duration=%s\n",
		os.Getpid(), *wavPath, opts.SampleRate, opts.NumChannels, len(pcm), *frameBytes, *duration)

	start := time.Now()
	nextReport := start.Add(*reportEvery)
	end := start.Add(*duration)

	var rounds int64
	var encodeCalls int64
	var outBytes int64

	for time.Now().Before(end) {
		rounds++
		for i := 0; i < len(pcm); i += *frameBytes {
			j := i + *frameBytes
			if j > len(pcm) {
				j = len(pcm)
			}

			frames, eerr := enc.EncodeOneFrame(pcm[i:j])
			if eerr != nil {
				fmt.Fprintf(os.Stderr, "encode failed: %v\n", eerr)
				os.Exit(1)
			}
			encodeCalls++
			for _, f := range frames {
				outBytes += int64(len(f))
			}
		}

		now := time.Now()
		if now.After(nextReport) {
			printStats(now.Sub(start), rounds, encodeCalls, outBytes)
			nextReport = now.Add(*reportEvery)
		}
	}

	printStats(time.Since(start), rounds, encodeCalls, outBytes)
}

func loadPCM(path string) ([]byte, *aac.Options, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()

	r := wav.NewReader(f)
	format, err := r.Format()
	if err != nil {
		return nil, nil, err
	}

	pcm := make([]byte, 0, 64*1024)
	buf := make([]byte, 8192)
	for {
		n, rerr := r.Read(buf)
		if n > 0 {
			pcm = append(pcm, buf[:n]...)
		}
		if errors.Is(rerr, io.EOF) {
			break
		}
		if rerr != nil {
			return nil, nil, rerr
		}
	}

	opts := &aac.Options{
		SampleRate:  int(format.SampleRate),
		NumChannels: int(format.NumChannels),
	}

	return pcm, opts, nil
}

func printStats(elapsed time.Duration, rounds, encodeCalls, outBytes int64) {
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)

	rssKB := currentRSSKB(os.Getpid())
	fmt.Printf("elapsed=%s rounds=%d calls=%d outBytes=%d goAlloc=%dKB goSys=%dKB rss=%dKB\n",
		elapsed.Truncate(time.Millisecond),
		rounds,
		encodeCalls,
		outBytes,
		ms.Alloc/1024,
		ms.Sys/1024,
		rssKB,
	)
}

func currentRSSKB(pid int) int64 {
	cmd := exec.Command("ps", "-o", "rss=", "-p", strconv.Itoa(pid))
	out, err := cmd.Output()
	if err != nil {
		return -1
	}

	v := strings.TrimSpace(string(out))
	rssKB, err := strconv.ParseInt(v, 10, 64)
	if err != nil {
		return -1
	}

	return rssKB
}
