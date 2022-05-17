package denylist

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/errors"
	"github.com/google/btree"
)

type Watcher struct {
	mu syncutil.Mutex

	nextID int64

	listeners *btree.BTree

	list *Denylist
}

type listener struct {
	mu struct {
		syncutil.Mutex

		denied func(error)
	}

	id int64

	connection ConnectionTags
}

type ConnectionTags struct {
	IP      string
	Cluster string
}

func NilWatcher() *Watcher {
	__antithesis_instrumentation__.Notify(21475)
	c := make(chan *Denylist)
	close(c)
	return newWatcher(emptyList(), c)
}

func WatcherFromFile(ctx context.Context, filename string, opts ...Option) *Watcher {
	__antithesis_instrumentation__.Notify(21476)
	dl, dlc := newDenylistWithFile(ctx, filename, opts...)
	return newWatcher(dl, dlc)
}

func newWatcher(list *Denylist, next chan *Denylist) *Watcher {
	__antithesis_instrumentation__.Notify(21477)
	w := &Watcher{
		listeners: btree.New(8),
		list:      list,
	}

	go func() {
		__antithesis_instrumentation__.Notify(21479)
		for n := range next {
			__antithesis_instrumentation__.Notify(21480)
			w.updateDenyList(n)
		}
	}()
	__antithesis_instrumentation__.Notify(21478)

	return w
}

func (w *Watcher) ListenForDenied(connection ConnectionTags, callback func(error)) (func(), error) {
	__antithesis_instrumentation__.Notify(21481)
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := checkConnection(connection, w.list); err != nil {
		__antithesis_instrumentation__.Notify(21483)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(21484)
	}
	__antithesis_instrumentation__.Notify(21482)

	id := w.nextID
	w.nextID++

	l := &listener{
		id:         id,
		connection: connection,
	}
	l.mu.denied = callback

	w.listeners.ReplaceOrInsert(l)
	return func() { __antithesis_instrumentation__.Notify(21485); w.removeListener(l) }, nil
}

func (w *Watcher) updateDenyList(list *Denylist) {
	__antithesis_instrumentation__.Notify(21486)
	w.mu.Lock()
	copy := w.listeners.Clone()
	w.list = list
	w.mu.Unlock()

	copy.Ascend(func(i btree.Item) bool {
		__antithesis_instrumentation__.Notify(21487)
		lst := i.(*listener)
		if err := checkConnection(lst.connection, list); err != nil {
			__antithesis_instrumentation__.Notify(21489)
			lst.mu.Lock()
			defer lst.mu.Unlock()
			if lst.mu.denied != nil {
				__antithesis_instrumentation__.Notify(21490)
				lst.mu.denied(err)
				lst.mu.denied = nil
			} else {
				__antithesis_instrumentation__.Notify(21491)
			}
		} else {
			__antithesis_instrumentation__.Notify(21492)
		}
		__antithesis_instrumentation__.Notify(21488)
		return true
	})
}

func (w *Watcher) removeListener(l *listener) {
	__antithesis_instrumentation__.Notify(21493)
	l.mu.Lock()
	defer l.mu.Unlock()
	w.mu.Lock()
	defer w.mu.Unlock()

	l.mu.denied = nil

	w.listeners.Delete(l)
}

func (l *listener) Less(than btree.Item) bool {
	__antithesis_instrumentation__.Notify(21494)
	return l.id < than.(*listener).id
}

func checkConnection(connection ConnectionTags, list *Denylist) error {
	__antithesis_instrumentation__.Notify(21495)
	ip := DenyEntity{Item: connection.IP, Type: IPAddrType}
	if err := list.Denied(ip); err != nil {
		__antithesis_instrumentation__.Notify(21498)
		return errors.Wrapf(err, "connection ip '%v' denied", connection.IP)
	} else {
		__antithesis_instrumentation__.Notify(21499)
	}
	__antithesis_instrumentation__.Notify(21496)
	cluster := DenyEntity{Item: connection.Cluster, Type: ClusterType}
	if err := list.Denied(cluster); err != nil {
		__antithesis_instrumentation__.Notify(21500)
		return errors.Wrapf(err, "connection cluster '%v' denied", connection.Cluster)
	} else {
		__antithesis_instrumentation__.Notify(21501)
	}
	__antithesis_instrumentation__.Notify(21497)
	return nil
}
