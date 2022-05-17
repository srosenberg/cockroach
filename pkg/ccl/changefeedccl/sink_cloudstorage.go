package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/kvevent"
	"github.com/cockroachdb/cockroach/pkg/cloud"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
	"github.com/google/btree"
)

func isCloudStorageSink(u *url.URL) bool {
	__antithesis_instrumentation__.Notify(18196)
	switch u.Scheme {
	case changefeedbase.SinkSchemeCloudStorageS3, changefeedbase.SinkSchemeCloudStorageGCS,
		changefeedbase.SinkSchemeCloudStorageNodelocal, changefeedbase.SinkSchemeCloudStorageHTTP,
		changefeedbase.SinkSchemeCloudStorageHTTPS, changefeedbase.SinkSchemeCloudStorageAzure:
		__antithesis_instrumentation__.Notify(18197)
		return true
	default:
		__antithesis_instrumentation__.Notify(18198)
		return false
	}
}

func cloudStorageFormatTime(ts hlc.Timestamp) string {
	__antithesis_instrumentation__.Notify(18199)

	const f = `20060102150405`
	t := ts.GoTime()
	return fmt.Sprintf(`%s%09d%010d`, t.Format(f), t.Nanosecond(), ts.Logical)
}

type cloudStorageSinkFile struct {
	cloudStorageSinkKey
	created     time.Time
	codec       io.WriteCloser
	rawSize     int
	numMessages int
	buf         bytes.Buffer
	alloc       kvevent.Alloc
	oldestMVCC  hlc.Timestamp
}

var _ io.Writer = &cloudStorageSinkFile{}

func (f *cloudStorageSinkFile) Write(p []byte) (int, error) {
	__antithesis_instrumentation__.Notify(18200)
	f.rawSize += len(p)
	f.numMessages++
	if f.codec != nil {
		__antithesis_instrumentation__.Notify(18202)
		return f.codec.Write(p)
	} else {
		__antithesis_instrumentation__.Notify(18203)
	}
	__antithesis_instrumentation__.Notify(18201)
	return f.buf.Write(p)
}

type cloudStorageSink struct {
	srcID             base.SQLInstanceID
	sinkID            int64
	targetMaxFileSize int64
	settings          *cluster.Settings
	partitionFormat   string
	topicNamer        *TopicNamer

	ext          string
	rowDelimiter []byte

	compression string

	es cloud.ExternalStorage

	fileID int64
	files  *btree.BTree

	timestampOracle timestampLowerBoundOracle
	jobSessionID    string

	dataFileTs        string
	dataFilePartition string
	prevFilename      string
	metrics           *sliMetrics
}

const sinkCompressionGzip = "gzip"

var cloudStorageSinkIDAtomic int64

var partitionDateFormats = map[string]string{
	"flat":   "/",
	"daily":  "2006-01-02/",
	"hourly": "2006-01-02/15/",
}
var defaultPartitionFormat = partitionDateFormats["daily"]

