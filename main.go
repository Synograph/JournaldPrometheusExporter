package main

import (
	"bufio"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Config struct for the JSON configuration
type Config struct {
	Debug       bool           `json:"debug"`
	MetricsPort int            `json:"metrics_port"`
	MetricsPath string         `json:"metrics_path"`
	Events      []EventMonitor `json:"events"`
}

// EventMonitor defines the structure for monitoring specific events
type EventMonitor struct {
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	LogCommand    []string `json:"log_command"`
	MatchPatterns []string `json:"match_patterns"`
}

var config Config
var metricsMap map[string]*prometheus.CounterVec
var wg sync.WaitGroup

// LoadConfig loads the configuration from a JSON file
func LoadConfig(filename string) error {
	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	return decoder.Decode(&config)
}

// monitorEvent creates a goroutine that watches system logs and increments Prometheus counters
func monitorEvent(ctx context.Context, event EventMonitor) {
	defer wg.Done()

	if config.Debug {
		// Log the command being executed for this event monitor
		fmt.Printf("Launching journalctl command: '%s'\n", strings.Join(event.LogCommand, " "))
	}

	cmd := exec.CommandContext(ctx, event.LogCommand[0], event.LogCommand[1:]...)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("Error creating pipe for journalctl: %v", err)
		return
	}

	if err := cmd.Start(); err != nil {
		log.Printf("Error starting journalctl: %v", err)
		return
	}

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		for _, pattern := range event.MatchPatterns {
			if strings.Contains(line, pattern) {
				metricsMap[event.Name].With(prometheus.Labels{"pattern": pattern}).Inc()
				if config.Debug {
					// Print the full line and matched pattern for debugging
					fmt.Printf("Matched pattern: '%s' in event: '%s'\nFull log line: %s\n", pattern, event.Name, line)
				}
			}
		}
	}

	if err := cmd.Wait(); err != nil {
		log.Printf("Error while waiting for command to finish: %v", err)
	}
}

func main() {
	// Command-line flag for configuration file
	configFile := flag.String("config", "config.json", "Path to the JSON configuration file")
	flag.Parse()

	// Load JSON configuration
	if err := LoadConfig(*configFile); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize Prometheus metrics registry
	metricsMap = make(map[string]*prometheus.CounterVec)

	// Create a Prometheus counter for each event defined in the configuration
	for _, event := range config.Events {
		counter := prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: event.Name,
				Help: event.Description,
			},
			[]string{"pattern"},
		)
		// Register the counter with Prometheus
		prometheus.MustRegister(counter)
		metricsMap[event.Name] = counter

		// Start monitoring the event in a separate goroutine
		wg.Add(1)
		go monitorEvent(context.Background(), event)
	}

	// Start the HTTP server for Prometheus metrics
	http.Handle(config.MetricsPath, promhttp.Handler())
	listenAddr := fmt.Sprintf(":%d", config.MetricsPort)
	log.Printf("Starting Prometheus metrics server on %s\n", listenAddr)

	// Run HTTP server in a separate goroutine
	go func() {
		log.Fatal(http.ListenAndServe(listenAddr, nil))
	}()

	// Wait for all event monitoring goroutines to finish
	wg.Wait()
}
