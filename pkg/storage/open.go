package storage

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/pebble/vfs"
)

type ConfigOption func(cfg *engineConfig) error

func CombineOptions(opts ...ConfigOption) ConfigOption {
	__antithesis_instrumentation__.Notify(641638)
	return func(cfg *engineConfig) error {
		__antithesis_instrumentation__.Notify(641639)
		for _, opt := range opts {
			__antithesis_instrumentation__.Notify(641641)
			if err := opt(cfg); err != nil {
				__antithesis_instrumentation__.Notify(641642)
				return err
			} else {
				__antithesis_instrumentation__.Notify(641643)
			}
		}
		__antithesis_instrumentation__.Notify(641640)
		return nil
	}
}

var ReadOnly ConfigOption = func(cfg *engineConfig) error {
	__antithesis_instrumentation__.Notify(641644)
	cfg.Opts.ReadOnly = true
	return nil
}

var MustExist ConfigOption = func(cfg *engineConfig) error {
	__antithesis_instrumentation__.Notify(641645)
	cfg.MustExist = true
	return nil
}

var DisableAutomaticCompactions ConfigOption = func(cfg *engineConfig) error {
	__antithesis_instrumentation__.Notify(641646)
	cfg.Opts.DisableAutomaticCompactions = true
	return nil
}

var ForTesting ConfigOption = func(cfg *engineConfig) error {
	__antithesis_instrumentation__.Notify(641647)
	if cfg.Settings == nil {
		__antithesis_instrumentation__.Notify(641649)
		cfg.Settings = cluster.MakeTestingClusterSettings()
	} else {
		__antithesis_instrumentation__.Notify(641650)
	}
	__antithesis_instrumentation__.Notify(641648)
	return nil
}

var ForStickyEngineTesting ConfigOption = func(cfg *engineConfig) error {
	__antithesis_instrumentation__.Notify(641651)
	if cfg.Settings == nil {
		__antithesis_instrumentation__.Notify(641653)
		cfg.Settings = cluster.MakeTestingClusterSettings()
	} else {
		__antithesis_instrumentation__.Notify(641654)
	}
	__antithesis_instrumentation__.Notify(641652)
	return nil
}

func Attributes(attrs roachpb.Attributes) ConfigOption {
	__antithesis_instrumentation__.Notify(641655)
	return func(cfg *engineConfig) error {
		__antithesis_instrumentation__.Notify(641656)
		cfg.Attrs = attrs
		return nil
	}
}

func MaxSize(size int64) ConfigOption {
	__antithesis_instrumentation__.Notify(641657)
	return func(cfg *engineConfig) error {
		__antithesis_instrumentation__.Notify(641658)
		cfg.MaxSize = size
		return nil
	}
}

func MaxOpenFiles(count int) ConfigOption {
	__antithesis_instrumentation__.Notify(641659)
	return func(cfg *engineConfig) error {
		__antithesis_instrumentation__.Notify(641660)
		cfg.Opts.MaxOpenFiles = count
		return nil
	}

}

func Settings(settings *cluster.Settings) ConfigOption {
	__antithesis_instrumentation__.Notify(641661)
	return func(cfg *engineConfig) error {
		__antithesis_instrumentation__.Notify(641662)
		cfg.Settings = settings
		return nil
	}
}

func CacheSize(size int64) ConfigOption {
	__antithesis_instrumentation__.Notify(641663)
	return func(cfg *engineConfig) error {
		__antithesis_instrumentation__.Notify(641664)
		cfg.cacheSize = &size
		return nil
	}
}

func EncryptionAtRest(encryptionOptions []byte) ConfigOption {
	__antithesis_instrumentation__.Notify(641665)
	return func(cfg *engineConfig) error {
		__antithesis_instrumentation__.Notify(641666)
		if len(encryptionOptions) > 0 {
			__antithesis_instrumentation__.Notify(641668)
			cfg.UseFileRegistry = true
			cfg.EncryptionOptions = encryptionOptions
		} else {
			__antithesis_instrumentation__.Notify(641669)
		}
		__antithesis_instrumentation__.Notify(641667)
		return nil
	}
}

func Hook(hookFunc func(*base.StorageConfig) error) ConfigOption {
	__antithesis_instrumentation__.Notify(641670)
	return func(cfg *engineConfig) error {
		__antithesis_instrumentation__.Notify(641671)
		if hookFunc == nil {
			__antithesis_instrumentation__.Notify(641673)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(641674)
		}
		__antithesis_instrumentation__.Notify(641672)
		return hookFunc(&cfg.PebbleConfig.StorageConfig)
	}
}

type Location struct {
	dir string
	fs  vfs.FS
}

func Filesystem(dir string) Location {
	__antithesis_instrumentation__.Notify(641675)
	return Location{
		dir: dir,
	}
}

func InMemory() Location {
	__antithesis_instrumentation__.Notify(641676)
	return Location{
		dir: "",
		fs:  vfs.NewMem(),
	}
}

type engineConfig struct {
	PebbleConfig

	cacheSize *int64
}

func Open(ctx context.Context, loc Location, opts ...ConfigOption) (*Pebble, error) {
	__antithesis_instrumentation__.Notify(641677)
	var cfg engineConfig
	cfg.Dir = loc.dir
	cfg.Opts = DefaultPebbleOptions()
	if loc.fs != nil {
		__antithesis_instrumentation__.Notify(641684)
		cfg.Opts.FS = loc.fs
	} else {
		__antithesis_instrumentation__.Notify(641685)
	}
	__antithesis_instrumentation__.Notify(641678)
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(641686)
		if err := opt(&cfg); err != nil {
			__antithesis_instrumentation__.Notify(641687)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(641688)
		}
	}
	__antithesis_instrumentation__.Notify(641679)
	if cfg.cacheSize != nil {
		__antithesis_instrumentation__.Notify(641689)
		cfg.Opts.Cache = pebble.NewCache(*cfg.cacheSize)
		defer cfg.Opts.Cache.Unref()
	} else {
		__antithesis_instrumentation__.Notify(641690)
	}
	__antithesis_instrumentation__.Notify(641680)
	if cfg.Settings == nil {
		__antithesis_instrumentation__.Notify(641691)
		cfg.Settings = cluster.MakeClusterSettings()
	} else {
		__antithesis_instrumentation__.Notify(641692)
	}
	__antithesis_instrumentation__.Notify(641681)
	p, err := NewPebble(ctx, cfg.PebbleConfig)
	if err != nil {
		__antithesis_instrumentation__.Notify(641693)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(641694)
	}
	__antithesis_instrumentation__.Notify(641682)

	if v := p.settings.Version.ActiveVersionOrEmpty(ctx).Version; v != (roachpb.Version{}) {
		__antithesis_instrumentation__.Notify(641695)
		if err := p.SetMinVersion(v); err != nil {
			__antithesis_instrumentation__.Notify(641696)
			p.Close()
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(641697)
		}
	} else {
		__antithesis_instrumentation__.Notify(641698)
	}
	__antithesis_instrumentation__.Notify(641683)
	return p, nil
}
