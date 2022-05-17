package descs

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/config/zonepb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/bootstrap"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/util/syncutil"
)

type systemDatabaseNamespaceCache struct {
	syncutil.RWMutex
	ns map[descpb.NameInfo]descpb.ID
}

func newSystemDatabaseNamespaceCache(codec keys.SQLCodec) *systemDatabaseNamespaceCache {
	__antithesis_instrumentation__.Notify(264893)
	nc := &systemDatabaseNamespaceCache{}
	nc.ns = make(map[descpb.NameInfo]descpb.ID)
	ms := bootstrap.MakeMetadataSchema(
		codec,
		zonepb.DefaultZoneConfigRef(),
		zonepb.DefaultSystemZoneConfigRef(),
	)
	_ = ms.ForEachCatalogDescriptor(func(desc catalog.Descriptor) error {
		__antithesis_instrumentation__.Notify(264895)
		if desc.GetID() < keys.MaxReservedDescID {
			__antithesis_instrumentation__.Notify(264897)
			nc.ns[descpb.NameInfo{
				ParentID:       desc.GetParentID(),
				ParentSchemaID: desc.GetParentSchemaID(),
				Name:           desc.GetName(),
			}] = desc.GetID()
		} else {
			__antithesis_instrumentation__.Notify(264898)
		}
		__antithesis_instrumentation__.Notify(264896)
		return nil
	})
	__antithesis_instrumentation__.Notify(264894)
	return nc
}

func (s *systemDatabaseNamespaceCache) lookup(schemaID descpb.ID, name string) descpb.ID {
	__antithesis_instrumentation__.Notify(264899)
	if s == nil {
		__antithesis_instrumentation__.Notify(264901)
		return descpb.InvalidID
	} else {
		__antithesis_instrumentation__.Notify(264902)
	}
	__antithesis_instrumentation__.Notify(264900)
	s.RLock()
	defer s.RUnlock()
	return s.ns[descpb.NameInfo{
		ParentID:       keys.SystemDatabaseID,
		ParentSchemaID: schemaID,
		Name:           name,
	}]
}

func (s *systemDatabaseNamespaceCache) add(info descpb.NameInfo, id descpb.ID) {
	__antithesis_instrumentation__.Notify(264903)
	if s == nil {
		__antithesis_instrumentation__.Notify(264905)
		return
	} else {
		__antithesis_instrumentation__.Notify(264906)
	}
	__antithesis_instrumentation__.Notify(264904)
	s.Lock()
	defer s.Unlock()
	s.ns[info] = id
}
