// Package validate contains all the descriptor validation logic.
package validate

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"

	"github.com/cockroachdb/cockroach/pkg/clusterversion"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/server/telemetry"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/errors"
)

func Self(version clusterversion.ClusterVersion, descriptors ...catalog.Descriptor) error {
	__antithesis_instrumentation__.Notify(265875)
	results := Validate(context.TODO(), version, nil, catalog.NoValidationTelemetry, catalog.ValidationLevelSelfOnly, descriptors...)
	return results.CombinedError()
}

func Validate(
	ctx context.Context,
	version clusterversion.ClusterVersion,
	vd ValidationDereferencer,
	telemetry catalog.ValidationTelemetry,
	targetLevel catalog.ValidationLevel,
	descriptors ...catalog.Descriptor,
) catalog.ValidationErrors {
	__antithesis_instrumentation__.Notify(265876)
	vea := validationErrorAccumulator{
		ValidationTelemetry: telemetry,
		targetLevel:         targetLevel,
		activeVersion:       version,
	}

	if !vea.validateDescriptorsAtLevel(
		catalog.ValidationLevelSelfOnly,
		descriptors,
		func(desc catalog.Descriptor) {
			__antithesis_instrumentation__.Notify(265883)
			desc.ValidateSelf(&vea)
			validateSchemaChangerState(desc, &vea)
		}) {
		__antithesis_instrumentation__.Notify(265884)
		return vea.errors
	} else {
		__antithesis_instrumentation__.Notify(265885)
	}
	__antithesis_instrumentation__.Notify(265877)

	vdg, descGetterErr := collectDescriptorsForValidation(ctx, vd, version, descriptors)
	if descGetterErr != nil {
		__antithesis_instrumentation__.Notify(265886)
		vea.reportDescGetterError(collectingReferencedDescriptors, descGetterErr)
		return vea.errors
	} else {
		__antithesis_instrumentation__.Notify(265887)
	}
	__antithesis_instrumentation__.Notify(265878)

	if !vea.validateDescriptorsAtLevel(
		catalog.ValidationLevelCrossReferences,
		descriptors,
		func(desc catalog.Descriptor) {
			__antithesis_instrumentation__.Notify(265888)
			if !desc.Dropped() {
				__antithesis_instrumentation__.Notify(265889)
				desc.ValidateCrossReferences(&vea, vdg)
			} else {
				__antithesis_instrumentation__.Notify(265890)
			}
		}) {
		__antithesis_instrumentation__.Notify(265891)
		return vea.errors
	} else {
		__antithesis_instrumentation__.Notify(265892)
	}
	__antithesis_instrumentation__.Notify(265879)

	if catalog.ValidationLevelNamespace&targetLevel != 0 {
		__antithesis_instrumentation__.Notify(265893)
		descGetterErr = vdg.addNamespaceEntries(ctx, descriptors, vd)
		if descGetterErr != nil {
			__antithesis_instrumentation__.Notify(265894)
			vea.reportDescGetterError(collectingNamespaceEntries, descGetterErr)
			return vea.errors
		} else {
			__antithesis_instrumentation__.Notify(265895)
		}
	} else {
		__antithesis_instrumentation__.Notify(265896)
	}
	__antithesis_instrumentation__.Notify(265880)

	if !vea.validateDescriptorsAtLevel(
		catalog.ValidationLevelNamespace,
		descriptors,
		func(desc catalog.Descriptor) {
			__antithesis_instrumentation__.Notify(265897)
			validateNamespace(desc, &vea, vdg.namespace)
		}) {
		__antithesis_instrumentation__.Notify(265898)
		return vea.errors
	} else {
		__antithesis_instrumentation__.Notify(265899)
	}
	__antithesis_instrumentation__.Notify(265881)

	_ = vea.validateDescriptorsAtLevel(
		catalog.ValidationLevelAllPreTxnCommit,
		descriptors,
		func(desc catalog.Descriptor) {
			__antithesis_instrumentation__.Notify(265900)
			if !desc.Dropped() {
				__antithesis_instrumentation__.Notify(265901)
				desc.ValidateTxnCommit(&vea, vdg)
			} else {
				__antithesis_instrumentation__.Notify(265902)
			}
		})
	__antithesis_instrumentation__.Notify(265882)
	return vea.errors
}

type ValidationDereferencer interface {
	DereferenceDescriptors(ctx context.Context, version clusterversion.ClusterVersion, reqs []descpb.ID) ([]catalog.Descriptor, error)
	DereferenceDescriptorIDs(ctx context.Context, requests []descpb.NameInfo) ([]descpb.ID, error)
}

