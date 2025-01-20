package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/gen2brain/aac-go"
	"github.com/youpy/go-wav"
)

func main() {
	file, err := os.Open("testdata/test.wav")
	if err != nil {
		panic(err)
	}

	wreader := wav.NewReader(file)
	f, err := wreader.Format()
	if err != nil {
		panic(err)
	}

	opts := &aac.Options{}
	opts.SampleRate = int(f.SampleRate)
	opts.NumChannels = int(f.NumChannels)

	enc, err := aac.NewEncoderV2(opts)
	if err != nil {
		panic(err)
	}
	outputFile, err := os.OpenFile("testdata/test.aac", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer outputFile.Close()
	inbuf := make([]byte, opts.NumChannels*2*1024)
	for {
		n, err := wreader.Read(inbuf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return
			}
		}
		if n == 0 {
			break
		}
		inFrame := inbuf[:n]
		outputFrame, err := enc.EncodeOneFrame(inFrame)
		if err != nil {
			panic(err)
		}
		for i := 0; i < len(outputFrame); i++ {
			outputFile.Write(outputFrame[i])
			if err != nil {
				panic(err)
			}
		}
		if errors.Is(err, io.EOF) {
			break
		}
	}
	err = enc.Close()
	if err != nil {
		panic(err)
	}
}
