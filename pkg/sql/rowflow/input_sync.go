package rowflow

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"container/heap"
	"context"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfra"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

type srcInfo struct {
	src execinfra.RowSource

	row rowenc.EncDatumRow
}

type srcIdx int

type serialSynchronizerState int

const (
	notInitialized serialSynchronizerState = iota

	returningRows

	draining

	drainBuffered
)

type serialSynchronizer interface {
	getSources() []srcInfo
}

type serialSynchronizerBase struct {
	state serialSynchronizerState

	types    []*types.T
	sources  []srcInfo
	rowAlloc rowenc.EncDatumRowAlloc
}

var _ serialSynchronizer = &serialSynchronizerBase{}

func (s *serialSynchronizerBase) getSources() []srcInfo {
	__antithesis_instrumentation__.Notify(575816)
	return s.sources
}

func (s *serialSynchronizerBase) Start(ctx context.Context) {
	__antithesis_instrumentation__.Notify(575817)
	for _, src := range s.sources {
		__antithesis_instrumentation__.Notify(575818)
		src.src.Start(ctx)
	}
}

func (s *serialSynchronizerBase) OutputTypes() []*types.T {
	__antithesis_instrumentation__.Notify(575819)
	return s.types
}

type serialOrderedSynchronizer struct {
	serialSynchronizerBase

	ordering colinfo.ColumnOrdering

	evalCtx *tree.EvalContext

	heap []srcIdx

	needsAdvance bool

	err error

	alloc tree.DatumAlloc

	metadata []*execinfrapb.ProducerMetadata
}

var _ execinfra.RowSource = &serialOrderedSynchronizer{}

func (s *serialOrderedSynchronizer) Len() int {
	__antithesis_instrumentation__.Notify(575820)
	return len(s.heap)
}

func (s *serialOrderedSynchronizer) Less(i, j int) bool {
	__antithesis_instrumentation__.Notify(575821)
	si := &s.sources[s.heap[i]]
	sj := &s.sources[s.heap[j]]
	cmp, err := si.row.Compare(s.types, &s.alloc, s.ordering, s.evalCtx, sj.row)
	if err != nil {
		__antithesis_instrumentation__.Notify(575823)
		s.err = err
		return false
	} else {
		__antithesis_instrumentation__.Notify(575824)
	}
	__antithesis_instrumentation__.Notify(575822)
	return cmp < 0
}

func (s *serialOrderedSynchronizer) Swap(i, j int) {
	__antithesis_instrumentation__.Notify(575825)
	s.heap[i], s.heap[j] = s.heap[j], s.heap[i]
}

func (s *serialOrderedSynchronizer) Push(x interface{}) {
	__antithesis_instrumentation__.Notify(575826)
	panic("unimplemented")
}

func (s *serialOrderedSynchronizer) Pop() interface{} {
	__antithesis_instrumentation__.Notify(575827)
	s.heap = s.heap[:len(s.heap)-1]
	return nil
}

func (s *serialOrderedSynchronizer) initHeap() error {
	__antithesis_instrumentation__.Notify(575828)

	var consumeErr error

	toDelete := 0
	for i, srcIdx := range s.heap {
		__antithesis_instrumentation__.Notify(575831)
		src := &s.sources[srcIdx]
		err := s.consumeMetadata(src, stopOnRowOrError)
		if err != nil {
			__antithesis_instrumentation__.Notify(575833)
			consumeErr = err
		} else {
			__antithesis_instrumentation__.Notify(575834)
		}
		__antithesis_instrumentation__.Notify(575832)
		if src.row == nil && func() bool {
			__antithesis_instrumentation__.Notify(575835)
			return err == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(575836)

			s.heap[toDelete], s.heap[i] = s.heap[i], s.heap[toDelete]
			toDelete++
		} else {
			__antithesis_instrumentation__.Notify(575837)
		}
	}
	__antithesis_instrumentation__.Notify(575829)
	s.heap = s.heap[toDelete:]
	if consumeErr != nil {
		__antithesis_instrumentation__.Notify(575838)
		return consumeErr
	} else {
		__antithesis_instrumentation__.Notify(575839)
	}
	__antithesis_instrumentation__.Notify(575830)
	heap.Init(s)

	return s.err
}

