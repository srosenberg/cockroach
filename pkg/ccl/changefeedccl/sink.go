package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/url"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/bufalloc"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type Sink interface {
	Dial() error

	EmitRow(
		ctx context.Context,
		topic TopicDescriptor,
		key, value []byte,
		updated, mvcc hlc.Timestamp,
		alloc kvevent.Alloc,
	) error

	EmitResolvedTimestamp(ctx context.Context, encoder Encoder, resolved hlc.Timestamp) error

	Flush(ctx context.Context) error

	Close() error
}

type SinkWithTopics interface {
	Sink
	Topics() []string
}

func getSink(
	ctx context.Context,
	serverCfg *execinfra.ServerConfig,
	feedCfg jobspb.ChangefeedDetails,
	timestampOracle timestampLowerBoundOracle,
	user security.SQLUsername,
	jobID jobspb.JobID,
	m *sliMetrics,
) (Sink, error) {
	__antithesis_instrumentation__.Notify(18050)
	u, err := url.Parse(feedCfg.SinkURI)
	if err != nil {
		__antithesis_instrumentation__.Notify(18058)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18059)
	}
	__antithesis_instrumentation__.Notify(18051)
	if scheme, ok := changefeedbase.NoLongerExperimental[u.Scheme]; ok {
		__antithesis_instrumentation__.Notify(18060)
		u.Scheme = scheme
	} else {
		__antithesis_instrumentation__.Notify(18061)
	}
	__antithesis_instrumentation__.Notify(18052)

	validateOptionsAndMakeSink := func(sinkSpecificOpts map[string]struct{}, makeSink func() (Sink, error)) (Sink, error) {
		__antithesis_instrumentation__.Notify(18062)
		err := validateSinkOptions(feedCfg.Opts, sinkSpecificOpts)
		if err != nil {
			__antithesis_instrumentation__.Notify(18064)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(18065)
		}
		__antithesis_instrumentation__.Notify(18063)
		return makeSink()
	}
	__antithesis_instrumentation__.Notify(18053)

	newSink := func() (Sink, error) {
		__antithesis_instrumentation__.Notify(18066)
		if feedCfg.SinkURI == "" {
			__antithesis_instrumentation__.Notify(18068)
			return &bufferSink{metrics: m}, nil
		} else {
			__antithesis_instrumentation__.Notify(18069)
		}
		__antithesis_instrumentation__.Notify(18067)

		switch {
		case u.Scheme == changefeedbase.SinkSchemeNull:
			__antithesis_instrumentation__.Notify(18070)
			return makeNullSink(sinkURL{URL: u}, m)
		case u.Scheme == changefeedbase.SinkSchemeKafka:
			__antithesis_instrumentation__.Notify(18071)
			return validateOptionsAndMakeSink(changefeedbase.KafkaValidOptions, func() (Sink, error) {
				__antithesis_instrumentation__.Notify(18078)
				return makeKafkaSink(ctx, sinkURL{URL: u}, AllTargets(feedCfg), feedCfg.Opts, m)
			})
		case isWebhookSink(u):
			__antithesis_instrumentation__.Notify(18072)
			return validateOptionsAndMakeSink(changefeedbase.WebhookValidOptions, func() (Sink, error) {
				__antithesis_instrumentation__.Notify(18079)
				return makeWebhookSink(ctx, sinkURL{URL: u}, feedCfg.Opts,
					defaultWorkerCount(), timeutil.DefaultTimeSource{}, m)
			})
		case isPubsubSink(u):
			__antithesis_instrumentation__.Notify(18073)

			return validateOptionsAndMakeSink(changefeedbase.PubsubValidOptions, func() (Sink, error) {
				__antithesis_instrumentation__.Notify(18080)
				return MakePubsubSink(ctx, u, feedCfg.Opts, AllTargets(feedCfg))
			})
		case isCloudStorageSink(u):
			__antithesis_instrumentation__.Notify(18074)
			return validateOptionsAndMakeSink(changefeedbase.CloudStorageValidOptions, func() (Sink, error) {
				__antithesis_instrumentation__.Notify(18081)
				return makeCloudStorageSink(
					ctx, sinkURL{URL: u}, serverCfg.NodeID.SQLInstanceID(), serverCfg.Settings,
					feedCfg.Opts, timestampOracle, serverCfg.ExternalStorageFromURI, user, m,
				)
			})
		case u.Scheme == changefeedbase.SinkSchemeExperimentalSQL:
			__antithesis_instrumentation__.Notify(18075)
			return validateOptionsAndMakeSink(changefeedbase.SQLValidOptions, func() (Sink, error) {
				__antithesis_instrumentation__.Notify(18082)
				return makeSQLSink(sinkURL{URL: u}, sqlSinkTableName, AllTargets(feedCfg), m)
			})
		case u.Scheme == "":
			__antithesis_instrumentation__.Notify(18076)
			return nil, errors.Errorf(`no scheme found for sink URL %q`, feedCfg.SinkURI)
		default:
			__antithesis_instrumentation__.Notify(18077)
			return nil, errors.Errorf(`unsupported sink: %s`, u.Scheme)
		}
	}
	__antithesis_instrumentation__.Notify(18054)

	sink, err := newSink()
	if err != nil {
		__antithesis_instrumentation__.Notify(18083)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18084)
	}
	__antithesis_instrumentation__.Notify(18055)

	if knobs, ok := serverCfg.TestingKnobs.Changefeed.(*TestingKnobs); ok && func() bool {
		__antithesis_instrumentation__.Notify(18085)
		return knobs.WrapSink != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(18086)
		sink = knobs.WrapSink(sink, jobID)
	} else {
		__antithesis_instrumentation__.Notify(18087)
	}
	__antithesis_instrumentation__.Notify(18056)

	if err := sink.Dial(); err != nil {
		__antithesis_instrumentation__.Notify(18088)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18089)
	}
	__antithesis_instrumentation__.Notify(18057)

	return sink, nil
}

