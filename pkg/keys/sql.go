package keys

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"math"

	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
)

func MakeTenantPrefix(tenID roachpb.TenantID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85831)
	if tenID == roachpb.SystemTenantID {
		__antithesis_instrumentation__.Notify(85833)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(85834)
	}
	__antithesis_instrumentation__.Notify(85832)
	return encoding.EncodeUvarintAscending(TenantPrefix, tenID.ToUint64())
}

func DecodeTenantPrefix(key roachpb.Key) ([]byte, roachpb.TenantID, error) {
	__antithesis_instrumentation__.Notify(85835)
	if len(key) == 0 {
		__antithesis_instrumentation__.Notify(85839)
		return nil, roachpb.SystemTenantID, nil
	} else {
		__antithesis_instrumentation__.Notify(85840)
	}
	__antithesis_instrumentation__.Notify(85836)
	if key[0] != tenantPrefixByte {
		__antithesis_instrumentation__.Notify(85841)
		return key, roachpb.SystemTenantID, nil
	} else {
		__antithesis_instrumentation__.Notify(85842)
	}
	__antithesis_instrumentation__.Notify(85837)
	rem, tenID, err := encoding.DecodeUvarintAscending(key[1:])
	if err != nil {
		__antithesis_instrumentation__.Notify(85843)
		return nil, roachpb.TenantID{}, err
	} else {
		__antithesis_instrumentation__.Notify(85844)
	}
	__antithesis_instrumentation__.Notify(85838)
	return rem, roachpb.MakeTenantID(tenID), nil
}

type SQLCodec struct {
	sqlEncoder
	sqlDecoder
	_ func()
}

type sqlEncoder struct {
	buf *roachpb.Key
}

type sqlDecoder struct {
	buf *roachpb.Key
}

func MakeSQLCodec(tenID roachpb.TenantID) SQLCodec {
	__antithesis_instrumentation__.Notify(85845)
	k := MakeTenantPrefix(tenID)
	k = k[:len(k):len(k)]
	return SQLCodec{
		sqlEncoder: sqlEncoder{&k},
		sqlDecoder: sqlDecoder{&k},
	}
}

var SystemSQLCodec = MakeSQLCodec(roachpb.SystemTenantID)

var TODOSQLCodec = MakeSQLCodec(roachpb.SystemTenantID)

func (e sqlEncoder) ForSystemTenant() bool {
	__antithesis_instrumentation__.Notify(85846)
	return len(e.TenantPrefix()) == 0
}

func (e sqlEncoder) TenantPrefix() roachpb.Key {
	__antithesis_instrumentation__.Notify(85847)
	return *e.buf
}

func (e sqlEncoder) TablePrefix(tableID uint32) roachpb.Key {
	__antithesis_instrumentation__.Notify(85848)
	k := e.TenantPrefix()
	return encoding.EncodeUvarintAscending(k, uint64(tableID))
}

func (e sqlEncoder) IndexPrefix(tableID, indexID uint32) roachpb.Key {
	__antithesis_instrumentation__.Notify(85849)
	k := e.TablePrefix(tableID)
	return encoding.EncodeUvarintAscending(k, uint64(indexID))
}

func (e sqlEncoder) DescMetadataPrefix() roachpb.Key {
	__antithesis_instrumentation__.Notify(85850)
	return e.IndexPrefix(DescriptorTableID, DescriptorTablePrimaryKeyIndexID)
}

func (e sqlEncoder) DescMetadataKey(descID uint32) roachpb.Key {
	__antithesis_instrumentation__.Notify(85851)
	k := e.DescMetadataPrefix()
	k = encoding.EncodeUvarintAscending(k, uint64(descID))
	return MakeFamilyKey(k, DescriptorTableDescriptorColFamID)
}

func (e sqlEncoder) TenantMetadataKey(tenID roachpb.TenantID) roachpb.Key {
	__antithesis_instrumentation__.Notify(85852)
	k := e.IndexPrefix(TenantsTableID, TenantsTablePrimaryKeyIndexID)
	k = encoding.EncodeUvarintAscending(k, tenID.ToUint64())
	return MakeFamilyKey(k, 0)
}

