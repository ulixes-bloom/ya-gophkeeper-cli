package grpc

import (
	"errors"
	"fmt"
	"io"

	pb "github.com/ulixes-bloom/ya-gophkeeper-cli/internal/infrastructure/proto/gen"
)

// secretStreamReader implements io.Reader for streaming secret data from gRPC.
// It buffers incoming chunks and provides a standard Read interface.
type secretStreamReader struct {
	stream pb.SecretService_GetLatestSecretStreamClient // gRPC stream
	buffer []byte                                       // current data buffer
	index  int                                          // current position in buffer
}

// NewSecretStreamReader creates a new io.Reader that reads from a gRPC secret stream.
// The reader handles chunked data from the stream and presents it as a continuous flow.
func NewSecretStreamReader(stream pb.SecretService_GetLatestSecretStreamClient) io.Reader {
	return &secretStreamReader{
		stream: stream,
	}
}

// Read implements io.Reader interface to read data from the gRPC stream.
// It fills the provided byte slice with data from the stream.
// Returns number of bytes read and any error encountered.
func (r *secretStreamReader) Read(p []byte) (int, error) {
	// Check if we have buffered data to return
	if r.index < len(r.buffer) {
		n := copy(p, r.buffer[r.index:])
		r.index += n
		return n, nil
	}

	// Get next chunk from stream
	response, err := r.stream.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return 0, io.EOF
		}
		return 0, fmt.Errorf("stream read failed: %w", err)
	}

	// Verify we got data chunk (not metadata or other message type)
	chunk := response.GetData()
	if chunk == nil {
		return 0, fmt.Errorf("expected data chunk, got %T", response.Chunk)
	}

	// Store new data in buffer
	r.buffer = chunk
	r.index = 0

	// Recursive call to read from the new buffer
	return r.Read(p)
}
