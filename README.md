# GophKeeper CLI

A secure command-line application for storing and managing sensitive data like credentials, payment cards, text notes, and files.

## Features

- Secure storage of sensitive data
- Multiple secret types supported
- Versioning support for secrets
- User authentication
- Simple command-line interface

## Installation

1. Download the latest release for your platform
2. Extract the binary
3. Move the binary to your PATH (e.g., `/usr/local/bin`)

## Authentication

| Command      | Description        | Required Flags                    |
|--------------|--------------------|-----------------------------------|
| `register`   | Create new account | `--login`/`-l`, `--password`/`-p` |
| `login`      | Authenticate       | `--login`/`-l`, `--password`/`-p` |

Before using the application, you need to register and login:

### Examples

```bash
# Register a new account
gophkeeper-cli register --login user_login --password user_password
gophkeeper-cli register -l user_login -p user_password

# Login to your account
gophkeeper-cli login --login user_login --password user_password
gophkeeper-cli login -l user_login -p user_password
```

## Secret Creation

| Command              | Description          | Required Flags                                   | Optional Flags    |
|----------------------|----------------------|--------------------------------------------------|-------------------|
| `create-credentials` | Store login/password | `--name`/`-n`, `--login`/`-l`, `--password`/`-p` | `--metadata`/`-m` |
| `create-paymentcard` | Store payment card   | `--name`/`-n`, `--number`/`-c`                   | `--metadata`/`-m` |
| `create-text`        | Store text content   | `--name`/`-n`                                    | `--metadata`/`-m` |
| `create-file`        | Store file           | `--name`/`-n`, `--file`/`-f`                     | `--metadata`/`-m` |

### Examples

```bash
gophkeeper-cli create-credentials \
  --name "github" \
  --login "user@example.com" \
  --password "s3cr3t" \
  --metadata "{"description": "Personal GitHub account"}"

gophkeeper-cli create-paymentcard \
  --name "visa-card" \
  --number "4111111111111111" \
  --metadata "Primary Visa card"

gophkeeper-cli create-text --name "private-notes"
# Type your text, then 'end' on a new line to finish

gophkeeper-cli create-file \
  --name "secret-document" \
  --file "/path/to/file.pdf"
```

## Secret Retrieval

| Command             | Description           | Required Flags  | Optional Flags   |
|---------------------|-----------------------|-----------------|------------------|
| `get-credentials`   | Get login/password    | `--name`/`-n`   | `--version`/`-v` |
| `get-paymentcard`   | Get payment card      | `--name`/`-n`   | `--version`/`-v` |
| `get-text`          | Get text content      | `--name`/`-n`   | `--version`/`-v` |
| `get-file`          | Get file              | `--name`/`-n`   | `--version`/`-v` |

### Examples

```bash
gophkeeper-cli get-credentials --name "github"

gophkeeper-cli get-credentials --name "github" --version 2

gophkeeper-cli get-paymentcard --name "visa-card"

gophkeeper-cli get-text --name "private-notes"

gophkeeper-cli get-text --name "private-notes" --version 2

gophkeeper-cli get-file-secret --name "secret-document"

```

## Secret Management

| Command    | Description       | Required Flags  |
|------------|-------------------|-----------------|
| `list`     | List all secrets  | None            |
| `delete`   | Delete a secret   | `--name`/`-n`   |


### Examples

```bash
gophkeeper-cli list

gophkeeper-cli delete --name "github"
```