package nstree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/internal/validate"
	"github.com/cockroachdb/errors"
)

type Catalog struct {
	underlying Map
}

func (c Catalog) ForEachDescriptorEntry(fn func(desc catalog.Descriptor) error) error {
	__antithesis_instrumentation__.Notify(266971)
	if !c.IsInitialized() {
		__antithesis_instrumentation__.Notify(266973)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266974)
	}
	__antithesis_instrumentation__.Notify(266972)
	return c.underlying.byID.ascend(func(e catalog.NameEntry) error {
		__antithesis_instrumentation__.Notify(266975)
		return fn(e.(catalog.Descriptor))
	})
}

func (c Catalog) ForEachNamespaceEntry(fn func(e catalog.NameEntry) error) error {
	__antithesis_instrumentation__.Notify(266976)
	if !c.IsInitialized() {
		__antithesis_instrumentation__.Notify(266978)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266979)
	}
	__antithesis_instrumentation__.Notify(266977)
	return c.underlying.byName.ascend(fn)
}

func (c Catalog) LookupDescriptorEntry(id descpb.ID) catalog.Descriptor {
	__antithesis_instrumentation__.Notify(266980)
	if !c.IsInitialized() || func() bool {
		__antithesis_instrumentation__.Notify(266983)
		return id == descpb.InvalidID == true
	}() == true {
		__antithesis_instrumentation__.Notify(266984)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266985)
	}
	__antithesis_instrumentation__.Notify(266981)
	e := c.underlying.byID.get(id)
	if e == nil {
		__antithesis_instrumentation__.Notify(266986)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266987)
	}
	__antithesis_instrumentation__.Notify(266982)
	return e.(catalog.Descriptor)
}

func (c Catalog) LookupNamespaceEntry(key catalog.NameKey) catalog.NameEntry {
	__antithesis_instrumentation__.Notify(266988)
	if !c.IsInitialized() || func() bool {
		__antithesis_instrumentation__.Notify(266990)
		return key == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(266991)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266992)
	}
	__antithesis_instrumentation__.Notify(266989)
	return c.underlying.byName.getByName(key.GetParentID(), key.GetParentSchemaID(), key.GetName())
}

func (c Catalog) OrderedDescriptors() []catalog.Descriptor {
	__antithesis_instrumentation__.Notify(266993)
	if !c.IsInitialized() {
		__antithesis_instrumentation__.Notify(266996)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(266997)
	}
	__antithesis_instrumentation__.Notify(266994)
	ret := make([]catalog.Descriptor, 0, c.underlying.byID.t.Len())
	_ = c.ForEachDescriptorEntry(func(desc catalog.Descriptor) error {
		__antithesis_instrumentation__.Notify(266998)
		ret = append(ret, desc)
		return nil
	})
	__antithesis_instrumentation__.Notify(266995)
	return ret
}

func (c Catalog) OrderedDescriptorIDs() []descpb.ID {
	__antithesis_instrumentation__.Notify(266999)
	if !c.IsInitialized() {
		__antithesis_instrumentation__.Notify(267002)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(267003)
	}
	__antithesis_instrumentation__.Notify(267000)
	ret := make([]descpb.ID, 0, c.underlying.byName.t.Len())
	_ = c.ForEachNamespaceEntry(func(e catalog.NameEntry) error {
		__antithesis_instrumentation__.Notify(267004)
		ret = append(ret, e.GetID())
		return nil
	})
	__antithesis_instrumentation__.Notify(267001)
	return ret
}

func (c Catalog) IsInitialized() bool {
	__antithesis_instrumentation__.Notify(267005)
	return c.underlying.initialized()
}

var _ validate.ValidationDereferencer = Catalog{}
var _ validate.ValidationDereferencer = MutableCatalog{}

func (c Catalog) DereferenceDescriptors(
	ctx context.Context, version clusterversion.ClusterVersion, reqs []descpb.ID,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(267006)
	ret := make([]catalog.Descriptor, len(reqs))
	for i, id := range reqs {
		__antithesis_instrumentation__.Notify(267008)
		ret[i] = c.LookupDescriptorEntry(id)
	}
	__antithesis_instrumentation__.Notify(267007)
	return ret, nil
}

