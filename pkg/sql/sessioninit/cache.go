package sessioninit

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"unsafe"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/cockroach/pkg/settings/cluster"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/mon"
	"github.com/cockroachdb/cockroach/pkg/util/stop"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil/singleflight"
	"github.com/cockroachdb/logtags"
)

var CacheEnabledSettingName = "server.authentication_cache.enabled"

var CacheEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	CacheEnabledSettingName,
	"enables a cache used during authentication to avoid lookups to system tables "+
		"when retrieving per-user authentication-related information",
	true,
).WithPublic()

type Cache struct {
	syncutil.Mutex
	usersTableVersion          descpb.DescriptorVersion
	roleOptionsTableVersion    descpb.DescriptorVersion
	dbRoleSettingsTableVersion descpb.DescriptorVersion
	boundAccount               mon.BoundAccount

	authInfoCache map[security.SQLUsername]AuthInfo

	settingsCache map[SettingsCacheKey][]string

	populateCacheGroup singleflight.Group
	stopper            *stop.Stopper
}

type AuthInfo struct {
	UserExists bool

	CanLoginSQL bool

	CanLoginDBConsole bool

	HashedPassword security.PasswordHash

	ValidUntil *tree.DTimestamp
}

type SettingsCacheKey struct {
	DatabaseID descpb.ID
	Username   security.SQLUsername
}

type SettingsCacheEntry struct {
	SettingsCacheKey
	Settings []string
}

func NewCache(account mon.BoundAccount, stopper *stop.Stopper) *Cache {
	__antithesis_instrumentation__.Notify(621583)
	return &Cache{
		boundAccount: account,
		stopper:      stopper,
	}
}

func (a *Cache) GetAuthInfo(
	ctx context.Context,
	settings *cluster.Settings,
	ie sqlutil.InternalExecutor,
	db *kv.DB,
	f *descs.CollectionFactory,
	username security.SQLUsername,
	readFromSystemTables func(
		ctx context.Context,
		ie sqlutil.InternalExecutor,
		username security.SQLUsername,
	) (AuthInfo, error),
) (aInfo AuthInfo, err error) {
	__antithesis_instrumentation__.Notify(621584)
	if !CacheEnabled.Get(&settings.SV) {
		__antithesis_instrumentation__.Notify(621591)
		return readFromSystemTables(ctx, ie, username)
	} else {
		__antithesis_instrumentation__.Notify(621592)
	}
	__antithesis_instrumentation__.Notify(621585)

	var usersTableDesc catalog.TableDescriptor
	var roleOptionsTableDesc catalog.TableDescriptor
	err = f.Txn(ctx, ie, db, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(621593)
		_, usersTableDesc, err = descriptors.GetImmutableTableByName(
			ctx,
			txn,
			UsersTableName,
			tree.ObjectLookupFlagsWithRequired(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(621595)
			return err
		} else {
			__antithesis_instrumentation__.Notify(621596)
		}
		__antithesis_instrumentation__.Notify(621594)
		_, roleOptionsTableDesc, err = descriptors.GetImmutableTableByName(
			ctx,
			txn,
			RoleOptionsTableName,
			tree.ObjectLookupFlagsWithRequired(),
		)
		return err
	})
	__antithesis_instrumentation__.Notify(621586)
	if err != nil {
		__antithesis_instrumentation__.Notify(621597)
		return AuthInfo{}, err
	} else {
		__antithesis_instrumentation__.Notify(621598)
	}
	__antithesis_instrumentation__.Notify(621587)

	usersTableVersion := usersTableDesc.GetVersion()
	roleOptionsTableVersion := roleOptionsTableDesc.GetVersion()

	var found bool
	aInfo, found = a.readAuthInfoFromCache(ctx, usersTableVersion, roleOptionsTableVersion, username)

	if found {
		__antithesis_instrumentation__.Notify(621599)
		return aInfo, nil
	} else {
		__antithesis_instrumentation__.Notify(621600)
	}
	__antithesis_instrumentation__.Notify(621588)

	val, err := a.loadValueOutsideOfCache(
		ctx, fmt.Sprintf("authinfo-%s-%d-%d", username.Normalized(), usersTableVersion, roleOptionsTableVersion),
		func(loadCtx context.Context) (interface{}, error) {
			__antithesis_instrumentation__.Notify(621601)
			return readFromSystemTables(loadCtx, ie, username)
		})
	__antithesis_instrumentation__.Notify(621589)
	if err != nil {
		__antithesis_instrumentation__.Notify(621602)
		return aInfo, err
	} else {
		__antithesis_instrumentation__.Notify(621603)
	}
	__antithesis_instrumentation__.Notify(621590)
	aInfo = val.(AuthInfo)

	a.maybeWriteAuthInfoBackToCache(
		ctx,
		usersTableVersion,
		roleOptionsTableVersion,
		aInfo,
		username,
	)

	return aInfo, err
}

