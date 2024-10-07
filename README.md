# Prometheus Journald Exporter

This Prometheus exporter monitors critical system events using `journalctl` and exposes metrics on a configurable HTTP endpoint for Prometheus scraping.

## Features

- Monitor important system events such as OOM killer, service failures, disk I/O errors, and more.
- Configurable via JSON for flexible log monitoring and metric creation.
- Debugging option to print matched log patterns for easier troubleshooting.
- Configurable Prometheus metrics port and path.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Exporter](#running-the-exporter)
- [Debugging](#debugging)
- [Prometheus Integration](#prometheus-integration)
- [License](#license)

## Installation

To install and run the exporter, follow these steps:

1. **Clone the repository**:
    ```bash
    git clone https://github.com/your-username/prometheus-system-event-exporter.git
    cd prometheus-system-event-exporter
    ```

2. **Build the Go exporter**:
    Ensure you have Go 1.18+ installed. Then run:
    ```bash
    go build -o system-event-exporter main.go
    ```

3. **Prepare the JSON configuration**:
    Modify or use the provided `config.json` file to monitor the desired system events.

## Configuration

The exporter is highly configurable through a JSON configuration file. Below is a sample configuration (`config.json`):

```json
{
  "debug": true,
  "metrics_port": 9100,
  "metrics_path": "/metrics",
  "events": [
    {
      "name": "oom_killer_events",
      "description": "Out of memory killer events.",
      "log_command": ["journalctl", "--follow", "-p", "3", "-k"],
      "match_patterns": ["Out of memory", "oom-kill"]
    },
    {
      "name": "service_failures",
      "description": "Systemd service failure events.",
      "log_command": ["journalctl", "--follow", "-p", "3", "-u", "*"],
      "match_patterns": ["Failed to start", "Service entered failed state"]
    }
    // ... Other events omitted for brevity
  ]
}

