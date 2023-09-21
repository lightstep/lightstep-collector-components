package requestsizeprocessor

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/processor"
	"go.opentelemetry.io/collector/processor/processorhelper"
)

const (
	// The value of "type" key in configuration.
	typeStr = "requestsize"
	// The stability level of the component.
	stability = component.StabilityLevelAlpha
)

// NewFactory creates a factory for the requestsize processor.
func NewFactory() processor.Factory {
	return processor.NewFactory(
		typeStr,
		CreateDefaultConfig,
		processor.WithTraces(createTracesProcessor, stability),
	)
}

// createTracesProcessor creates an instance of requestsize for processing traces
func createTracesProcessor(
	ctx context.Context,
	set processor.CreateSettings,
	cfg component.Config,
	next consumer.Traces,
) (processor.Traces, error) {
	oCfg := cfg.(*Config)
	processor := NewRequestSizeProcessor(oCfg, set.Logger)
	processor.nextTraces = next
	return processorhelper.NewTracesProcessor(
		ctx,
		set,
		cfg,
		next,
		processor.processTraces,
		processorhelper.WithCapabilities(processor.Capabilities()),
		processorhelper.WithStart(processor.Start),
		processorhelper.WithShutdown(processor.Shutdown))
}