func (c Catalog) DereferenceDescriptorIDs(
	_ context.Context, reqs []descpb.NameInfo,
) ([]descpb.ID, error) {
	__antithesis_instrumentation__.Notify(267009)
	ret := make([]descpb.ID, len(reqs))
	for i, req := range reqs {
		__antithesis_instrumentation__.Notify(267011)
		ne := c.LookupNamespaceEntry(req)
		if ne == nil {
			__antithesis_instrumentation__.Notify(267013)
			continue
		} else {
			__antithesis_instrumentation__.Notify(267014)
		}
		__antithesis_instrumentation__.Notify(267012)
		ret[i] = ne.GetID()
	}
	__antithesis_instrumentation__.Notify(267010)
	return ret, nil
}

func (c Catalog) Validate(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	telemetry catalog.ValidationTelemetry,
	targetLevel catalog.ValidationLevel,
	descriptors ...catalog.Descriptor,
) (ve catalog.ValidationErrors) {
	__antithesis_instrumentation__.Notify(267015)
	return validate.Validate(ctx, version, c, telemetry, targetLevel, descriptors...)
}

func (c Catalog) ValidateNamespaceEntry(key catalog.NameKey) error {
	__antithesis_instrumentation__.Notify(267016)
	ne := c.LookupNamespaceEntry(key)
	if ne == nil {
		__antithesis_instrumentation__.Notify(267023)
		return errors.New("invalid descriptor ID")
	} else {
		__antithesis_instrumentation__.Notify(267024)
	}
	__antithesis_instrumentation__.Notify(267017)

	switch ne.GetID() {
	case descpb.InvalidID:
		__antithesis_instrumentation__.Notify(267025)
		return errors.New("invalid descriptor ID")
	case keys.PublicSchemaID:
		__antithesis_instrumentation__.Notify(267026)

		return nil
	default:
		__antithesis_instrumentation__.Notify(267027)
		isSchema := ne.GetParentID() != keys.RootNamespaceID && func() bool {
			__antithesis_instrumentation__.Notify(267028)
			return ne.GetParentSchemaID() == keys.RootNamespaceID == true
		}() == true
		if isSchema && func() bool {
			__antithesis_instrumentation__.Notify(267029)
			return strings.HasPrefix(ne.GetName(), "pg_temp_") == true
		}() == true {
			__antithesis_instrumentation__.Notify(267030)

			return nil
		} else {
			__antithesis_instrumentation__.Notify(267031)
		}
	}
	__antithesis_instrumentation__.Notify(267018)

	desc := c.LookupDescriptorEntry(ne.GetID())
	if desc == nil {
		__antithesis_instrumentation__.Notify(267032)
		return catalog.ErrDescriptorNotFound
	} else {
		__antithesis_instrumentation__.Notify(267033)
	}
	__antithesis_instrumentation__.Notify(267019)

	for _, dn := range desc.GetDrainingNames() {
		__antithesis_instrumentation__.Notify(267034)
		if ne.GetParentID() == dn.GetParentID() && func() bool {
			__antithesis_instrumentation__.Notify(267035)
			return ne.GetParentSchemaID() == dn.GetParentSchemaID() == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(267036)
			return ne.GetName() == dn.GetName() == true
		}() == true {
			__antithesis_instrumentation__.Notify(267037)
			return nil
		} else {
			__antithesis_instrumentation__.Notify(267038)
		}
	}
	__antithesis_instrumentation__.Notify(267020)
	if desc.Dropped() {
		__antithesis_instrumentation__.Notify(267039)
		return errors.Newf("no matching name info in draining names of dropped %s",
			desc.DescriptorType())
	} else {
		__antithesis_instrumentation__.Notify(267040)
	}
	__antithesis_instrumentation__.Notify(267021)
	if ne.GetParentID() == desc.GetParentID() && func() bool {
		__antithesis_instrumentation__.Notify(267041)
		return ne.GetParentSchemaID() == desc.GetParentSchemaID() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(267042)
		return ne.GetName() == desc.GetName() == true
	}() == true {
		__antithesis_instrumentation__.Notify(267043)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(267044)
	}
	__antithesis_instrumentation__.Notify(267022)
	return errors.Newf("no matching name info found in non-dropped %s %q",
		desc.DescriptorType(), desc.GetName())
}

