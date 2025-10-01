// internal/matter/client.go
package matter

import (
	"context"
	"fmt"
	"matter2mqtt/internal/matter/chiptool"
	"os"
	"os/exec"
)

type Client struct {
	storagePath string
	chiptool    chiptool.Client // NOT *chiptool.Client
	sessions    map[uint64]*Session
}

func NewClient(storagePath string, deviceNodeIDs []uint64) (*Client, error) {
	var client chiptool.Client

	if os.Getenv("MOCK_CHIPTOOL") == "true" {
		fmt.Println("Using mock chip-tool client")
		client = chiptool.NewMockClient()

		// Auto-commission all devices in mock mode
		ctx := context.Background()
		for _, nodeID := range deviceNodeIDs {
			if err := client.Commission(ctx, nodeID, "MOCK-SETUP-CODE"); err != nil {
				fmt.Printf("Warning: failed to commission mock device %d: %v\n", nodeID, err)
			}
		}
	} else {
		chipToolPath, err := findChipTool()
		if err != nil {
			return nil, fmt.Errorf("chip-tool not found: %w", err)
		}
		client = chiptool.NewClient(chipToolPath, storagePath)
	}

	return &Client{
		storagePath: storagePath,
		chiptool:    client,
		sessions:    make(map[uint64]*Session),
	}, nil
}

func (c *Client) Connect(nodeID uint64) (*Session, error) {
	if session, exists := c.sessions[nodeID]; exists {
		return session, nil
	}

	session := &Session{
		nodeID:   nodeID,
		chiptool: c.chiptool, // No pointer
		ctx:      context.Background(),
	}

	c.sessions[nodeID] = session
	return session, nil
}

func (c *Client) Close() error {
	return c.chiptool.Close()
}

func findChipTool() (string, error) {
	paths := []string{
		"./bin/chip-tool",
		"/usr/local/bin/chip-tool",
		"/opt/homebrew/bin/chip-tool",
		"./chip-tool",
		"chip-tool",
	}

	for _, path := range paths {
		if _, err := exec.LookPath(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("chip-tool binary not found in common locations")
}
