package sqlutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"

	"github.com/cockroachdb/cockroach-go/v2/crdb"
	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/cockroachdb/errors"
)

func InjectDescriptors(
	ctx context.Context, db *gosql.DB, input []*descpb.Descriptor, force bool,
) error {
	__antithesis_instrumentation__.Notify(646085)
	cloneInput := func() []*descpb.Descriptor {
		__antithesis_instrumentation__.Notify(646090)
		cloned := make([]*descpb.Descriptor, 0, len(input))
		for _, d := range input {
			__antithesis_instrumentation__.Notify(646092)
			cloned = append(cloned, protoutil.Clone(d).(*descpb.Descriptor))
		}
		__antithesis_instrumentation__.Notify(646091)
		return cloned
	}
	__antithesis_instrumentation__.Notify(646086)
	findDatabases := func(descs []*descpb.Descriptor) (dbs, others []*descpb.Descriptor) {
		__antithesis_instrumentation__.Notify(646093)
		for _, d := range descs {
			__antithesis_instrumentation__.Notify(646095)
			if _, ok := d.Union.(*descpb.Descriptor_Database); ok {
				__antithesis_instrumentation__.Notify(646096)
				dbs = append(dbs, d)
			} else {
				__antithesis_instrumentation__.Notify(646097)
				others = append(others, d)
			}
		}
		__antithesis_instrumentation__.Notify(646094)
		return dbs, others
	}
	__antithesis_instrumentation__.Notify(646087)
	injectDescriptor := func(tx *gosql.Tx, id descpb.ID, desc *descpb.Descriptor) error {
		__antithesis_instrumentation__.Notify(646098)
		resetVersionAndModificationTime(desc)
		encoded, err := protoutil.Marshal(desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(646100)
			return err
		} else {
			__antithesis_instrumentation__.Notify(646101)
		}
		__antithesis_instrumentation__.Notify(646099)
		_, err = tx.Exec(
			"SELECT crdb_internal.unsafe_upsert_descriptor($1, $2, $3)",
			id, encoded, force,
		)
		return err
	}
	__antithesis_instrumentation__.Notify(646088)
	injectNamespaceEntry := func(
		tx *gosql.Tx, parent, schema descpb.ID, name string, id descpb.ID,
	) error {
		__antithesis_instrumentation__.Notify(646102)
		_, err := tx.Exec(
			"SELECT crdb_internal.unsafe_upsert_namespace_entry($1, $2, $3, $4)",
			parent, schema, name, id,
		)
		return err
	}
	__antithesis_instrumentation__.Notify(646089)
	return crdb.ExecuteTx(ctx, db, nil, func(tx *gosql.Tx) error {
		__antithesis_instrumentation__.Notify(646103)
		descriptors := cloneInput()

		databases, others := findDatabases(descriptors)
		for _, db := range databases {
			__antithesis_instrumentation__.Notify(646107)
			id, _, name, _, _, err := descpb.GetDescriptorMetadata(db)
			if err != nil {
				__antithesis_instrumentation__.Notify(646111)
				return err
			} else {
				__antithesis_instrumentation__.Notify(646112)
			}
			__antithesis_instrumentation__.Notify(646108)
			if err := injectDescriptor(tx, id, db); err != nil {
				__antithesis_instrumentation__.Notify(646113)
				return errors.Wrapf(err, "failed to inject database descriptor %d", id)
			} else {
				__antithesis_instrumentation__.Notify(646114)
			}
			__antithesis_instrumentation__.Notify(646109)
			if err := injectNamespaceEntry(tx, 0, 0, name, id); err != nil {
				__antithesis_instrumentation__.Notify(646115)
				return errors.Wrapf(err, "failed to inject namespace entry for database %d", id)
			} else {
				__antithesis_instrumentation__.Notify(646116)
			}
			__antithesis_instrumentation__.Notify(646110)
			if err := injectNamespaceEntry(
				tx, id, 0, tree.PublicSchema, keys.PublicSchemaID,
			); err != nil {
				__antithesis_instrumentation__.Notify(646117)
				return errors.Wrapf(err, "failed to inject namespace entry for public schema in %d", id)
			} else {
				__antithesis_instrumentation__.Notify(646118)
			}
		}
		__antithesis_instrumentation__.Notify(646104)

		for _, d := range others {
			__antithesis_instrumentation__.Notify(646119)
			id, _, _, _, _, err := descpb.GetDescriptorMetadata(d)
			if err != nil {
				__antithesis_instrumentation__.Notify(646121)
				return err
			} else {
				__antithesis_instrumentation__.Notify(646122)
			}
			__antithesis_instrumentation__.Notify(646120)
			if err := injectDescriptor(tx, id, d); err != nil {
				__antithesis_instrumentation__.Notify(646123)
				return errors.Wrapf(err, "failed to inject descriptor %d", id)
			} else {
				__antithesis_instrumentation__.Notify(646124)
			}
		}
		__antithesis_instrumentation__.Notify(646105)

		for _, d := range others {
			__antithesis_instrumentation__.Notify(646125)
			id, _, name, _, _, err := descpb.GetDescriptorMetadata(d)
			if err != nil {
				__antithesis_instrumentation__.Notify(646127)
				return err
			} else {
				__antithesis_instrumentation__.Notify(646128)
			}
			__antithesis_instrumentation__.Notify(646126)
			parent, schema := getDescriptorParentAndSchema(d)
			if err := injectNamespaceEntry(tx, parent, schema, name, id); err != nil {
				__antithesis_instrumentation__.Notify(646129)
				return errors.Wrapf(err, "failed to inject namespace entry (%d, %d, %s) %d",
					parent, schema, name, id)
			} else {
				__antithesis_instrumentation__.Notify(646130)
			}
		}
		__antithesis_instrumentation__.Notify(646106)
		return nil
	})
}