func (c Catalog) ValidateWithRecover(
	ctx context.Context, version clusterversion.ClusterVersion, desc catalog.Descriptor,
) (ve catalog.ValidationErrors) {
	__antithesis_instrumentation__.Notify(267045)
	defer func() {
		__antithesis_instrumentation__.Notify(267047)
		if r := recover(); r != nil {
			__antithesis_instrumentation__.Notify(267048)
			err, ok := r.(error)
			if !ok {
				__antithesis_instrumentation__.Notify(267050)
				err = errors.Newf("%v", r)
			} else {
				__antithesis_instrumentation__.Notify(267051)
			}
			__antithesis_instrumentation__.Notify(267049)
			err = errors.WithAssertionFailure(errors.Wrap(err, "validation"))
			ve = append(ve, err)
		} else {
			__antithesis_instrumentation__.Notify(267052)
		}
	}()
	__antithesis_instrumentation__.Notify(267046)
	return c.Validate(ctx, version, catalog.NoValidationTelemetry, catalog.ValidationLevelAllPreTxnCommit, desc)
}

type MutableCatalog struct {
	Catalog
}

func (mc *MutableCatalog) UpsertDescriptorEntry(desc catalog.Descriptor) {
	__antithesis_instrumentation__.Notify(267053)
	if desc == nil || func() bool {
		__antithesis_instrumentation__.Notify(267055)
		return desc.GetID() == descpb.InvalidID == true
	}() == true {
		__antithesis_instrumentation__.Notify(267056)
		return
	} else {
		__antithesis_instrumentation__.Notify(267057)
	}
	__antithesis_instrumentation__.Notify(267054)
	mc.underlying.maybeInitialize()
	mc.underlying.byID.upsert(desc)
}

func (mc *MutableCatalog) DeleteDescriptorEntry(id descpb.ID) {
	__antithesis_instrumentation__.Notify(267058)
	if id == descpb.InvalidID || func() bool {
		__antithesis_instrumentation__.Notify(267060)
		return !mc.IsInitialized() == true
	}() == true {
		__antithesis_instrumentation__.Notify(267061)
		return
	} else {
		__antithesis_instrumentation__.Notify(267062)
	}
	__antithesis_instrumentation__.Notify(267059)
	mc.underlying.maybeInitialize()
	mc.underlying.byID.delete(id)
}

func (mc *MutableCatalog) UpsertNamespaceEntry(key catalog.NameKey, id descpb.ID) {
	__antithesis_instrumentation__.Notify(267063)
	if key == nil || func() bool {
		__antithesis_instrumentation__.Notify(267065)
		return id == descpb.InvalidID == true
	}() == true {
		__antithesis_instrumentation__.Notify(267066)
		return
	} else {
		__antithesis_instrumentation__.Notify(267067)
	}
	__antithesis_instrumentation__.Notify(267064)
	mc.underlying.maybeInitialize()
	mc.underlying.byName.upsert(&namespaceEntry{
		NameInfo: descpb.NameInfo{
			ParentID:       key.GetParentID(),
			ParentSchemaID: key.GetParentSchemaID(),
			Name:           key.GetName(),
		},
		ID: id,
	})
}

func (mc *MutableCatalog) DeleteNamespaceEntry(key catalog.NameKey) {
	__antithesis_instrumentation__.Notify(267068)
	if key == nil || func() bool {
		__antithesis_instrumentation__.Notify(267070)
		return !mc.IsInitialized() == true
	}() == true {
		__antithesis_instrumentation__.Notify(267071)
		return
	} else {
		__antithesis_instrumentation__.Notify(267072)
	}
	__antithesis_instrumentation__.Notify(267069)
	mc.underlying.maybeInitialize()
	mc.underlying.byName.delete(key)
}

func (mc *MutableCatalog) Clear() {
	__antithesis_instrumentation__.Notify(267073)
	mc.underlying.Clear()
}

type namespaceEntry struct {
	descpb.NameInfo
	descpb.ID
}

var _ catalog.NameEntry = namespaceEntry{}

func (e namespaceEntry) GetID() descpb.ID {
	__antithesis_instrumentation__.Notify(267074)
	return e.ID
}