func makeCloudStorageSink(
	ctx context.Context,
	u sinkURL,
	srcID base.SQLInstanceID,
	settings *cluster.Settings,
	opts map[string]string,
	timestampOracle timestampLowerBoundOracle,
	makeExternalStorageFromURI cloud.ExternalStorageFromURIFactory,
	user security.SQLUsername,
	m *sliMetrics,
) (Sink, error) {
	__antithesis_instrumentation__.Notify(18204)
	var targetMaxFileSize int64 = 16 << 20
	if fileSizeParam := u.consumeParam(changefeedbase.SinkParamFileSize); fileSizeParam != `` {
		__antithesis_instrumentation__.Notify(18215)
		var err error
		if targetMaxFileSize, err = humanizeutil.ParseBytes(fileSizeParam); err != nil {
			__antithesis_instrumentation__.Notify(18216)
			return nil, pgerror.Wrapf(err, pgcode.Syntax, `parsing %s`, fileSizeParam)
		} else {
			__antithesis_instrumentation__.Notify(18217)
		}
	} else {
		__antithesis_instrumentation__.Notify(18218)
	}
	__antithesis_instrumentation__.Notify(18205)
	u.Scheme = strings.TrimPrefix(u.Scheme, `experimental-`)

	sinkID := atomic.AddInt64(&cloudStorageSinkIDAtomic, 1)
	sessID, err := generateChangefeedSessionID()
	if err != nil {
		__antithesis_instrumentation__.Notify(18219)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18220)
	}
	__antithesis_instrumentation__.Notify(18206)

	tn, err := MakeTopicNamer([]jobspb.ChangefeedTargetSpecification{}, WithJoinByte('+'))
	if err != nil {
		__antithesis_instrumentation__.Notify(18221)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18222)
	}
	__antithesis_instrumentation__.Notify(18207)

	s := &cloudStorageSink{
		srcID:             srcID,
		sinkID:            sinkID,
		settings:          settings,
		targetMaxFileSize: targetMaxFileSize,
		files:             btree.New(8),
		partitionFormat:   defaultPartitionFormat,
		timestampOracle:   timestampOracle,

		jobSessionID: sessID,
		metrics:      m,
		topicNamer:   tn,
	}

	if partitionFormat := u.consumeParam(changefeedbase.SinkParamPartitionFormat); partitionFormat != "" {
		__antithesis_instrumentation__.Notify(18223)
		dateFormat, ok := partitionDateFormats[partitionFormat]
		if !ok {
			__antithesis_instrumentation__.Notify(18225)
			return nil, errors.Errorf("invalid partition_format of %s", partitionFormat)
		} else {
			__antithesis_instrumentation__.Notify(18226)
		}
		__antithesis_instrumentation__.Notify(18224)

		s.partitionFormat = dateFormat
	} else {
		__antithesis_instrumentation__.Notify(18227)
	}
	__antithesis_instrumentation__.Notify(18208)

	if s.timestampOracle != nil {
		__antithesis_instrumentation__.Notify(18228)
		s.dataFileTs = cloudStorageFormatTime(s.timestampOracle.inclusiveLowerBoundTS())
		s.dataFilePartition = s.timestampOracle.inclusiveLowerBoundTS().GoTime().Format(s.partitionFormat)
	} else {
		__antithesis_instrumentation__.Notify(18229)
	}
	__antithesis_instrumentation__.Notify(18209)

	switch changefeedbase.FormatType(opts[changefeedbase.OptFormat]) {
	case changefeedbase.OptFormatJSON:
		__antithesis_instrumentation__.Notify(18230)

		s.ext = `.ndjson`
		s.rowDelimiter = []byte{'\n'}
	default:
		__antithesis_instrumentation__.Notify(18231)
		return nil, errors.Errorf(`this sink is incompatible with %s=%s`,
			changefeedbase.OptFormat, opts[changefeedbase.OptFormat])
	}
	__antithesis_instrumentation__.Notify(18210)

	switch changefeedbase.EnvelopeType(opts[changefeedbase.OptEnvelope]) {
	case changefeedbase.OptEnvelopeWrapped:
		__antithesis_instrumentation__.Notify(18232)
	default:
		__antithesis_instrumentation__.Notify(18233)
		return nil, errors.Errorf(`this sink is incompatible with %s=%s`,
			changefeedbase.OptEnvelope, opts[changefeedbase.OptEnvelope])
	}
	__antithesis_instrumentation__.Notify(18211)

	if _, ok := opts[changefeedbase.OptKeyInValue]; !ok {
		__antithesis_instrumentation__.Notify(18234)
		return nil, errors.Errorf(`this sink requires the WITH %s option`, changefeedbase.OptKeyInValue)
	} else {
		__antithesis_instrumentation__.Notify(18235)
	}
	__antithesis_instrumentation__.Notify(18212)

	if codec, ok := opts[changefeedbase.OptCompression]; ok && func() bool {
		__antithesis_instrumentation__.Notify(18236)
		return codec != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(18237)
		if strings.EqualFold(codec, "gzip") {
			__antithesis_instrumentation__.Notify(18238)
			s.compression = sinkCompressionGzip
			s.ext = s.ext + ".gz"
		} else {
			__antithesis_instrumentation__.Notify(18239)
			return nil, errors.Errorf(`unsupported compression codec %q`, codec)
		}
	} else {
		__antithesis_instrumentation__.Notify(18240)
	}
	__antithesis_instrumentation__.Notify(18213)

	if s.es, err = makeExternalStorageFromURI(ctx, u.String(), user); err != nil {
		__antithesis_instrumentation__.Notify(18241)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(18242)
	}
	__antithesis_instrumentation__.Notify(18214)

	return s, nil
}

