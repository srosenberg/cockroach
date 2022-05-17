package ledger

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"math/rand"
	"strconv"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/cockroach/pkg/workload/histogram"
	"github.com/cockroachdb/errors"
)

type worker struct {
	config *ledger
	hists  *histogram.Histograms
	db     *gosql.DB

	rng      *rand.Rand
	deckPerm []int
	permIdx  int
}

type ledgerTx interface {
	run(config *ledger, db *gosql.DB, rng *rand.Rand) (interface{}, error)
}

type tx struct {
	ledgerTx
	weight int
	name   string
}

var allTxs = [...]tx{
	{
		ledgerTx: balance{}, name: "balance",
	},
	{
		ledgerTx: withdrawal{}, name: "withdrawal",
	},
	{
		ledgerTx: deposit{}, name: "deposit",
	},
	{
		ledgerTx: reversal{}, name: "reversal",
	},
}

func initializeMix(config *ledger) error {
	__antithesis_instrumentation__.Notify(694726)
	config.txs = append([]tx(nil), allTxs[0:]...)
	nameToTx := make(map[string]int, len(allTxs))
	for i, tx := range config.txs {
		__antithesis_instrumentation__.Notify(694730)
		nameToTx[tx.name] = i
	}
	__antithesis_instrumentation__.Notify(694727)

	items := strings.Split(config.mix, `,`)
	totalWeight := 0
	for _, item := range items {
		__antithesis_instrumentation__.Notify(694731)
		kv := strings.Split(item, `=`)
		if len(kv) != 2 {
			__antithesis_instrumentation__.Notify(694735)
			return errors.Errorf(`Invalid mix %s: %s is not a k=v pair`, config.mix, item)
		} else {
			__antithesis_instrumentation__.Notify(694736)
		}
		__antithesis_instrumentation__.Notify(694732)
		txName, weightStr := kv[0], kv[1]

		weight, err := strconv.Atoi(weightStr)
		if err != nil {
			__antithesis_instrumentation__.Notify(694737)
			return errors.Errorf(
				`Invalid percentage mix %s: %s is not an integer`, config.mix, weightStr)
		} else {
			__antithesis_instrumentation__.Notify(694738)
		}
		__antithesis_instrumentation__.Notify(694733)

		i, ok := nameToTx[txName]
		if !ok {
			__antithesis_instrumentation__.Notify(694739)
			return errors.Errorf(
				`Invalid percentage mix %s: no such transaction %s`, config.mix, txName)
		} else {
			__antithesis_instrumentation__.Notify(694740)
		}
		__antithesis_instrumentation__.Notify(694734)

		config.txs[i].weight = weight
		totalWeight += weight
	}
	__antithesis_instrumentation__.Notify(694728)

	config.deck = make([]int, 0, totalWeight)
	for i, t := range config.txs {
		__antithesis_instrumentation__.Notify(694741)
		for j := 0; j < t.weight; j++ {
			__antithesis_instrumentation__.Notify(694742)
			config.deck = append(config.deck, i)
		}
	}
	__antithesis_instrumentation__.Notify(694729)

	return nil
}

func (w *worker) run(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(694743)
	if w.permIdx == len(w.deckPerm) {
		__antithesis_instrumentation__.Notify(694746)
		rand.Shuffle(len(w.deckPerm), func(i, j int) {
			__antithesis_instrumentation__.Notify(694748)
			w.deckPerm[i], w.deckPerm[j] = w.deckPerm[j], w.deckPerm[i]
		})
		__antithesis_instrumentation__.Notify(694747)
		w.permIdx = 0
	} else {
		__antithesis_instrumentation__.Notify(694749)
	}
	__antithesis_instrumentation__.Notify(694744)

	opIdx := w.deckPerm[w.permIdx]
	t := w.config.txs[opIdx]
	w.permIdx++

	start := timeutil.Now()
	if _, err := t.run(w.config, w.db, w.rng); err != nil {
		__antithesis_instrumentation__.Notify(694750)
		return errors.Wrapf(err, "error in %s", t.name)
	} else {
		__antithesis_instrumentation__.Notify(694751)
	}
	__antithesis_instrumentation__.Notify(694745)
	elapsed := timeutil.Since(start)
	w.hists.Get(t.name).Record(elapsed)
	return nil
}
