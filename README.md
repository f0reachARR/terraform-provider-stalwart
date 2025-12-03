# Terraform Provider for Stalwart Email Server

This Terraform provider allows you to manage [Stalwart Email Server](https://stalw.art/) resources using Infrastructure as Code.

## Features

- **Auto-generated API Client**: Uses `oapi-codegen` to generate type-safe API client code from the official [Stalwart OpenAPI Specification](https://github.com/stalwartlabs/stalwart/blob/main/api/v1/openapi.yml)
- **Principal Management**: Create and manage users, groups, domains, and mailing lists
- **DKIM Signatures**: Configure DKIM signing for domains
- **Data Sources**: Query existing principals and configurations

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)
- Access to a Stalwart Email Server instance

## Installation

### Using the Provider

```hcl
terraform {
  required_providers {
    stalwart = {
      source = "f0reachARR/stalwart"
      version = "~> 0.1"
    }
  }
}

provider "stalwart" {
  endpoint = "https://mail.example.org/api"
  username = "admin"
  password = var.stalwart_password
}
```

### Provider Configuration

The provider supports the following configuration options:

- `endpoint` - (Required) The API endpoint URL for your Stalwart server. Can also be set via `STALWART_ENDPOINT` environment variable.
- `username` - (Required) Username for API authentication. Can also be set via `STALWART_USERNAME` environment variable.
- `password` - (Required) Password for API authentication. Can also be set via `STALWART_PASSWORD` environment variable.

## Resources

### `stalwart_principal`

Manages a principal (user, group, domain, or list).

```hcl
resource "stalwart_principal" "user" {
  type        = "individual"
  name        = "john"
  description = "John Doe"
  quota       = 10737418240  # 10GB
  emails      = ["john@example.org"]
  roles       = ["user"]
  lists       = ["all"]
}
```

### `stalwart_dkim_signature`

Creates a DKIM signature for a domain.

```hcl
resource "stalwart_dkim_signature" "example" {
  algorithm = "Ed25519"
  domain    = "example.org"
  selector  = "default"
}
```

## Data Sources

### `stalwart_principals`

Fetches a list of principals.

```hcl
data "stalwart_principals" "users" {
  types = "individual"
}

output "user_count" {
  value = length(data.stalwart_principals.users.principals)
}
```

## Development

### Building the Provider

```bash
go build -o terraform-provider-stalwart
```

### Generating API Client

The API client is auto-generated from the OpenAPI specification:

```bash
make generate
```

This runs `oapi-codegen` to regenerate `internal/client/client_gen.go` from `openapi.yml`.

### Installing Locally

```bash
make install
```

This builds and installs the provider to your local Terraform plugins directory.

### Running Tests

```bash
make test
```

### Code Generation Workflow

1. The OpenAPI spec is downloaded from [Stalwart's repository](https://github.com/stalwartlabs/stalwart/blob/main/api/v1/openapi.yml)
2. `oapi-codegen` generates type-safe Go code including:
   - Request/response types
   - Client interface
   - Client with typed responses
3. Custom authentication wrapper (`internal/client/auth.go`) adds OAuth token management
4. Terraform resources use the generated client for all API operations

### Project Structure

```
.
├── internal/
│   ├── client/
│   │   ├── client_gen.go  # Auto-generated from OpenAPI spec
│   │   └── auth.go         # Authentication wrapper
│   └── provider/
│       ├── provider.go
│       ├── principal_resource.go
│       ├── dkim_signature_resource.go
│       └── principals_data_source.go
├── examples/              # Example Terraform configurations
├── openapi.yml           # Stalwart OpenAPI specification
├── oapi-codegen.yaml     # Code generation configuration
├── main.go
├── Makefile
└── README.md
```

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the terms specified in the LICENSE file.

## Acknowledgments

- [Stalwart Labs](https://stalw.art/) for the excellent email server
- [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) for OpenAPI code generation
- [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) for the provider SDK