func (s *cloudStorageSink) getOrCreateFile(
	topic TopicDescriptor, eventMVCC hlc.Timestamp,
) *cloudStorageSinkFile {
	__antithesis_instrumentation__.Notify(18243)
	name, _ := s.topicNamer.Name(topic)
	key := cloudStorageSinkKey{name, int64(topic.GetVersion())}
	if item := s.files.Get(key); item != nil {
		__antithesis_instrumentation__.Notify(18246)
		f := item.(*cloudStorageSinkFile)
		if eventMVCC.Less(f.oldestMVCC) {
			__antithesis_instrumentation__.Notify(18248)
			f.oldestMVCC = eventMVCC
		} else {
			__antithesis_instrumentation__.Notify(18249)
		}
		__antithesis_instrumentation__.Notify(18247)
		return f
	} else {
		__antithesis_instrumentation__.Notify(18250)
	}
	__antithesis_instrumentation__.Notify(18244)
	f := &cloudStorageSinkFile{
		created:             timeutil.Now(),
		cloudStorageSinkKey: key,
		oldestMVCC:          eventMVCC,
	}
	switch s.compression {
	case sinkCompressionGzip:
		__antithesis_instrumentation__.Notify(18251)
		f.codec = gzip.NewWriter(&f.buf)
	default:
		__antithesis_instrumentation__.Notify(18252)
	}
	__antithesis_instrumentation__.Notify(18245)
	s.files.ReplaceOrInsert(f)
	return f
}

func (s *cloudStorageSink) EmitRow(
	ctx context.Context,
	topic TopicDescriptor,
	key, value []byte,
	updated, mvcc hlc.Timestamp,
	alloc kvevent.Alloc,
) error {
	__antithesis_instrumentation__.Notify(18253)
	if s.files == nil {
		__antithesis_instrumentation__.Notify(18258)
		return errors.New(`cannot EmitRow on a closed sink`)
	} else {
		__antithesis_instrumentation__.Notify(18259)
	}
	__antithesis_instrumentation__.Notify(18254)

	s.metrics.recordMessageSize(int64(len(key) + len(value)))
	file := s.getOrCreateFile(topic, mvcc)
	file.alloc.Merge(&alloc)

	if _, err := file.Write(value); err != nil {
		__antithesis_instrumentation__.Notify(18260)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18261)
	}
	__antithesis_instrumentation__.Notify(18255)
	if _, err := file.Write(s.rowDelimiter); err != nil {
		__antithesis_instrumentation__.Notify(18262)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18263)
	}
	__antithesis_instrumentation__.Notify(18256)

	if int64(file.buf.Len()) > s.targetMaxFileSize {
		__antithesis_instrumentation__.Notify(18264)
		if err := s.flushTopicVersions(ctx, file.topic, file.schemaID); err != nil {
			__antithesis_instrumentation__.Notify(18265)
			return err
		} else {
			__antithesis_instrumentation__.Notify(18266)
		}
	} else {
		__antithesis_instrumentation__.Notify(18267)
	}
	__antithesis_instrumentation__.Notify(18257)
	return nil
}