type validationErrorAccumulator struct {
	errors catalog.ValidationErrors

	catalog.ValidationTelemetry
	targetLevel       catalog.ValidationLevel
	activeVersion     clusterversion.ClusterVersion
	currentState      validationErrorAccumulatorState
	currentLevel      catalog.ValidationLevel
	currentDescriptor catalog.Descriptor
}

type validationErrorAccumulatorState int

const (
	validatingDescriptor validationErrorAccumulatorState = iota
	collectingReferencedDescriptors
	collectingNamespaceEntries
)

var _ catalog.ValidationErrorAccumulator = &validationErrorAccumulator{}

func (vea *validationErrorAccumulator) Report(err error) {
	__antithesis_instrumentation__.Notify(265903)
	if err == nil {
		__antithesis_instrumentation__.Notify(265905)
		return
	} else {
		__antithesis_instrumentation__.Notify(265906)
	}
	__antithesis_instrumentation__.Notify(265904)
	vea.errors = append(vea.errors, vea.decorate(err))
}

func (vea *validationErrorAccumulator) validateDescriptorsAtLevel(
	level catalog.ValidationLevel,
	descs []catalog.Descriptor,
	validationFn func(descriptor catalog.Descriptor),
) bool {
	__antithesis_instrumentation__.Notify(265907)
	vea.currentState = validatingDescriptor
	vea.currentLevel = level
	if vea.currentLevel&vea.targetLevel != 0 {
		__antithesis_instrumentation__.Notify(265911)
		for _, desc := range descs {
			__antithesis_instrumentation__.Notify(265912)
			if desc == nil {
				__antithesis_instrumentation__.Notify(265914)
				continue
			} else {
				__antithesis_instrumentation__.Notify(265915)
			}
			__antithesis_instrumentation__.Notify(265913)
			vea.currentDescriptor = desc
			validationFn(desc)
		}
	} else {
		__antithesis_instrumentation__.Notify(265916)
	}
	__antithesis_instrumentation__.Notify(265908)
	vea.currentDescriptor = nil

	if level == catalog.ValidationLevelSelfOnly && func() bool {
		__antithesis_instrumentation__.Notify(265917)
		return len(vea.errors) > 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(265918)
		return false
	} else {
		__antithesis_instrumentation__.Notify(265919)
	}
	__antithesis_instrumentation__.Notify(265909)

	if vea.targetLevel <= vea.currentLevel {
		__antithesis_instrumentation__.Notify(265920)
		return false
	} else {
		__antithesis_instrumentation__.Notify(265921)
	}
	__antithesis_instrumentation__.Notify(265910)
	return true
}

func (vea *validationErrorAccumulator) IsActive(version clusterversion.Key) bool {
	__antithesis_instrumentation__.Notify(265922)
	return vea.activeVersion.IsActive(version)
}

func (vea *validationErrorAccumulator) reportDescGetterError(
	state validationErrorAccumulatorState, err error,
) {
	__antithesis_instrumentation__.Notify(265923)
	vea.currentState = state

	vea.errors = append(make([]error, 1, 1+len(vea.errors)), vea.errors...)
	vea.errors[0] = vea.decorate(err)
}

