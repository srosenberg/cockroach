package kvnemesis

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"sync/atomic"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/util/ctxgroup"
	"github.com/cockroachdb/cockroach/pkg/util/log"
)

func RunNemesis(
	ctx context.Context,
	rng *rand.Rand,
	env *Env,
	config GeneratorConfig,
	numSteps int,
	dbs ...*kv.DB,
) ([]error, error) {
	__antithesis_instrumentation__.Notify(90393)
	const concurrency = 5
	if numSteps <= 0 {
		__antithesis_instrumentation__.Notify(90403)
		return nil, fmt.Errorf("numSteps must be >0, got %v", numSteps)
	} else {
		__antithesis_instrumentation__.Notify(90404)
	}
	__antithesis_instrumentation__.Notify(90394)

	g, err := MakeGenerator(config, newGetReplicasFn(dbs...))
	if err != nil {
		__antithesis_instrumentation__.Notify(90405)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(90406)
	}
	__antithesis_instrumentation__.Notify(90395)
	a := MakeApplier(env, dbs...)
	w, err := Watch(ctx, env, dbs, GeneratorDataSpan())
	if err != nil {
		__antithesis_instrumentation__.Notify(90407)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(90408)
	}
	__antithesis_instrumentation__.Notify(90396)
	defer func() { __antithesis_instrumentation__.Notify(90409); _ = w.Finish() }()
	__antithesis_instrumentation__.Notify(90397)

	var stepsStartedAtomic int64
	stepsByWorker := make([][]Step, concurrency)

	workerFn := func(ctx context.Context, workerIdx int) error {
		__antithesis_instrumentation__.Notify(90410)
		workerName := fmt.Sprintf(`%d`, workerIdx)
		var buf strings.Builder
		for atomic.AddInt64(&stepsStartedAtomic, 1) <= int64(numSteps) {
			__antithesis_instrumentation__.Notify(90412)
			step := g.RandStep(rng)

			buf.Reset()
			fmt.Fprintf(&buf, "step:")
			step.format(&buf, formatCtx{indent: `  ` + workerName + ` PRE  `})
			trace, err := a.Apply(ctx, &step)
			buf.WriteString(trace.String())
			step.Trace = buf.String()
			if err != nil {
				__antithesis_instrumentation__.Notify(90414)
				buf.Reset()
				step.format(&buf, formatCtx{indent: `  ` + workerName + ` ERR `})
				log.Infof(ctx, "error: %+v\n\n%s", err, buf.String())
				return err
			} else {
				__antithesis_instrumentation__.Notify(90415)
			}
			__antithesis_instrumentation__.Notify(90413)
			buf.Reset()
			fmt.Fprintf(&buf, "\n  before: %s", step.Before)
			step.format(&buf, formatCtx{indent: `  ` + workerName + ` OP  `})
			fmt.Fprintf(&buf, "\n  after: %s", step.After)
			log.Infof(ctx, "%v", buf.String())
			stepsByWorker[workerIdx] = append(stepsByWorker[workerIdx], step)
		}
		__antithesis_instrumentation__.Notify(90411)
		return nil
	}
	__antithesis_instrumentation__.Notify(90398)
	if err := ctxgroup.GroupWorkers(ctx, concurrency, workerFn); err != nil {
		__antithesis_instrumentation__.Notify(90416)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(90417)
	}
	__antithesis_instrumentation__.Notify(90399)

	allSteps := make(steps, 0, numSteps)
	for _, steps := range stepsByWorker {
		__antithesis_instrumentation__.Notify(90418)
		allSteps = append(allSteps, steps...)
	}
	__antithesis_instrumentation__.Notify(90400)

	if err := w.WaitForFrontier(ctx, allSteps.After()); err != nil {
		__antithesis_instrumentation__.Notify(90419)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(90420)
	}
	__antithesis_instrumentation__.Notify(90401)
	kvs := w.Finish()
	defer kvs.Close()
	failures := Validate(allSteps, kvs)

	if len(failures) > 0 {
		__antithesis_instrumentation__.Notify(90421)
		log.Infof(ctx, "reproduction steps:\n%s", printRepro(stepsByWorker))
		log.Infof(ctx, "kvs (recorded from rangefeed):\n%s", kvs.DebugPrint("  "))

		span := GeneratorDataSpan()
		scanKVs, err := dbs[0].Scan(ctx, span.Key, span.EndKey, -1)
		if err != nil {
			__antithesis_instrumentation__.Notify(90422)
			log.Infof(ctx, "could not scan actual latest values: %+v", err)
		} else {
			__antithesis_instrumentation__.Notify(90423)
			var kvsBuf strings.Builder
			for _, kv := range scanKVs {
				__antithesis_instrumentation__.Notify(90425)
				fmt.Fprintf(&kvsBuf, "  %s %s -> %s\n", kv.Key, kv.Value.Timestamp, kv.Value.PrettyPrint())
			}
			__antithesis_instrumentation__.Notify(90424)
			log.Infof(ctx, "kvs (scan of latest values according to crdb):\n%s", kvsBuf.String())
		}
	} else {
		__antithesis_instrumentation__.Notify(90426)
	}
	__antithesis_instrumentation__.Notify(90402)

	return failures, nil
}

func printRepro(stepsByWorker [][]Step) string {
	__antithesis_instrumentation__.Notify(90427)

	var buf strings.Builder
	buf.WriteString("g := ctxgroup.WithContext(ctx)\n")
	for _, steps := range stepsByWorker {
		__antithesis_instrumentation__.Notify(90429)
		buf.WriteString("g.GoCtx(func(ctx context.Context) error {")
		for _, step := range steps {
			__antithesis_instrumentation__.Notify(90431)
			fctx := formatCtx{receiver: fmt.Sprintf(`db%d`, step.DBID), indent: "  "}
			buf.WriteString("\n")
			buf.WriteString(fctx.indent)
			step.Op.format(&buf, fctx)
			buf.WriteString(step.Trace)
			buf.WriteString("\n")
		}
		__antithesis_instrumentation__.Notify(90430)
		buf.WriteString("\n  return nil\n")
		buf.WriteString("})\n")
	}
	__antithesis_instrumentation__.Notify(90428)
	buf.WriteString("g.Wait()\n")
	return buf.String()
}