func validateSinkOptions(opts map[string]string, sinkSpecificOpts map[string]struct{}) error {
	__antithesis_instrumentation__.Notify(18090)
	for opt := range opts {
		__antithesis_instrumentation__.Notify(18092)
		if _, ok := changefeedbase.CommonOptions[opt]; ok {
			__antithesis_instrumentation__.Notify(18095)
			continue
		} else {
			__antithesis_instrumentation__.Notify(18096)
		}
		__antithesis_instrumentation__.Notify(18093)
		if sinkSpecificOpts != nil {
			__antithesis_instrumentation__.Notify(18097)
			if _, ok := sinkSpecificOpts[opt]; ok {
				__antithesis_instrumentation__.Notify(18098)
				continue
			} else {
				__antithesis_instrumentation__.Notify(18099)
			}
		} else {
			__antithesis_instrumentation__.Notify(18100)
		}
		__antithesis_instrumentation__.Notify(18094)
		return errors.Errorf("this sink is incompatible with option %s", opt)
	}
	__antithesis_instrumentation__.Notify(18091)
	return nil
}

type sinkURL struct {
	*url.URL
	q url.Values
}

func (u *sinkURL) consumeParam(p string) string {
	__antithesis_instrumentation__.Notify(18101)
	if u.q == nil {
		__antithesis_instrumentation__.Notify(18103)
		u.q = u.Query()
	} else {
		__antithesis_instrumentation__.Notify(18104)
	}
	__antithesis_instrumentation__.Notify(18102)
	v := u.q.Get(p)
	u.q.Del(p)
	return v
}

func (u *sinkURL) addParam(p string, value string) {
	__antithesis_instrumentation__.Notify(18105)
	if u.q == nil {
		__antithesis_instrumentation__.Notify(18107)
		u.q = u.Query()
	} else {
		__antithesis_instrumentation__.Notify(18108)
	}
	__antithesis_instrumentation__.Notify(18106)
	u.q.Add(p, value)
}

func (u *sinkURL) consumeBool(param string, dest *bool) (wasSet bool, err error) {
	__antithesis_instrumentation__.Notify(18109)
	if paramVal := u.consumeParam(param); paramVal != "" {
		__antithesis_instrumentation__.Notify(18111)
		wasSet, err := strToBool(paramVal, dest)
		if err != nil {
			__antithesis_instrumentation__.Notify(18113)
			return false, errors.Wrapf(err, "param %s must be a bool", param)
		} else {
			__antithesis_instrumentation__.Notify(18114)
		}
		__antithesis_instrumentation__.Notify(18112)
		return wasSet, err
	} else {
		__antithesis_instrumentation__.Notify(18115)
	}
	__antithesis_instrumentation__.Notify(18110)
	return false, nil
}

func (u *sinkURL) decodeBase64(param string, dest *[]byte) error {
	__antithesis_instrumentation__.Notify(18116)

	val := u.consumeParam(param)
	err := decodeBase64FromString(val, dest)
	if err != nil {
		__antithesis_instrumentation__.Notify(18118)
		return errors.Wrapf(err, `param %s must be base 64 encoded`, param)
	} else {
		__antithesis_instrumentation__.Notify(18119)
	}
	__antithesis_instrumentation__.Notify(18117)
	return nil
}

func (u *sinkURL) remainingQueryParams() (res []string) {
	__antithesis_instrumentation__.Notify(18120)
	for p := range u.q {
		__antithesis_instrumentation__.Notify(18122)
		res = append(res, p)
	}
	__antithesis_instrumentation__.Notify(18121)
	return
}

