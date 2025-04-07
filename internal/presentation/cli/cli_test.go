package cli_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/domain"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/mocks"
	"github.com/ulixes-bloom/ya-gophkeeper-cli/internal/presentation/cli"
)

func TestCLI_RegisterCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	tests := []struct {
		name           string
		login          string
		password       string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name:     "successful registration",
			login:    "testuser",
			password: "testpass",
			setupMock: func() {
				mockAuthService.EXPECT().Register(ctx, "testuser", "testpass").
					Return(nil)
			},
			expectedOutput: "Successfully registered user\n",
		},
		{
			name:     "login already exists",
			login:    "existinguser",
			password: "testpass",
			setupMock: func() {
				mockAuthService.EXPECT().Register(ctx, "existinguser", "testpass").
					Return(domain.ErrLoginAlreayExists)
			},
			expectedError: domain.ErrLoginAlreayExists,
		},
		{
			name:     "other registration error",
			login:    "testuser",
			password: "testpass",
			setupMock: func() {
				mockAuthService.EXPECT().Register(ctx, "testuser", "testpass").
					Return(fmt.Errorf("failed to register"))
			},
			expectedError: errors.New("failed to register"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Set the flags
			cmd.SetArgs([]string{"register", "-l", tt.login, "-p", tt.password})

			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestCLI_LoginCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	tests := []struct {
		name           string
		login          string
		password       string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name:     "successful login",
			login:    "testuser",
			password: "testpass",
			setupMock: func() {
				mockAuthService.EXPECT().Login(ctx, "testuser", "testpass").
					Return(nil)
			},
			expectedOutput: "Successfully logged in\n",
		},
		{
			name:     "login not found",
			login:    "notexistinguser",
			password: "testpass",
			setupMock: func() {
				mockAuthService.EXPECT().Login(ctx, "notexistinguser", "testpass").
					Return(domain.ErrUserNotFound)
			},
			expectedError: domain.ErrUserNotFound,
		},
		{
			name:     "invalid credentials",
			login:    "testuser",
			password: "invalidpass",
			setupMock: func() {
				mockAuthService.EXPECT().Login(ctx, "testuser", "invalidpass").
					Return(domain.ErrInvalidCredentials)
			},
			expectedError: domain.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			// Set the flags
			cmd.SetArgs([]string{"login", "-l", tt.login, "-p", tt.password})

			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestCLI_CreateCredentialsSecretCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	tests := []struct {
		name           string
		args           []string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name: "successful credentials creation",
			args: []string{"create-credentials", "-n", "testcreds", "-l", "testuser", "-p", "testpass"},
			setupMock: func() {
				expectedSecret := domain.Secret{
					Info: domain.SecretInfo{
						Name: "testcreds",
						Type: domain.CredentialsSecretType,
					},
				}
				expectedData, _ := json.Marshal(domain.CredentialsSecret{
					Login:    "testuser",
					Password: "testpass",
				})
				mockSecretService.EXPECT().
					CreateSecret(ctx, expectedSecret, gomock.Any()).
					DoAndReturn(func(ctx context.Context, secret domain.Secret, r io.Reader) error {
						data, _ := io.ReadAll(r)
						assert.Equal(t, expectedData, data)
						return nil
					})
			},
			expectedOutput: "Successfully stored credentials for 'testcreds'\n",
		},
		{
			name: "creation error",
			args: []string{"create-credentials", "-n", "testcreds", "-l", "testuser", "-p", "testpass"},
			setupMock: func() {
				mockSecretService.EXPECT().
					CreateSecret(gomock.Any(), gomock.Any(), gomock.Any()).
					Return(fmt.Errorf("storage error"))
			},
			expectedError: errors.New("failed to create secret: storage error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			cmd.SetArgs(tt.args)
			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestCLI_CreatePaymentCardSecretCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	tests := []struct {
		name           string
		args           []string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name: "successful payment card creation",
			args: []string{"create-paymentcard", "-n", "testcard", "-c", "1234567890123456"},
			setupMock: func() {
				expectedSecret := domain.Secret{
					Info: domain.SecretInfo{
						Name: "testcard",
						Type: domain.PaymentCardSecretType,
					},
				}
				expectedData, _ := json.Marshal(domain.PaymentCardSecret{
					Number: "1234567890123456",
				})
				mockSecretService.EXPECT().
					CreateSecret(ctx, expectedSecret, gomock.Any()).
					DoAndReturn(func(ctx context.Context, secret domain.Secret, r io.Reader) error {
						data, _ := io.ReadAll(r)
						assert.Equal(t, expectedData, data)
						return nil
					})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			cmd.SetArgs(tt.args)
			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestCLI_CreateTextSecretCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	// Mock stdin for interactive input
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	tests := []struct {
		name           string
		args           []string
		input          string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name:  "successful text creation",
			args:  []string{"create-text", "-n", "testtext"},
			input: "line1\nline2\nend\n",
			setupMock: func() {
				expectedSecret := domain.Secret{
					Info: domain.SecretInfo{
						Name: "testtext",
						Type: domain.TextSecretType,
					},
				}
				mockSecretService.EXPECT().
					CreateSecret(ctx, expectedSecret, gomock.Any()).
					DoAndReturn(func(ctx context.Context, secret domain.Secret, r io.Reader) error {
						data, _ := io.ReadAll(r)
						assert.Equal(t, "line1\nline2\n", string(data))
						return nil
					})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input != "" {
				r, w, _ := os.Pipe()
				os.Stdin = r
				go func() {
					w.Write([]byte(tt.input))
					w.Close()
				}()
			}

			if tt.setupMock != nil {
				tt.setupMock()
			}

			cmd.SetArgs(tt.args)
			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestCLI_CreateFileSecretCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("test content")
	tmpFile.Close()

	tests := []struct {
		name           string
		args           []string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name: "successful file creation",
			args: []string{"create-file", "-n", "testfile", "-f", tmpFile.Name()},
			setupMock: func() {
				expectedSecret := domain.Secret{
					Info: domain.SecretInfo{
						Name: "testfile",
						Type: domain.FileSecretType,
					},
				}
				mockSecretService.EXPECT().
					CreateSecret(ctx, expectedSecret, gomock.Any()).
					DoAndReturn(func(ctx context.Context, secret domain.Secret, r io.Reader) error {
						data, _ := io.ReadAll(r)
						assert.Equal(t, "test content", string(data))
						return nil
					})
			},
			expectedOutput: fmt.Sprintf("Successfully stored file '%s' as 'testfile'\n", tmpFile.Name()),
		},
		{
			name:          "file not found",
			args:          []string{"create-file", "-n", "testfile", "-f", "nonexistent"},
			expectedError: errors.New("failed to open file 'nonexistent'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			cmd.SetArgs(tt.args)
			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestCLI_ListSecretsCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	tests := []struct {
		name           string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name: "successful list with secrets",
			setupMock: func() {
				mockSecretService.EXPECT().
					ListSecrets(ctx).
					Return([]string{"secret1", "secret2"}, nil)
			},
			expectedOutput: "Stored secrets:\n  - secret1\n  - secret2\n",
		},
		{
			name: "list error",
			setupMock: func() {
				mockSecretService.EXPECT().
					ListSecrets(ctx).
					Return(nil, fmt.Errorf("list error"))
			},
			expectedError: errors.New("failed to list secrets"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			cmd.SetArgs([]string{"list"})
			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
				fmt.Println(output)
			}
		})
	}
}

func TestCLI_GetCredentialsSecretCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	testCreds := domain.CredentialsSecret{
		Login:    "testuser",
		Password: "testpass",
	}
	credsData, _ := json.Marshal(testCreds)

	tests := []struct {
		name           string
		args           []string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name: "successful get latest credentials",
			args: []string{"get-credentials", "-n", "testcreds"},
			setupMock: func() {
				mockSecretService.EXPECT().
					GetLatestSecret(ctx, "testcreds").
					Return(&domain.Secret{
						Info: domain.SecretInfo{
							Name:    "testcreds",
							Version: 1,
						},
						Data: string(credsData),
					}, nil)
			},
			expectedOutput: "Name: testcreds\nVersion: 1\nLogin: testuser\nPassword: testpass\n",
		},
		{
			name: "successful get specific version",
			args: []string{"get-credentials", "-n", "testcreds", "-v", "2"},
			setupMock: func() {
				mockSecretService.EXPECT().
					GetSecretByVersion(ctx, "testcreds", int32(2)).
					Return(&domain.Secret{
						Info: domain.SecretInfo{
							Name:     "testcreds",
							Version:  2,
							Metadata: "test metadata",
						},
						Data: string(credsData),
					}, nil)
			},
			expectedOutput: "Name: testcreds\nVersion: 2\nMetadata: test metadata\nLogin: testuser\nPassword: testpass\n",
		},
		{
			name: "secret not found",
			args: []string{"get-credentials", "-n", "testcreds", "-v", "2"},
			setupMock: func() {
				mockSecretService.EXPECT().
					GetSecretByVersion(ctx, "testcreds", int32(2)).
					Return(nil, fmt.Errorf("not found"))
			},
			expectedError: errors.New("failed to retrieve credentials"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			cmd.SetArgs(tt.args)
			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestCLI_GetPaymentCardSecretCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	testCard := domain.PaymentCardSecret{
		Number: "1234567890123456",
	}
	cardData, _ := json.Marshal(testCard)

	tests := []struct {
		name           string
		args           []string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name: "successful get payment card",
			args: []string{"get-paymentcard", "-n", "testcard"},
			setupMock: func() {
				mockSecretService.EXPECT().
					GetLatestSecret(ctx, "testcard").
					Return(&domain.Secret{
						Info: domain.SecretInfo{
							Name:    "testcard",
							Version: 1,
						},
						Data: string(cardData),
					}, nil)
			},
			expectedOutput: "Name: testcard\nVersion: 1\nCard Number: 1234567890123456\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			cmd.SetArgs(tt.args)
			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestCLI_GetTextSecretCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	tests := []struct {
		name           string
		args           []string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name: "successful get text",
			args: []string{"get-text", "-n", "testtext"},
			setupMock: func() {
				mockSecretService.EXPECT().
					GetLatestSecretStream(ctx, "testtext").
					Return(
						io.NopCloser(strings.NewReader("text content")),
						&domain.SecretInfo{
							Name:    "testtext",
							Version: 1,
						},
						nil,
					)
			},
			expectedOutput: "Name: testtext\nVersion: 1\n\nContent:\ntext content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			cmd.SetArgs(tt.args)
			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

func TestCLI_GetFileSecretCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	// Clean up test files after
	defer os.Remove("testfile")

	tests := []struct {
		name           string
		args           []string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name: "successful get file",
			args: []string{"get-file-secret", "-n", "testfile"},
			setupMock: func() {
				mockSecretService.EXPECT().
					GetLatestSecretStream(ctx, "testfile").
					Return(
						io.NopCloser(strings.NewReader("file content")),
						&domain.SecretInfo{
							Name:    "testfile",
							Version: 1,
						},
						nil,
					)
			},
			expectedOutput: "Successfully downloaded 'testfile' version 1\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupMock != nil {
				tt.setupMock()
			}

			cmd.SetArgs(tt.args)
			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
				if tt.name == "successful get file" {
					content, _ := os.ReadFile("testfile")
					assert.Equal(t, "file content", string(content))
				}
			}
		})
	}
}

func TestCLI_DeleteSecretCmd(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockAuthService := mocks.NewMockAuthService(ctrl)
	mockSecretService := mocks.NewMockSecretService(ctrl)

	ctx := context.Background()
	cmd := cli.NewCLI(ctx, mockSecretService, mockAuthService)

	// Mock stdin for confirmation
	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	tests := []struct {
		name           string
		args           []string
		input          string
		setupMock      func()
		expectedOutput string
		expectedError  error
	}{
		{
			name:  "successful delete with confirmation",
			args:  []string{"delete", "-n", "testsecret"},
			input: "y\n",
			setupMock: func() {
				mockSecretService.EXPECT().
					DeleteSecret(ctx, "testsecret").
					Return(nil)
			},
			expectedOutput: "Are you sure you want to delete 'testsecret'? (y/n): Successfully deleted 'testsecret'\n",
		},
		{
			name:           "delete cancelled",
			args:           []string{"delete", "-n", "testsecret"},
			input:          "n\n",
			expectedOutput: "Are you sure you want to delete 'testsecret'? (y/n): Deletion cancelled\n",
		},
		{
			name:  "delete error",
			args:  []string{"delete", "-n", "testsecret"},
			input: "y\n",
			setupMock: func() {
				mockSecretService.EXPECT().
					DeleteSecret(ctx, "testsecret").
					Return(fmt.Errorf("delete error"))
			},
			expectedError: errors.New("failed to delete secret 'testsecret'"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.input != "" {
				r, w, _ := os.Pipe()
				os.Stdin = r
				go func() {
					w.Write([]byte(tt.input))
					w.Close()
				}()
			}

			if tt.setupMock != nil {
				tt.setupMock()
			}

			cmd.SetArgs(tt.args)
			output, err := executeCommand(cmd)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedOutput, output)
			}
		})
	}
}

// executeCommand executes the command and returns the output
func executeCommand(cmd *cobra.Command) (string, error) {
	// Backup the original stdout and stderr
	oldStdout := os.Stdout
	oldStderr := os.Stderr
	defer func() {
		os.Stdout = oldStdout
		os.Stderr = oldStderr
	}()

	// Create pipes to capture output
	rOut, wOut, _ := os.Pipe()
	rErr, wErr, _ := os.Pipe()
	os.Stdout = wOut
	os.Stderr = wErr
	cmd.SetOut(wOut)
	cmd.SetErr(wErr)

	// Channel to collect the output
	outChan := make(chan string)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, rOut)
		_, _ = io.Copy(&buf, rErr)
		outChan <- buf.String()
	}()

	// Execute the command
	err := cmd.Execute()

	// Close the writers and wait for output
	wOut.Close()
	wErr.Close()
	output := <-outChan

	return output, err
}
