package metrics

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	pb "github.com/libp2p/go-libp2p-pubsub/pb"
	libp2pmetrics "github.com/libp2p/go-libp2p/core/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/rpc"

	"github.com/ethereum-optimism/optimism/op-node/eth"
)

const (
	Namespace = "op_node"

	RPCServerSubsystem = "rpc_server"
	RPCClientSubsystem = "rpc_client"

	BatchMethod = "<batch>"
)

type Metricer interface {
	RecordInfo(version string)
	RecordUp()
	RecordRPCServerRequest(method string) func()
	RecordRPCClientRequest(method string) func(err error)
	RecordRPCClientResponse(method string, err error)
	SetDerivationIdle(status bool)
	RecordPipelineReset()
	RecordSequencingError()
	RecordPublishingError()
	RecordDerivationError()
	RecordReceivedUnsafePayload(payload *eth.ExecutionPayload)
	recordRef(layer string, name string, num uint64, timestamp uint64, h common.Hash)
	RecordL1Ref(name string, ref eth.L1BlockRef)
	RecordL2Ref(name string, ref eth.L2BlockRef)
	RecordUnsafePayloadsBuffer(length uint64, memSize uint64, next eth.BlockID)
	CountSequencedTxs(count int)
	RecordL1ReorgDepth(d uint64)
	RecordGossipEvent(evType int32)
	IncPeerCount()
	DecPeerCount()
	IncStreamCount()
	DecStreamCount()
	RecordBandwidth(ctx context.Context, bwc *libp2pmetrics.BandwidthCounter)
	RecordSequencerBuildingDiffTime(duration time.Duration)
	RecordSequencerSealingTime(duration time.Duration)
}

type Metrics struct {
	Info *prometheus.GaugeVec
	Up   prometheus.Gauge

	RPCServerRequestsTotal          *prometheus.CounterVec
	RPCServerRequestDurationSeconds *prometheus.HistogramVec
	RPCClientRequestsTotal          *prometheus.CounterVec
	RPCClientRequestDurationSeconds *prometheus.HistogramVec
	RPCClientResponsesTotal         *prometheus.CounterVec

	L1SourceCache *CacheMetrics
	L2SourceCache *CacheMetrics

	DerivationIdle prometheus.Gauge

	PipelineResets   *EventMetrics
	UnsafePayloads   *EventMetrics
	DerivationErrors *EventMetrics
	SequencingErrors *EventMetrics
	PublishingErrors *EventMetrics

	SequencerBuildingDiffDurationSeconds prometheus.Histogram
	SequencerBuildingDiffTotal           prometheus.Counter

	SequencerSealingDurationSeconds prometheus.Histogram
	SequencerSealingTotal           prometheus.Counter

	UnsafePayloadsBufferLen     prometheus.Gauge
	UnsafePayloadsBufferMemSize prometheus.Gauge

	RefsNumber  *prometheus.GaugeVec
	RefsTime    *prometheus.GaugeVec
	RefsHash    *prometheus.GaugeVec
	RefsSeqNr   *prometheus.GaugeVec
	RefsLatency *prometheus.GaugeVec
	// hash of the last seen block per name, so we don't reduce/increase latency on updates of the same data,
	// and only count the first occurrence
	LatencySeen map[string]common.Hash

	L1ReorgDepth prometheus.Histogram

	TransactionsSequencedTotal prometheus.Counter

	// P2P Metrics
	PeerCount         prometheus.Gauge
	StreamCount       prometheus.Gauge
	GossipEventsTotal *prometheus.CounterVec
	BandwidthTotal    *prometheus.GaugeVec

	registry *prometheus.Registry
}

var _ Metricer = (*Metrics)(nil)