func (s *cloudStorageSink) EmitResolvedTimestamp(
	ctx context.Context, encoder Encoder, resolved hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(18268)
	if s.files == nil {
		__antithesis_instrumentation__.Notify(18272)
		return errors.New(`cannot EmitRow on a closed sink`)
	} else {
		__antithesis_instrumentation__.Notify(18273)
	}
	__antithesis_instrumentation__.Notify(18269)

	defer s.metrics.recordResolvedCallback()()

	var noTopic string
	payload, err := encoder.EncodeResolvedTimestamp(ctx, noTopic, resolved)
	if err != nil {
		__antithesis_instrumentation__.Notify(18274)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18275)
	}
	__antithesis_instrumentation__.Notify(18270)

	part := resolved.GoTime().Format(s.partitionFormat)
	filename := fmt.Sprintf(`%s.RESOLVED`, cloudStorageFormatTime(resolved))
	if log.V(1) {
		__antithesis_instrumentation__.Notify(18276)
		log.Infof(ctx, "writing file %s %s", filename, resolved.AsOfSystemTime())
	} else {
		__antithesis_instrumentation__.Notify(18277)
	}
	__antithesis_instrumentation__.Notify(18271)
	return cloud.WriteFile(ctx, s.es, filepath.Join(part, filename), bytes.NewReader(payload))
}

func (s *cloudStorageSink) flushTopicVersions(
	ctx context.Context, topic string, maxVersionToFlush int64,
) (err error) {
	__antithesis_instrumentation__.Notify(18278)
	var toRemoveAlloc [2]int64
	toRemove := toRemoveAlloc[:0]
	gte := cloudStorageSinkKey{topic: topic}
	lt := cloudStorageSinkKey{topic: topic, schemaID: maxVersionToFlush + 1}
	s.files.AscendRange(gte, lt, func(i btree.Item) (wantMore bool) {
		__antithesis_instrumentation__.Notify(18281)
		f := i.(*cloudStorageSinkFile)
		if err = s.flushFile(ctx, f); err == nil {
			__antithesis_instrumentation__.Notify(18283)
			toRemove = append(toRemove, f.schemaID)
		} else {
			__antithesis_instrumentation__.Notify(18284)
		}
		__antithesis_instrumentation__.Notify(18282)
		return err == nil
	})
	__antithesis_instrumentation__.Notify(18279)
	for _, v := range toRemove {
		__antithesis_instrumentation__.Notify(18285)
		s.files.Delete(cloudStorageSinkKey{topic: topic, schemaID: v})
	}
	__antithesis_instrumentation__.Notify(18280)
	return err
}

func (s *cloudStorageSink) Flush(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(18286)
	if s.files == nil {
		__antithesis_instrumentation__.Notify(18290)
		return errors.New(`cannot Flush on a closed sink`)
	} else {
		__antithesis_instrumentation__.Notify(18291)
	}
	__antithesis_instrumentation__.Notify(18287)

	s.metrics.recordFlushRequestCallback()()

	var err error
	s.files.Ascend(func(i btree.Item) (wantMore bool) {
		__antithesis_instrumentation__.Notify(18292)
		err = s.flushFile(ctx, i.(*cloudStorageSinkFile))
		return err == nil
	})
	__antithesis_instrumentation__.Notify(18288)
	if err != nil {
		__antithesis_instrumentation__.Notify(18293)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18294)
	}
	__antithesis_instrumentation__.Notify(18289)
	s.files.Clear(true)

	s.dataFileTs = cloudStorageFormatTime(s.timestampOracle.inclusiveLowerBoundTS())
	s.dataFilePartition = s.timestampOracle.inclusiveLowerBoundTS().GoTime().Format(s.partitionFormat)
	return nil
}

