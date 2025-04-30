package handlers

func (c *SeedLinkClient) SetSequence(sequence int64) {
	c.sequenceMutex.Lock()
	defer c.sequenceMutex.Unlock()
	c.sequence = sequence
}

func (c *SeedLinkClient) GetSequence() int64 {
	c.sequenceMutex.Lock()
	defer c.sequenceMutex.Unlock()
	return c.sequence
}
