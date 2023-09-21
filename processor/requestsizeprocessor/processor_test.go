package requestsizeprocessor

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/ptrace"
	"go.uber.org/zap"
	"google.golang.org/grpc/metadata"
)

var (
	TestSpanStartTime      = time.Date(2020, 2, 11, 20, 26, 12, 321, time.UTC)
	TestSpanStartTimestamp = pcommon.NewTimestampFromTime(TestSpanStartTime)

	TestSpanEventTime      = time.Date(2020, 2, 11, 20, 26, 13, 123, time.UTC)
	TestSpanEventTimestamp = pcommon.NewTimestampFromTime(TestSpanEventTime)

	TestSpanEndTime      = time.Date(2020, 2, 11, 20, 26, 13, 789, time.UTC)
	TestSpanEndTimestamp = pcommon.NewTimestampFromTime(TestSpanEndTime)

	defaultTraces = GenerateTraces()
	pm            = &ptrace.ProtoMarshaler{}
	defaultSize   = pm.TracesSize(defaultTraces)
)

func initResource(r pcommon.Resource) {
	r.Attributes().PutStr("resource-attr", "resource-attr-val-1")
	r.Attributes().PutStr("service.name", "auth-service")
}

func GenerateTraces() ptrace.Traces {
	td := ptrace.NewTraces()

	rs0 := td.ResourceSpans().AppendEmpty()
	initResource(rs0.Resource())

	rs0ils0 := rs0.ScopeSpans().AppendEmpty()
	fillSpanOne(rs0ils0.Spans().AppendEmpty())

	return td
}

func GenSpanIDTraceID(seed int64) (pcommon.SpanID, pcommon.TraceID) {
	var sid pcommon.SpanID
	var tid pcommon.TraceID

	src := rand.NewSource(seed)
	rnd := rand.New(src)

	rnd.Read(sid[:])
	rnd.Read(tid[:])

	return sid, tid
}

func fillSpanOne(span ptrace.Span) {
	span.SetName("operationA")
	spanID, traceID := GenSpanIDTraceID(1)
	span.SetSpanID(spanID)
	span.SetTraceID(traceID)
	span.SetStartTimestamp(TestSpanStartTimestamp)
	span.SetEndTimestamp(TestSpanEndTimestamp)
	span.SetDroppedAttributesCount(1)

	evs := span.Events()
	ev0 := evs.AppendEmpty()
	ev0.SetTimestamp(TestSpanEventTimestamp)
	ev0.SetName("event-with-attr")
	ev0.Attributes().PutStr("span-event-attr", "span-event-attr-val")
	ev0.SetDroppedAttributesCount(2)

	ev1 := evs.AppendEmpty()
	ev1.SetTimestamp(TestSpanEventTimestamp)
	ev1.SetName("event")
	ev1.SetDroppedAttributesCount(2)

	span.SetDroppedEventsCount(1)
	status := span.Status()
	status.SetCode(ptrace.StatusCodeError)
	status.SetMessage("status-cancelled")
}

type ctxTraceConsumer struct {
	ctx context.Context
}

func (n *ctxTraceConsumer) ConsumeTraces(ctx context.Context, td ptrace.Traces) error {
	outgoing, ok := metadata.FromOutgoingContext(ctx)
	if !ok {
		return fmt.Errorf("failed to get context data")
	}
	bytes := outgoing.Get("lightstep-uncompressed-bytes")
	if len(bytes) != 1 {
		return fmt.Errorf("incorrect number of values for lightstep-uncompressed-bytes: %v", len(bytes))
	}
	n.ctx = metadata.AppendToOutgoingContext(n.ctx, "lightstep-uncompressed-bytes", bytes[0])
	return nil
}

func (n *ctxTraceConsumer) Capabilities() consumer.Capabilities {
	return consumer.Capabilities{MutatesData: false}
}

func TestProcessTraces(t *testing.T) {
	ctx := context.Background() 

	cfg := CreateDefaultConfig()

	rl := NewRequestSizeProcessor(cfg.(*Config), zap.NewNop())
	rl.nextTraces = &ctxTraceConsumer{ctx: ctx}

	td, err := rl.processTraces(ctx, defaultTraces)
	assert.NoError(t, err)

	// processor simply adds a header and does not alter traces at all.
	assert.Equal(t, defaultTraces, td)

	outgoing, ok := metadata.FromOutgoingContext(rl.nextTraces.(*ctxTraceConsumer).ctx)
	assert.True(t, ok)

	bytes := outgoing.Get("lightstep-uncompressed-bytes")
	assert.Len(t, bytes, 1)

	actualBytes, err := strconv.Atoi(bytes[0])
	assert.NoError(t, err)

	assert.Equal(t, defaultSize, actualBytes)
}
