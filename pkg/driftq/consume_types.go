package driftq

type Routing struct {
	Label string            `json:"label,omitempty"`
	Meta  map[string]string `json:"meta,omitempty"`
}

type ConsumedMessage struct {
	Partition int       `json:"partition"`
	Offset    int64     `json:"offset"`
	Attempts  int       `json:"attempts"`
	Key       string    `json:"key,omitempty"`
	Value     string    `json:"value,omitempty"`
	LastError string    `json:"last_error,omitempty"`
	Routing   *Routing  `json:"routing,omitempty"`
	Envelope  *Envelope `json:"envelope,omitempty"`
}

type ConsumeOptions struct {
	Topic   string
	Group   string
	Owner   string
	LeaseMS int // optional
}
