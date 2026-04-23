//go:build !ignore

//nolint:goanalysis_metalinter

// Package aac provides AAC codec encoder based on VisualOn AAC encoder library.
package aac

//#include <stdlib.h>
//#include <string.h>
import "C"

import (
	"errors"
	"fmt"
	"unsafe"

	"github.com/gen2brain/aac-go/aacenc"
)

// Options represent encoding options.
type Options struct {
	// Audio file sample rate
	SampleRate int
	// Encoder bit rate in bits/sec
	BitRate int
	// Number of channels on input (1,2)
	NumChannels int
}

// Encoder type.
type Encoder struct {
	handle   aacenc.VoHandle
	inputBuf unsafe.Pointer
	inputCap int
	outBuf   unsafe.Pointer
	outCap   int
	closed   bool
}

// NewEncoder returns new AAC encoder.
func NewEncoderV2(opts *Options) (*Encoder, error) {
	if opts == nil {
		return nil, fmt.Errorf("aac: options is nil")
	}

	e := &Encoder{}
	if opts.BitRate == 0 {
		opts.BitRate = 64000
	}

	handle, ret := aacenc.Init(aacenc.VoAudioCodingAac)
	err := aacenc.ErrorFromResult(ret)
	if err != nil {
		return nil, fmt.Errorf("aac: %w", err)
	}

	var params aacenc.Param
	params.SampleRate = int32(opts.SampleRate)
	params.BitRate = int32(opts.BitRate)
	params.NChannels = int16(opts.NumChannels)
	params.AdtsUsed = 1
	e.handle = handle

	ret = aacenc.SetParam(handle, aacenc.VoPidAacEncparam, unsafe.Pointer(&params))
	err = aacenc.ErrorFromResult(ret)
	if err != nil {
		e.Close()
		return nil, fmt.Errorf("aac: %w", err)
	}
	return e, nil
}

// Encode encodes data from inbuf.
func (e *Encoder) EncodeOneFrame(inbuf []byte) ([][]byte, error) {
	if e == nil {
		return nil, fmt.Errorf("aac: encoder is nil")
	}
	if e.closed {
		return nil, fmt.Errorf("aac: encoder is closed")
	}
	var outinfo aacenc.VoAudioOutputinfo
	var input, output aacenc.VoCodecBuffer

	if err := e.ensureInputBuffer(len(inbuf)); err != nil {
		return nil, err
	}
	if len(inbuf) > 0 {
		inSlice := unsafe.Slice((*byte)(e.inputBuf), len(inbuf))
		copy(inSlice, inbuf)
	}
	input.Buffer = e.inputBuf
	if input.Buffer == nil {
		return nil, fmt.Errorf("aac: memory allocation failed")
	}
	input.Length = uint64(len(inbuf))
	ret := aacenc.SetInputData(e.handle, &input)
	err := aacenc.ErrorFromResult(ret)
	if err != nil {
		return nil, fmt.Errorf("aac: %w", err)
	}
	const bufferSize = 20480
	var outDataList [][]byte
	if err := e.ensureOutputBuffer(bufferSize); err != nil {
		return nil, err
	}
	output.Buffer = e.outBuf
	if output.Buffer == nil {
		return nil, fmt.Errorf("aac: memory allocation failed")
	}
	for {
		output.Length = uint64(e.outCap)
		ret = aacenc.GetOutputData(e.handle, &output, &outinfo)
		err = aacenc.ErrorFromResult(ret)
		if err != nil {
			if !errors.Is(err, aacenc.ErrInputBufferSmall) {
				return nil, fmt.Errorf("aac: %w", err)
			}
			break
		}

		outData := C.GoBytes(output.Buffer, C.int(output.Length))
		outDataList = append(outDataList, outData)
	}

	return outDataList, nil
}

func (e *Encoder) ensureInputBuffer(size int) error {
	if e.inputCap >= size && e.inputBuf != nil {
		return nil
	}

	if e.inputBuf != nil {
		C.free(e.inputBuf)
		e.inputBuf = nil
		e.inputCap = 0
	}

	allocSize := size
	if allocSize == 0 {
		allocSize = 1
	}

	e.inputBuf = C.malloc(C.size_t(allocSize))
	if e.inputBuf == nil {
		return fmt.Errorf("aac: memory allocation failed")
	}
	e.inputCap = size

	return nil
}

func (e *Encoder) ensureOutputBuffer(size int) error {
	if e.outCap >= size && e.outBuf != nil {
		return nil
	}

	if e.outBuf != nil {
		C.free(e.outBuf)
		e.outBuf = nil
		e.outCap = 0
	}

	e.outBuf = C.malloc(C.size_t(size))
	if e.outBuf == nil {
		return fmt.Errorf("aac: memory allocation failed")
	}
	e.outCap = size

	return nil
}

// Close closes encoder.
func (e *Encoder) Close() error {
	if e == nil || e.closed {
		return nil
	}
	e.closed = true

	if e.inputBuf != nil {
		C.free(e.inputBuf)
		e.inputBuf = nil
		e.inputCap = 0
	}

	if e.outBuf != nil {
		C.free(e.outBuf)
		e.outBuf = nil
		e.outCap = 0
	}

	if e.handle == nil {
		return nil
	}

	ret := aacenc.Uninit(e.handle)
	e.handle = nil
	return aacenc.ErrorFromResult(ret)
}
