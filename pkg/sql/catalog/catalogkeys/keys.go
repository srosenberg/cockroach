package catalogkeys

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catconstants"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/errors"
)

const (
	DefaultDatabaseName = "defaultdb"

	PgDatabaseName = "postgres"
)

var DefaultUserDBs = []string{
	DefaultDatabaseName, PgDatabaseName,
}

func IndexKeyValDirs(index catalog.Index) []encoding.Direction {
	__antithesis_instrumentation__.Notify(247306)
	if index == nil {
		__antithesis_instrumentation__.Notify(247309)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(247310)
	}
	__antithesis_instrumentation__.Notify(247307)

	dirs := make([]encoding.Direction, 0, 2+index.NumKeyColumns())

	dirs = append(dirs, encoding.Ascending, encoding.Ascending)

	for colIdx := 0; colIdx < index.NumKeyColumns(); colIdx++ {
		__antithesis_instrumentation__.Notify(247311)
		d, err := index.GetKeyColumnDirection(colIdx).ToEncodingDirection()
		if err != nil {
			__antithesis_instrumentation__.Notify(247313)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(247314)
		}
		__antithesis_instrumentation__.Notify(247312)
		dirs = append(dirs, d)
	}
	__antithesis_instrumentation__.Notify(247308)

	return dirs
}

func PrettyKey(valDirs []encoding.Direction, key roachpb.Key, skip int) string {
	__antithesis_instrumentation__.Notify(247315)
	p := key.StringWithDirs(valDirs, 0)
	for i := 0; i <= skip; i++ {
		__antithesis_instrumentation__.Notify(247317)
		n := strings.IndexByte(p[1:], '/')
		if n == -1 {
			__antithesis_instrumentation__.Notify(247319)
			return ""
		} else {
			__antithesis_instrumentation__.Notify(247320)
		}
		__antithesis_instrumentation__.Notify(247318)
		p = p[n+1:]
	}
	__antithesis_instrumentation__.Notify(247316)
	return p
}

func PrettySpan(valDirs []encoding.Direction, span roachpb.Span, skip int) string {
	__antithesis_instrumentation__.Notify(247321)
	var b strings.Builder
	b.WriteString(PrettyKey(valDirs, span.Key, skip))
	if span.EndKey != nil {
		__antithesis_instrumentation__.Notify(247323)
		b.WriteByte('-')
		b.WriteString(PrettyKey(valDirs, span.EndKey, skip))
	} else {
		__antithesis_instrumentation__.Notify(247324)
	}
	__antithesis_instrumentation__.Notify(247322)
	return b.String()
}

func PrettySpans(index catalog.Index, spans []roachpb.Span, skip int) string {
	__antithesis_instrumentation__.Notify(247325)
	if len(spans) == 0 {
		__antithesis_instrumentation__.Notify(247328)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(247329)
	}
	__antithesis_instrumentation__.Notify(247326)

	valDirs := IndexKeyValDirs(index)

	var b strings.Builder
	for i, span := range spans {
		__antithesis_instrumentation__.Notify(247330)
		if i > 0 {
			__antithesis_instrumentation__.Notify(247332)
			b.WriteString(" ")
		} else {
			__antithesis_instrumentation__.Notify(247333)
		}
		__antithesis_instrumentation__.Notify(247331)
		b.WriteString(PrettySpan(valDirs, span, skip))
	}
	__antithesis_instrumentation__.Notify(247327)
	return b.String()
}

func NewNameKeyComponents(
	parentID descpb.ID, parentSchemaID descpb.ID, name string,
) catalog.NameKey {
	__antithesis_instrumentation__.Notify(247334)
	return descpb.NameInfo{ParentID: parentID, ParentSchemaID: parentSchemaID, Name: name}
}

func MakeObjectNameKey(
	codec keys.SQLCodec, parentID, parentSchemaID descpb.ID, name string,
) roachpb.Key {
	__antithesis_instrumentation__.Notify(247335)
	return EncodeNameKey(codec, NewNameKeyComponents(parentID, parentSchemaID, name))
}

func MakePublicObjectNameKey(codec keys.SQLCodec, parentID descpb.ID, name string) roachpb.Key {
	__antithesis_instrumentation__.Notify(247336)
	return EncodeNameKey(codec, NewNameKeyComponents(parentID, keys.PublicSchemaID, name))
}

