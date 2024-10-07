package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
)

// execCommand is a variable to allow overriding exec.Command for testing purposes.
var execCommand = exec.Command

// TestLoadConfig tests if the config is loaded properly.
func TestLoadConfig(t *testing.T) {
	// Create a temporary JSON configuration file
	configData := `{
		"debug": true,
		"metrics_port": 9091,
		"metrics_path": "/metrics",
		"events": [
			{
				"name": "test_event",
				"description": "Test event monitoring",
				"log_command": ["journalctl", "-f", "-n", "0"],
				"match_patterns": ["test pattern"]
			}
		]
	}`

	tempFile, err := os.CreateTemp("", "config_test.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tempFile.Name())

	_, err = tempFile.Write([]byte(configData))
	if err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tempFile.Close()

	// Test LoadConfig function
	err = LoadConfig(tempFile.Name())
	if err != nil {
		t.Errorf("LoadConfig failed: %v", err)
	}

	// Validate loaded configuration
	if !config.Debug {
		t.Errorf("Expected Debug to be true, got %v", config.Debug)
	}
	if config.MetricsPort != 9091 {
		t.Errorf("Expected MetricsPort to be 9091, got %d", config.MetricsPort)
	}
	if len(config.Events) != 1 || config.Events[0].Name != "test_event" {
		t.Errorf("Expected event 'test_event', got %+v", config.Events)
	}
}

// TestMonitorEvent writes a log into the journal and verifies Prometheus captures it.
func TestMonitorEvent(t *testing.T) {
	// Initialize a fake event to monitor
	event := EventMonitor{
		Name:          "test_event",
		Description:   "Test event",
		LogCommand:    []string{"journalctl", "-f", "-n", "0"},
		MatchPatterns: []string{"test pattern"},
	}

	// Initialize metrics map and register a test counter
	metricsMap = make(map[string]*prometheus.CounterVec)
	metricsMap[event.Name] = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: event.Name,
			Help: event.Description,
		},
		[]string{"pattern"},
	)
	prometheus.MustRegister(metricsMap[event.Name])

	// Create a cancellable context to pass to monitorEvent
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure the context is canceled to clean up resources

	// Run monitorEvent function in a separate goroutine with context
	wg.Add(1)
	go monitorEvent(ctx, event)

	// Give monitorEvent time to start
	time.Sleep(2 * time.Second)

	// Write a log entry into journalctl using the logger command
	testLogEntry := "This is a test pattern"
	cmd := exec.Command("logger", testLogEntry)
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to write to journal: %v", err)
	}

	// Give the log entry time to be processed
	time.Sleep(5 * time.Second)

	// Check if the counter was incremented
	counter := metricsMap[event.Name].With(prometheus.Labels{"pattern": "test pattern"})
	metric := &io_prometheus_client.Metric{}
	err := counter.Write(metric)
	if err != nil {
		t.Fatalf("Failed to write metric: %v", err)
	}

	// Get the counter value
	counterValue := metric.GetCounter().GetValue()
	if counterValue != 1 {
		t.Errorf("Expected counter to be 1, got %v", counterValue)
	}

	// Clean up by canceling the monitoring goroutine
	cancel()
	wg.Wait()
}

// TestMetricsEndpoint verifies that the metrics endpoint is correctly exposed
func TestMetricsEndpoint(t *testing.T) {
	// Start the Prometheus metrics server in a separate goroutine
	go func() {
		main()
	}()

	// Give the server time to start
	time.Sleep(2 * time.Second)

	// Perform an HTTP GET request on the metrics endpoint
	resp, err := http.Get(fmt.Sprintf("http://localhost:%d%s", config.MetricsPort, config.MetricsPath))
	if err != nil {
		t.Fatalf("Failed to send HTTP request to metrics endpoint: %v", err)
	}
	defer resp.Body.Close()

	// Check if the HTTP response status is 200 OK
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected HTTP 200 OK, got %v", resp.Status)
	}
}