func getDescriptorParentAndSchema(d *descpb.Descriptor) (parent, schema descpb.ID) {
	__antithesis_instrumentation__.Notify(646131)
	switch d := d.Union.(type) {
	case *descpb.Descriptor_Database:
		__antithesis_instrumentation__.Notify(646132)
		return 0, 0
	case *descpb.Descriptor_Schema:
		__antithesis_instrumentation__.Notify(646133)
		return d.Schema.ParentID, 0
	case *descpb.Descriptor_Type:
		__antithesis_instrumentation__.Notify(646134)
		return d.Type.ParentID, d.Type.ParentSchemaID
	case *descpb.Descriptor_Table:
		__antithesis_instrumentation__.Notify(646135)
		schema := d.Table.UnexposedParentSchemaID

		if schema == 0 {
			__antithesis_instrumentation__.Notify(646138)
			schema = keys.PublicSchemaID
		} else {
			__antithesis_instrumentation__.Notify(646139)
		}
		__antithesis_instrumentation__.Notify(646136)
		return d.Table.ParentID, schema
	default:
		__antithesis_instrumentation__.Notify(646137)
		panic(errors.Errorf("unknown descriptor type %T", d))
	}
}

func resetVersionAndModificationTime(d *descpb.Descriptor) {
	__antithesis_instrumentation__.Notify(646140)
	switch d := d.Union.(type) {
	case *descpb.Descriptor_Database:
		__antithesis_instrumentation__.Notify(646141)
		d.Database.Version = 1
	case *descpb.Descriptor_Schema:
		__antithesis_instrumentation__.Notify(646142)
		d.Schema.Version = 1
	case *descpb.Descriptor_Type:
		__antithesis_instrumentation__.Notify(646143)
		d.Type.Version = 1
	case *descpb.Descriptor_Table:
		__antithesis_instrumentation__.Notify(646144)
		d.Table.Version = 1
	}
}