type consumeMetadataOption int

const (
	stopOnRowOrError consumeMetadataOption = iota

	drain
)

func (s *serialOrderedSynchronizer) consumeMetadata(
	src *srcInfo, mode consumeMetadataOption,
) error {
	__antithesis_instrumentation__.Notify(575840)
	for {
		__antithesis_instrumentation__.Notify(575841)
		row, meta := src.src.Next()
		if meta != nil {
			__antithesis_instrumentation__.Notify(575844)
			if meta.Err != nil && func() bool {
				__antithesis_instrumentation__.Notify(575846)
				return mode == stopOnRowOrError == true
			}() == true {
				__antithesis_instrumentation__.Notify(575847)
				return meta.Err
			} else {
				__antithesis_instrumentation__.Notify(575848)
			}
			__antithesis_instrumentation__.Notify(575845)
			s.metadata = append(s.metadata, meta)
			continue
		} else {
			__antithesis_instrumentation__.Notify(575849)
		}
		__antithesis_instrumentation__.Notify(575842)
		if mode == stopOnRowOrError {
			__antithesis_instrumentation__.Notify(575850)
			if row != nil {
				__antithesis_instrumentation__.Notify(575852)
				row = s.rowAlloc.CopyRow(row)
			} else {
				__antithesis_instrumentation__.Notify(575853)
			}
			__antithesis_instrumentation__.Notify(575851)
			src.row = row
			return nil
		} else {
			__antithesis_instrumentation__.Notify(575854)
		}
		__antithesis_instrumentation__.Notify(575843)
		if row == nil {
			__antithesis_instrumentation__.Notify(575855)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(575856)
		}
	}
}

func (s *serialOrderedSynchronizer) advanceRoot() error {
	__antithesis_instrumentation__.Notify(575857)
	if s.state != returningRows {
		__antithesis_instrumentation__.Notify(575863)
		return errors.Errorf("advanceRoot() called in unsupported state: %d", s.state)
	} else {
		__antithesis_instrumentation__.Notify(575864)
	}
	__antithesis_instrumentation__.Notify(575858)
	if len(s.heap) == 0 {
		__antithesis_instrumentation__.Notify(575865)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(575866)
	}
	__antithesis_instrumentation__.Notify(575859)
	src := &s.sources[s.heap[0]]
	if src.row == nil {
		__antithesis_instrumentation__.Notify(575867)
		return errors.Errorf("trying to advance closed source")
	} else {
		__antithesis_instrumentation__.Notify(575868)
	}
	__antithesis_instrumentation__.Notify(575860)

	oldRow := src.row
	if err := s.consumeMetadata(src, stopOnRowOrError); err != nil {
		__antithesis_instrumentation__.Notify(575869)
		return err
	} else {
		__antithesis_instrumentation__.Notify(575870)
	}
	__antithesis_instrumentation__.Notify(575861)

	if src.row == nil {
		__antithesis_instrumentation__.Notify(575871)
		heap.Remove(s, 0)
	} else {
		__antithesis_instrumentation__.Notify(575872)
		heap.Fix(s, 0)

		if cmp, err := oldRow.Compare(s.types, &s.alloc, s.ordering, s.evalCtx, src.row); err != nil {
			__antithesis_instrumentation__.Notify(575873)
			return err
		} else {
			__antithesis_instrumentation__.Notify(575874)
			if cmp > 0 {
				__antithesis_instrumentation__.Notify(575875)
				return errors.Errorf(
					"incorrectly ordered stream %s after %s (ordering: %v)",
					src.row.String(s.types), oldRow.String(s.types), s.ordering,
				)
			} else {
				__antithesis_instrumentation__.Notify(575876)
			}
		}
	}
	__antithesis_instrumentation__.Notify(575862)

	return s.err
}

