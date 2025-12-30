package driftq

type Routing struct {
	Label string            `json:"label"`
	Meta  map[string]string `json:"meta,omitempty"`
}

type ConsumeMessage struct {
	Partition int       `json:"partition"`
	Offset    int64     `json:"offset"`
	Attempts  int       `json:"attempts"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	LastError string    `json:"last_error,omitempty"`
	Routing   *Routing  `json:"routing,omitempty"`
	Envelope  *Envelope `json:"envelope,omitempty"`
}

type ConsumeOptions struct {
	Topic   string
	Group   string
	Owner   string
	LeaseMS int64 // optional; 0 = server default
}
