package backupccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"sort"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/execinfrapb"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

type prefixRewrite struct {
	OldPrefix []byte
	NewPrefix []byte
	noop      bool
}

type prefixRewriter struct {
	rewrites []prefixRewrite
	last     int
}

func (p prefixRewriter) rewriteKey(key []byte) ([]byte, bool) {
	__antithesis_instrumentation__.Notify(9769)
	if len(p.rewrites) < 1 {
		__antithesis_instrumentation__.Notify(9774)
		return key, false
	} else {
		__antithesis_instrumentation__.Notify(9775)
	}
	__antithesis_instrumentation__.Notify(9770)

	found := p.last
	if !bytes.HasPrefix(key, p.rewrites[found].OldPrefix) {
		__antithesis_instrumentation__.Notify(9776)

		found = sort.Search(len(p.rewrites), func(i int) bool {
			__antithesis_instrumentation__.Notify(9778)
			return bytes.HasPrefix(key, p.rewrites[i].OldPrefix) || func() bool {
				__antithesis_instrumentation__.Notify(9779)
				return bytes.Compare(key, p.rewrites[i].OldPrefix) < 0 == true
			}() == true
		})
		__antithesis_instrumentation__.Notify(9777)
		if found == len(p.rewrites) || func() bool {
			__antithesis_instrumentation__.Notify(9780)
			return !bytes.HasPrefix(key, p.rewrites[found].OldPrefix) == true
		}() == true {
			__antithesis_instrumentation__.Notify(9781)
			return key, false
		} else {
			__antithesis_instrumentation__.Notify(9782)
		}
	} else {
		__antithesis_instrumentation__.Notify(9783)
	}
	__antithesis_instrumentation__.Notify(9771)

	p.last = found
	rewrite := p.rewrites[found]
	if rewrite.noop {
		__antithesis_instrumentation__.Notify(9784)
		return key, true
	} else {
		__antithesis_instrumentation__.Notify(9785)
	}
	__antithesis_instrumentation__.Notify(9772)
	if len(rewrite.OldPrefix) == len(rewrite.NewPrefix) {
		__antithesis_instrumentation__.Notify(9786)
		copy(key[:len(rewrite.OldPrefix)], rewrite.NewPrefix)
		return key, true
	} else {
		__antithesis_instrumentation__.Notify(9787)
	}
	__antithesis_instrumentation__.Notify(9773)

	newKey := make([]byte, 0, len(rewrite.NewPrefix)+len(key)-len(rewrite.OldPrefix))
	newKey = append(newKey, rewrite.NewPrefix...)
	newKey = append(newKey, key[len(rewrite.OldPrefix):]...)
	return newKey, true
}

type KeyRewriter struct {
	codec keys.SQLCodec

	lastKeyTenant struct {
		id     roachpb.TenantID
		prefix []byte
	}

	fromSystemTenant bool

	prefixes prefixRewriter
	tenants  prefixRewriter
	descs    map[descpb.ID]catalog.TableDescriptor
}

func makeKeyRewriterFromRekeys(
	codec keys.SQLCodec, tableRekeys []execinfrapb.TableRekey, tenantRekeys []execinfrapb.TenantRekey,
) (*KeyRewriter, error) {
	__antithesis_instrumentation__.Notify(9788)
	descs := make(map[descpb.ID]catalog.TableDescriptor)
	for _, rekey := range tableRekeys {
		__antithesis_instrumentation__.Notify(9790)

		if rekey.OldID == 0 {
			__antithesis_instrumentation__.Notify(9794)
			continue
		} else {
			__antithesis_instrumentation__.Notify(9795)
		}
		__antithesis_instrumentation__.Notify(9791)
		var desc descpb.Descriptor
		if err := protoutil.Unmarshal(rekey.NewDesc, &desc); err != nil {
			__antithesis_instrumentation__.Notify(9796)
			return nil, errors.Wrapf(err, "unmarshalling rekey descriptor for old table id %d", rekey.OldID)
		} else {
			__antithesis_instrumentation__.Notify(9797)
		}
		__antithesis_instrumentation__.Notify(9792)
		table, _, _, _ := descpb.FromDescriptor(&desc)
		if table == nil {
			__antithesis_instrumentation__.Notify(9798)
			return nil, errors.New("expected a table descriptor")
		} else {
			__antithesis_instrumentation__.Notify(9799)
		}
		__antithesis_instrumentation__.Notify(9793)
		descs[descpb.ID(rekey.OldID)] = tabledesc.NewBuilder(table).BuildImmutableTable()
	}
	__antithesis_instrumentation__.Notify(9789)

	return makeKeyRewriter(codec, descs, tenantRekeys)
}