func (u *sinkURL) String() string {
	__antithesis_instrumentation__.Notify(18123)
	if u.q != nil {
		__antithesis_instrumentation__.Notify(18125)

		u.URL.RawQuery = u.q.Encode()
		u.q = nil
	} else {
		__antithesis_instrumentation__.Notify(18126)
	}
	__antithesis_instrumentation__.Notify(18124)
	return u.URL.String()
}

type errorWrapperSink struct {
	wrapped Sink
}

func (s errorWrapperSink) EmitRow(
	ctx context.Context,
	topic TopicDescriptor,
	key, value []byte,
	updated, mvcc hlc.Timestamp,
	alloc kvevent.Alloc,
) error {
	__antithesis_instrumentation__.Notify(18127)
	if err := s.wrapped.EmitRow(ctx, topic, key, value, updated, mvcc, alloc); err != nil {
		__antithesis_instrumentation__.Notify(18129)
		return changefeedbase.MarkRetryableError(err)
	} else {
		__antithesis_instrumentation__.Notify(18130)
	}
	__antithesis_instrumentation__.Notify(18128)
	return nil
}

func (s errorWrapperSink) EmitResolvedTimestamp(
	ctx context.Context, encoder Encoder, resolved hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(18131)
	if err := s.wrapped.EmitResolvedTimestamp(ctx, encoder, resolved); err != nil {
		__antithesis_instrumentation__.Notify(18133)
		return changefeedbase.MarkRetryableError(err)
	} else {
		__antithesis_instrumentation__.Notify(18134)
	}
	__antithesis_instrumentation__.Notify(18132)
	return nil
}

func (s errorWrapperSink) Flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(18135)
	if err := s.wrapped.Flush(ctx); err != nil {
		__antithesis_instrumentation__.Notify(18137)
		return changefeedbase.MarkRetryableError(err)
	} else {
		__antithesis_instrumentation__.Notify(18138)
	}
	__antithesis_instrumentation__.Notify(18136)
	return nil
}

func (s errorWrapperSink) Close() error {
	__antithesis_instrumentation__.Notify(18139)
	if err := s.wrapped.Close(); err != nil {
		__antithesis_instrumentation__.Notify(18141)
		return changefeedbase.MarkRetryableError(err)
	} else {
		__antithesis_instrumentation__.Notify(18142)
	}
	__antithesis_instrumentation__.Notify(18140)
	return nil
}

func (s errorWrapperSink) Dial() error {
	__antithesis_instrumentation__.Notify(18143)
	return s.wrapped.Dial()
}

type encDatumRowBuffer []rowenc.EncDatumRow

func (b *encDatumRowBuffer) IsEmpty() bool {
	__antithesis_instrumentation__.Notify(18144)
	return b == nil || func() bool {
		__antithesis_instrumentation__.Notify(18145)
		return len(*b) == 0 == true
	}() == true
}
func (b *encDatumRowBuffer) Push(r rowenc.EncDatumRow) {
	__antithesis_instrumentation__.Notify(18146)
	*b = append(*b, r)
}
func (b *encDatumRowBuffer) Pop() rowenc.EncDatumRow {
	__antithesis_instrumentation__.Notify(18147)
	ret := (*b)[0]
	*b = (*b)[1:]
	return ret
}

type bufferSink struct {
	buf     encDatumRowBuffer
	alloc   tree.DatumAlloc
	scratch bufalloc.ByteAllocator
	closed  bool
	metrics *sliMetrics
}

func (s *bufferSink) EmitRow(
	ctx context.Context,
	topic TopicDescriptor,
	key, value []byte,
	updated, mvcc hlc.Timestamp,
	r kvevent.Alloc,
) error {
	__antithesis_instrumentation__.Notify(18148)
	defer r.Release(ctx)
	defer s.metrics.recordOneMessage()(mvcc, len(key)+len(value), sinkDoesNotCompress)

	if s.closed {
		__antithesis_instrumentation__.Notify(18150)
		return errors.New(`cannot EmitRow on a closed sink`)
	} else {
		__antithesis_instrumentation__.Notify(18151)
	}
	__antithesis_instrumentation__.Notify(18149)

	s.buf.Push(rowenc.EncDatumRow{
		{Datum: tree.DNull},
		{Datum: s.getTopicDatum(topic)},
		{Datum: s.alloc.NewDBytes(tree.DBytes(key))},
		{Datum: s.alloc.NewDBytes(tree.DBytes(value))},
	})
	return nil
}

