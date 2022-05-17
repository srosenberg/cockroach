package denylist

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"io"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"gopkg.in/yaml.v2"
)

const (
	defaultPollingInterval   = time.Minute
	defaultEmptyDenylistText = "SequenceNumber: 0"
)

type File struct {
	Seq      int64        `yaml:"SequenceNumber"`
	Denylist []*DenyEntry `yaml:"denylist"`
}

type Option func(*fileOptions)

type fileOptions struct {
	pollingInterval time.Duration
	timeSource      timeutil.TimeSource
}

func newDenylistWithFile(
	ctx context.Context, filename string, opts ...Option,
) (*Denylist, chan *Denylist) {
	__antithesis_instrumentation__.Notify(21427)
	options := &fileOptions{
		pollingInterval: defaultPollingInterval,
		timeSource:      timeutil.DefaultTimeSource{},
	}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(21430)
		opt(options)
	}
	__antithesis_instrumentation__.Notify(21428)
	ret, err := readDenyList(ctx, options, filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(21431)

		log.Errorf(ctx, "error when reading from file %s: %v", filename, err)

		ret = &Denylist{
			entries: make(map[DenyEntity]*DenyEntry),
		}
	} else {
		__antithesis_instrumentation__.Notify(21432)
	}
	__antithesis_instrumentation__.Notify(21429)

	return ret, watchForUpdate(ctx, options, filename)
}

func Deserialize(reader io.Reader) (*File, error) {
	__antithesis_instrumentation__.Notify(21433)
	decoder := yaml.NewDecoder(reader)
	var denylistFile File
	err := decoder.Decode(&denylistFile)
	if err != nil {
		__antithesis_instrumentation__.Notify(21435)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21436)
	}
	__antithesis_instrumentation__.Notify(21434)
	return &denylistFile, nil
}

func (dlf *File) Serialize() ([]byte, error) {
	__antithesis_instrumentation__.Notify(21437)
	return yaml.Marshal(dlf)
}

var strToTypeMap = map[string]Type{
	"ip":      IPAddrType,
	"cluster": ClusterType,
}

var typeToStrMap = map[Type]string{
	IPAddrType:  "ip",
	ClusterType: "cluster",
}

func (typ *Type) UnmarshalYAML(unmarshal func(interface{}) error) error {
	__antithesis_instrumentation__.Notify(21438)
	var raw string
	err := unmarshal(&raw)
	if err != nil {
		__antithesis_instrumentation__.Notify(21441)
		return err
	} else {
		__antithesis_instrumentation__.Notify(21442)
	}
	__antithesis_instrumentation__.Notify(21439)

	normalized := strings.ToLower(raw)
	t, ok := strToTypeMap[normalized]
	if !ok {
		__antithesis_instrumentation__.Notify(21443)
		*typ = UnknownType
	} else {
		__antithesis_instrumentation__.Notify(21444)
		*typ = t
	}
	__antithesis_instrumentation__.Notify(21440)

	return nil
}

func (typ Type) MarshalYAML() (interface{}, error) {
	__antithesis_instrumentation__.Notify(21445)
	return typ.String(), nil
}

func (typ Type) String() string {
	__antithesis_instrumentation__.Notify(21446)
	s, ok := typeToStrMap[typ]
	if !ok {
		__antithesis_instrumentation__.Notify(21448)
		return "UNKNOWN"
	} else {
		__antithesis_instrumentation__.Notify(21449)
	}
	__antithesis_instrumentation__.Notify(21447)
	return s
}

func WithPollingInterval(d time.Duration) Option {
	__antithesis_instrumentation__.Notify(21450)
	return func(op *fileOptions) {
		__antithesis_instrumentation__.Notify(21451)
		op.pollingInterval = d
	}
}

func WithTimeSource(t timeutil.TimeSource) Option {
	__antithesis_instrumentation__.Notify(21452)
	return func(op *fileOptions) {
		__antithesis_instrumentation__.Notify(21453)
		op.timeSource = t
	}
}

func readDenyList(ctx context.Context, options *fileOptions, filename string) (*Denylist, error) {
	__antithesis_instrumentation__.Notify(21454)
	handle, err := os.Open(filename)
	if err != nil {
		__antithesis_instrumentation__.Notify(21458)
		log.Errorf(ctx, "open file %s: %v", filename, err)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21459)
	}
	__antithesis_instrumentation__.Notify(21455)
	defer handle.Close()

	dlf, err := Deserialize(handle)
	if err != nil {
		__antithesis_instrumentation__.Notify(21460)
		stat, _ := handle.Stat()
		if stat != nil {
			__antithesis_instrumentation__.Notify(21462)
			log.Errorf(ctx, "error updating denylist from file %s modified at %s: %v",
				filename, stat.ModTime(), err)
		} else {
			__antithesis_instrumentation__.Notify(21463)
			log.Errorf(ctx, "error updating denylist from file %s: %v",
				filename, err)
		}
		__antithesis_instrumentation__.Notify(21461)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21464)
	}
	__antithesis_instrumentation__.Notify(21456)
	dl := &Denylist{
		entries:    make(map[DenyEntity]*DenyEntry),
		timeSource: options.timeSource,
	}
	for _, entry := range dlf.Denylist {
		__antithesis_instrumentation__.Notify(21465)
		dl.entries[entry.Entity] = entry
	}
	__antithesis_instrumentation__.Notify(21457)
	return dl, nil
}

func watchForUpdate(ctx context.Context, options *fileOptions, filename string) chan *Denylist {
	__antithesis_instrumentation__.Notify(21466)
	result := make(chan *Denylist)
	go func() {
		__antithesis_instrumentation__.Notify(21468)

		t := timeutil.NewTimer()
		defer t.Stop()
		for {
			__antithesis_instrumentation__.Notify(21469)
			t.Reset(options.pollingInterval)
			select {
			case <-ctx.Done():
				__antithesis_instrumentation__.Notify(21470)
				close(result)
				log.Errorf(ctx, "WatchList daemon stopped: %v", ctx.Err())
				return
			case <-t.C:
				__antithesis_instrumentation__.Notify(21471)
				t.Read = true
				list, err := readDenyList(ctx, options, filename)
				if err != nil {
					__antithesis_instrumentation__.Notify(21473)

					continue
				} else {
					__antithesis_instrumentation__.Notify(21474)
				}
				__antithesis_instrumentation__.Notify(21472)
				result <- list
			}
		}
	}()
	__antithesis_instrumentation__.Notify(21467)
	return result
}
