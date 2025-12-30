package driftq

import (
	"context"
	"errors"
	"sync"
	"time"
)

type StepHandler interface {
	Handle(ctx context.Context, msg ConsumeMessage) error
}

type StepFunc func(ctx context.Context, msg ConsumeMessage) error

func (f StepFunc) Handle(ctx context.Context, msg ConsumeMessage) error { return f(ctx, msg) }

type WorkerConfig struct {
	Client             *Client
	Consume            ConsumeOptions
	Handler            StepHandler
	Concurrency        int
	OnError            func(error)
	NackReason         func(ctx context.Context, msg ConsumeMessage, err error) string
	MaxNackReasonBytes int
}

type Worker struct {
	c   *Client
	opt ConsumeOptions
	h   StepHandler

	concurrency int
	onError     func(error)

	nackReason func(ctx context.Context, msg ConsumeMessage, err error) string
	maxReason  int
}

func NewWorker(cfg WorkerConfig) (*Worker, error) {
	if cfg.Client == nil {
		return nil, errors.New("worker: Client is required")
	}

	if cfg.Handler == nil {
		return nil, errors.New("worker: Handler is required")
	}

	if cfg.Consume.Topic == "" || cfg.Consume.Group == "" || cfg.Consume.Owner == "" {
		return nil, errors.New("worker: ConsumeOptions requires Topic, Group, Owner")
	}

	conc := cfg.Concurrency
	if conc <= 0 {
		conc = 1
	}

	maxReason := cfg.MaxNackReasonBytes
	if maxReason <= 0 {
		maxReason = 1024
	}

	nrf := cfg.NackReason
	if nrf == nil {
		nrf = func(_ context.Context, _ ConsumeMessage, err error) string {
			if err == nil {
				return ""
			}
			return err.Error()
		}
	}

	return &Worker{
		c:           cfg.Client,
		opt:         cfg.Consume,
		h:           cfg.Handler,
		concurrency: conc,
		onError:     cfg.OnError,
		nackReason:  nrf,
		maxReason:   maxReason,
	}, nil
}

// Run starts consuming and processing until ctx is cancelled or the server closes the stream
// ctx cancellation is treated as a normal shutdown (Run returns nil)
func (w *Worker) Run(ctx context.Context) error {
	msgs, errs, err := w.c.ConsumeStream(ctx, w.opt)
	if err != nil {
		return err
	}

	sem := make(chan struct{}, w.concurrency)
	var wg sync.WaitGroup

	wait := func() {
		wg.Wait()
	}

	for {
		select {
		case <-ctx.Done():
			wait()
			return nil

		case err, ok := <-errs:
			if !ok || err == nil {
				continue
			}

			w.report(err)
			wait()
			return err

		case m, ok := <-msgs:
			if !ok {
				wait()
				return nil
			}

			sem <- struct{}{}
			wg.Add(1)

			go func(msg ConsumeMessage) {
				defer wg.Done()
				defer func() { <-sem }()

				w.handleOne(ctx, msg)
			}(m)
		}
	}
}

func (w *Worker) handleOne(ctx context.Context, msg ConsumeMessage) {
	// Derive per-message ctx:
	// - If message envelope has a deadline, honor it (earlier deadline wins)
	hctx := ctx
	var cancel func()
	if dl := envelopeDeadline(msg); !dl.IsZero() {
		if cur, ok := ctx.Deadline(); !ok || dl.Before(cur) {
			hctx, cancel = context.WithDeadline(ctx, dl)
		}
	}

	if cancel != nil {
		defer cancel()
	}

	err := w.h.Handle(hctx, msg)

	if err == nil {
		ackErr := w.c.Ack(ctx, AckRequest{
			Topic:     w.opt.Topic,
			Group:     w.opt.Group,
			Owner:     w.opt.Owner,
			Partition: msg.Partition,
			Offset:    msg.Offset,
		})
		if ackErr != nil {
			w.report(ackErr)
		}
		return
	}

	reason := w.nackReason(hctx, msg, err)
	if len(reason) > w.maxReason {
		reason = reason[:w.maxReason]
	}

	nackErr := w.c.Nack(ctx, NackRequest{
		Topic:     w.opt.Topic,
		Group:     w.opt.Group,
		Owner:     w.opt.Owner,
		Partition: msg.Partition,
		Offset:    msg.Offset,
		Reason:    reason,
	})

	if nackErr != nil {
		w.report(nackErr)
	}
}

func (w *Worker) report(err error) {
	if err == nil || w.onError == nil {
		return
	}
	w.onError(err)
}

func envelopeDeadline(msg ConsumeMessage) time.Time {
	if msg.Envelope == nil || msg.Envelope.Deadline == nil {
		return time.Time{}
	}
	return *msg.Envelope.Deadline
}