func NewMetrics(procName string) *Metrics {
	if procName == "" {
		procName = "default"
	}
	ns := Namespace + "_" + procName

	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	registry.MustRegister(collectors.NewGoCollector())
	return &Metrics{
		Info: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "info",
			Help:      "Pseudo-metric tracking version and config info",
		}, []string{
			"version",
		}),
		Up: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "up",
			Help:      "1 if the op node has finished starting up",
		}),

		RPCServerRequestsTotal: promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: RPCServerSubsystem,
			Name:      "requests_total",
			Help:      "Total requests to the RPC server",
		}, []string{
			"method",
		}),
		RPCServerRequestDurationSeconds: promauto.With(registry).NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: RPCServerSubsystem,
			Name:      "request_duration_seconds",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			Help:      "Histogram of RPC server request durations",
		}, []string{
			"method",
		}),
		RPCClientRequestsTotal: promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: RPCClientSubsystem,
			Name:      "requests_total",
			Help:      "Total RPC requests initiated by the opnode's RPC client",
		}, []string{
			"method",
		}),
		RPCClientRequestDurationSeconds: promauto.With(registry).NewHistogramVec(prometheus.HistogramOpts{
			Namespace: ns,
			Subsystem: RPCClientSubsystem,
			Name:      "request_duration_seconds",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			Help:      "Histogram of RPC client request durations",
		}, []string{
			"method",
		}),
		RPCClientResponsesTotal: promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: RPCClientSubsystem,
			Name:      "responses_total",
			Help:      "Total RPC request responses received by the opnode's RPC client",
		}, []string{
			"method",
			"error",
		}),

		L1SourceCache: NewCacheMetrics(registry, ns, "l1_source_cache", "L1 Source cache"),
		L2SourceCache: NewCacheMetrics(registry, ns, "l2_source_cache", "L2 Source cache"),

		DerivationIdle: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "derivation_idle",
			Help:      "1 if the derivation pipeline is idle",
		}),

		PipelineResets:   NewEventMetrics(registry, ns, "pipeline_resets", "derivation pipeline resets"),
		UnsafePayloads:   NewEventMetrics(registry, ns, "unsafe_payloads", "unsafe payloads"),
		DerivationErrors: NewEventMetrics(registry, ns, "derivation_errors", "derivation errors"),
		SequencingErrors: NewEventMetrics(registry, ns, "sequencing_errors", "sequencing errors"),
		PublishingErrors: NewEventMetrics(registry, ns, "publishing_errors", "p2p publishing errors"),

		UnsafePayloadsBufferLen: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "unsafe_payloads_buffer_len",
			Help:      "Number of buffered L2 unsafe payloads",
		}),
		UnsafePayloadsBufferMemSize: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "unsafe_payloads_buffer_mem_size",
			Help:      "Total estimated memory size of buffered L2 unsafe payloads",
		}),

		RefsNumber: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "refs_number",
			Help:      "Gauge representing the different L1/L2 reference block numbers",
		}, []string{
			"layer",
			"type",
		}),
		RefsTime: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "refs_time",
			Help:      "Gauge representing the different L1/L2 reference block timestamps",
		}, []string{
			"layer",
			"type",
		}),
		RefsHash: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "refs_hash",
			Help:      "Gauge representing the different L1/L2 reference block hashes truncated to float values",
		}, []string{
			"layer",
			"type",
		}),
		RefsSeqNr: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "refs_seqnr",
			Help:      "Gauge representing the different L2 reference sequence numbers",
		}, []string{
			"type",
		}),
		RefsLatency: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "refs_latency",
			Help:      "Gauge representing the different L1/L2 reference block timestamps minus current time, in seconds",
		}, []string{
			"layer",
			"type",
		}),
		LatencySeen: make(map[string]common.Hash),

		L1ReorgDepth: promauto.With(registry).NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Name:      "l1_reorg_depth",
			Buckets:   []float64{0.5, 1.5, 2.5, 3.5, 4.5, 5.5, 6.5, 7.5, 8.5, 9.5, 10.5, 20.5, 50.5, 100.5},
			Help:      "Histogram of L1 Reorg Depths",
		}),

		TransactionsSequencedTotal: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Name:      "transactions_sequenced_total",
			Help:      "Count of total transactions sequenced",
		}),

		PeerCount: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Subsystem: "p2p",
			Name:      "peer_count",
			Help:      "Count of currently connected p2p peers",
		}),
		StreamCount: promauto.With(registry).NewGauge(prometheus.GaugeOpts{
			Namespace: ns,
			Subsystem: "p2p",
			Name:      "stream_count",
			Help:      "Count of currently connected p2p streams",
		}),
		GossipEventsTotal: promauto.With(registry).NewCounterVec(prometheus.CounterOpts{
			Namespace: ns,
			Subsystem: "p2p",
			Name:      "gossip_events_total",
			Help:      "Count of gossip events by type",
		}, []string{
			"type",
		}),
		BandwidthTotal: promauto.With(registry).NewGaugeVec(prometheus.GaugeOpts{
			Namespace: ns,
			Subsystem: "p2p",
			Name:      "bandwidth_bytes_total",
			Help:      "P2P bandwidth by direction",
		}, []string{
			"direction",
		}),

		SequencerBuildingDiffDurationSeconds: promauto.With(registry).NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Name:      "sequencer_building_diff_seconds",
			Buckets: []float64{
				-10, -5, -2.5, -1, -.5, -.25, -.1, -0.05, -0.025, -0.01, -0.005,
				.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			Help: "Histogram of Sequencer building time, minus block time",
		}),
		SequencerBuildingDiffTotal: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Name:      "sequencer_building_diff_total",
			Help:      "Number of sequencer block building jobs",
		}),
		SequencerSealingDurationSeconds: promauto.With(registry).NewHistogram(prometheus.HistogramOpts{
			Namespace: ns,
			Name:      "sequencer_sealing_seconds",
			Buckets:   []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
			Help:      "Histogram of Sequencer block sealing time",
		}),
		SequencerSealingTotal: promauto.With(registry).NewCounter(prometheus.CounterOpts{
			Namespace: ns,
			Name:      "sequencer_sealing_total",
			Help:      "Number of sequencer block sealing jobs",
		}),

		registry: registry,
	}
}