func (s *cloudStorageSink) flushFile(ctx context.Context, file *cloudStorageSinkFile) error {
	__antithesis_instrumentation__.Notify(18295)
	defer file.alloc.Release(ctx)

	if file.rawSize == 0 {
		__antithesis_instrumentation__.Notify(18300)

		return nil
	} else {
		__antithesis_instrumentation__.Notify(18301)
	}
	__antithesis_instrumentation__.Notify(18296)

	if file.codec != nil {
		__antithesis_instrumentation__.Notify(18302)
		if err := file.codec.Close(); err != nil {
			__antithesis_instrumentation__.Notify(18303)
			return err
		} else {
			__antithesis_instrumentation__.Notify(18304)
		}
	} else {
		__antithesis_instrumentation__.Notify(18305)
	}
	__antithesis_instrumentation__.Notify(18297)

	fileID := s.fileID
	s.fileID++

	filename := fmt.Sprintf(`%s-%s-%d-%d-%08x-%s-%x%s`, s.dataFileTs,
		s.jobSessionID, s.srcID, s.sinkID, fileID, file.topic, file.schemaID, s.ext)
	if s.prevFilename != "" && func() bool {
		__antithesis_instrumentation__.Notify(18306)
		return filename < s.prevFilename == true
	}() == true {
		__antithesis_instrumentation__.Notify(18307)
		return errors.AssertionFailedf("error: detected a filename %s that lexically "+
			"precedes a file emitted before: %s", filename, s.prevFilename)
	} else {
		__antithesis_instrumentation__.Notify(18308)
	}
	__antithesis_instrumentation__.Notify(18298)
	s.prevFilename = filename
	compressedBytes := file.buf.Len()
	if err := cloud.WriteFile(ctx, s.es, filepath.Join(s.dataFilePartition, filename), bytes.NewReader(file.buf.Bytes())); err != nil {
		__antithesis_instrumentation__.Notify(18309)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18310)
	}
	__antithesis_instrumentation__.Notify(18299)
	s.metrics.recordEmittedBatch(file.created, file.numMessages, file.oldestMVCC, file.rawSize, compressedBytes)

	return nil
}

func (s *cloudStorageSink) Close() error {
	__antithesis_instrumentation__.Notify(18311)
	s.files = nil
	return s.es.Close()
}

func (s *cloudStorageSink) Dial() error {
	__antithesis_instrumentation__.Notify(18312)
	return nil
}

type cloudStorageSinkKey struct {
	topic    string
	schemaID int64
}

func (k cloudStorageSinkKey) Less(other btree.Item) bool {
	__antithesis_instrumentation__.Notify(18313)
	switch other := other.(type) {
	case *cloudStorageSinkFile:
		__antithesis_instrumentation__.Notify(18314)
		return keyLess(k, other.cloudStorageSinkKey)
	case cloudStorageSinkKey:
		__antithesis_instrumentation__.Notify(18315)
		return keyLess(k, other)
	default:
		__antithesis_instrumentation__.Notify(18316)
		panic(errors.Errorf("unexpected item type %T", other))
	}
}

func keyLess(a, b cloudStorageSinkKey) bool {
	__antithesis_instrumentation__.Notify(18317)
	if a.topic == b.topic {
		__antithesis_instrumentation__.Notify(18319)
		return a.schemaID < b.schemaID
	} else {
		__antithesis_instrumentation__.Notify(18320)
	}
	__antithesis_instrumentation__.Notify(18318)
	return a.topic < b.topic
}

func generateChangefeedSessionID() (string, error) {
	__antithesis_instrumentation__.Notify(18321)

	const size = 8
	p := make([]byte, size)
	buf := make([]byte, hex.EncodedLen(size))
	if _, err := rand.Read(p); err != nil {
		__antithesis_instrumentation__.Notify(18323)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(18324)
	}
	__antithesis_instrumentation__.Notify(18322)
	hex.Encode(buf, p)
	return string(buf), nil
}
