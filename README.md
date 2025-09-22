# XConf Admin

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.23%2B-blue.svg)](https://golang.org/)
[![Build Status](https://img.shields.io/badge/Build-Passing-green.svg)]()

XConf Admin is a comprehensive configuration management server designed for RDK (Reference Design Kit) devices. It provides a centralized platform for managing device configurations, firmware updates, telemetry settings, and various administrative functions across RDK deployments.

## 🚀 Features

- **Configuration Management**: Centralized management of device configurations
- **Firmware Management**: Control firmware distribution and updates
- **Telemetry Services**: Manage telemetry profiles and data collection
- **Device Control Manager (DCM)**: Handle device control and settings
- **Feature Management**: Control feature flags and rules
- **Authentication & Authorization**: JWT-based security with role-based access control
- **RESTful API**: Comprehensive REST API for all operations
- **Metrics & Monitoring**: Built-in Prometheus metrics support
- **Canary Deployments**: Support for gradual rollouts

## 📋 Prerequisites

- **Go 1.23+**: This project requires Go version 1.23 or later
- **Cassandra**: For data persistence (configure in config file)
- **Git**: For version control and building

## 🛠️ Installation

### 1. Clone the Repository

```bash
git clone https://github.com/rdkcentral/xconfadmin.git
cd xconfadmin
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Build the Application

```bash
make build
```

This will create a binary `bin/xconfadmin-{OS}-{ARCH}` (e.g., `bin/xconfadmin-linux-amd64`)

## ⚙️ Configuration

### Environment Variables

Set the following required environment variables:

```bash
export SAT_CLIENT_ID='your_client_id'
export SAT_CLIENT_SECRET='your_client_secret'
export SECURITY_TOKEN_KEY='your_security_token_key'
```

### Configuration File

Create a configuration file based on the sample provided:

```bash
cp config/sample_xconfadmin.conf config/xconfadmin.conf
```

Edit the configuration file to match your environment settings including:
- Server port and timeouts
- Database connection details
- Logging configuration
- Authentication settings
- Service endpoints

## 🚀 Running the Application

### 1. Create Log Directory

```bash
mkdir -p /app/logs/xconfadmin
```

### 2. Start the Server

```bash
bin/xconfadmin-linux-amd64 -f config/xconfadmin.conf
```

### 3. Verify Installation

Test the server is running:

```bash
curl http://localhost:9001/api/v1/version
```

Expected response:
```json
{
  "status": 200,
  "message": "OK",
  "data": {
    "code_git_commit": "abc123",
    "build_time": "2025-09-08T10:00:00Z",
    "binary_version": "v1.0.0",
    "binary_branch": "main",
    "binary_build_time": "2025-09-08_10:00:00_UTC"
  }
}
```

## 📖 API Documentation

The XConf Admin server provides several API endpoints organized by functionality:

### Core APIs

- **Version**: `GET /api/v1/version` - Get application version info
- **Health**: `GET /health` - Health check endpoint
- **Metrics**: `GET /metrics` - Prometheus metrics

### Administrative APIs

- **Authentication**: `/auth/*` - Authentication and authorization
- **Firmware**: `/firmware/*` - Firmware management
- **DCM**: `/dcm/*` - Device Control Manager
- **Telemetry**: `/telemetry/*` - Telemetry configuration
- **Features**: `/feature/*` - Feature management
- **Settings**: `/setting/*` - Various device settings

### Example API Calls

```bash
# Get firmware configurations
curl -H "Authorization: Bearer <token>" http://localhost:9001/api/firmware/configs

# Update device settings
curl -X POST -H "Authorization: Bearer <token>" \
     -H "Content-Type: application/json" \
     -d '{"key":"value"}' \
     http://localhost:9001/api/dcm/settings
```

## 🏗️ Project Structure

```
xconfadmin/
├── adminapi/           # Admin API handlers and services
│   ├── auth/          # Authentication and authorization
│   ├── canary/        # Canary deployment management
│   ├── change/        # Change management
│   ├── dcm/           # Device Control Manager
│   ├── firmware/      # Firmware management
│   ├── queries/       # Query handlers
│   ├── rfc/           # Remote Feature Control
│   ├── setting/       # Settings management
│   └── telemetry/     # Telemetry services
├── common/            # Common utilities and constants
├── config/            # Configuration files
├── http/              # HTTP utilities and middleware
├── shared/            # Shared components
├── taggingapi/        # Tagging API
└── util/              # Utility functions
```

## 🧪 Testing

### Run All Tests

```bash
make test
```

### Run Tests Locally

```bash
make localtest
```

### Generate Coverage Report

```bash
make cover
make html
```

## 🔧 Development

### Build for Development

```bash
make build
```

### Clean Build Artifacts

```bash
make clean
```

### Release Build

```bash
make release
```

## 📊 Monitoring

XConf Admin includes built-in monitoring capabilities:

- **Prometheus Metrics**: Available at `/metrics` endpoint
- **Health Checks**: Available at `/health` endpoint
- **Structured Logging**: JSON-formatted logs with configurable levels
- **Request Tracing**: Optional OpenTelemetry integration

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

### Development Workflow

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Run the test suite
6. Submit a pull request

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:

- Create an issue in the GitHub repository
- Check the [documentation](docs/)
- Review existing issues and discussions

## 🔗 Related Projects

- [xconfwebconfig](https://github.com/rdkcentral/xconfwebconfig) - Web configuration service
- [RDK Central](https://github.com/rdkcentral) - RDK Central organization

---

**Note**: This is a configuration management server for RDK devices. Ensure proper security measures are in place when deploying in production environments.