#!/bin/bash
mockgen -source=internal/domain/secret.go -destination=internal/mocks/mock_secret_service.go -package=mocks
mockgen -source=internal/domain/auth.go -destination=internal/mocks/mock_auth_service.go -package=mocks