package chiptool

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strconv"
	"strings"
	"sync"
)

type Client struct {
	binaryPath    string
	storagePath   string
	mu            sync.RWMutex
	subscriptions map[string]*exec.Cmd // key -> running process
}

func NewClient(binaryPath, storagePath string) *Client {
	return &Client{
		binaryPath:    binaryPath,
		storagePath:   storagePath,
		subscriptions: make(map[string]*exec.Cmd),
	}
}

// Commission pairs a new device
func (c *Client) Commission(ctx context.Context, nodeID uint64, setupCode string) error {
	cmd := exec.CommandContext(ctx, c.binaryPath,
		"pairing", "code",
		fmt.Sprintf("%d", nodeID),
		setupCode,
		"--storage-directory", c.storagePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("commission failed: %w\nOutput: %s", err, output)
	}

	return nil
}

// SubscribeOccupancy subscribes to occupancy sensor changes
func (c *Client) SubscribeOccupancy(ctx context.Context, nodeID uint64, endpoint uint8, callback func(bool, error)) error {
	key := fmt.Sprintf("occupancy-%d-%d", nodeID, endpoint)

	cmd := exec.CommandContext(ctx, c.binaryPath,
		"occupancysensing", "subscribe", "occupancy",
		"1", "60", // min/max interval in seconds
		fmt.Sprintf("%d", nodeID),
		fmt.Sprintf("%d", endpoint),
		"--storage-directory", c.storagePath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	// Track subscription
	c.mu.Lock()
	c.subscriptions[key] = cmd
	c.mu.Unlock()

	// Parse output in goroutines
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.parseOccupancyOutput(stdout, callback)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		c.parseErrors(stderr, callback)
	}()

	// Cleanup when done
	go func() {
		wg.Wait()
		cmd.Wait()

		c.mu.Lock()
		delete(c.subscriptions, key)
		c.mu.Unlock()

		// Notify callback of disconnection
		callback(false, fmt.Errorf("subscription ended"))
	}()

	return nil
}

func (c *Client) parseOccupancyOutput(r io.Reader, callback func(bool, error)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		// chip-tool output format:
		// [timestamp][nodeID] CHIP: [TOO]   Occupancy: 1
		if strings.Contains(line, "Occupancy:") {
			parts := strings.Split(line, "Occupancy:")
			if len(parts) >= 2 {
				valueStr := strings.TrimSpace(parts[1])
				value := valueStr == "1" || strings.ToLower(valueStr) == "true"
				callback(value, nil)
			}
		}
	}
}

func (c *Client) parseErrors(r io.Reader, callback func(bool, error)) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "ERROR") || strings.Contains(line, "Error") {
			callback(false, fmt.Errorf("chip-tool error: %s", line))
		}
	}
}

// ReadOccupancy reads current occupancy state
func (c *Client) ReadOccupancy(ctx context.Context, nodeID uint64, endpoint uint8) (bool, error) {
	cmd := exec.CommandContext(ctx, c.binaryPath,
		"occupancysensing", "read", "occupancy",
		fmt.Sprintf("%d", nodeID),
		fmt.Sprintf("%d", endpoint),
		"--storage-directory", c.storagePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return false, fmt.Errorf("read failed: %w\nOutput: %s", err, output)
	}

	// Parse output
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "Occupancy:") {
			parts := strings.Split(line, "Occupancy:")
			if len(parts) >= 2 {
				valueStr := strings.TrimSpace(parts[1])
				return valueStr == "1" || strings.ToLower(valueStr) == "true", nil
			}
		}
	}

	return false, fmt.Errorf("could not parse occupancy from output")
}

// SubscribeIlluminance subscribes to illuminance sensor changes
func (c *Client) SubscribeIlluminance(ctx context.Context, nodeID uint64, endpoint uint8, callback func(uint16, error)) error {
	key := fmt.Sprintf("illuminance-%d-%d", nodeID, endpoint)

	cmd := exec.CommandContext(ctx, c.binaryPath,
		"illuminancemeasurement", "subscribe", "measured-value",
		"1", "60",
		fmt.Sprintf("%d", nodeID),
		fmt.Sprintf("%d", endpoint),
		"--storage-directory", c.storagePath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	c.mu.Lock()
	c.subscriptions[key] = cmd
	c.mu.Unlock()

	go func() {
		defer func() {
			cmd.Wait()
			c.mu.Lock()
			delete(c.subscriptions, key)
			c.mu.Unlock()
		}()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "MeasuredValue:") {
				parts := strings.Split(line, "MeasuredValue:")
				if len(parts) >= 2 {
					valueStr := strings.TrimSpace(parts[1])
					if value, err := strconv.ParseUint(valueStr, 10, 16); err == nil {
						callback(uint16(value), nil)
					}
				}
			}
		}
	}()

	return nil
}

// ReadIlluminance reads current illuminance value
func (c *Client) ReadIlluminance(ctx context.Context, nodeID uint64, endpoint uint8) (uint16, error) {
	cmd := exec.CommandContext(ctx, c.binaryPath,
		"illuminancemeasurement", "read", "measured-value",
		fmt.Sprintf("%d", nodeID),
		fmt.Sprintf("%d", endpoint),
		"--storage-directory", c.storagePath,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return 0, fmt.Errorf("read failed: %w", err)
	}

	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		if strings.Contains(line, "MeasuredValue:") {
			parts := strings.Split(line, "MeasuredValue:")
			if len(parts) >= 2 {
				valueStr := strings.TrimSpace(parts[1])
				if value, err := strconv.ParseUint(valueStr, 10, 16); err == nil {
					return uint16(value), nil
				}
			}
		}
	}

	return 0, fmt.Errorf("could not parse illuminance from output")
}

// SubscribeBattery subscribes to battery percentage changes
func (c *Client) SubscribeBattery(ctx context.Context, nodeID uint64, endpoint uint8, callback func(uint8, error)) error {
	key := fmt.Sprintf("battery-%d-%d", nodeID, endpoint)

	cmd := exec.CommandContext(ctx, c.binaryPath,
		"powersource", "subscribe", "bat-percent-remaining",
		"1", "300", // Battery updates less frequently
		fmt.Sprintf("%d", nodeID),
		fmt.Sprintf("%d", endpoint),
		"--storage-directory", c.storagePath,
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	c.mu.Lock()
	c.subscriptions[key] = cmd
	c.mu.Unlock()

	go func() {
		defer func() {
			cmd.Wait()
			c.mu.Lock()
			delete(c.subscriptions, key)
			c.mu.Unlock()
		}()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()

			if strings.Contains(line, "BatPercentRemaining:") {
				parts := strings.Split(line, "BatPercentRemaining:")
				if len(parts) >= 2 {
					valueStr := strings.TrimSpace(parts[1])
					if value, err := strconv.ParseUint(valueStr, 10, 8); err == nil {
						// Matter reports 0-200 (0-200%), convert to 0-100
						callback(uint8(value/2), nil)
					}
				}
			}
		}
	}()

	return nil
}

// Unsubscribe stops a specific subscription
func (c *Client) Unsubscribe(key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	cmd, exists := c.subscriptions[key]
	if !exists {
		return fmt.Errorf("subscription %s not found", key)
	}

	if cmd.Process != nil {
		if err := cmd.Process.Kill(); err != nil {
			return err
		}
	}

	delete(c.subscriptions, key)
	return nil
}

// Close stops all subscriptions
func (c *Client) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for key, cmd := range c.subscriptions {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		delete(c.subscriptions, key)
	}

	return nil
}