func (s *serialOrderedSynchronizer) drainSources() {
	__antithesis_instrumentation__.Notify(575877)
	for _, srcIdx := range s.heap {
		__antithesis_instrumentation__.Notify(575878)
		if err := s.consumeMetadata(&s.sources[srcIdx], drain); err != nil {
			__antithesis_instrumentation__.Notify(575879)
			log.Fatalf(context.TODO(), "unexpected draining error: %s", err)
		} else {
			__antithesis_instrumentation__.Notify(575880)
		}
	}
}

func (s *serialOrderedSynchronizer) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(575881)
	if s.state == notInitialized {
		__antithesis_instrumentation__.Notify(575886)
		if err := s.initHeap(); err != nil {
			__antithesis_instrumentation__.Notify(575888)
			s.ConsumerDone()
			return nil, &execinfrapb.ProducerMetadata{Err: err}
		} else {
			__antithesis_instrumentation__.Notify(575889)
		}
		__antithesis_instrumentation__.Notify(575887)
		s.state = returningRows
	} else {
		__antithesis_instrumentation__.Notify(575890)
		if s.state == returningRows && func() bool {
			__antithesis_instrumentation__.Notify(575891)
			return s.needsAdvance == true
		}() == true {
			__antithesis_instrumentation__.Notify(575892)

			if err := s.advanceRoot(); err != nil {
				__antithesis_instrumentation__.Notify(575893)
				s.ConsumerDone()
				return nil, &execinfrapb.ProducerMetadata{Err: err}
			} else {
				__antithesis_instrumentation__.Notify(575894)
			}
		} else {
			__antithesis_instrumentation__.Notify(575895)
		}
	}
	__antithesis_instrumentation__.Notify(575882)

	if s.state == draining {
		__antithesis_instrumentation__.Notify(575896)

		s.drainSources()
		s.state = drainBuffered
		s.heap = nil
	} else {
		__antithesis_instrumentation__.Notify(575897)
	}
	__antithesis_instrumentation__.Notify(575883)

	if len(s.metadata) != 0 {
		__antithesis_instrumentation__.Notify(575898)

		var meta *execinfrapb.ProducerMetadata
		meta, s.metadata = s.metadata[0], s.metadata[1:]
		s.needsAdvance = false
		return nil, meta
	} else {
		__antithesis_instrumentation__.Notify(575899)
	}
	__antithesis_instrumentation__.Notify(575884)

	if len(s.heap) == 0 {
		__antithesis_instrumentation__.Notify(575900)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(575901)
	}
	__antithesis_instrumentation__.Notify(575885)

	s.needsAdvance = true
	return s.sources[s.heap[0]].row, nil
}

func (s *serialOrderedSynchronizer) ConsumerDone() {
	__antithesis_instrumentation__.Notify(575902)

	if s.state != draining {
		__antithesis_instrumentation__.Notify(575903)
		s.consumerStatusChanged(draining, execinfra.RowSource.ConsumerDone)
	} else {
		__antithesis_instrumentation__.Notify(575904)
	}
}

func (s *serialOrderedSynchronizer) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(575905)

	s.consumerStatusChanged(drainBuffered, execinfra.RowSource.ConsumerClosed)
}

func (s *serialOrderedSynchronizer) consumerStatusChanged(
	newState serialSynchronizerState, f func(execinfra.RowSource),
) {
	__antithesis_instrumentation__.Notify(575906)
	if s.state == notInitialized {
		__antithesis_instrumentation__.Notify(575908)
		for i := range s.sources {
			__antithesis_instrumentation__.Notify(575909)
			f(s.sources[i].src)
		}
	} else {
		__antithesis_instrumentation__.Notify(575910)

		for _, sIdx := range s.heap {
			__antithesis_instrumentation__.Notify(575911)
			f(s.sources[sIdx].src)
		}
	}
	__antithesis_instrumentation__.Notify(575907)
	s.state = newState
}

