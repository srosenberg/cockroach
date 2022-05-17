package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/storage"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

var debugSyncTestCmd = &cobra.Command{
	Use:   "synctest <empty-dir> <nemesis-script>",
	Short: "Run a log-like workload that can help expose filesystem anomalies",
	Long: `
synctest is a tool to verify filesystem consistency in the presence of I/O errors.
It takes a directory (required to be initially empty and created on demand) into
which data will be written and a nemesis script which receives a single argument
that is either "on" or "off".

The nemesis script will be called with a parameter of "on" when the filesystem
underlying the given directory should be "disturbed". It is called with "off"
to restore the undisturbed state (note that "off" must be idempotent).

synctest will run run across multiple "epochs", each terminated by an I/O error
injected by the nemesis. After each epoch, the nemesis is turned off and the
written data is reopened, checked for data loss, new data is written, and
the nemesis turned back on. In the absence of unexpected error or user interrupt,
this process continues indefinitely.
`,
	Args: cobra.ExactArgs(2),
	RunE: runDebugSyncTest,
}

type scriptNemesis string

func (sn scriptNemesis) exec(arg string) error {
	__antithesis_instrumentation__.Notify(31713)
	b, err := exec.Command(string(sn), arg).CombinedOutput()
	if err != nil {
		__antithesis_instrumentation__.Notify(31715)
		return errors.WithDetailf(err, "command output:\n%s", string(b))
	} else {
		__antithesis_instrumentation__.Notify(31716)
	}
	__antithesis_instrumentation__.Notify(31714)
	fmt.Fprintf(stderr, "%s %s: %s", sn, arg, b)
	return nil
}

func (sn scriptNemesis) On() error {
	__antithesis_instrumentation__.Notify(31717)
	return sn.exec("on")
}

func (sn scriptNemesis) Off() error {
	__antithesis_instrumentation__.Notify(31718)
	return sn.exec("off")
}

func runDebugSyncTest(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(31719)

	duration := 10 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), duration)
	defer cancel()

	nem := scriptNemesis(args[1])
	if err := nem.Off(); err != nil {
		__antithesis_instrumentation__.Notify(31721)
		return errors.Wrap(err, "unable to disable nemesis at beginning of run")
	} else {
		__antithesis_instrumentation__.Notify(31722)
	}
	__antithesis_instrumentation__.Notify(31720)

	var generation int
	var lastSeq int64
	for {
		__antithesis_instrumentation__.Notify(31723)
		dir := filepath.Join(args[0], strconv.Itoa(generation))
		curLastSeq, err := runSyncer(ctx, dir, lastSeq, nem)
		if err != nil {
			__antithesis_instrumentation__.Notify(31725)
			return err
		} else {
			__antithesis_instrumentation__.Notify(31726)
		}
		__antithesis_instrumentation__.Notify(31724)
		lastSeq = curLastSeq
		if curLastSeq == 0 {
			__antithesis_instrumentation__.Notify(31727)
			if ctx.Err() != nil {
				__antithesis_instrumentation__.Notify(31729)

				return nil
			} else {
				__antithesis_instrumentation__.Notify(31730)
			}
			__antithesis_instrumentation__.Notify(31728)

			generation++
			continue
		} else {
			__antithesis_instrumentation__.Notify(31731)
		}
	}
}

type nemesisI interface {
	On() error
	Off() error
}