func (vea *validationErrorAccumulator) decorate(err error) error {
	__antithesis_instrumentation__.Notify(265924)
	var tkSuffix string
	switch vea.currentState {
	case collectingReferencedDescriptors:
		__antithesis_instrumentation__.Notify(265927)
		err = errors.Wrap(err, "collecting referenced descriptors")
		tkSuffix = "read_referenced_descriptors"
	case collectingNamespaceEntries:
		__antithesis_instrumentation__.Notify(265928)
		err = errors.Wrap(err, "collecting namespace table entries")
		tkSuffix = "read_namespace_table"
	case validatingDescriptor:
		__antithesis_instrumentation__.Notify(265929)
		name := vea.currentDescriptor.GetName()
		id := vea.currentDescriptor.GetID()

		switch vea.currentDescriptor.DescriptorType() {
		case catalog.Table:
			__antithesis_instrumentation__.Notify(265933)
			err = errors.Wrapf(err, catalog.Table+" %q (%d)", name, id)
		case catalog.Database:
			__antithesis_instrumentation__.Notify(265934)
			err = errors.Wrapf(err, catalog.Database+" %q (%d)", name, id)
		case catalog.Schema:
			__antithesis_instrumentation__.Notify(265935)
			err = errors.Wrapf(err, catalog.Schema+" %q (%d)", name, id)
		case catalog.Type:
			__antithesis_instrumentation__.Notify(265936)
			err = errors.Wrapf(err, catalog.Type+" %q (%d)", name, id)
		default:
			__antithesis_instrumentation__.Notify(265937)
			return err
		}
		__antithesis_instrumentation__.Notify(265930)
		switch vea.currentLevel {
		case catalog.ValidationLevelSelfOnly:
			__antithesis_instrumentation__.Notify(265938)
			tkSuffix = "self"
		case catalog.ValidationLevelCrossReferences:
			__antithesis_instrumentation__.Notify(265939)
			tkSuffix = "cross_references"
		case catalog.ValidationLevelNamespace:
			__antithesis_instrumentation__.Notify(265940)
			tkSuffix = "namespace"
		case catalog.ValidationLevelAllPreTxnCommit:
			__antithesis_instrumentation__.Notify(265941)
			tkSuffix = "pre_txn_commit"
		default:
			__antithesis_instrumentation__.Notify(265942)
			return err
		}
		__antithesis_instrumentation__.Notify(265931)
		tkSuffix += "." + string(vea.currentDescriptor.DescriptorType())
	default:
		__antithesis_instrumentation__.Notify(265932)
	}
	__antithesis_instrumentation__.Notify(265925)
	switch vea.ValidationTelemetry {
	case catalog.ValidationReadTelemetry:
		__antithesis_instrumentation__.Notify(265943)
		tkSuffix = "read." + tkSuffix
	case catalog.ValidationWriteTelemetry:
		__antithesis_instrumentation__.Notify(265944)
		tkSuffix = "write." + tkSuffix
	default:
		__antithesis_instrumentation__.Notify(265945)
		return err
	}
	__antithesis_instrumentation__.Notify(265926)
	return errors.WithTelemetry(err, telemetry.ValidationTelemetryKeyPrefix+tkSuffix)
}

type validationDescGetterImpl struct {
	descriptors map[descpb.ID]catalog.Descriptor
	namespace   map[descpb.NameInfo]descpb.ID
}

var _ catalog.ValidationDescGetter = (*validationDescGetterImpl)(nil)

func (vdg *validationDescGetterImpl) GetDatabaseDescriptor(
	id descpb.ID,
) (catalog.DatabaseDescriptor, error) {
	__antithesis_instrumentation__.Notify(265946)
	desc, found := vdg.descriptors[id]
	if !found || func() bool {
		__antithesis_instrumentation__.Notify(265948)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(265949)
		return nil, catalog.WrapDatabaseDescRefErr(id, catalog.ErrReferencedDescriptorNotFound)
	} else {
		__antithesis_instrumentation__.Notify(265950)
	}
	__antithesis_instrumentation__.Notify(265947)
	return catalog.AsDatabaseDescriptor(desc)
}

func (vdg *validationDescGetterImpl) GetSchemaDescriptor(
	id descpb.ID,
) (catalog.SchemaDescriptor, error) {
	__antithesis_instrumentation__.Notify(265951)
	desc, found := vdg.descriptors[id]
	if !found || func() bool {
		__antithesis_instrumentation__.Notify(265953)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(265954)
		return nil, catalog.WrapSchemaDescRefErr(id, catalog.ErrReferencedDescriptorNotFound)
	} else {
		__antithesis_instrumentation__.Notify(265955)
	}
	__antithesis_instrumentation__.Notify(265952)
	return catalog.AsSchemaDescriptor(desc)
}

func (vdg *validationDescGetterImpl) GetTableDescriptor(
	id descpb.ID,
) (catalog.TableDescriptor, error) {
	__antithesis_instrumentation__.Notify(265956)
	desc, found := vdg.descriptors[id]
	if !found || func() bool {
		__antithesis_instrumentation__.Notify(265958)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(265959)
		return nil, catalog.WrapTableDescRefErr(id, catalog.ErrReferencedDescriptorNotFound)
	} else {
		__antithesis_instrumentation__.Notify(265960)
	}
	__antithesis_instrumentation__.Notify(265957)
	return catalog.AsTableDescriptor(desc)
}