type serialUnorderedSynchronizer struct {
	serialSynchronizerBase

	srcIndex int
}

var _ execinfra.RowSource = &serialUnorderedSynchronizer{}

func (u *serialUnorderedSynchronizer) Next() (rowenc.EncDatumRow, *execinfrapb.ProducerMetadata) {
	__antithesis_instrumentation__.Notify(575912)
	for u.srcIndex < len(u.sources) {
		__antithesis_instrumentation__.Notify(575914)
		row, metadata := u.sources[u.srcIndex].src.Next()

		if u.state == draining && func() bool {
			__antithesis_instrumentation__.Notify(575918)
			return row != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(575919)
			continue
		} else {
			__antithesis_instrumentation__.Notify(575920)
		}
		__antithesis_instrumentation__.Notify(575915)

		if row == nil && func() bool {
			__antithesis_instrumentation__.Notify(575921)
			return metadata == nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(575922)
			u.srcIndex++
			continue
		} else {
			__antithesis_instrumentation__.Notify(575923)
		}
		__antithesis_instrumentation__.Notify(575916)

		if row != nil {
			__antithesis_instrumentation__.Notify(575924)
			row = u.rowAlloc.CopyRow(row)
		} else {
			__antithesis_instrumentation__.Notify(575925)
		}
		__antithesis_instrumentation__.Notify(575917)

		return row, metadata
	}
	__antithesis_instrumentation__.Notify(575913)

	return nil, nil
}

func (u *serialUnorderedSynchronizer) ConsumerDone() {
	__antithesis_instrumentation__.Notify(575926)
	if u.state != draining {
		__antithesis_instrumentation__.Notify(575927)
		for i := range u.sources {
			__antithesis_instrumentation__.Notify(575929)
			u.sources[i].src.ConsumerDone()
		}
		__antithesis_instrumentation__.Notify(575928)
		u.state = draining
	} else {
		__antithesis_instrumentation__.Notify(575930)
	}
}

func (u *serialUnorderedSynchronizer) ConsumerClosed() {
	__antithesis_instrumentation__.Notify(575931)
	for i := range u.sources {
		__antithesis_instrumentation__.Notify(575932)
		u.sources[i].src.ConsumerClosed()
	}
}

func makeSerialSync(
	ordering colinfo.ColumnOrdering, evalCtx *tree.EvalContext, sources []execinfra.RowSource,
) (execinfra.RowSource, error) {
	__antithesis_instrumentation__.Notify(575933)
	if len(sources) < 2 {
		__antithesis_instrumentation__.Notify(575937)
		return nil, errors.Errorf("only %d sources for serial synchronizer", len(sources))
	} else {
		__antithesis_instrumentation__.Notify(575938)
	}
	__antithesis_instrumentation__.Notify(575934)

	base := serialSynchronizerBase{
		state:   notInitialized,
		sources: make([]srcInfo, len(sources)),
		types:   sources[0].OutputTypes(),
	}

	for i := range base.sources {
		__antithesis_instrumentation__.Notify(575939)
		base.sources[i].src = sources[i]
	}
	__antithesis_instrumentation__.Notify(575935)

	var sync execinfra.RowSource

	if len(ordering) > 0 {
		__antithesis_instrumentation__.Notify(575940)
		os := &serialOrderedSynchronizer{
			serialSynchronizerBase: base,
			heap:                   make([]srcIdx, 0, len(sources)),
			ordering:               ordering,
			evalCtx:                evalCtx,
		}
		for i := range os.sources {
			__antithesis_instrumentation__.Notify(575942)
			os.heap = append(os.heap, srcIdx(i))
		}
		__antithesis_instrumentation__.Notify(575941)
		sync = os
	} else {
		__antithesis_instrumentation__.Notify(575943)
		sync = &serialUnorderedSynchronizer{
			serialSynchronizerBase: base,
		}
	}
	__antithesis_instrumentation__.Notify(575936)

	return sync, nil
}
