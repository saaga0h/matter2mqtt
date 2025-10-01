package matter

type Subscription struct {
	NodeID      uint64
	ClusterID   uint32
	AttributeID uint32
	Callback    func(value interface{})
}

type Session struct {
	nodeID uint64
	// TODO: Connection state
}

func (s *Session) Subscribe(cluster, attribute uint32, callback func(interface{})) (*Subscription, error) {
	// TODO: Subscribe to attribute changes
	// TODO: When change occurs, invoke callback
	sub := &Subscription{
		NodeID:      s.nodeID,
		ClusterID:   cluster,
		AttributeID: attribute,
		Callback:    callback,
	}
	return sub, nil
}

func (s *Session) ReadAttribute(cluster, attribute uint32) (interface{}, error) {
	// TODO: Read current attribute value
	return nil, nil
}