func (e sqlEncoder) SequenceKey(tableID uint32) roachpb.Key {
	__antithesis_instrumentation__.Notify(85853)
	k := e.IndexPrefix(tableID, SequenceIndexID)
	k = encoding.EncodeUvarintAscending(k, 0)
	k = MakeFamilyKey(k, SequenceColumnFamilyID)
	return k
}

func (e sqlEncoder) DescIDSequenceKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85854)
	if e.ForSystemTenant() {
		__antithesis_instrumentation__.Notify(85856)

		return descIDGenerator
	} else {
		__antithesis_instrumentation__.Notify(85857)
	}
	__antithesis_instrumentation__.Notify(85855)
	return e.SequenceKey(DescIDSequenceID)
}

func (e sqlEncoder) ZoneKeyPrefix(id uint32) roachpb.Key {
	__antithesis_instrumentation__.Notify(85858)
	k := e.IndexPrefix(ZonesTableID, ZonesTablePrimaryIndexID)
	return encoding.EncodeUvarintAscending(k, uint64(id))
}

func (e sqlEncoder) ZoneKey(id uint32) roachpb.Key {
	__antithesis_instrumentation__.Notify(85859)
	k := e.ZoneKeyPrefix(id)
	return MakeFamilyKey(k, uint32(ZonesTableConfigColumnID))
}

func (e sqlEncoder) MigrationKeyPrefix() roachpb.Key {
	__antithesis_instrumentation__.Notify(85860)
	return append(e.TenantPrefix(), MigrationPrefix...)
}

func (e sqlEncoder) MigrationLeaseKey() roachpb.Key {
	__antithesis_instrumentation__.Notify(85861)
	return append(e.TenantPrefix(), MigrationLease...)
}

func (d sqlDecoder) tenantPrefix() roachpb.Key {
	__antithesis_instrumentation__.Notify(85862)
	return *d.buf
}

func (d sqlDecoder) StripTenantPrefix(key roachpb.Key) ([]byte, error) {
	__antithesis_instrumentation__.Notify(85863)
	tenPrefix := d.tenantPrefix()
	if !bytes.HasPrefix(key, tenPrefix) {
		__antithesis_instrumentation__.Notify(85865)
		return nil, errors.Errorf("invalid tenant id prefix: %q", key)
	} else {
		__antithesis_instrumentation__.Notify(85866)
	}
	__antithesis_instrumentation__.Notify(85864)
	return key[len(tenPrefix):], nil
}

func (d sqlDecoder) DecodeTablePrefix(key roachpb.Key) ([]byte, uint32, error) {
	__antithesis_instrumentation__.Notify(85867)
	key, err := d.StripTenantPrefix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85870)
		return nil, 0, err
	} else {
		__antithesis_instrumentation__.Notify(85871)
	}
	__antithesis_instrumentation__.Notify(85868)
	if encoding.PeekType(key) != encoding.Int {
		__antithesis_instrumentation__.Notify(85872)
		return nil, 0, errors.Errorf("invalid key prefix: %q", key)
	} else {
		__antithesis_instrumentation__.Notify(85873)
	}
	__antithesis_instrumentation__.Notify(85869)
	key, tableID, err := encoding.DecodeUvarintAscending(key)
	return key, uint32(tableID), err
}

func (d sqlDecoder) DecodeIndexPrefix(key roachpb.Key) ([]byte, uint32, uint32, error) {
	__antithesis_instrumentation__.Notify(85874)
	key, tableID, err := d.DecodeTablePrefix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85877)
		return nil, 0, 0, err
	} else {
		__antithesis_instrumentation__.Notify(85878)
	}
	__antithesis_instrumentation__.Notify(85875)
	if encoding.PeekType(key) != encoding.Int {
		__antithesis_instrumentation__.Notify(85879)
		return nil, 0, 0, errors.Errorf("invalid key prefix: %q", key)
	} else {
		__antithesis_instrumentation__.Notify(85880)
	}
	__antithesis_instrumentation__.Notify(85876)
	key, indexID, err := encoding.DecodeUvarintAscending(key)
	return key, tableID, uint32(indexID), err
}