func (a *Cache) readAuthInfoFromCache(
	ctx context.Context,
	usersTableVersion descpb.DescriptorVersion,
	roleOptionsTableVersion descpb.DescriptorVersion,
	username security.SQLUsername,
) (AuthInfo, bool) {
	__antithesis_instrumentation__.Notify(621604)
	a.Lock()
	defer a.Unlock()

	isEligibleForCache := a.clearCacheIfStale(ctx, usersTableVersion, roleOptionsTableVersion, a.dbRoleSettingsTableVersion)
	if !isEligibleForCache {
		__antithesis_instrumentation__.Notify(621606)
		return AuthInfo{}, false
	} else {
		__antithesis_instrumentation__.Notify(621607)
	}
	__antithesis_instrumentation__.Notify(621605)
	ai, foundAuthInfo := a.authInfoCache[username]
	return ai, foundAuthInfo
}

func (a *Cache) loadValueOutsideOfCache(
	ctx context.Context, requestKey string, fn func(loadCtx context.Context) (interface{}, error),
) (interface{}, error) {
	__antithesis_instrumentation__.Notify(621608)
	ch, _ := a.populateCacheGroup.DoChan(requestKey, func() (interface{}, error) {
		__antithesis_instrumentation__.Notify(621610)

		loadCtx, cancel := a.stopper.WithCancelOnQuiesce(
			logtags.WithTags(context.Background(), logtags.FromContext(ctx)),
		)
		defer cancel()
		return fn(loadCtx)
	})
	__antithesis_instrumentation__.Notify(621609)
	select {
	case res := <-ch:
		__antithesis_instrumentation__.Notify(621611)
		if res.Err != nil {
			__antithesis_instrumentation__.Notify(621614)
			return AuthInfo{}, res.Err
		} else {
			__antithesis_instrumentation__.Notify(621615)
		}
		__antithesis_instrumentation__.Notify(621612)
		return res.Val, nil
	case <-ctx.Done():
		__antithesis_instrumentation__.Notify(621613)
		return AuthInfo{}, ctx.Err()
	}
}

