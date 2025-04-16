package grpc

import (
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/domain"
	pb "github.com/ulixes-bloom/ya-gophkeeper-cli/internal/infrastructure/proto/gen"
)

// mapProtoGetSecretInfoResponseToDomainSecretInfo converts protobuf SecretInfo to domain model
func mapProtoGetSecretInfoResponseToDomainSecretInfo(info *pb.GetSecretInfoResponse) domain.SecretInfo {
	return domain.SecretInfo{
		Name:      info.GetName(),
		Metadata:  info.GetMetadata(),
		Version:   info.GetVersion(),
		Type:      mapProtoSecretTypeToDomain(info.GetType()),
		CreatedAt: info.GetCreatedAt().AsTime(),
	}
}

// mapProtoGetSecretResponseToDomainSecret converts protobuf Secret to domain model
func mapProtoGetSecretResponseToDomainSecret(resp *pb.GetSecretResponse) *domain.Secret {
	return &domain.Secret{
		Info: mapProtoGetSecretInfoResponseToDomainSecretInfo(resp.GetInfo()),
		Data: resp.GetData(),
	}
}

// mapDomainSecretInfoToProtoCreateSecretInfoRequest converts domain SecretInfo to protobuf request
func mapDomainSecretInfoToProtoCreateSecretInfoRequest(secretInfo domain.SecretInfo) *pb.CreateSecretInfoRequest {
	return &pb.CreateSecretInfoRequest{
		Name:     secretInfo.Name,
		Type:     mapDomainSecretTypeToProto(secretInfo.Type),
		Metadata: secretInfo.Metadata,
	}
}

// mapProtoSecretTypeToDomain converts protobuf SecretType to domain SecretType
func mapProtoSecretTypeToDomain(stype pb.SecretType) domain.SecretType {
	switch stype {
	case pb.SecretType_CREDENTIALS:
		return domain.CredentialsSecretType
	case pb.SecretType_PAYMENT_CARD:
		return domain.PaymentCardSecretType
	case pb.SecretType_BINARY:
		return domain.FileSecretType
	case pb.SecretType_TEXT:
		return domain.TextSecretType
	default:
		return "unknown"
	}
}

// mapDomainSecretTypeToProto converts domain SecretType to protobuf SecretType
func mapDomainSecretTypeToProto(domainType domain.SecretType) pb.SecretType {
	switch domainType {
	case domain.CredentialsSecretType:
		return pb.SecretType_CREDENTIALS
	case domain.TextSecretType:
		return pb.SecretType_TEXT
	case domain.FileSecretType:
		return pb.SecretType_BINARY
	case domain.PaymentCardSecretType:
		return pb.SecretType_PAYMENT_CARD
	default:
		return -1
	}
}