func (d sqlDecoder) DecodeDescMetadataID(key roachpb.Key) (uint32, error) {
	__antithesis_instrumentation__.Notify(85881)

	remaining, tableID, _, err := d.DecodeIndexPrefix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85886)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(85887)
	}
	__antithesis_instrumentation__.Notify(85882)
	if tableID != DescriptorTableID {
		__antithesis_instrumentation__.Notify(85888)
		return 0, errors.Errorf("key is not a descriptor table entry: %v", key)
	} else {
		__antithesis_instrumentation__.Notify(85889)
	}
	__antithesis_instrumentation__.Notify(85883)

	_, id, err := encoding.DecodeUvarintAscending(remaining)
	if err != nil {
		__antithesis_instrumentation__.Notify(85890)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(85891)
	}
	__antithesis_instrumentation__.Notify(85884)
	if id > math.MaxUint32 {
		__antithesis_instrumentation__.Notify(85892)
		return 0, errors.Errorf("descriptor ID %d exceeds uint32 bounds", id)
	} else {
		__antithesis_instrumentation__.Notify(85893)
	}
	__antithesis_instrumentation__.Notify(85885)
	return uint32(id), nil
}

func (d sqlDecoder) DecodeTenantMetadataID(key roachpb.Key) (roachpb.TenantID, error) {
	__antithesis_instrumentation__.Notify(85894)

	remaining, tableID, _, err := d.DecodeIndexPrefix(key)
	if err != nil {
		__antithesis_instrumentation__.Notify(85898)
		return roachpb.TenantID{}, err
	} else {
		__antithesis_instrumentation__.Notify(85899)
	}
	__antithesis_instrumentation__.Notify(85895)
	if tableID != TenantsTableID {
		__antithesis_instrumentation__.Notify(85900)
		return roachpb.TenantID{}, errors.Errorf("key is not a tenant table entry: %v", key)
	} else {
		__antithesis_instrumentation__.Notify(85901)
	}
	__antithesis_instrumentation__.Notify(85896)

	_, id, err := encoding.DecodeUvarintAscending(remaining)
	if err != nil {
		__antithesis_instrumentation__.Notify(85902)
		return roachpb.TenantID{}, err
	} else {
		__antithesis_instrumentation__.Notify(85903)
	}
	__antithesis_instrumentation__.Notify(85897)
	return roachpb.MakeTenantID(id), nil
}

func RewriteSpanToTenantPrefix(sp roachpb.Span, prefix roachpb.Key) (roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(85904)
	var err error
	sp.Key, err = rewriteKeyToTenantPrefix(sp.Key, prefix)
	if err != nil {
		__antithesis_instrumentation__.Notify(85906)
		return sp, err
	} else {
		__antithesis_instrumentation__.Notify(85907)
	}
	__antithesis_instrumentation__.Notify(85905)
	sp.EndKey, err = rewriteKeyToTenantPrefix(sp.EndKey, prefix)
	return sp, err
}

func rewriteKeyToTenantPrefix(key roachpb.Key, prefix roachpb.Key) (roachpb.Key, error) {
	__antithesis_instrumentation__.Notify(85908)

	if len(prefix) == 0 && func() bool {
		__antithesis_instrumentation__.Notify(85910)
		return bytes.HasPrefix(key, TenantPrefix) == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(85911)
		return !bytes.HasPrefix(key, prefix) == true
	}() == true {
		__antithesis_instrumentation__.Notify(85912)
		suffix, _, err := DecodeTenantPrefix(key)
		if err != nil {
			__antithesis_instrumentation__.Notify(85914)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(85915)
		}
		__antithesis_instrumentation__.Notify(85913)

		if extra := len(key) - len(prefix) - len(suffix); extra >= 0 {
			__antithesis_instrumentation__.Notify(85916)
			key = key[extra:]
			copy(key, prefix)
		} else {
			__antithesis_instrumentation__.Notify(85917)
			key = make(roachpb.Key, len(prefix)+len(suffix))
			n := copy(key, prefix)
			copy(key[n:], suffix)
		}
	} else {
		__antithesis_instrumentation__.Notify(85918)
	}
	__antithesis_instrumentation__.Notify(85909)
	return key, nil
}
