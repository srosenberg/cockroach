package denylist

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/util/timeutil"
	"github.com/cockroachdb/errors"
)

type Denylist struct {
	timeSource timeutil.TimeSource
	entries    map[DenyEntity]*DenyEntry
}

type DenyEntry struct {
	Entity     DenyEntity `yaml:"entity"`
	Expiration time.Time  `yaml:"expiration"`
	Reason     string     `yaml:"reason"`
}

type DenyEntity struct {
	Item string `yaml:"item"`
	Type Type   `yaml:"type"`
}

type Type int

const (
	IPAddrType Type = iota + 1
	ClusterType
	UnknownType
)

func (dl *Denylist) Denied(entity DenyEntity) error {
	__antithesis_instrumentation__.Notify(21420)
	if ent, ok := dl.entries[entity]; ok && func() bool {
		__antithesis_instrumentation__.Notify(21422)
		return (ent.Expiration.IsZero() || func() bool {
			__antithesis_instrumentation__.Notify(21423)
			return !ent.Expiration.Before(dl.timeSource.Now()) == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(21424)
		return errors.Newf("%s", ent.Reason)
	} else {
		__antithesis_instrumentation__.Notify(21425)
	}
	__antithesis_instrumentation__.Notify(21421)
	return nil
}

func emptyList() *Denylist {
	__antithesis_instrumentation__.Notify(21426)
	return &Denylist{
		timeSource: timeutil.DefaultTimeSource{},
		entries:    make(map[DenyEntity]*DenyEntry)}
}
