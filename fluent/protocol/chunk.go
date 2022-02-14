// MIT License

// Copyright contributors to the fluent-forward-go project

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package protocol

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/tinylib/msgp/msgp"
)

// ChunkEncoder wraps methods to encode a message and generate
// "chunk" IDs for use with Fluent's chunk-ack protocol. See
// https://github.com/fluent/fluentd/wiki/Forward-Protocol-Specification-v1#response
// for more information.
type ChunkEncoder interface {
	Chunk() (string, error)
	EncodeMsg(*msgp.Writer) error
}

func makeChunkID() (string, error) {
	b, err := uuid.New().MarshalBinary()
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

//msgp:ignore GzipCompressor
type ChunkReader struct {
	br *bytes.Reader
	R  *msgp.Reader
}

func (cr *ChunkReader) Reset(b []byte) {
	if cr.br == nil {
		cr.br = bytes.NewReader(b)
		cr.R = msgp.NewReader(cr.br)

		return
	}

	cr.br.Reset(b)
	cr.R.Reset(cr.br)
}

var chunkKeyBits = []byte("chunk")

// GetChunk searches a marshaled message for the "chunk"
// option value and returns it. The chunk can be used for
// ack checks without the overhead of unmarshalling.
// GetChunk returns an error if no value is found.
func GetChunk(b []byte) (string, error) {
	chunkReader := chunkReaderPool.Get().(*ChunkReader)
	chunkReader.Reset(b)
	reader := chunkReader.R

	defer func() {
		chunkReaderPool.Put(chunkReader)
	}()

	sz, err := reader.ReadArrayHeader()
	if err != nil {
		return "", fmt.Errorf("read array header: %w", err)
	}

	if sz == 2 {
		return "", errors.New("chunk not found")
	}

	if err = reader.Skip(); err != nil {
		return "", fmt.Errorf("skip tag: %w", err)
	}

	t, err := reader.NextType()
	if err != nil {
		return "", fmt.Errorf("next type: %w", err)
	}

	if t == msgp.ExtensionType || t == msgp.IntType {
		// this is Message or MessageExt, which is sz 3
		// when there are no options
		if sz == 3 {
			return "", errors.New("chunk not found")
		}

		if err = reader.Skip(); err != nil {
			return "", fmt.Errorf("skip timestamp: %w", err)
		}
	}

	if err = reader.Skip(); err != nil {
		return "", fmt.Errorf("skip records: %w", err)
	}

	if t, err = reader.NextType(); t != msgp.MapType || err != nil {
		return "", fmt.Errorf("chunk not found: %w", err)
	}

	sz, err = reader.ReadMapHeader()
	if err != nil {
		return "", fmt.Errorf("read map header: %w", err)
	}

	for i := uint32(0); i < sz; i++ {
		keyBits, err := reader.ReadMapKeyPtr()
		if err != nil {
			return "", fmt.Errorf("read map key: %w", err)
		}

		if bytes.Equal(keyBits, chunkKeyBits) {
			v, err := reader.ReadMapKey(nil)
			return string(v), err
		}

		// didn't find "chunk", so skip to next key
		if err = reader.Skip(); err != nil {
			return "", fmt.Errorf("skip value: %w", err)
		}
	}

	return "", errors.New("chunk not found")
}