func runSyncer(
	ctx context.Context, dir string, expSeq int64, nemesis nemesisI,
) (lastSeq int64, _ error) {
	__antithesis_instrumentation__.Notify(31732)
	stopper := stop.NewStopper()
	defer stopper.Stop(ctx)

	db, err := OpenEngine(dir, stopper, OpenEngineOptions{})
	if err != nil {
		__antithesis_instrumentation__.Notify(31740)
		if expSeq == 0 {
			__antithesis_instrumentation__.Notify(31742)

			return 0, err
		} else {
			__antithesis_instrumentation__.Notify(31743)
		}
		__antithesis_instrumentation__.Notify(31741)
		fmt.Fprintln(stderr, "RocksDB directory", dir, "corrupted:", err)
		return 0, nil
	} else {
		__antithesis_instrumentation__.Notify(31744)
	}
	__antithesis_instrumentation__.Notify(31733)

	buf := make([]byte, 128)
	var seq int64
	key := func() roachpb.Key {
		__antithesis_instrumentation__.Notify(31745)
		seq++
		return encoding.EncodeUvarintAscending(buf[:0:0], uint64(seq))
	}
	__antithesis_instrumentation__.Notify(31734)

	check := func(kv storage.MVCCKeyValue) error {
		__antithesis_instrumentation__.Notify(31746)
		expKey := key()
		if !bytes.Equal(kv.Key.Key, expKey) {
			__antithesis_instrumentation__.Notify(31748)
			return errors.Errorf(
				"found unexpected key %q (expected %q)", kv.Key.Key, expKey,
			)
		} else {
			__antithesis_instrumentation__.Notify(31749)
		}
		__antithesis_instrumentation__.Notify(31747)
		return nil
	}
	__antithesis_instrumentation__.Notify(31735)

	fmt.Fprintf(stderr, "verifying existing sequence numbers...")
	if err := db.MVCCIterate(roachpb.KeyMin, roachpb.KeyMax, storage.MVCCKeyAndIntentsIterKind, check); err != nil {
		__antithesis_instrumentation__.Notify(31750)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(31751)
	}
	__antithesis_instrumentation__.Notify(31736)

	if expSeq != 0 && func() bool {
		__antithesis_instrumentation__.Notify(31752)
		return seq < expSeq == true
	}() == true {
		__antithesis_instrumentation__.Notify(31753)
		return 0, errors.Errorf("highest persisted sequence number is %d, but expected at least %d", seq, expSeq)
	} else {
		__antithesis_instrumentation__.Notify(31754)
	}
	__antithesis_instrumentation__.Notify(31737)
	fmt.Fprintf(stderr, "done (seq=%d).\nWriting new entries:\n", seq)

	waitFailure := time.After(time.Duration(rand.Int63n(5 * time.Second.Nanoseconds())))

	if err := stopper.RunAsyncTask(ctx, "syncer", func(ctx context.Context) {
		__antithesis_instrumentation__.Notify(31755)
		<-waitFailure
		if err := nemesis.On(); err != nil {
			__antithesis_instrumentation__.Notify(31758)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(31759)
		}
		__antithesis_instrumentation__.Notify(31756)
		defer func() {
			__antithesis_instrumentation__.Notify(31760)
			if err := nemesis.Off(); err != nil {
				__antithesis_instrumentation__.Notify(31761)
				panic(err)
			} else {
				__antithesis_instrumentation__.Notify(31762)
			}
		}()
		__antithesis_instrumentation__.Notify(31757)
		<-stopper.ShouldQuiesce()
	}); err != nil {
		__antithesis_instrumentation__.Notify(31763)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(31764)
	}
	__antithesis_instrumentation__.Notify(31738)

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, drainSignals...)

	write := func() (_ int64, err error) {
		__antithesis_instrumentation__.Notify(31765)
		defer func() {
			__antithesis_instrumentation__.Notify(31768)

			if r := recover(); r != nil {
				__antithesis_instrumentation__.Notify(31769)
				if err == nil {
					__antithesis_instrumentation__.Notify(31771)
					err = errors.New("recovered panic on write")
				} else {
					__antithesis_instrumentation__.Notify(31772)
				}
				__antithesis_instrumentation__.Notify(31770)
				err = errors.Wrapf(err, "%v", r)
			} else {
				__antithesis_instrumentation__.Notify(31773)
			}
		}()
		__antithesis_instrumentation__.Notify(31766)

		k, v := key(), []byte("payload")
		switch seq % 2 {
		case 0:
			__antithesis_instrumentation__.Notify(31774)
			if err := db.PutUnversioned(k, v); err != nil {
				__antithesis_instrumentation__.Notify(31778)
				seq--
				return seq, err
			} else {
				__antithesis_instrumentation__.Notify(31779)
			}
			__antithesis_instrumentation__.Notify(31775)
			if err := db.Flush(); err != nil {
				__antithesis_instrumentation__.Notify(31780)
				seq--
				return seq, err
			} else {
				__antithesis_instrumentation__.Notify(31781)
			}
		default:
			__antithesis_instrumentation__.Notify(31776)
			b := db.NewBatch()
			if err := b.PutUnversioned(k, v); err != nil {
				__antithesis_instrumentation__.Notify(31782)
				seq--
				return seq, err
			} else {
				__antithesis_instrumentation__.Notify(31783)
			}
			__antithesis_instrumentation__.Notify(31777)
			if err := b.Commit(true); err != nil {
				__antithesis_instrumentation__.Notify(31784)
				seq--
				return seq, err
			} else {
				__antithesis_instrumentation__.Notify(31785)
			}
		}
		__antithesis_instrumentation__.Notify(31767)
		return seq, nil
	}
	__antithesis_instrumentation__.Notify(31739)

	for {
		__antithesis_instrumentation__.Notify(31786)
		if lastSeq, err := write(); err != nil {
			__antithesis_instrumentation__.Notify(31788)

			for n := rand.Intn(3); n >= 0; n-- {
				__antithesis_instrumentation__.Notify(31790)
				if n == 1 {
					__antithesis_instrumentation__.Notify(31792)
					if err := nemesis.Off(); err != nil {
						__antithesis_instrumentation__.Notify(31793)
						return 0, err
					} else {
						__antithesis_instrumentation__.Notify(31794)
					}
				} else {
					__antithesis_instrumentation__.Notify(31795)
				}
				__antithesis_instrumentation__.Notify(31791)
				fmt.Fprintf(stderr, "error after seq %d (trying %d additional writes): %v\n", lastSeq, n, err)
				lastSeq, err = write()
			}
			__antithesis_instrumentation__.Notify(31789)
			fmt.Fprintf(stderr, "error after seq %d: %v\n", lastSeq, err)

			return lastSeq, nil
		} else {
			__antithesis_instrumentation__.Notify(31796)
		}
		__antithesis_instrumentation__.Notify(31787)
		select {
		case sig := <-ch:
			__antithesis_instrumentation__.Notify(31797)
			return seq, errors.Errorf("interrupted (%v)", sig)
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(31798)
			return 0, nil
		default:
			__antithesis_instrumentation__.Notify(31799)
		}
	}
}
