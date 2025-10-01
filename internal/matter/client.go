// internal/matter/client.go
package matter

import (
	"context"
	"fmt"
	"matter2mqtt/internal/matter/chiptool"
	"os/exec"
)

type Client struct {
	storagePath string
	chiptool    *chiptool.Client
	sessions    map[uint64]*Session
}

func NewClient(storagePath string) (*Client, error) {
	// Find chip-tool binary
	chipToolPath, err := findChipTool()
	if err != nil {
		return nil, fmt.Errorf("chip-tool not found: %w", err)
	}

	return &Client{
		storagePath: storagePath,
		chiptool:    chiptool.NewClient(chipToolPath, storagePath),
		sessions:    make(map[uint64]*Session),
	}, nil
}

func findChipTool() (string, error) {
	// Check common locations
	paths := []string{
		"/usr/local/bin/chip-tool",
		"/usr/bin/chip-tool",
		"./bin/chip-tool",
		"chip-tool", // in PATH
	}

	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("chip-tool binary not found in common locations")
}

func (c *Client) Connect(nodeID uint64) (*Session, error) {
	if session, exists := c.sessions[nodeID]; exists {
		return session, nil
	}

	session := &Session{
		nodeID:   nodeID,
		chiptool: c.chiptool,
		ctx:      context.Background(),
	}

	c.sessions[nodeID] = session
	return session, nil
}

func (c *Client) Close() error {
	return c.chiptool.Close()
}
