package requestsizeprocessor

import (
	"context"
	"fmt"
	"strconv"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

type requestsize struct {
	nextTraces consumer.Traces
	logger *zap.Logger

	cfg *Config
}

func NewRequestSizeProcessor(cfg *Config, logger *zap.Logger) *requestsize {
	return &requestsize{
		logger: logger,
		cfg:    cfg,
	}
}

// ConsumeTraces is required by the processor.Traces interface.
func (rl *requestsize) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	_, err := rl.processTraces(ctx, td)
	return err
}

func (rl *requestsize) processTraces(ctx context.Context, batch ptrace.Traces) (ptrace.Traces, error) {
	pm := &ptrace.ProtoMarshaler{}
	sz := pm.TracesSize(batch)
	fmt.Println("HEADER NAME")
	fmt.Println(rl.cfg.HeaderName)
	fmt.Println(sz)

	ctx = metadata.AppendToOutgoingContext(
		ctx,
		rl.cfg.HeaderName,
		strconv.Itoa(sz),
	)
	return batch, rl.nextTraces.ConsumeTraces(ctx, batch)
}

func (rl *requestsize) Start(ctx context.Context, host component.Host) error {
	return nil
}

func (rl *requestsize) Shutdown(ctx context.Context) error {
	return nil
}

// Capabilities specifies what this processor does, such as whether it mutates data
func (rl *requestsize) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}