var (
	isBackupFromSystemTenantRekey = execinfrapb.TenantRekey{
		OldID: roachpb.SystemTenantID,
		NewID: roachpb.SystemTenantID,
	}
)

func makeKeyRewriter(
	codec keys.SQLCodec,
	descs map[descpb.ID]catalog.TableDescriptor,
	tenants []execinfrapb.TenantRekey,
) (*KeyRewriter, error) {
	__antithesis_instrumentation__.Notify(9800)
	var prefixes prefixRewriter
	var tenantPrefixes prefixRewriter
	tenantPrefixes.rewrites = make([]prefixRewrite, 0, len(tenants))

	seenPrefixes := make(map[string]bool)
	for oldID, desc := range descs {
		__antithesis_instrumentation__.Notify(9805)

		for _, index := range desc.NonDropIndexes() {
			__antithesis_instrumentation__.Notify(9806)
			oldPrefix := roachpb.Key(MakeKeyRewriterPrefixIgnoringInterleaved(oldID, index.GetID()))
			newPrefix := roachpb.Key(MakeKeyRewriterPrefixIgnoringInterleaved(desc.GetID(), index.GetID()))
			if !seenPrefixes[string(oldPrefix)] {
				__antithesis_instrumentation__.Notify(9808)
				seenPrefixes[string(oldPrefix)] = true
				prefixes.rewrites = append(prefixes.rewrites, prefixRewrite{
					OldPrefix: oldPrefix,
					NewPrefix: newPrefix,
					noop:      bytes.Equal(oldPrefix, newPrefix),
				})
			} else {
				__antithesis_instrumentation__.Notify(9809)
			}
			__antithesis_instrumentation__.Notify(9807)

			oldPrefix = oldPrefix.PrefixEnd()
			newPrefix = newPrefix.PrefixEnd()
			if !seenPrefixes[string(oldPrefix)] {
				__antithesis_instrumentation__.Notify(9810)
				seenPrefixes[string(oldPrefix)] = true
				prefixes.rewrites = append(prefixes.rewrites, prefixRewrite{
					OldPrefix: oldPrefix,
					NewPrefix: newPrefix,
					noop:      bytes.Equal(oldPrefix, newPrefix),
				})
			} else {
				__antithesis_instrumentation__.Notify(9811)
			}
		}
	}
	__antithesis_instrumentation__.Notify(9801)
	sort.Slice(prefixes.rewrites, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(9812)
		return bytes.Compare(prefixes.rewrites[i].OldPrefix, prefixes.rewrites[j].OldPrefix) < 0
	})
	__antithesis_instrumentation__.Notify(9802)
	fromSystemTenant := false
	for i := range tenants {
		__antithesis_instrumentation__.Notify(9813)
		if tenants[i] == isBackupFromSystemTenantRekey {
			__antithesis_instrumentation__.Notify(9815)
			fromSystemTenant = true
			continue
		} else {
			__antithesis_instrumentation__.Notify(9816)
		}
		__antithesis_instrumentation__.Notify(9814)
		from, to := keys.MakeSQLCodec(tenants[i].OldID).TenantPrefix(), keys.MakeSQLCodec(tenants[i].NewID).TenantPrefix()
		tenantPrefixes.rewrites = append(tenantPrefixes.rewrites, prefixRewrite{
			OldPrefix: from, NewPrefix: to, noop: bytes.Equal(from, to),
		})
	}
	__antithesis_instrumentation__.Notify(9803)
	sort.Slice(tenantPrefixes.rewrites, func(i, j int) bool {
		__antithesis_instrumentation__.Notify(9817)
		return bytes.Compare(tenantPrefixes.rewrites[i].OldPrefix, tenantPrefixes.rewrites[j].OldPrefix) < 0
	})
	__antithesis_instrumentation__.Notify(9804)
	return &KeyRewriter{
		codec:            codec,
		prefixes:         prefixes,
		descs:            descs,
		tenants:          tenantPrefixes,
		fromSystemTenant: fromSystemTenant,
	}, nil
}