// RecordInfo sets a pseudo-metric that contains versioning and
// config info for the opnode.
func (m *Metrics) RecordInfo(version string) {
	m.Info.WithLabelValues(version).Set(1)
}

// RecordUp sets the up metric to 1.
func (m *Metrics) RecordUp() {
	prometheus.MustRegister()
	m.Up.Set(1)
}

// RecordRPCServerRequest is a helper method to record an incoming RPC
// call to the opnode's RPC server. It bumps the requests metric,
// and tracks how long it takes to serve a response.
func (m *Metrics) RecordRPCServerRequest(method string) func() {
	m.RPCServerRequestsTotal.WithLabelValues(method).Inc()
	timer := prometheus.NewTimer(m.RPCServerRequestDurationSeconds.WithLabelValues(method))
	return func() {
		timer.ObserveDuration()
	}
}

// RecordRPCClientRequest is a helper method to record an RPC client
// request. It bumps the requests metric, tracks the response
// duration, and records the response's error code.
func (m *Metrics) RecordRPCClientRequest(method string) func(err error) {
	m.RPCClientRequestsTotal.WithLabelValues(method).Inc()
	timer := prometheus.NewTimer(m.RPCClientRequestDurationSeconds.WithLabelValues(method))
	return func(err error) {
		m.RecordRPCClientResponse(method, err)
		timer.ObserveDuration()
	}
}

// RecordRPCClientResponse records an RPC response. It will
// convert the passed-in error into something metrics friendly.
// Nil errors get converted into <nil>, RPC errors are converted
// into rpc_<error code>, HTTP errors are converted into
// http_<status code>, and everything else is converted into
// <unknown>.
func (m *Metrics) RecordRPCClientResponse(method string, err error) {
	var errStr string
	var rpcErr rpc.Error
	var httpErr rpc.HTTPError
	if err == nil {
		errStr = "<nil>"
	} else if errors.As(err, &rpcErr) {
		errStr = fmt.Sprintf("rpc_%d", rpcErr.ErrorCode())
	} else if errors.As(err, &httpErr) {
		errStr = fmt.Sprintf("http_%d", httpErr.StatusCode)
	} else if errors.Is(err, ethereum.NotFound) {
		errStr = "<not found>"
	} else {
		errStr = "<unknown>"
	}
	m.RPCClientResponsesTotal.WithLabelValues(method, errStr).Inc()
}

