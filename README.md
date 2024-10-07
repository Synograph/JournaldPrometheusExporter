
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
```

## Configuration Fields:

- `debug`: Set to true to enable detailed log output for debugging purposes.
- `metrics_port`: The port on which the Prometheus server will scrape metrics (default: 9100).
- `metrics_path`: The HTTP path to expose metrics (default: /metrics).
- `events`: A list of system events to monitor:
- `name`: The metric name exposed in Prometheus.
- `description`: Description of the metric.
- `log_command`: Command to follow logs (via journalctl) for each event.
- `match_patterns`: Log patterns that trigger a Prometheus counter increase.
You can add, remove, or customize these events to suit your specific needs.

## Running the Exporter
To start the exporter:
- Ensure the config.json file is correctly configured.
Run the exporter:
```bash
    ./system-event-exporter -config config.json
```
This will start the exporter with the specified configuration and expose the metrics at the defined port and path (default: http://localhost:9100/metrics).

## Debugging
If `debug` is enabled in the configuration, the exporter will print matched log patterns for each monitored event to the console.

```json
{
  "debug": true
  [...]
}
```

Use this feature to verify that the correct logs are being monitored and matched.

## Prometheus Integration
To monitor the exported metrics with Prometheus, add the following to your Prometheus configuration (prometheus.yml):

```yaml
scrape_configs:
  - job_name: "system_event_exporter"
    static_configs:
      - targets: ["localhost:9100"]
      - 
```
## Example Metrics
Once running, the exporter will expose Prometheus metrics for each event defined in the configuration. Example metrics:

```
# HELP oom_killer_events Out of memory killer events.
# TYPE oom_killer_events counter
oom_killer_events 3

# HELP service_failures Systemd service failure events.
# TYPE service_failures counter
service_failures 1
```
These metrics can then be scraped by Prometheus and used for alerting and visualization. 


