package driftq

import "time"

type RetryPolicy struct {
	MaxAttempts  int   `json:"max_attempts,omitempty"`
	BackoffMs    int64 `json:"backoff_ms,omitempty"`
	MaxBackoffMs int64 `json:"max_backoff_ms,omitempty"`
}

type Envelope struct {
	RunID             string       `json:"run_id,omitempty"`
	StepID            string       `json:"step_id,omitempty"`
	ParentStepID      string       `json:"parent_step_id,omitempty"`
	TenantID          string       `json:"tenant_id,omitempty"`
	IdempotencyKey    string       `json:"idempotency_key,omitempty"`
	TargetTopic       string       `json:"target_topic,omitempty"`
	Deadline          *time.Time   `json:"deadline,omitempty"`
	PartitionOverride *int         `json:"partition_override,omitempty"`
	RetryPolicy       *RetryPolicy `json:"retry_policy,omitempty"`
}

type ProduceRequest struct {
	Topic    string    `json:"topic"`
	Key      string    `json:"key,omitempty"`
	Value    string    `json:"value"`
	Envelope *Envelope `json:"envelope,omitempty"`
}

type ProduceResponse struct {
	Status string `json:"status"`
	Topic  string `json:"topic"`
}

type AckRequest struct {
	Topic     string `json:"topic"`
	Group     string `json:"group"`
	Owner     string `json:"owner"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset"`
}

type NackRequest struct {
	Topic     string `json:"topic"`
	Group     string `json:"group"`
	Owner     string `json:"owner"`
	Partition int    `json:"partition"`
	Offset    int64  `json:"offset"`
	Reason    string `json:"reason,omitempty"`
}