func (m *Metrics) SetDerivationIdle(status bool) {
	var val float64
	if status {
		val = 1
	}
	m.DerivationIdle.Set(val)
}

func (m *Metrics) RecordPipelineReset() {
	m.PipelineResets.RecordEvent()
}

func (m *Metrics) RecordSequencingError() {
	m.SequencingErrors.RecordEvent()
}

func (m *Metrics) RecordPublishingError() {
	m.PublishingErrors.RecordEvent()
}

func (m *Metrics) RecordDerivationError() {
	m.DerivationErrors.RecordEvent()
}

func (m *Metrics) RecordReceivedUnsafePayload(payload *eth.ExecutionPayload) {
	m.UnsafePayloads.RecordEvent()
	m.recordRef("l2", "received_payload", uint64(payload.BlockNumber), uint64(payload.Timestamp), payload.BlockHash)
}

func (m *Metrics) recordRef(layer string, name string, num uint64, timestamp uint64, h common.Hash) {
	m.RefsNumber.WithLabelValues(layer, name).Set(float64(num))
	if timestamp != 0 {
		m.RefsTime.WithLabelValues(layer, name).Set(float64(timestamp))
		// only meter the latency when we first see this hash for the given label name
		if m.LatencySeen[name] != h {
			m.LatencySeen[name] = h
			m.RefsLatency.WithLabelValues(layer, name).Set(float64(timestamp) - (float64(time.Now().UnixNano()) / 1e9))
		}
	}
	// we map the first 8 bytes to a float64, so we can graph changes of the hash to find divergences visually.
	// We don't do math.Float64frombits, just a regular conversion, to keep the value within a manageable range.
	m.RefsHash.WithLabelValues(layer, name).Set(float64(binary.LittleEndian.Uint64(h[:])))
}

func (m *Metrics) RecordL1Ref(name string, ref eth.L1BlockRef) {
	m.recordRef("l1", name, ref.Number, ref.Time, ref.Hash)
}

func (m *Metrics) RecordL2Ref(name string, ref eth.L2BlockRef) {
	m.recordRef("l2", name, ref.Number, ref.Time, ref.Hash)
	m.recordRef("l1_origin", name, ref.L1Origin.Number, 0, ref.L1Origin.Hash)
	m.RefsSeqNr.WithLabelValues(name).Set(float64(ref.SequenceNumber))
}

func (m *Metrics) RecordUnsafePayloadsBuffer(length uint64, memSize uint64, next eth.BlockID) {
	m.recordRef("l2", "l2_buffer_unsafe", next.Number, 0, next.Hash)
	m.UnsafePayloadsBufferLen.Set(float64(length))
	m.UnsafePayloadsBufferMemSize.Set(float64(memSize))
}

func (m *Metrics) CountSequencedTxs(count int) {
	m.TransactionsSequencedTotal.Add(float64(count))
}

func (m *Metrics) RecordL1ReorgDepth(d uint64) {
	m.L1ReorgDepth.Observe(float64(d))
}

func (m *Metrics) RecordGossipEvent(evType int32) {
	m.GossipEventsTotal.WithLabelValues(pb.TraceEvent_Type_name[evType]).Inc()
}

func (m *Metrics) IncPeerCount() {
	m.PeerCount.Inc()
}

func (m *Metrics) DecPeerCount() {
	m.PeerCount.Dec()
}

func (m *Metrics) IncStreamCount() {
	m.StreamCount.Inc()
}

func (m *Metrics) DecStreamCount() {
	m.StreamCount.Dec()
}

