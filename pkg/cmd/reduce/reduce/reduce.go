// Package reduce implements a reducer core for reducing the size of test
// failure cases.
//
// See: https://blog.regehr.org/archives/1678.
package reduce

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"log"
	"math/rand"
	"runtime"

	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type Pass interface {
	New(string) State

	Transform(string, State) (string, Result, error)

	Advance(string, State) State

	Name() string
}

type Result int

const (
	OK Result = iota

	STOP
)

type State interface{}

type InterestingFn func(context.Context, string) (_ bool, logOriginalHint func())

type Mode int

const (
	ModeSize Mode = iota

	ModeInteresting
)

func Reduce(
	logger *log.Logger,
	originalTestCase string,
	isInteresting InterestingFn,
	numGoroutines int,
	mode Mode,
	chunkReducer ChunkReducer,
	passList ...Pass,
) (string, error) {
	__antithesis_instrumentation__.Notify(41835)
	if numGoroutines < 1 {
		__antithesis_instrumentation__.Notify(41843)
		numGoroutines = runtime.GOMAXPROCS(0)
	} else {
		__antithesis_instrumentation__.Notify(41844)
	}
	__antithesis_instrumentation__.Notify(41836)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if interesting, logHint := isInteresting(ctx, originalTestCase); !interesting {
		__antithesis_instrumentation__.Notify(41845)
		if logHint != nil {
			__antithesis_instrumentation__.Notify(41847)
			logHint()
		} else {
			__antithesis_instrumentation__.Notify(41848)
		}
		__antithesis_instrumentation__.Notify(41846)
		return "", errors.New("original test case not interesting")
	} else {
		__antithesis_instrumentation__.Notify(41849)
	}
	__antithesis_instrumentation__.Notify(41837)

	chunkReducedTestCase, err := attemptChunkReduction(logger, originalTestCase, isInteresting, chunkReducer)
	if err != nil {
		__antithesis_instrumentation__.Notify(41850)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(41851)
	}
	__antithesis_instrumentation__.Notify(41838)

	findNextInteresting := func(vs varState) (*varState, error) {
		__antithesis_instrumentation__.Notify(41852)
		ctx := context.Background()
		g := ctxgroup.WithContext(ctx)
		variants := make(chan varState)
		g.GoCtx(func(ctx context.Context) error {
			__antithesis_instrumentation__.Notify(41856)

			defer close(variants)
			current := vs.file
			state := vs.s
			var done, prev chan struct{}

			prev = make(chan struct{}, 1)
			prev <- struct{}{}
			for pi := vs.pi; pi < len(passList); pi++ {
				__antithesis_instrumentation__.Notify(41858)
				p := passList[pi]
				if state == nil {
					__antithesis_instrumentation__.Notify(41860)
					state = p.New(current)
				} else {
					__antithesis_instrumentation__.Notify(41861)
				}
				__antithesis_instrumentation__.Notify(41859)
				for {
					__antithesis_instrumentation__.Notify(41862)
					variant, result, err := p.Transform(current, state)
					if err != nil || func() bool {
						__antithesis_instrumentation__.Notify(41865)
						return result != OK == true
					}() == true {
						__antithesis_instrumentation__.Notify(41866)
						state = nil
						break
					} else {
						__antithesis_instrumentation__.Notify(41867)
					}
					__antithesis_instrumentation__.Notify(41863)

					done = make(chan struct{}, 1)
					select {
					case variants <- varState{
						pi:   pi,
						file: variant,
						s:    state,
						done: done,
						prev: prev,
					}:
						__antithesis_instrumentation__.Notify(41868)
						prev = done
					case <-ctx.Done():
						__antithesis_instrumentation__.Notify(41869)
						return nil
					}
					__antithesis_instrumentation__.Notify(41864)
					state = p.Advance(current, state)
				}
			}
			__antithesis_instrumentation__.Notify(41857)
			return nil
		})
		__antithesis_instrumentation__.Notify(41853)

		for i := 0; i < numGoroutines; i++ {
			__antithesis_instrumentation__.Notify(41870)
			g.GoCtx(func(ctx context.Context) error {
				__antithesis_instrumentation__.Notify(41871)
				for vs := range variants {
					__antithesis_instrumentation__.Notify(41873)
					if interesting, _ := isInteresting(ctx, vs.file); interesting {
						__antithesis_instrumentation__.Notify(41875)

						select {
						case <-ctx.Done():
							__antithesis_instrumentation__.Notify(41876)
							return nil
						case <-vs.prev:
							__antithesis_instrumentation__.Notify(41877)

							return errInteresting(vs)
						}
					} else {
						__antithesis_instrumentation__.Notify(41878)
					}
					__antithesis_instrumentation__.Notify(41874)
					vs.done <- struct{}{}
				}
				__antithesis_instrumentation__.Notify(41872)
				return nil
			})
		}
		__antithesis_instrumentation__.Notify(41854)

		if err := g.Wait(); err != nil {
			__antithesis_instrumentation__.Notify(41879)
			var ierr errInteresting
			if errors.As(err, &ierr) {
				__antithesis_instrumentation__.Notify(41881)
				vs := varState(ierr)
				if logger != nil {
					__antithesis_instrumentation__.Notify(41883)
					logger.Printf("\tpass %d of %d (%s): %d bytes\n", vs.pi+1, len(passList),
						passList[vs.pi].Name(), len(vs.file))
				} else {
					__antithesis_instrumentation__.Notify(41884)
				}
				__antithesis_instrumentation__.Notify(41882)
				return &vs, nil
			} else {
				__antithesis_instrumentation__.Notify(41885)
			}
			__antithesis_instrumentation__.Notify(41880)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(41886)
		}
		__antithesis_instrumentation__.Notify(41855)
		return nil, nil
	}
	__antithesis_instrumentation__.Notify(41839)

	start := timeutil.Now()
	vs := varState{
		file: chunkReducedTestCase,
	}
	if logger != nil {
		__antithesis_instrumentation__.Notify(41887)
		logger.Printf("size: %d\n", len(vs.file))
	} else {
		__antithesis_instrumentation__.Notify(41888)
	}
	__antithesis_instrumentation__.Notify(41840)
	for {
		__antithesis_instrumentation__.Notify(41889)
		sizeAtStart := len(vs.file)
		foundInteresting := false
		for {
			__antithesis_instrumentation__.Notify(41893)
			next, err := findNextInteresting(vs)
			if err != nil {
				__antithesis_instrumentation__.Notify(41896)
				if logger != nil {
					__antithesis_instrumentation__.Notify(41898)
					logger.Printf("unexpected error: %s", err)
				} else {
					__antithesis_instrumentation__.Notify(41899)
				}
				__antithesis_instrumentation__.Notify(41897)

				return "", nil
			} else {
				__antithesis_instrumentation__.Notify(41900)
			}
			__antithesis_instrumentation__.Notify(41894)
			if next == nil {
				__antithesis_instrumentation__.Notify(41901)
				break
			} else {
				__antithesis_instrumentation__.Notify(41902)
			}
			__antithesis_instrumentation__.Notify(41895)
			foundInteresting = true
			vs = *next
		}
		__antithesis_instrumentation__.Notify(41890)
		done := false
		switch mode {
		case ModeSize:
			__antithesis_instrumentation__.Notify(41903)
			if len(vs.file) >= sizeAtStart {
				__antithesis_instrumentation__.Notify(41906)
				done = true
			} else {
				__antithesis_instrumentation__.Notify(41907)
			}
		case ModeInteresting:
			__antithesis_instrumentation__.Notify(41904)
			done = !foundInteresting
		default:
			__antithesis_instrumentation__.Notify(41905)
			panic("unknown mode")
		}
		__antithesis_instrumentation__.Notify(41891)
		if done {
			__antithesis_instrumentation__.Notify(41908)
			break
		} else {
			__antithesis_instrumentation__.Notify(41909)
		}
		__antithesis_instrumentation__.Notify(41892)

		vs = varState{
			file: vs.file,
		}
	}
	__antithesis_instrumentation__.Notify(41841)
	if logger != nil {
		__antithesis_instrumentation__.Notify(41910)
		logger.Printf("total time: %v\n", timeutil.Since(start))
		logger.Printf("original size: %v\n", len(originalTestCase))
		if chunkReducer != nil {
			__antithesis_instrumentation__.Notify(41912)
			logger.Printf("chunk-reduced size: %v\n", len(chunkReducedTestCase))
		} else {
			__antithesis_instrumentation__.Notify(41913)
		}
		__antithesis_instrumentation__.Notify(41911)
		logger.Printf("final size: %v\n", len(vs.file))
		logger.Printf("reduction: %v%%\n", 100-int(100*float64(len(vs.file))/float64(len(originalTestCase))))
	} else {
		__antithesis_instrumentation__.Notify(41914)
	}
	__antithesis_instrumentation__.Notify(41842)
	return vs.file, nil
}