func (vdg *validationDescGetterImpl) GetTypeDescriptor(
	id descpb.ID,
) (catalog.TypeDescriptor, error) {
	__antithesis_instrumentation__.Notify(265961)
	desc, found := vdg.descriptors[id]
	if !found || func() bool {
		__antithesis_instrumentation__.Notify(265964)
		return desc == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(265965)
		return nil, catalog.WrapTypeDescRefErr(id, catalog.ErrReferencedDescriptorNotFound)
	} else {
		__antithesis_instrumentation__.Notify(265966)
	}
	__antithesis_instrumentation__.Notify(265962)
	descriptor, err := catalog.AsTypeDescriptor(desc)
	if err != nil {
		__antithesis_instrumentation__.Notify(265967)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265968)
	}
	__antithesis_instrumentation__.Notify(265963)
	return descriptor, err
}

func (vdg *validationDescGetterImpl) addNamespaceEntries(
	ctx context.Context, descriptors []catalog.Descriptor, vd ValidationDereferencer,
) error {
	__antithesis_instrumentation__.Notify(265969)
	reqs := make([]descpb.NameInfo, 0, len(descriptors))
	for _, desc := range descriptors {
		__antithesis_instrumentation__.Notify(265973)
		if desc == nil {
			__antithesis_instrumentation__.Notify(265975)
			continue
		} else {
			__antithesis_instrumentation__.Notify(265976)
		}
		__antithesis_instrumentation__.Notify(265974)
		reqs = append(reqs, descpb.NameInfo{
			ParentID:       desc.GetParentID(),
			ParentSchemaID: desc.GetParentSchemaID(),
			Name:           desc.GetName(),
		})
		reqs = append(reqs, desc.GetDrainingNames()...)
	}
	__antithesis_instrumentation__.Notify(265970)

	ids, err := vd.DereferenceDescriptorIDs(ctx, reqs)
	if err != nil {
		__antithesis_instrumentation__.Notify(265977)
		return err
	} else {
		__antithesis_instrumentation__.Notify(265978)
	}
	__antithesis_instrumentation__.Notify(265971)
	for i, r := range reqs {
		__antithesis_instrumentation__.Notify(265979)
		vdg.namespace[r] = ids[i]
	}
	__antithesis_instrumentation__.Notify(265972)
	return nil
}

type collectorState struct {
	vdg          validationDescGetterImpl
	referencedBy catalog.DescriptorIDSet
}

func (cs *collectorState) addDirectReferences(desc catalog.Descriptor) error {
	__antithesis_instrumentation__.Notify(265980)
	cs.vdg.descriptors[desc.GetID()] = desc
	idSet, err := desc.GetReferencedDescIDs()
	if err != nil {
		__antithesis_instrumentation__.Notify(265982)
		return err
	} else {
		__antithesis_instrumentation__.Notify(265983)
	}
	__antithesis_instrumentation__.Notify(265981)
	idSet.ForEach(cs.referencedBy.Add)
	return nil
}

func (cs *collectorState) getMissingDescs(
	ctx context.Context, vd ValidationDereferencer, version clusterversion.ClusterVersion,
) ([]catalog.Descriptor, error) {
	__antithesis_instrumentation__.Notify(265984)
	reqs := make([]descpb.ID, 0, cs.referencedBy.Len())
	for _, id := range cs.referencedBy.Ordered() {
		__antithesis_instrumentation__.Notify(265989)
		if _, exists := cs.vdg.descriptors[id]; !exists {
			__antithesis_instrumentation__.Notify(265990)
			reqs = append(reqs, id)
		} else {
			__antithesis_instrumentation__.Notify(265991)
		}
	}
	__antithesis_instrumentation__.Notify(265985)
	if len(reqs) == 0 {
		__antithesis_instrumentation__.Notify(265992)
		return nil, nil
	} else {
		__antithesis_instrumentation__.Notify(265993)
	}
	__antithesis_instrumentation__.Notify(265986)
	resps, err := vd.DereferenceDescriptors(ctx, version, reqs)
	if err != nil {
		__antithesis_instrumentation__.Notify(265994)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(265995)
	}
	__antithesis_instrumentation__.Notify(265987)
	for _, desc := range resps {
		__antithesis_instrumentation__.Notify(265996)
		if desc != nil {
			__antithesis_instrumentation__.Notify(265997)
			cs.vdg.descriptors[desc.GetID()] = desc
		} else {
			__antithesis_instrumentation__.Notify(265998)
		}
	}
	__antithesis_instrumentation__.Notify(265988)
	return resps, nil
}

