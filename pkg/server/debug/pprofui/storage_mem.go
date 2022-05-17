package pprofui

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"sync/atomic"
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type record struct {
	id string
	t  time.Time
	b  []byte
}

type MemStorage struct {
	mu struct {
		syncutil.Mutex
		records []record
	}
	idGen        int32
	keepDuration time.Duration
	keepNumber   int
}

func (s *MemStorage) getRecords() []record {
	__antithesis_instrumentation__.Notify(190417)
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]record(nil), s.mu.records...)
}

var _ Storage = &MemStorage{}

func NewMemStorage(n int, d time.Duration) *MemStorage {
	__antithesis_instrumentation__.Notify(190418)
	return &MemStorage{
		keepNumber:   n,
		keepDuration: d,
	}
}

func (s *MemStorage) ID() string {
	__antithesis_instrumentation__.Notify(190419)
	return fmt.Sprint(atomic.AddInt32(&s.idGen, 1))
}

func (s *MemStorage) cleanLocked() {
	__antithesis_instrumentation__.Notify(190420)
	if l, m := len(s.mu.records), s.keepNumber; l > m && func() bool {
		__antithesis_instrumentation__.Notify(190422)
		return m != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(190423)
		s.mu.records = append([]record(nil), s.mu.records[l-m:]...)
	} else {
		__antithesis_instrumentation__.Notify(190424)
	}
	__antithesis_instrumentation__.Notify(190421)
	now := timeutil.Now()
	if pos := sort.Search(len(s.mu.records), func(i int) bool {
		__antithesis_instrumentation__.Notify(190425)
		return s.mu.records[i].t.Add(s.keepDuration).After(now)
	}); pos < len(s.mu.records) && func() bool {
		__antithesis_instrumentation__.Notify(190426)
		return s.keepDuration != 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(190427)
		s.mu.records = append([]record(nil), s.mu.records[pos:]...)
	} else {
		__antithesis_instrumentation__.Notify(190428)
	}
}

func (s *MemStorage) Store(id string, write func(io.Writer) error) error {
	__antithesis_instrumentation__.Notify(190429)
	var b bytes.Buffer
	if err := write(&b); err != nil {
		__antithesis_instrumentation__.Notify(190432)
		return err
	} else {
		__antithesis_instrumentation__.Notify(190433)
	}
	__antithesis_instrumentation__.Notify(190430)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mu.records = append(s.mu.records, record{id: id, t: timeutil.Now(), b: b.Bytes()})
	sort.Slice(s.mu.records, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(190434)
		return s.mu.records[i].t.Before(s.mu.records[j].t)
	})
	__antithesis_instrumentation__.Notify(190431)
	s.cleanLocked()
	return nil
}

func (s *MemStorage) Get(id string, read func(io.Reader) error) error {
	__antithesis_instrumentation__.Notify(190435)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, v := range s.mu.records {
		__antithesis_instrumentation__.Notify(190437)
		if v.id == id {
			__antithesis_instrumentation__.Notify(190438)
			return read(bytes.NewReader(v.b))
		} else {
			__antithesis_instrumentation__.Notify(190439)
		}
	}
	__antithesis_instrumentation__.Notify(190436)
	return errors.Errorf("profile not found; it may have expired, please regenerate the profile.\n" +
		"To generate profile for a node, use the profile generation link from the Advanced Debug page.\n" +
		"Attempting to generate a profile by modifying the node query parameter in the URL will not work.",
	)
}
