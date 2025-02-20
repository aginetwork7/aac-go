//go:build !ignore

//nolint:goanalysis_metalinter

// Package aac provides AAC codec encoder based on VisualOn AAC encoder library.
package aac

//#include <stdlib.h>
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
	handleID int
}

// NewEncoder returns new AAC encoder.
func NewEncoderV2(opts *Options) (*Encoder, error) {
	e := &Encoder{}
	if opts.BitRate == 0 {
		opts.BitRate = 64000
	}

	handleID, ret := aacenc.Init(aacenc.VoAudioCodingAac)
	err := aacenc.ErrorFromResult(ret)
	if err != nil {
		return nil, fmt.Errorf("aac: %w", err)
	}

	var params aacenc.Param
	params.SampleRate = int32(opts.SampleRate)
	params.BitRate = int32(opts.BitRate)
	params.NChannels = int16(opts.NumChannels)
	params.AdtsUsed = 1

	ret = aacenc.SetParam(handleID, aacenc.VoPidAacEncparam, unsafe.Pointer(&params))
	err = aacenc.ErrorFromResult(ret)
	if err != nil {
		return nil, fmt.Errorf("aac: %w", err)
	}
	e.handleID = handleID
	return e, nil
}

// Encode encodes data from inbuf.
func (e *Encoder) EncodeOneFrame(inbuf []byte) ([][]byte, error) {
	var outinfo aacenc.VoAudioOutputinfo
	var input, output aacenc.VoCodecBuffer

	input.Buffer = C.CBytes(inbuf)
	if input.Buffer == nil {
		return nil, fmt.Errorf("aac: memory allocation failed")
	}
	defer C.free(input.Buffer)
	input.Length = uint64(len(inbuf))
	ret := aacenc.SetInputData(e.handleID, &input)
	err := aacenc.ErrorFromResult(ret)
	if err != nil {
		return nil, fmt.Errorf("aac: %w", err)
	}
	const bufferSize = 20480
	var outDataList [][]byte
	output.Buffer = C.malloc(C.size_t(bufferSize))
	if output.Buffer == nil {
		return nil, fmt.Errorf("aac: memory allocation failed")
	}
	defer C.free(output.Buffer)
	for {
		output.Length = bufferSize
		ret = aacenc.GetOutputData(e.handleID, &output, &outinfo)
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

// Close closes encoder.
func (e *Encoder) Close() error {
	ret := aacenc.Uninit(e.handleID)
	return aacenc.ErrorFromResult(ret)
}
