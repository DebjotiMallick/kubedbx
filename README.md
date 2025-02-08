# Database Backup Tool

A robust, Kubernetes-native tool for automated database backups supporting multiple database types including MongoDB, MySQL, PostgreSQL, MinIO, and MariaDB.

## Features

- Multi-database support:
  - MongoDB
  - MySQL
  - PostgreSQL
  - MinIO
  - MariaDB
- Kubernetes-native implementation
- Configurable backup schedules
- Slack notifications for backup status
- YAML-based configuration
- Automatic error handling and retries
- Support for different storage backends

## Prerequisites

- Kubernetes cluster access
- `kubectl` configured with appropriate permissions
- Access to database instances
- Storage backend configured (for backup storage)
- Slack webhook URL (for notifications)

## Configuration

The tool uses a YAML configuration file to define database backup settings. By default, it looks for the configuration at `/etc/config/databases.yaml`, but you can override this using the `DATABASE_CONFIG_FILE` environment variable.

Example configuration structure:

```yaml
databases:
  mongodb:
    - hostname: mongodb-instance
      namespace: app-namespace
  mysql:
    - hostname: mysql-instance
      namespace: app-namespace
  postgresql:
    - hostname: postgres-instance
      namespace: app-namespace
```

## Installation

1. Clone the repository:
```bash
git clone https://github.com/yourusername/database-backup.git
cd database-backup
```

2. Build the tool:
```bash
go build -o backup-tool cmd/backup\ tool/main.go
```

3. Deploy configuration:
```bash
kubectl create configmap database-config --from-file=config/databases.yaml
```

## Usage

Run the backup tool:

```bash
./backup-tool
```

Or using environment variable for custom config location:

```bash
DATABASE_CONFIG_FILE=/path/to/config.yaml ./backup-tool
```

## Project Structure

```
database-backup/
├── cmd/
│   └── backup tool/
│       └── main.go         # Main application entry point
├── internal/
│   ├── backup/            # Database-specific backup implementations
│   ├── config/            # Configuration handling
│   ├── kubernetes/        # Kubernetes client and operations
│   ├── notification/      # Notification services (Slack)
│   └── storage/          # Storage backend implementations
├── config/
│   └── databases.yaml    # Database configuration file
└── README.md
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

For support, please open an issue in the GitHub repository or contact the maintainers.