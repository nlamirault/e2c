# e2c - AWS EC2 Terminal UI Manager

`e2c` is a terminal-based UI application for managing AWS EC2 instances, inspired by [k9s](https://github.com/derailed/k9s) for Kubernetes and [e1s](https://github.com/keidarcy/e1s/) for ECS.

![e2c screenshot](docs/images/screenshot.png)

## Features

- Terminal-based UI for managing EC2 instances
- View instance details, status, and resource utilization
- Start, stop, reboot, and terminate instances
- Connect to instances via SSH
- Filter and search for instances across multiple regions
- Monitor resource metrics
- View instance logs and console output
- Support for multiple AWS profiles and regions

## Installation

### Using Go

```bash
go install github.com/nlamirault/e2c/cmd/e2c@latest
```

### Binary Release

Download the latest binary from the [releases page](https://github.com/nlamirault/e2c/releases).

## Usage

```bash
# Start e2c with default profile
e2c

# Start with a specific AWS region
e2c --region eu-west-1

# Show help
e2c --help
```

## Keyboard Shortcuts

## Key Shortcuts

| Key   | Action                               |
| ----- | ------------------------------------ |
| `?`   | Help                                 |
| `q`   | Quit                                 |
| `Esc` | Back/Close Dialog                    |
| `f`   | Filter instances                     |
| `r`   | Refresh                              |
| `s`   | Start selected instance              |
| `p`   | Stop selected instance               |
| `b`   | Reboot selected instance             |
| `t`   | Terminate selected instance          |
| `c`   | Connect to selected instance via SSH |
| `l`   | View instance logs                   |
| `/`   | Search                               |

## Configuration

e2c uses the AWS SDK's default credential chain, supporting:

- Environment variables
- AWS credentials file
- IAM roles for EC2/ECS

Configuration file located at `~/.config/e2c/config.yaml`:

```yaml
aws:
  default_region: eu-west-1
  refresh_interval: 30s

ui:
  # Compact mode reduces whitespace in the UI
  compact: false
```

### Environment Variables

The following environment variables can be used to configure e2c:

- `E2C_LOG_LEVEL`: Set the logging level (debug, info, warn, error)
- `E2C_LOG_FORMAT`: Set the log format ("json" or "text"). Default is text format with colors

Examples:

```bash
# Set environment variables before running e2c
E2C_LOG_FORMAT=json E2C_LOG_LEVEL=debug e2c
```

Note: Command line flags take precedence over environment variables.

## Requirements

- AWS credentials configured
- Appropriate IAM permissions to list and manage EC2 instances

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

Apache Version 2.0