func MakeSchemaNameKey(codec keys.SQLCodec, parentID descpb.ID, name string) roachpb.Key {
	__antithesis_instrumentation__.Notify(247337)
	return EncodeNameKey(codec, NewNameKeyComponents(parentID, keys.RootNamespaceID, name))
}

func MakeDatabaseNameKey(codec keys.SQLCodec, name string) roachpb.Key {
	__antithesis_instrumentation__.Notify(247338)
	return EncodeNameKey(codec, NewNameKeyComponents(keys.RootNamespaceID, keys.RootNamespaceID, name))
}

func EncodeNameKey(codec keys.SQLCodec, nameKey catalog.NameKey) roachpb.Key {
	__antithesis_instrumentation__.Notify(247339)
	r := codec.IndexPrefix(keys.NamespaceTableID, catconstants.NamespaceTablePrimaryIndexID)
	r = encoding.EncodeUvarintAscending(r, uint64(nameKey.GetParentID()))
	r = encoding.EncodeUvarintAscending(r, uint64(nameKey.GetParentSchemaID()))
	if nameKey.GetName() != "" {
		__antithesis_instrumentation__.Notify(247341)
		r = encoding.EncodeBytesAscending(r, []byte(nameKey.GetName()))
		r = keys.MakeFamilyKey(r, catconstants.NamespaceTableFamilyID)
	} else {
		__antithesis_instrumentation__.Notify(247342)
	}
	__antithesis_instrumentation__.Notify(247340)
	return r
}

func DecodeNameMetadataKey(
	codec keys.SQLCodec, k roachpb.Key,
) (nameKey descpb.NameInfo, err error) {
	__antithesis_instrumentation__.Notify(247343)
	k, _, err = codec.DecodeTablePrefix(k)
	if err != nil {
		__antithesis_instrumentation__.Notify(247350)
		return nameKey, err
	} else {
		__antithesis_instrumentation__.Notify(247351)
	}
	__antithesis_instrumentation__.Notify(247344)

	var buf uint64
	k, buf, err = encoding.DecodeUvarintAscending(k)
	if err != nil {
		__antithesis_instrumentation__.Notify(247352)
		return nameKey, err
	} else {
		__antithesis_instrumentation__.Notify(247353)
	}
	__antithesis_instrumentation__.Notify(247345)
	if buf != uint64(catconstants.NamespaceTablePrimaryIndexID) {
		__antithesis_instrumentation__.Notify(247354)
		return nameKey, errors.Newf("tried get table %d, but got %d", catconstants.NamespaceTablePrimaryIndexID, buf)
	} else {
		__antithesis_instrumentation__.Notify(247355)
	}
	__antithesis_instrumentation__.Notify(247346)

	k, buf, err = encoding.DecodeUvarintAscending(k)
	if err != nil {
		__antithesis_instrumentation__.Notify(247356)
		return nameKey, err
	} else {
		__antithesis_instrumentation__.Notify(247357)
	}
	__antithesis_instrumentation__.Notify(247347)
	nameKey.ParentID = descpb.ID(buf)

	k, buf, err = encoding.DecodeUvarintAscending(k)
	if err != nil {
		__antithesis_instrumentation__.Notify(247358)
		return nameKey, err
	} else {
		__antithesis_instrumentation__.Notify(247359)
	}
	__antithesis_instrumentation__.Notify(247348)
	nameKey.ParentSchemaID = descpb.ID(buf)

	var bytesBuf []byte
	_, bytesBuf, err = encoding.DecodeBytesAscending(k, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(247360)
		return nameKey, err
	} else {
		__antithesis_instrumentation__.Notify(247361)
	}
	__antithesis_instrumentation__.Notify(247349)
	nameKey.Name = string(bytesBuf)

	return nameKey, nil
}

func MakeAllDescsMetadataKey(codec keys.SQLCodec) roachpb.Key {
	__antithesis_instrumentation__.Notify(247362)
	return codec.DescMetadataPrefix()
}

func MakeDescMetadataKey(codec keys.SQLCodec, descID descpb.ID) roachpb.Key {
	__antithesis_instrumentation__.Notify(247363)
	return codec.DescMetadataKey(uint32(descID))
}
