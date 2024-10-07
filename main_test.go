package main

import (
	"context"
	"os"
	"os/exec"
	"testing"

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
		"metrics_port": 9090,
		"metrics_path": "/metrics",
		"events": [
			{
				"name": "test_event",
				"description": "Test event monitoring",
				"log_command": ["echo", "test log"],
				"match_patterns": ["test"]
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
	if config.MetricsPort != 9090 {
		t.Errorf("Expected MetricsPort to be 9090, got %d", config.MetricsPort)
	}
	if len(config.Events) != 1 || config.Events[0].Name != "test_event" {
		t.Errorf("Expected event 'test_event', got %+v", config.Events)
	}
}

// fakeExecCommand simulates the behavior of exec.Command for testing.
func fakeExecCommand(command string, args ...string) *exec.Cmd {
	cmd := exec.Command("echo", "This is a test log with pattern")
	return cmd
}

func TestMonitorEvent(t *testing.T) {
	// Mock exec.Command with fakeExecCommand
	execCommand = fakeExecCommand

	// Initialize a fake event to monitor
	event := EventMonitor{
		Name:          "test_event",
		Description:   "Test event",
		LogCommand:    []string{"journalctl", "-f"},
		MatchPatterns: []string{"pattern"},
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

	// Wait for the goroutine to finish
	wg.Wait()

	// Check if the counter was incremented
	counter := metricsMap[event.Name].With(prometheus.Labels{"pattern": "pattern"})
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
}
