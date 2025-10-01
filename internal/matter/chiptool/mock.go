// internal/matter/chiptool/mock.go
package chiptool

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type MockClient struct {
	mu            sync.RWMutex
	subscriptions map[string]context.CancelFunc
	commissioned  map[uint64]bool
}

func NewMockClient() *MockClient {
	return &MockClient{
		subscriptions: make(map[string]context.CancelFunc),
		commissioned:  make(map[uint64]bool),
	}
}

func (m *MockClient) Commission(ctx context.Context, nodeID uint64, setupCode string) error {
	// Simulate commissioning delay
	time.Sleep(500 * time.Millisecond)

	m.mu.Lock()
	m.commissioned[nodeID] = true
	m.mu.Unlock()

	fmt.Printf("[MOCK] Commissioned device %d with code %s\n", nodeID, setupCode)
	return nil
}

func (m *MockClient) SubscribeOccupancy(ctx context.Context, nodeID uint64, endpoint uint8, callback func(bool, error)) error {
	key := fmt.Sprintf("occupancy-%d-%d", nodeID, endpoint)

	m.mu.Lock()
	if !m.commissioned[nodeID] {
		m.mu.Unlock()
		return fmt.Errorf("device %d not commissioned", nodeID)
	}
	m.mu.Unlock()

	// Create cancellable context for this subscription
	subCtx, cancel := context.WithCancel(ctx)

	m.mu.Lock()
	m.subscriptions[key] = cancel
	m.mu.Unlock()

	go func() {
		fmt.Printf("[MOCK] Started occupancy subscription for node %d endpoint %d\n", nodeID, endpoint)

		// Initial state
		occupied := false
		callback(occupied, nil)

		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-subCtx.Done():
				fmt.Printf("[MOCK] Stopped occupancy subscription for node %d\n", nodeID)
				return
			case <-ticker.C:
				// Simulate realistic occupancy patterns
				// 30% chance to toggle state
				if rand.Float32() < 0.3 {
					occupied = !occupied
					fmt.Printf("[MOCK] Occupancy changed: %v (node %d)\n", occupied, nodeID)
					callback(occupied, nil)
				}
			}
		}
	}()

	return nil
}

func (m *MockClient) ReadOccupancy(ctx context.Context, nodeID uint64, endpoint uint8) (bool, error) {
	m.mu.RLock()
	commissioned := m.commissioned[nodeID]
	m.mu.RUnlock()

	if !commissioned {
		return false, fmt.Errorf("device %d not commissioned", nodeID)
	}

	// Simulate read delay
	time.Sleep(100 * time.Millisecond)

	// Random state
	value := rand.Float32() > 0.5
	fmt.Printf("[MOCK] Read occupancy: %v (node %d)\n", value, nodeID)
	return value, nil
}

func (m *MockClient) SubscribeIlluminance(ctx context.Context, nodeID uint64, endpoint uint8, callback func(uint16, error)) error {
	key := fmt.Sprintf("illuminance-%d-%d", nodeID, endpoint)

	m.mu.Lock()
	if !m.commissioned[nodeID] {
		m.mu.Unlock()
		return fmt.Errorf("device %d not commissioned", nodeID)
	}
	m.mu.Unlock()

	subCtx, cancel := context.WithCancel(ctx)

	m.mu.Lock()
	m.subscriptions[key] = cancel
	m.mu.Unlock()

	go func() {
		fmt.Printf("[MOCK] Started illuminance subscription for node %d endpoint %d\n", nodeID, endpoint)

		// Start with medium brightness
		currentLux := uint16(500)
		callback(currentLux, nil)

		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-subCtx.Done():
				fmt.Printf("[MOCK] Stopped illuminance subscription for node %d\n", nodeID)
				return
			case <-ticker.C:
				// Simulate gradual changes in light level
				change := int16(rand.Intn(200) - 100) // -100 to +100
				newLux := int16(currentLux) + change

				// Keep in valid range
				if newLux < 0 {
					newLux = 0
				} else if newLux > 1000 {
					newLux = 1000
				}

				currentLux = uint16(newLux)
				fmt.Printf("[MOCK] Illuminance changed: %d lux (node %d)\n", currentLux, nodeID)
				callback(currentLux, nil)
			}
		}
	}()

	return nil
}

func (m *MockClient) ReadIlluminance(ctx context.Context, nodeID uint64, endpoint uint8) (uint16, error) {
	m.mu.RLock()
	commissioned := m.commissioned[nodeID]
	m.mu.RUnlock()

	if !commissioned {
		return 0, fmt.Errorf("device %d not commissioned", nodeID)
	}

	time.Sleep(100 * time.Millisecond)

	value := uint16(rand.Intn(1000))
	fmt.Printf("[MOCK] Read illuminance: %d lux (node %d)\n", value, nodeID)
	return value, nil
}

func (m *MockClient) SubscribeBattery(ctx context.Context, nodeID uint64, endpoint uint8, callback func(uint8, error)) error {
	key := fmt.Sprintf("battery-%d-%d", nodeID, endpoint)

	m.mu.Lock()
	if !m.commissioned[nodeID] {
		m.mu.Unlock()
		return fmt.Errorf("device %d not commissioned", nodeID)
	}
	m.mu.Unlock()

	subCtx, cancel := context.WithCancel(ctx)

	m.mu.Lock()
	m.subscriptions[key] = cancel
	m.mu.Unlock()

	go func() {
		fmt.Printf("[MOCK] Started battery subscription for node %d endpoint %d\n", nodeID, endpoint)

		// Start at 87% battery
		battery := uint8(87)
		callback(battery, nil)

		// Battery updates very slowly (every 5 minutes in mock)
		ticker := time.NewTicker(5 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-subCtx.Done():
				fmt.Printf("[MOCK] Stopped battery subscription for node %d\n", nodeID)
				return
			case <-ticker.C:
				// Slowly drain battery
				if battery > 0 && rand.Float32() < 0.5 {
					battery--
					fmt.Printf("[MOCK] Battery level: %d%% (node %d)\n", battery, nodeID)
					callback(battery, nil)
				}
			}
		}
	}()

	return nil
}

func (m *MockClient) Unsubscribe(key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cancel, exists := m.subscriptions[key]
	if !exists {
		return fmt.Errorf("subscription %s not found", key)
	}

	cancel()
	delete(m.subscriptions, key)
	fmt.Printf("[MOCK] Unsubscribed: %s\n", key)
	return nil
}

func (m *MockClient) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	fmt.Println("[MOCK] Closing all subscriptions")
	for key, cancel := range m.subscriptions {
		cancel()
		delete(m.subscriptions, key)
	}

	return nil
}