func (a *Cache) maybeWriteAuthInfoBackToCache(
	ctx context.Context,
	usersTableVersion descpb.DescriptorVersion,
	roleOptionsTableVersion descpb.DescriptorVersion,
	aInfo AuthInfo,
	username security.SQLUsername,
) bool {
	__antithesis_instrumentation__.Notify(621616)
	a.Lock()
	defer a.Unlock()

	if a.usersTableVersion != usersTableVersion || func() bool {
		__antithesis_instrumentation__.Notify(621620)
		return a.roleOptionsTableVersion != roleOptionsTableVersion == true
	}() == true {
		__antithesis_instrumentation__.Notify(621621)
		return false
	} else {
		__antithesis_instrumentation__.Notify(621622)
	}
	__antithesis_instrumentation__.Notify(621617)

	const sizeOfUsername = int(unsafe.Sizeof(security.SQLUsername{}))
	const sizeOfAuthInfo = int(unsafe.Sizeof(AuthInfo{}))
	const sizeOfTimestamp = int(unsafe.Sizeof(tree.DTimestamp{}))

	hpSize := 0
	if aInfo.HashedPassword != nil {
		__antithesis_instrumentation__.Notify(621623)
		hpSize = aInfo.HashedPassword.Size()
	} else {
		__antithesis_instrumentation__.Notify(621624)
	}
	__antithesis_instrumentation__.Notify(621618)

	sizeOfEntry := sizeOfUsername + len(username.Normalized()) +
		sizeOfAuthInfo + hpSize +
		sizeOfTimestamp
	if err := a.boundAccount.Grow(ctx, int64(sizeOfEntry)); err != nil {
		__antithesis_instrumentation__.Notify(621625)

		log.Ops.Warningf(ctx, "no memory available to cache authentication info: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(621626)
		a.authInfoCache[username] = aInfo
	}
	__antithesis_instrumentation__.Notify(621619)
	return true
}

func (a *Cache) GetDefaultSettings(
	ctx context.Context,
	settings *cluster.Settings,
	ie sqlutil.InternalExecutor,
	db *kv.DB,
	f *descs.CollectionFactory,
	username security.SQLUsername,
	databaseName string,
	readFromSystemTables func(
		ctx context.Context,
		ie sqlutil.InternalExecutor,
		username security.SQLUsername,
		databaseID descpb.ID,
	) ([]SettingsCacheEntry, error),
) (settingsEntries []SettingsCacheEntry, err error) {
	__antithesis_instrumentation__.Notify(621627)
	var dbRoleSettingsTableDesc catalog.TableDescriptor
	var databaseID descpb.ID
	err = f.Txn(ctx, ie, db, func(
		ctx context.Context, txn *kv.Txn, descriptors *descs.Collection,
	) error {
		__antithesis_instrumentation__.Notify(621634)
		_, dbRoleSettingsTableDesc, err = descriptors.GetImmutableTableByName(
			ctx,
			txn,
			DatabaseRoleSettingsTableName,
			tree.ObjectLookupFlagsWithRequired(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(621637)
			return err
		} else {
			__antithesis_instrumentation__.Notify(621638)
		}
		__antithesis_instrumentation__.Notify(621635)
		databaseID = descpb.ID(0)
		if databaseName != "" {
			__antithesis_instrumentation__.Notify(621639)
			dbDesc, err := descriptors.GetImmutableDatabaseByName(ctx, txn, databaseName, tree.DatabaseLookupFlags{})
			if err != nil {
				__antithesis_instrumentation__.Notify(621641)
				return err
			} else {
				__antithesis_instrumentation__.Notify(621642)
			}
			__antithesis_instrumentation__.Notify(621640)

			if dbDesc != nil {
				__antithesis_instrumentation__.Notify(621643)
				databaseID = dbDesc.GetID()
			} else {
				__antithesis_instrumentation__.Notify(621644)
			}
		} else {
			__antithesis_instrumentation__.Notify(621645)
		}
		__antithesis_instrumentation__.Notify(621636)
		return nil
	})
	__antithesis_instrumentation__.Notify(621628)
	if err != nil {
		__antithesis_instrumentation__.Notify(621646)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(621647)
	}
	__antithesis_instrumentation__.Notify(621629)

	if !CacheEnabled.Get(&settings.SV) {
		__antithesis_instrumentation__.Notify(621648)
		settingsEntries, err = readFromSystemTables(
			ctx,
			ie,
			username,
			databaseID,
		)
		return settingsEntries, err
	} else {
		__antithesis_instrumentation__.Notify(621649)
	}
	__antithesis_instrumentation__.Notify(621630)

	dbRoleSettingsTableVersion := dbRoleSettingsTableDesc.GetVersion()

	var found bool
	settingsEntries, found = a.readDefaultSettingsFromCache(ctx, dbRoleSettingsTableVersion, username, databaseID)

	if found {
		__antithesis_instrumentation__.Notify(621650)
		return settingsEntries, nil
	} else {
		__antithesis_instrumentation__.Notify(621651)
	}
	__antithesis_instrumentation__.Notify(621631)

	val, err := a.loadValueOutsideOfCache(
		ctx, fmt.Sprintf("defaultsettings-%s-%d-%d", username.Normalized(), databaseID, dbRoleSettingsTableVersion),
		func(loadCtx context.Context) (interface{}, error) {
			__antithesis_instrumentation__.Notify(621652)
			return readFromSystemTables(loadCtx, ie, username, databaseID)
		},
	)
	__antithesis_instrumentation__.Notify(621632)
	if err != nil {
		__antithesis_instrumentation__.Notify(621653)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(621654)
	}
	__antithesis_instrumentation__.Notify(621633)
	settingsEntries = val.([]SettingsCacheEntry)

	a.maybeWriteDefaultSettingsBackToCache(
		ctx,
		dbRoleSettingsTableVersion,
		settingsEntries,
	)
	return settingsEntries, err
}

func (a *Cache) readDefaultSettingsFromCache(
	ctx context.Context,
	dbRoleSettingsTableVersion descpb.DescriptorVersion,
	username security.SQLUsername,
	databaseID descpb.ID,
) ([]SettingsCacheEntry, bool) {
	__antithesis_instrumentation__.Notify(621655)
	a.Lock()
	defer a.Unlock()

	isEligibleForCache := a.clearCacheIfStale(
		ctx, a.usersTableVersion, a.roleOptionsTableVersion, dbRoleSettingsTableVersion,
	)
	if !isEligibleForCache {
		__antithesis_instrumentation__.Notify(621658)
		return nil, false
	} else {
		__antithesis_instrumentation__.Notify(621659)
	}
	__antithesis_instrumentation__.Notify(621656)
	foundAllDefaultSettings := true
	var sEntries []SettingsCacheEntry

	for _, k := range GenerateSettingsCacheKeys(databaseID, username) {
		__antithesis_instrumentation__.Notify(621660)
		s, ok := a.settingsCache[k]
		if !ok {
			__antithesis_instrumentation__.Notify(621662)
			foundAllDefaultSettings = false
			break
		} else {
			__antithesis_instrumentation__.Notify(621663)
		}
		__antithesis_instrumentation__.Notify(621661)
		sEntries = append(sEntries, SettingsCacheEntry{k, s})
	}
	__antithesis_instrumentation__.Notify(621657)
	return sEntries, foundAllDefaultSettings
}

func (a *Cache) maybeWriteDefaultSettingsBackToCache(
	ctx context.Context,
	dbRoleSettingsTableVersion descpb.DescriptorVersion,
	settingsEntries []SettingsCacheEntry,
) bool {
	__antithesis_instrumentation__.Notify(621664)
	a.Lock()
	defer a.Unlock()

	if a.dbRoleSettingsTableVersion > dbRoleSettingsTableVersion {
		__antithesis_instrumentation__.Notify(621668)
		return false
	} else {
		__antithesis_instrumentation__.Notify(621669)
	}
	__antithesis_instrumentation__.Notify(621665)

	const sizeOfSettingsCacheEntry = int(unsafe.Sizeof(SettingsCacheEntry{}))
	sizeOfSettings := 0
	for _, sEntry := range settingsEntries {
		__antithesis_instrumentation__.Notify(621670)
		if _, ok := a.settingsCache[sEntry.SettingsCacheKey]; ok {
			__antithesis_instrumentation__.Notify(621672)

			continue
		} else {
			__antithesis_instrumentation__.Notify(621673)
		}
		__antithesis_instrumentation__.Notify(621671)
		sizeOfSettings += sizeOfSettingsCacheEntry
		sizeOfSettings += len(sEntry.SettingsCacheKey.Username.Normalized())
		for _, s := range sEntry.Settings {
			__antithesis_instrumentation__.Notify(621674)
			sizeOfSettings += len(s)
		}
	}
	__antithesis_instrumentation__.Notify(621666)
	if err := a.boundAccount.Grow(ctx, int64(sizeOfSettings)); err != nil {
		__antithesis_instrumentation__.Notify(621675)

		log.Ops.Warningf(ctx, "no memory available to cache authentication info: %v", err)
	} else {
		__antithesis_instrumentation__.Notify(621676)
		for _, sEntry := range settingsEntries {
			__antithesis_instrumentation__.Notify(621677)

			if _, ok := a.settingsCache[sEntry.SettingsCacheKey]; !ok {
				__antithesis_instrumentation__.Notify(621678)
				a.settingsCache[sEntry.SettingsCacheKey] = sEntry.Settings
			} else {
				__antithesis_instrumentation__.Notify(621679)
			}
		}
	}
	__antithesis_instrumentation__.Notify(621667)
	return true
}

func (a *Cache) clearCacheIfStale(
	ctx context.Context,
	usersTableVersion descpb.DescriptorVersion,
	roleOptionsTableVersion descpb.DescriptorVersion,
	dbRoleSettingsTableVersion descpb.DescriptorVersion,
) (isEligibleForCache bool) {
	__antithesis_instrumentation__.Notify(621680)
	if a.usersTableVersion < usersTableVersion || func() bool {
		__antithesis_instrumentation__.Notify(621682)
		return a.roleOptionsTableVersion < roleOptionsTableVersion == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(621683)
		return a.dbRoleSettingsTableVersion < dbRoleSettingsTableVersion == true
	}() == true {
		__antithesis_instrumentation__.Notify(621684)

		a.usersTableVersion = usersTableVersion
		a.roleOptionsTableVersion = roleOptionsTableVersion
		a.dbRoleSettingsTableVersion = dbRoleSettingsTableVersion
		a.authInfoCache = make(map[security.SQLUsername]AuthInfo)
		a.settingsCache = make(map[SettingsCacheKey][]string)
		a.boundAccount.Empty(ctx)
	} else {
		__antithesis_instrumentation__.Notify(621685)
		if a.usersTableVersion > usersTableVersion || func() bool {
			__antithesis_instrumentation__.Notify(621686)
			return a.roleOptionsTableVersion > roleOptionsTableVersion == true
		}() == true || func() bool {
			__antithesis_instrumentation__.Notify(621687)
			return a.dbRoleSettingsTableVersion > dbRoleSettingsTableVersion == true
		}() == true {
			__antithesis_instrumentation__.Notify(621688)

			return false
		} else {
			__antithesis_instrumentation__.Notify(621689)
		}
	}
	__antithesis_instrumentation__.Notify(621681)
	return true
}

func GenerateSettingsCacheKeys(
	databaseID descpb.ID, username security.SQLUsername,
) []SettingsCacheKey {
	__antithesis_instrumentation__.Notify(621690)
	return []SettingsCacheKey{
		{
			DatabaseID: databaseID,
			Username:   username,
		},
		{
			DatabaseID: defaultDatabaseID,
			Username:   username,
		},
		{
			DatabaseID: databaseID,
			Username:   defaultUsername,
		},
		{
			DatabaseID: defaultDatabaseID,
			Username:   defaultUsername,
		},
	}
}