func (s *bufferSink) EmitResolvedTimestamp(
	ctx context.Context, encoder Encoder, resolved hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(18152)
	if s.closed {
		__antithesis_instrumentation__.Notify(18155)
		return errors.New(`cannot EmitResolvedTimestamp on a closed sink`)
	} else {
		__antithesis_instrumentation__.Notify(18156)
	}
	__antithesis_instrumentation__.Notify(18153)
	defer s.metrics.recordResolvedCallback()()

	var noTopic string
	payload, err := encoder.EncodeResolvedTimestamp(ctx, noTopic, resolved)
	if err != nil {
		__antithesis_instrumentation__.Notify(18157)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18158)
	}
	__antithesis_instrumentation__.Notify(18154)
	s.scratch, payload = s.scratch.Copy(payload, 0)
	s.buf.Push(rowenc.EncDatumRow{
		{Datum: tree.DNull},
		{Datum: tree.DNull},
		{Datum: tree.DNull},
		{Datum: s.alloc.NewDBytes(tree.DBytes(payload))},
	})
	return nil
}

func (s *bufferSink) Flush(_ context.Context) error {
	__antithesis_instrumentation__.Notify(18159)
	defer s.metrics.recordFlushRequestCallback()()
	return nil
}

func (s *bufferSink) Close() error {
	__antithesis_instrumentation__.Notify(18160)
	s.closed = true
	return nil
}

func (s *bufferSink) Dial() error {
	__antithesis_instrumentation__.Notify(18161)
	return nil
}

func (s *bufferSink) getTopicDatum(t TopicDescriptor) *tree.DString {
	__antithesis_instrumentation__.Notify(18162)
	return s.alloc.NewDString(tree.DString(strings.Join(t.GetNameComponents(), ".")))
}

type nullSink struct {
	ticker  *time.Ticker
	metrics *sliMetrics
}

var _ Sink = (*nullSink)(nil)

func makeNullSink(u sinkURL, m *sliMetrics) (Sink, error) {
	__antithesis_instrumentation__.Notify(18163)
	var pacer *time.Ticker
	if delay := u.consumeParam(`delay`); delay != "" {
		__antithesis_instrumentation__.Notify(18165)
		pace, err := time.ParseDuration(delay)
		if err != nil {
			__antithesis_instrumentation__.Notify(18167)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(18168)
		}
		__antithesis_instrumentation__.Notify(18166)
		pacer = time.NewTicker(pace)
	} else {
		__antithesis_instrumentation__.Notify(18169)
	}
	__antithesis_instrumentation__.Notify(18164)
	return &nullSink{ticker: pacer, metrics: m}, nil
}

func (n *nullSink) pace(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(18170)
	if n.ticker != nil {
		__antithesis_instrumentation__.Notify(18172)
		select {
		case <-n.ticker.C:
			__antithesis_instrumentation__.Notify(18173)
			return nil
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(18174)
			return ctx.Err()
		}
	} else {
		__antithesis_instrumentation__.Notify(18175)
	}
	__antithesis_instrumentation__.Notify(18171)
	return nil
}

func (n *nullSink) EmitRow(
	ctx context.Context,
	topic TopicDescriptor,
	key, value []byte,
	updated, mvcc hlc.Timestamp,
	r kvevent.Alloc,
) error {
	__antithesis_instrumentation__.Notify(18176)
	defer r.Release(ctx)
	defer n.metrics.recordOneMessage()(mvcc, len(key)+len(value), sinkDoesNotCompress)
	if err := n.pace(ctx); err != nil {
		__antithesis_instrumentation__.Notify(18179)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18180)
	}
	__antithesis_instrumentation__.Notify(18177)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(18181)
		log.Infof(ctx, "emitting row %s@%s", key, updated.String())
	} else {
		__antithesis_instrumentation__.Notify(18182)
	}
	__antithesis_instrumentation__.Notify(18178)
	return nil
}

func (n *nullSink) EmitResolvedTimestamp(
	ctx context.Context, encoder Encoder, resolved hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(18183)
	defer n.metrics.recordResolvedCallback()()
	if err := n.pace(ctx); err != nil {
		__antithesis_instrumentation__.Notify(18186)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18187)
	}
	__antithesis_instrumentation__.Notify(18184)
	if log.V(2) {
		__antithesis_instrumentation__.Notify(18188)
		log.Infof(ctx, "emitting resolved %s", resolved.String())
	} else {
		__antithesis_instrumentation__.Notify(18189)
	}
	__antithesis_instrumentation__.Notify(18185)

	return nil
}

func (n *nullSink) Flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(18190)
	defer n.metrics.recordFlushRequestCallback()()
	if log.V(2) {
		__antithesis_instrumentation__.Notify(18192)
		log.Info(ctx, "flushing")
	} else {
		__antithesis_instrumentation__.Notify(18193)
	}
	__antithesis_instrumentation__.Notify(18191)

	return nil
}

func (n *nullSink) Close() error {
	__antithesis_instrumentation__.Notify(18194)
	return nil
}

func (n *nullSink) Dial() error {
	__antithesis_instrumentation__.Notify(18195)
	return nil
}
