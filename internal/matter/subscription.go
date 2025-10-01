package matter

import (
	"context"
	"fmt"
	"matter2mqtt/internal/matter/chiptool"
)

type Subscription struct {
	NodeID      uint64
	ClusterID   uint32
	AttributeID uint32
	Callback    func(value interface{})
	key         string
}

type Session struct {
	nodeID   uint64
	chiptool chiptool.Client // NOT *chiptool.Client
	ctx      context.Context
}

func (s *Session) Subscribe(cluster, attribute uint32, callback func(interface{})) (*Subscription, error) {
	endpoint := uint8(1) // Most devices use endpoint 1

	sub := &Subscription{
		NodeID:      s.nodeID,
		ClusterID:   cluster,
		AttributeID: attribute,
		Callback:    callback,
	}

	// Route to appropriate subscription based on cluster
	switch cluster {
	case ClusterOccupancySensing:
		sub.key = fmt.Sprintf("occupancy-%d-%d", s.nodeID, endpoint)
		err := s.chiptool.SubscribeOccupancy(s.ctx, s.nodeID, endpoint, func(value bool, err error) {
			if err != nil {
				// Log error but don't stop
				return
			}
			callback(value)
		})
		return sub, err

	case ClusterIlluminanceMeasurement:
		sub.key = fmt.Sprintf("illuminance-%d-%d", s.nodeID, endpoint)
		err := s.chiptool.SubscribeIlluminance(s.ctx, s.nodeID, endpoint, func(value uint16, err error) {
			if err != nil {
				return
			}
			callback(value)
		})
		return sub, err

	case ClusterPowerSource:
		sub.key = fmt.Sprintf("battery-%d-%d", s.nodeID, endpoint)
		err := s.chiptool.SubscribeBattery(s.ctx, s.nodeID, endpoint, func(value uint8, err error) {
			if err != nil {
				return
			}
			callback(value)
		})
		return sub, err

	default:
		return nil, fmt.Errorf("unsupported cluster: 0x%04X", cluster)
	}
}

func (s *Session) ReadAttribute(cluster, attribute uint32) (interface{}, error) {
	endpoint := uint8(1)

	switch cluster {
	case ClusterOccupancySensing:
		return s.chiptool.ReadOccupancy(s.ctx, s.nodeID, endpoint)
	case ClusterIlluminanceMeasurement:
		return s.chiptool.ReadIlluminance(s.ctx, s.nodeID, endpoint)
	default:
		return nil, fmt.Errorf("unsupported cluster for read: 0x%04X", cluster)
	}
}

func (s *Session) Unsubscribe(sub *Subscription) error {
	return s.chiptool.Unsubscribe(sub.key)
}
