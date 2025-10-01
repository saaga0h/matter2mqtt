// internal/matter/chiptool/interface.go
package chiptool

import "context"

// Client interface that both real and mock implement
type Client interface {
	Commission(ctx context.Context, nodeID uint64, setupCode string) error
	SubscribeOccupancy(ctx context.Context, nodeID uint64, endpoint uint8, callback func(bool, error)) error
	ReadOccupancy(ctx context.Context, nodeID uint64, endpoint uint8) (bool, error)
	SubscribeIlluminance(ctx context.Context, nodeID uint64, endpoint uint8, callback func(uint16, error)) error
	ReadIlluminance(ctx context.Context, nodeID uint64, endpoint uint8) (uint16, error)
	SubscribeBattery(ctx context.Context, nodeID uint64, endpoint uint8, callback func(uint8, error)) error
	Unsubscribe(key string) error
	Close() error
}