func collectDescriptorsForValidation(
	ctx context.Context,
	vd ValidationDereferencer,
	version clusterversion.ClusterVersion,
	descriptors []catalog.Descriptor,
) (*validationDescGetterImpl, error) {
	__antithesis_instrumentation__.Notify(265999)
	cs := collectorState{
		vdg: validationDescGetterImpl{
			descriptors: make(map[descpb.ID]catalog.Descriptor, len(descriptors)),
			namespace:   make(map[descpb.NameInfo]descpb.ID, len(descriptors)),
		},
		referencedBy: catalog.MakeDescriptorIDSet(),
	}
	for _, desc := range descriptors {
		__antithesis_instrumentation__.Notify(266004)
		if desc == nil {
			__antithesis_instrumentation__.Notify(266006)
			continue
		} else {
			__antithesis_instrumentation__.Notify(266007)
		}
		__antithesis_instrumentation__.Notify(266005)
		if err := cs.addDirectReferences(desc); err != nil {
			__antithesis_instrumentation__.Notify(266008)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(266009)
		}
	}
	__antithesis_instrumentation__.Notify(266000)
	newDescs, err := cs.getMissingDescs(ctx, vd, version)
	if err != nil {
		__antithesis_instrumentation__.Notify(266010)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(266011)
	}
	__antithesis_instrumentation__.Notify(266001)
	for _, newDesc := range newDescs {
		__antithesis_instrumentation__.Notify(266012)
		if newDesc == nil {
			__antithesis_instrumentation__.Notify(266014)
			continue
		} else {
			__antithesis_instrumentation__.Notify(266015)
		}
		__antithesis_instrumentation__.Notify(266013)
		switch newDesc.(type) {
		case catalog.DatabaseDescriptor, catalog.TypeDescriptor:
			__antithesis_instrumentation__.Notify(266016)
			if err := cs.addDirectReferences(newDesc); err != nil {
				__antithesis_instrumentation__.Notify(266017)
				return nil, err
			} else {
				__antithesis_instrumentation__.Notify(266018)
			}
		}
	}
	__antithesis_instrumentation__.Notify(266002)
	_, err = cs.getMissingDescs(ctx, vd, version)
	if err != nil {
		__antithesis_instrumentation__.Notify(266019)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(266020)
	}
	__antithesis_instrumentation__.Notify(266003)
	return &cs.vdg, nil
}

func validateNamespace(
	desc catalog.Descriptor,
	vea catalog.ValidationErrorAccumulator,
	namespace map[descpb.NameInfo]descpb.ID,
) {
	__antithesis_instrumentation__.Notify(266021)
	if desc.GetID() == keys.NamespaceTableID || func() bool {
		__antithesis_instrumentation__.Notify(266024)
		return desc.GetID() == keys.DeprecatedNamespaceTableID == true
	}() == true {
		__antithesis_instrumentation__.Notify(266025)
		return
	} else {
		__antithesis_instrumentation__.Notify(266026)
	}
	__antithesis_instrumentation__.Notify(266022)

	key := descpb.NameInfo{
		ParentID:       desc.GetParentID(),
		ParentSchemaID: desc.GetParentSchemaID(),
		Name:           desc.GetName(),
	}
	id := namespace[key]

	if !desc.Dropped() {
		__antithesis_instrumentation__.Notify(266027)
		if id == descpb.InvalidID {
			__antithesis_instrumentation__.Notify(266028)
			vea.Report(errors.Errorf("expected matching namespace entry, found none"))
		} else {
			__antithesis_instrumentation__.Notify(266029)
			if id != desc.GetID() {
				__antithesis_instrumentation__.Notify(266030)
				vea.Report(errors.Errorf("expected matching namespace entry value, instead found %d", id))
			} else {
				__antithesis_instrumentation__.Notify(266031)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(266032)
	}
	__antithesis_instrumentation__.Notify(266023)

	for _, dn := range desc.GetDrainingNames() {
		__antithesis_instrumentation__.Notify(266033)
		id := namespace[dn]
		if id == descpb.InvalidID {
			__antithesis_instrumentation__.Notify(266034)
			vea.Report(errors.Errorf("expected matching namespace entry for draining name (%d, %d, %s), found none",
				dn.ParentID, dn.ParentSchemaID, dn.Name))
		} else {
			__antithesis_instrumentation__.Notify(266035)
			if id != desc.GetID() {
				__antithesis_instrumentation__.Notify(266036)
				vea.Report(errors.Errorf("expected matching namespace entry value for draining name (%d, %d, %s), instead found %d",
					dn.ParentID, dn.ParentSchemaID, dn.Name, id))
			} else {
				__antithesis_instrumentation__.Notify(266037)
			}
		}
	}
}
