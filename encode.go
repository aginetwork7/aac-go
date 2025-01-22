//go:build !ignore

// Package aac provides AAC codec encoder based on VisualOn AAC encoder library.
package aac

//#include <stdlib.h>
import "C"

import (
	"errors"
	"fmt"
	"io"
	"unsafe"

	"github.com/aginetwork7/aac-go/aacenc"
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
	w io.Writer

	insize int
	inbuf  []byte
	outbuf []byte
}

// NewEncoder returns new AAC encoder.
func NewEncoder(w io.Writer, opts *Options) (*Encoder, error) {
	e := &Encoder{}
	e.w = w

	if opts.BitRate == 0 {
		opts.BitRate = 64000
	}

	ret := aacenc.Init(aacenc.VoAudioCodingAac)
	err := aacenc.ErrorFromResult(ret)
	if err != nil {
		return nil, fmt.Errorf("aac: %w", err)
	}

	var params aacenc.Param
	params.SampleRate = int32(opts.SampleRate)
	params.BitRate = int32(opts.BitRate)
	params.NChannels = int16(opts.NumChannels)
	params.AdtsUsed = 1

	ret = aacenc.SetParam(aacenc.VoPidAacEncparam, unsafe.Pointer(&params))
	err = aacenc.ErrorFromResult(ret)
	if err != nil {
		return nil, fmt.Errorf("aac: %w", err)
	}

	e.insize = opts.NumChannels * 2 * 1024

	e.inbuf = make([]byte, e.insize)
	e.outbuf = make([]byte, 20480)

	return e, nil
}

// NewEncoder returns new AAC encoder.
func NewEncoderV2(opts *Options) (*Encoder, error) {
	e := &Encoder{}
	if opts.BitRate == 0 {
		opts.BitRate = 64000
	}

	ret := aacenc.Init(aacenc.VoAudioCodingAac)
	err := aacenc.ErrorFromResult(ret)
	if err != nil {
		return nil, fmt.Errorf("aac: %w", err)
	}

	var params aacenc.Param
	params.SampleRate = int32(opts.SampleRate)
	params.BitRate = int32(opts.BitRate)
	params.NChannels = int16(opts.NumChannels)
	params.AdtsUsed = 1

	ret = aacenc.SetParam(aacenc.VoPidAacEncparam, unsafe.Pointer(&params))
	err = aacenc.ErrorFromResult(ret)
	if err != nil {
		return nil, fmt.Errorf("aac: %w", err)
	}

	e.outbuf = make([]byte, 20480)

	return e, nil
}

// Encode encodes data from reader.
func (e *Encoder) Encode(r io.Reader) error {
	inputEmpty := false
	for !inputEmpty {
		n, err := r.Read(e.inbuf)
		if err != nil {
			if !errors.Is(err, io.EOF) {
				return fmt.Errorf("aac: %w", err)
			}
			inputEmpty = true
		}

		if n < e.insize {
			inputEmpty = true
		}

		var outinfo aacenc.VoAudioOutputinfo
		var input, output aacenc.VoCodecBuffer

		input.Buffer = C.CBytes(e.inbuf)
		input.Length = uint64(n)

		ret := aacenc.SetInputData(&input)
		err = aacenc.ErrorFromResult(ret)
		if err != nil {
			return fmt.Errorf("aac: %w", err)
		}

		outputEmpty := false
		for !outputEmpty {
			output.Buffer = C.CBytes(e.outbuf)
			output.Length = uint64(len(e.outbuf))

			ret = aacenc.GetOutputData(&output, &outinfo)
			err = aacenc.ErrorFromResult(ret)
			if err != nil {
				if !errors.Is(err, aacenc.ErrInputBufferSmall) {
					return fmt.Errorf("aac: %w", err)
				}
				outputEmpty = true
			}

			_, err = e.w.Write(C.GoBytes(output.Buffer, C.int(output.Length)))
			if err != nil {
				return fmt.Errorf("aac: %w", err)
			}
			C.free(output.Buffer)
		}

		C.free(input.Buffer)
	}

	return nil
}

// Encode encodes data from reader.
func (e *Encoder) EncodeOneFrame(inbuf []byte) ([][]byte, error) {
	var outinfo aacenc.VoAudioOutputinfo
	var input, output aacenc.VoCodecBuffer

	input.Buffer = C.CBytes(inbuf)
	input.Length = uint64(len(inbuf))

	ret := aacenc.SetInputData(&input)
	err := aacenc.ErrorFromResult(ret)
	if err != nil {
		return nil, fmt.Errorf("aac: %w", err)
	}

	outputEmpty := false
	var outDataList [][]byte
	for !outputEmpty {
		output.Buffer = C.CBytes(e.outbuf)
		output.Length = uint64(len(e.outbuf))

		ret = aacenc.GetOutputData(&output, &outinfo)
		err = aacenc.ErrorFromResult(ret)
		if err != nil {
			if !errors.Is(err, aacenc.ErrInputBufferSmall) {
				return nil, fmt.Errorf("aac: %w", err)
			}
			outputEmpty = true
		}

		outData := C.GoBytes(output.Buffer, C.int(output.Length))
		outDataList = append(outDataList, outData)
		C.free(output.Buffer)
	}

	C.free(input.Buffer)

	return outDataList, nil
}

// Close closes encoder.
func (e *Encoder) Close() error {
	ret := aacenc.Uninit()
	return aacenc.ErrorFromResult(ret)
}
