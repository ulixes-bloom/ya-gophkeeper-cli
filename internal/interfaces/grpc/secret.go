package grpc

import (
	"context"
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/domain"
	pb "github.com/ulixes-bloom/ya-gophkeeper-cli/internal/infrastructure/proto/gen"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/interfaces/grpc/interceptors"
)

const (
	chunkSize = 1024 * 512 // 500B chunk size for streaming
)

// SecretClient provides methods to interact with the gRPC secret service.
type SecretClient struct {
	client pb.SecretServiceClient
	conn   *grpc.ClientConn
}

// NewSecretClient initializes a new SecretClient with authentication interceptors.
func NewSecretClient(serverAddr string, repo domain.TokenRepository, creds credentials.TransportCredentials) (*SecretClient, error) {
	authInterceptor := interceptors.NewAuthInterceptor(repo)

	conn, err := grpc.NewClient(serverAddr,
		grpc.WithTransportCredentials(creds),
		grpc.WithUnaryInterceptor(authInterceptor.UnaryInterceptor),
		grpc.WithStreamInterceptor(authInterceptor.StreamInterceptor))
	if err != nil {
		return nil, fmt.Errorf("grpc.NewSecretClient: failed to dial gRPC server: %w", err)
	}

	client := pb.NewSecretServiceClient(conn)
	return &SecretClient{client: client, conn: conn}, nil
}

// Close terminates the gRPC connection.
func (c *SecretClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

// CreateSecret sends a request to create a new secret.
func (c *SecretClient) CreateSecret(ctx context.Context, secret domain.Secret) error {
	_, err := c.client.CreateSecret(ctx, &pb.CreateSecretRequest{
		Info: mapDomainSecretInfoToProtoCreateSecretInfoRequest(secret.Info),
		Data: secret.Data,
	})
	if err != nil {
		return err
	}

	return nil
}

// CreateSecretStream streams a large secret to the server in chunks.
func (c *SecretClient) CreateSecretStream(ctx context.Context, secret domain.Secret, reader io.Reader) error {
	stream, err := c.client.CreateSecretStream(ctx)
	if err != nil {
		return fmt.Errorf("client.CreateSecretStream: %w", err)
	}

	if err := c.sendMetadataChunk(stream, secret); err != nil {
		return fmt.Errorf("failed to send metadata chunk: %w", err)
	}

	if err := c.sendDataChunks(stream, reader); err != nil {
		return fmt.Errorf("failed to send data chunks: %w", err)
	}

	if _, err = stream.CloseAndRecv(); err != nil {
		return fmt.Errorf("failed to close stream and recieve server's response: %w", err)
	}

	return nil
}

// sendMetadataChunk sends the secret metadata as the first chunk.
func (c *SecretClient) sendMetadataChunk(stream pb.SecretService_CreateSecretStreamClient, secret domain.Secret) error {
	return stream.Send(&pb.CreateSecretChunkRequest{
		Chunk: &pb.CreateSecretChunkRequest_Info{
			Info: mapDomainSecretInfoToProtoCreateSecretInfoRequest(secret.Info),
		},
	})
}

// sendDataChunks streams the secret data in chunks.
func (c *SecretClient) sendDataChunks(stream pb.SecretService_CreateSecretStreamClient, reader io.Reader) error {
	buf := make([]byte, chunkSize)
	for {
		n, err := reader.Read(buf)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		if n > 0 {
			if err := stream.Send(&pb.CreateSecretChunkRequest{
				Chunk: &pb.CreateSecretChunkRequest_Data{
					Data: buf[:n],
				},
			}); err != nil {
				return err
			}
		}
	}
	return nil
}

// ListSecrets retrieves all secret names from the server.
func (c *SecretClient) ListSecrets(ctx context.Context) ([]string, error) {
	resp, err := c.client.ListSecrets(ctx, &emptypb.Empty{})
	if err != nil {
		return nil, fmt.Errorf("client.ListSecrets: %w", err)
	}

	return resp.Data, nil
}

// GetLatestSecret retrieves the latest version of a secret.
func (c *SecretClient) GetLatestSecret(ctx context.Context, secretName string) (*domain.Secret, error) {
	resp, err := c.client.GetLatestSecret(ctx, &pb.GetLatestSecretRequest{Name: secretName})
	if err != nil {
		return nil, fmt.Errorf("client.GetLatestSecret: %w", err)
	}

	secret := mapProtoGetSecretResponseToDomainSecret(resp)
	return secret, nil
}

// GetLatestSecretStream retrieves the latest version of a secret as a stream.
func (c *SecretClient) GetLatestSecretStream(ctx context.Context, secretName string) (io.Reader, *domain.SecretInfo, error) {
	stream, err := c.client.GetLatestSecretStream(ctx, &pb.GetLatestSecretRequest{Name: secretName})
	if err != nil {
		return nil, nil, fmt.Errorf("client.GetLatestSecretStream: %w", err)
	}

	// Read the first chunk (metadata)
	firstChunk, err := stream.Recv()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to receive first chunk with metadata: %w", err)
	}
	secretInfo := firstChunk.GetInfo()
	if secretInfo == nil {
		return nil, nil, fmt.Errorf("first chunk must contain secret info")
	}

	domainSecretInfo := mapProtoGetSecretInfoResponseToDomainSecretInfo(secretInfo)

	return NewSecretStreamReader(stream), &domainSecretInfo, nil
}

// GetSecretByVersion retrieves a specific version of a secret.
func (c *SecretClient) GetSecretByVersion(ctx context.Context, secretName string, version int32) (*domain.Secret, error) {
	resp, err := c.client.GetSecretByVersion(ctx, &pb.GetSecretByVersionRequest{
		Name:    secretName,
		Version: version,
	})
	if err != nil {
		return nil, fmt.Errorf("client.GetSecretByVersion: %w", err)
	}

	secret := mapProtoGetSecretResponseToDomainSecret(resp)

	return secret, nil
}

// GetSecretStreamByVersion retrieves a specific version of a secret as a stream.
func (c *SecretClient) GetSecretStreamByVersion(ctx context.Context, secretName string, version int32) (io.Reader, *domain.SecretInfo, error) {
	stream, err := c.client.GetSecretStreamByVersion(ctx, &pb.GetSecretByVersionRequest{
		Name:    secretName,
		Version: version,
	})
	if err != nil {
		return nil, nil, fmt.Errorf("client.GetSecretStreamByVersion: %w", err)
	}

	firstChunk, err := stream.Recv()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to recieve metadata: %w", err)
	}
	secretInfo := firstChunk.GetInfo()
	if secretInfo == nil {
		return nil, nil, fmt.Errorf("first chunk must contain secret info")
	}

	domainSecretInfo := mapProtoGetSecretInfoResponseToDomainSecretInfo(secretInfo)

	return NewSecretStreamReader(stream), &domainSecretInfo, nil
}

// DeleteSecret removes a secret by its name.
func (c *SecretClient) DeleteSecret(ctx context.Context, secretName string) error {
	_, err := c.client.DeleteSecret(ctx, &pb.DeleteSecretRequest{
		Name: secretName,
	})
	if err != nil {
		return fmt.Errorf("client.DeleteSecret: %w", err)
	}

	return nil
}