func (m *Metrics) RecordBandwidth(ctx context.Context, bwc *libp2pmetrics.BandwidthCounter) {
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			bwTotals := bwc.GetBandwidthTotals()
			m.BandwidthTotal.WithLabelValues("in").Set(float64(bwTotals.TotalIn))
			m.BandwidthTotal.WithLabelValues("out").Set(float64(bwTotals.TotalOut))
		case <-ctx.Done():
			return
		}
	}
}

// RecordSequencerBuildingDiffTime tracks the amount of time the sequencer was allowed between
// start to finish, incl. sealing, minus the block time.
// Ideally this is 0, realistically the sequencer scheduler may be busy with other jobs like syncing sometimes.
func (m *Metrics) RecordSequencerBuildingDiffTime(duration time.Duration) {
	m.SequencerBuildingDiffTotal.Inc()
	m.SequencerBuildingDiffDurationSeconds.Observe(float64(duration) / float64(time.Second))
}

// RecordSequencerSealingTime tracks the amount of time the sequencer took to finish sealing the block.
// Ideally this is 0, realistically it may take some time.
func (m *Metrics) RecordSequencerSealingTime(duration time.Duration) {
	m.SequencerSealingTotal.Inc()
	m.SequencerSealingDurationSeconds.Observe(float64(duration) / float64(time.Second))
}

// Serve starts the metrics server on the given hostname and port.
// The server will be closed when the passed-in context is cancelled.
func (m *Metrics) Serve(ctx context.Context, hostname string, port int) error {
	addr := net.JoinHostPort(hostname, strconv.Itoa(port))
	server := &http.Server{
		Addr: addr,
		Handler: promhttp.InstrumentMetricHandler(
			m.registry, promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{}),
		),
	}
	go func() {
		<-ctx.Done()
		server.Close()
	}()
	return server.ListenAndServe()
}

type noopMetricer struct{}

var NoopMetrics Metricer = new(noopMetricer)

func (n *noopMetricer) RecordInfo(version string) {
}

func (n *noopMetricer) RecordUp() {
}

func (n *noopMetricer) RecordRPCServerRequest(method string) func() {
	return func() {}
}

func (n *noopMetricer) RecordRPCClientRequest(method string) func(err error) {
	return func(err error) {}
}

func (n *noopMetricer) RecordRPCClientResponse(method string, err error) {
}

func (n *noopMetricer) SetDerivationIdle(status bool) {
}

func (n *noopMetricer) RecordPipelineReset() {
}

func (n *noopMetricer) RecordSequencingError() {
}

func (n *noopMetricer) RecordPublishingError() {
}

func (n *noopMetricer) RecordDerivationError() {
}

func (n *noopMetricer) RecordReceivedUnsafePayload(payload *eth.ExecutionPayload) {
}

func (n *noopMetricer) recordRef(layer string, name string, num uint64, timestamp uint64, h common.Hash) {
}

func (n *noopMetricer) RecordL1Ref(name string, ref eth.L1BlockRef) {
}

func (n *noopMetricer) RecordL2Ref(name string, ref eth.L2BlockRef) {
}

func (n *noopMetricer) RecordUnsafePayloadsBuffer(length uint64, memSize uint64, next eth.BlockID) {
}

func (n *noopMetricer) CountSequencedTxs(count int) {
}

func (n *noopMetricer) RecordL1ReorgDepth(d uint64) {
}

func (n *noopMetricer) RecordGossipEvent(evType int32) {
}

func (n *noopMetricer) IncPeerCount() {
}

func (n *noopMetricer) DecPeerCount() {
}

func (n *noopMetricer) IncStreamCount() {
}

func (n *noopMetricer) DecStreamCount() {
}

func (n *noopMetricer) RecordBandwidth(ctx context.Context, bwc *libp2pmetrics.BandwidthCounter) {
}

func (n *noopMetricer) RecordSequencerBuildingDiffTime(duration time.Duration) {
}

func (n *noopMetricer) RecordSequencerSealingTime(duration time.Duration) {
}