func MakeKeyRewriterPrefixIgnoringInterleaved(tableID descpb.ID, indexID descpb.IndexID) []byte {
	__antithesis_instrumentation__.Notify(9818)
	return keys.SystemSQLCodec.IndexPrefix(uint32(tableID), uint32(indexID))
}

func (kr *KeyRewriter) RewriteKey(key []byte) ([]byte, bool, error) {
	__antithesis_instrumentation__.Notify(9819)

	if kr.fromSystemTenant && func() bool {
		__antithesis_instrumentation__.Notify(9825)
		return bytes.HasPrefix(key, keys.TenantPrefix) == true
	}() == true {
		__antithesis_instrumentation__.Notify(9826)
		k, ok := kr.tenants.rewriteKey(key)
		return k, ok, nil
	} else {
		__antithesis_instrumentation__.Notify(9827)
	}
	__antithesis_instrumentation__.Notify(9820)

	noTenantPrefix, oldTenantID, err := keys.DecodeTenantPrefix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(9828)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(9829)
	}
	__antithesis_instrumentation__.Notify(9821)

	rekeyed, ok, err := kr.rewriteTableKey(noTenantPrefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(9830)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(9831)
	}
	__antithesis_instrumentation__.Notify(9822)

	if kr.lastKeyTenant.id != oldTenantID {
		__antithesis_instrumentation__.Notify(9832)
		kr.lastKeyTenant.id = oldTenantID
		kr.lastKeyTenant.prefix = keys.MakeSQLCodec(oldTenantID).TenantPrefix()
	} else {
		__antithesis_instrumentation__.Notify(9833)
	}
	__antithesis_instrumentation__.Notify(9823)

	newTenantPrefix := kr.codec.TenantPrefix()
	if len(newTenantPrefix) == len(kr.lastKeyTenant.prefix) {
		__antithesis_instrumentation__.Notify(9834)
		keyTenantPrefix := key[:len(kr.lastKeyTenant.prefix)]
		copy(keyTenantPrefix, newTenantPrefix)
		rekeyed = append(keyTenantPrefix, rekeyed...)
	} else {
		__antithesis_instrumentation__.Notify(9835)
		rekeyed = append(newTenantPrefix, rekeyed...)
	}
	__antithesis_instrumentation__.Notify(9824)

	return rekeyed, ok, err
}

func (kr *KeyRewriter) rewriteTableKey(key []byte) ([]byte, bool, error) {
	__antithesis_instrumentation__.Notify(9836)

	_, tableID, _ := keys.SystemSQLCodec.DecodeTablePrefix(key)

	key, ok := kr.prefixes.rewriteKey(key)
	if !ok {
		__antithesis_instrumentation__.Notify(9839)
		return nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(9840)
	}
	__antithesis_instrumentation__.Notify(9837)
	desc := kr.descs[descpb.ID(tableID)]
	if desc == nil {
		__antithesis_instrumentation__.Notify(9841)
		return nil, false, errors.Errorf("missing descriptor for table %d", tableID)
	} else {
		__antithesis_instrumentation__.Notify(9842)
	}
	__antithesis_instrumentation__.Notify(9838)
	return key, true, nil
}
