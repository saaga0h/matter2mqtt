package matter

type Client struct {
	fabricPath string
	sessions   map[uint64]*Session
}

func NewClient(fabricPath string) (*Client, error) {
	// TODO: Load fabric credentials
	// TODO: Initialize Matter stack (or wrap chip-tool)
	return &Client{
		fabricPath: fabricPath,
		sessions:   make(map[uint64]*Session),
	}, nil
}

func (c *Client) Connect(nodeID uint64) (*Session, error) {
	// TODO: Establish CASE session to device
	// For now, return a stub session
	session := &Session{
		nodeID: nodeID,
	}
	c.sessions[nodeID] = session
	return session, nil
}
