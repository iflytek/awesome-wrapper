package mock

// Span define span as a fixed size `map[string]string`
type Span struct {
	// identifier for a trace, set on all items within it.
	TraceId []byte
	// span name , rpc method for example.
	Name string
	// identifier of this span within a trace.
	// Id []byte

	// span id for address & ts
	spanIdTs []byte
	// short span id for compute
	spanIdHierarchy []byte
	// meta for TraceId+spanIdTs+spanIdHierarchy
	meta []byte

	// epoch microseconds of the start of this span.
	Timestamp int64
	// measurement in microseconds of the critical path.
	Duration int64

	// annotations
	annotations map[string]int64
	// tags
	tags map[string]string
	// span type.
	spanType int32
	// current child id
	currentChildId int32
}

func NewSpan(spanType int32, abandon bool) *Span {
	return nil
}

func (span *Span) Next(spanType int32) *Span {
	return span
}

func FromMeta(meta string, spanType int32) *Span {
	return nil
}

func (span *Span) WithName(name string) *Span {
	return span
}

func (span *Span) Start() *Span {
	return span
}

// End records the duration of rpc span.
func (span *Span) End() *Span {
	return span
}

// Send records the message send timestamp of mq span.
func (span *Span) Send() *Span {
	return span
}

// Recv records the message receive timestamp of mq span.
func (span *Span) Recv() *Span {
	return span
}

// WithTag set custom tag.
func (span *Span) WithTag(key string, value string) *Span {
	return span
}

// WithRetTag set ret
func (span *Span) WithRetTag(value string) *Span {
	return span
}

// WithErrorTag set error
func (span *Span) WithErrorTag(value string) *Span {
	return span
}

// WithLocalComponent set local component.
func (span *Span) WithLocalComponent() *Span {
	return span
}

// WithClientAddr set client address.
func (span *Span) WithClientAddr() *Span {
	return span
}

// WithServerAddr set server address.
func (span *Span) WithServerAddr() *Span {
	return span
}

// WithMessageAddr set message address.
func (span *Span) WithMessageAddr() *Span {
	return span
}

// WithDescf set desc with format, `fmt.Sprintf` will cause performance
// Deprecated: this function will cause performance duo to `fmt.Printf()`
func (span *Span) WithDescf(format string, values ...interface{}) *Span {
	return span
}

// public property functions
// Meta gets meta string, format: <traceId>#<id>.
func (span *Span) Meta() string {
	return ""
}

func (span *Span) ToString() string {
	return ""
}