type errInteresting varState

func (e errInteresting) Error() string {
	__antithesis_instrumentation__.Notify(41915)
	return "interesting"
}

type varState struct {
	pi   int
	file string
	s    State

	done, prev chan struct{}
}

type ChunkReducer interface {
	HaltAfter() int

	Init(string) error

	NumSegments() int

	DeleteSegments(start, end int) string
}

func attemptChunkReduction(
	logger *log.Logger,
	originalTestCase string,
	isInteresting InterestingFn,
	chunkReducer ChunkReducer,
) (string, error) {
	__antithesis_instrumentation__.Notify(41916)
	if chunkReducer == nil {
		__antithesis_instrumentation__.Notify(41919)
		return originalTestCase, nil
	} else {
		__antithesis_instrumentation__.Notify(41920)
	}
	__antithesis_instrumentation__.Notify(41917)

	ctx := context.Background()
	reduced := originalTestCase

	failedAttempts := 0
	for failedAttempts < chunkReducer.HaltAfter() {
		__antithesis_instrumentation__.Notify(41921)
		err := chunkReducer.Init(reduced)
		if err != nil {
			__antithesis_instrumentation__.Notify(41923)
			return "", err
		} else {
			__antithesis_instrumentation__.Notify(41924)
		}
		__antithesis_instrumentation__.Notify(41922)

		start := rand.Intn(chunkReducer.NumSegments())
		end := rand.Intn(chunkReducer.NumSegments()-start) + start + 1

		localReduced := chunkReducer.DeleteSegments(start, end)
		if interesting, _ := isInteresting(ctx, localReduced); interesting {
			__antithesis_instrumentation__.Notify(41925)
			reduced = localReduced
			if logger != nil {
				__antithesis_instrumentation__.Notify(41927)
				logger.Printf("\tchunk reduction: %d bytes\n", len(reduced))
			} else {
				__antithesis_instrumentation__.Notify(41928)
			}
			__antithesis_instrumentation__.Notify(41926)
			failedAttempts = 0
		} else {
			__antithesis_instrumentation__.Notify(41929)
			failedAttempts++
		}
	}
	__antithesis_instrumentation__.Notify(41918)

	return reduced, nil
}
