package rewrite

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"go/constant"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemadesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/schemaexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/tabledesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/typedesc"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/screl"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
	"github.com/lib/pq/oid"
)

func TableDescs(
	tables []*tabledesc.Mutable, descriptorRewrites jobspb.DescRewriteMap, overrideDB string,
) error {
	__antithesis_instrumentation__.Notify(267356)
	for _, table := range tables {
		__antithesis_instrumentation__.Notify(267358)
		tableRewrite, ok := descriptorRewrites[table.ID]
		if !ok {
			__antithesis_instrumentation__.Notify(267372)
			return errors.Errorf("missing table rewrite for table %d", table.ID)
		} else {
			__antithesis_instrumentation__.Notify(267373)
		}
		__antithesis_instrumentation__.Notify(267359)

		table.Version = 1
		table.ModificationTime = hlc.Timestamp{}

		if table.IsView() && func() bool {
			__antithesis_instrumentation__.Notify(267374)
			return overrideDB != "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(267375)

			if err := rewriteViewQueryDBNames(table, overrideDB); err != nil {
				__antithesis_instrumentation__.Notify(267376)
				return err
			} else {
				__antithesis_instrumentation__.Notify(267377)
			}
		} else {
			__antithesis_instrumentation__.Notify(267378)
		}
		__antithesis_instrumentation__.Notify(267360)
		if err := rewriteSchemaChangerState(table, descriptorRewrites); err != nil {
			__antithesis_instrumentation__.Notify(267379)
			return err
		} else {
			__antithesis_instrumentation__.Notify(267380)
		}
		__antithesis_instrumentation__.Notify(267361)

		table.ID = tableRewrite.ID
		table.UnexposedParentSchemaID = tableRewrite.ParentSchemaID
		table.ParentID = tableRewrite.ParentID

		if err := tabledesc.ForEachExprStringInTableDesc(table, func(expr *string) error {
			__antithesis_instrumentation__.Notify(267381)
			newExpr, err := rewriteTypesInExpr(*expr, descriptorRewrites)
			if err != nil {
				__antithesis_instrumentation__.Notify(267384)
				return err
			} else {
				__antithesis_instrumentation__.Notify(267385)
			}
			__antithesis_instrumentation__.Notify(267382)
			*expr = newExpr

			newExpr, err = rewriteSequencesInExpr(*expr, descriptorRewrites)
			if err != nil {
				__antithesis_instrumentation__.Notify(267386)
				return err
			} else {
				__antithesis_instrumentation__.Notify(267387)
			}
			__antithesis_instrumentation__.Notify(267383)
			*expr = newExpr
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(267388)
			return err
		} else {
			__antithesis_instrumentation__.Notify(267389)
		}
		__antithesis_instrumentation__.Notify(267362)

		if table.IsView() {
			__antithesis_instrumentation__.Notify(267390)
			viewQuery, err := rewriteSequencesInView(table.ViewQuery, descriptorRewrites)
			if err != nil {
				__antithesis_instrumentation__.Notify(267392)
				return err
			} else {
				__antithesis_instrumentation__.Notify(267393)
			}
			__antithesis_instrumentation__.Notify(267391)
			table.ViewQuery = viewQuery
		} else {
			__antithesis_instrumentation__.Notify(267394)
		}
		__antithesis_instrumentation__.Notify(267363)

		origFKs := table.OutboundFKs
		table.OutboundFKs = nil
		for i := range origFKs {
			__antithesis_instrumentation__.Notify(267395)
			fk := &origFKs[i]
			to := fk.ReferencedTableID
			if indexRewrite, ok := descriptorRewrites[to]; ok {
				__antithesis_instrumentation__.Notify(267397)
				fk.ReferencedTableID = indexRewrite.ID
				fk.OriginTableID = tableRewrite.ID
			} else {
				__antithesis_instrumentation__.Notify(267398)

				continue
			}
			__antithesis_instrumentation__.Notify(267396)

			table.OutboundFKs = append(table.OutboundFKs, *fk)
		}
		__antithesis_instrumentation__.Notify(267364)

		origInboundFks := table.InboundFKs
		table.InboundFKs = nil
		for i := range origInboundFks {
			__antithesis_instrumentation__.Notify(267399)
			ref := &origInboundFks[i]
			if refRewrite, ok := descriptorRewrites[ref.OriginTableID]; ok {
				__antithesis_instrumentation__.Notify(267400)
				ref.ReferencedTableID = tableRewrite.ID
				ref.OriginTableID = refRewrite.ID
				table.InboundFKs = append(table.InboundFKs, *ref)
			} else {
				__antithesis_instrumentation__.Notify(267401)
			}
		}
		__antithesis_instrumentation__.Notify(267365)

		for i, dest := range table.DependsOn {
			__antithesis_instrumentation__.Notify(267402)
			if depRewrite, ok := descriptorRewrites[dest]; ok {
				__antithesis_instrumentation__.Notify(267403)
				table.DependsOn[i] = depRewrite.ID
			} else {
				__antithesis_instrumentation__.Notify(267404)

				return errors.AssertionFailedf(
					"cannot restore %q because referenced table %d was not found",
					table.Name, dest)
			}
		}
		__antithesis_instrumentation__.Notify(267366)
		for i, dest := range table.DependsOnTypes {
			__antithesis_instrumentation__.Notify(267405)
			if depRewrite, ok := descriptorRewrites[dest]; ok {
				__antithesis_instrumentation__.Notify(267406)
				table.DependsOnTypes[i] = depRewrite.ID
			} else {
				__antithesis_instrumentation__.Notify(267407)

				return errors.AssertionFailedf(
					"cannot restore %q because referenced type %d was not found",
					table.Name, dest)
			}
		}
		__antithesis_instrumentation__.Notify(267367)
		origRefs := table.DependedOnBy
		table.DependedOnBy = nil
		for _, ref := range origRefs {
			__antithesis_instrumentation__.Notify(267408)
			if refRewrite, ok := descriptorRewrites[ref.ID]; ok {
				__antithesis_instrumentation__.Notify(267409)
				ref.ID = refRewrite.ID
				table.DependedOnBy = append(table.DependedOnBy, ref)
			} else {
				__antithesis_instrumentation__.Notify(267410)
			}
		}
		__antithesis_instrumentation__.Notify(267368)

		if table.IsSequence() && func() bool {
			__antithesis_instrumentation__.Notify(267411)
			return table.SequenceOpts.HasOwner() == true
		}() == true {
			__antithesis_instrumentation__.Notify(267412)
			if ownerRewrite, ok := descriptorRewrites[table.SequenceOpts.SequenceOwner.OwnerTableID]; ok {
				__antithesis_instrumentation__.Notify(267413)
				table.SequenceOpts.SequenceOwner.OwnerTableID = ownerRewrite.ID
			} else {
				__antithesis_instrumentation__.Notify(267414)

				table.SequenceOpts.SequenceOwner = descpb.TableDescriptor_SequenceOpts_SequenceOwner{}
			}
		} else {
			__antithesis_instrumentation__.Notify(267415)
		}
		__antithesis_instrumentation__.Notify(267369)

		rewriteCol := func(col *descpb.ColumnDescriptor) error {
			__antithesis_instrumentation__.Notify(267416)

			if err := rewriteIDsInTypesT(col.Type, descriptorRewrites); err != nil {
				__antithesis_instrumentation__.Notify(267420)
				return err
			} else {
				__antithesis_instrumentation__.Notify(267421)
			}
			__antithesis_instrumentation__.Notify(267417)
			var newUsedSeqRefs []descpb.ID
			for _, seqID := range col.UsesSequenceIds {
				__antithesis_instrumentation__.Notify(267422)
				if rewrite, ok := descriptorRewrites[seqID]; ok {
					__antithesis_instrumentation__.Notify(267423)
					newUsedSeqRefs = append(newUsedSeqRefs, rewrite.ID)
				} else {
					__antithesis_instrumentation__.Notify(267424)

					newUsedSeqRefs = []descpb.ID{}
					col.DefaultExpr = nil
					break
				}
			}
			__antithesis_instrumentation__.Notify(267418)
			col.UsesSequenceIds = newUsedSeqRefs

			var newOwnedSeqRefs []descpb.ID
			for _, seqID := range col.OwnsSequenceIds {
				__antithesis_instrumentation__.Notify(267425)

				if rewrite, ok := descriptorRewrites[seqID]; ok {
					__antithesis_instrumentation__.Notify(267426)
					newOwnedSeqRefs = append(newOwnedSeqRefs, rewrite.ID)
				} else {
					__antithesis_instrumentation__.Notify(267427)
				}
			}
			__antithesis_instrumentation__.Notify(267419)
			col.OwnsSequenceIds = newOwnedSeqRefs

			return nil
		}
		__antithesis_instrumentation__.Notify(267370)

		for idx := range table.Columns {
			__antithesis_instrumentation__.Notify(267428)
			if err := rewriteCol(&table.Columns[idx]); err != nil {
				__antithesis_instrumentation__.Notify(267429)
				return err
			} else {
				__antithesis_instrumentation__.Notify(267430)
			}
		}
		__antithesis_instrumentation__.Notify(267371)
		for idx := range table.Mutations {
			__antithesis_instrumentation__.Notify(267431)
			if col := table.Mutations[idx].GetColumn(); col != nil {
				__antithesis_instrumentation__.Notify(267432)
				if err := rewriteCol(col); err != nil {
					__antithesis_instrumentation__.Notify(267433)
					return err
				} else {
					__antithesis_instrumentation__.Notify(267434)
				}
			} else {
				__antithesis_instrumentation__.Notify(267435)
			}
		}
	}
	__antithesis_instrumentation__.Notify(267357)

	return nil
}

func rewriteViewQueryDBNames(table *tabledesc.Mutable, newDB string) error {
	__antithesis_instrumentation__.Notify(267436)
	stmt, err := parser.ParseOne(table.ViewQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(267439)
		return pgerror.Wrapf(err, pgcode.Syntax,
			"failed to parse underlying query from view %q", table.Name)
	} else {
		__antithesis_instrumentation__.Notify(267440)
	}
	__antithesis_instrumentation__.Notify(267437)

	f := tree.NewFmtCtx(
		tree.FmtParsable,
		tree.FmtReformatTableNames(func(ctx *tree.FmtCtx, tn *tree.TableName) {
			__antithesis_instrumentation__.Notify(267441)

			if tn.CatalogName != "" {
				__antithesis_instrumentation__.Notify(267443)
				tn.CatalogName = tree.Name(newDB)
			} else {
				__antithesis_instrumentation__.Notify(267444)
			}
			__antithesis_instrumentation__.Notify(267442)
			ctx.WithReformatTableNames(nil, func() {
				__antithesis_instrumentation__.Notify(267445)
				ctx.FormatNode(tn)
			})
		}),
	)
	__antithesis_instrumentation__.Notify(267438)
	f.FormatNode(stmt.AST)
	table.ViewQuery = f.CloseAndGetString()
	return nil
}

func rewriteTypesInExpr(expr string, rewrites jobspb.DescRewriteMap) (string, error) {
	__antithesis_instrumentation__.Notify(267446)
	parsed, err := parser.ParseExpr(expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(267450)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(267451)
	}
	__antithesis_instrumentation__.Notify(267447)

	ctx := tree.NewFmtCtx(
		tree.FmtSerializable,
		tree.FmtIndexedTypeFormat(func(ctx *tree.FmtCtx, ref *tree.OIDTypeReference) {
			__antithesis_instrumentation__.Notify(267452)
			newRef := ref
			var id descpb.ID
			id, err = typedesc.UserDefinedTypeOIDToID(ref.OID)
			if err != nil {
				__antithesis_instrumentation__.Notify(267455)
				return
			} else {
				__antithesis_instrumentation__.Notify(267456)
			}
			__antithesis_instrumentation__.Notify(267453)
			if rw, ok := rewrites[id]; ok {
				__antithesis_instrumentation__.Notify(267457)
				newRef = &tree.OIDTypeReference{OID: typedesc.TypeIDToOID(rw.ID)}
			} else {
				__antithesis_instrumentation__.Notify(267458)
			}
			__antithesis_instrumentation__.Notify(267454)
			ctx.WriteString(newRef.SQLString())
		}),
	)
	__antithesis_instrumentation__.Notify(267448)
	if err != nil {
		__antithesis_instrumentation__.Notify(267459)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(267460)
	}
	__antithesis_instrumentation__.Notify(267449)
	ctx.FormatNode(parsed)
	return ctx.CloseAndGetString(), nil
}

func rewriteSequencesInExpr(expr string, rewrites jobspb.DescRewriteMap) (string, error) {
	__antithesis_instrumentation__.Notify(267461)
	parsed, err := parser.ParseExpr(expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(267465)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(267466)
	}
	__antithesis_instrumentation__.Notify(267462)
	rewriteFunc := func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(267467)
		id, ok := schemaexpr.GetSeqIDFromExpr(expr)
		if !ok {
			__antithesis_instrumentation__.Notify(267471)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(267472)
		}
		__antithesis_instrumentation__.Notify(267468)
		annotateTypeExpr, ok := expr.(*tree.AnnotateTypeExpr)
		if !ok {
			__antithesis_instrumentation__.Notify(267473)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(267474)
		}
		__antithesis_instrumentation__.Notify(267469)

		rewrite, ok := rewrites[descpb.ID(id)]
		if !ok {
			__antithesis_instrumentation__.Notify(267475)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(267476)
		}
		__antithesis_instrumentation__.Notify(267470)
		annotateTypeExpr.Expr = tree.NewNumVal(
			constant.MakeInt64(int64(rewrite.ID)),
			strconv.Itoa(int(rewrite.ID)),
			false,
		)
		return false, annotateTypeExpr, nil
	}
	__antithesis_instrumentation__.Notify(267463)

	newExpr, err := tree.SimpleVisit(parsed, rewriteFunc)
	if err != nil {
		__antithesis_instrumentation__.Notify(267477)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(267478)
	}
	__antithesis_instrumentation__.Notify(267464)
	return newExpr.String(), nil
}

func rewriteSequencesInView(viewQuery string, rewrites jobspb.DescRewriteMap) (string, error) {
	__antithesis_instrumentation__.Notify(267479)
	rewriteFunc := func(expr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		__antithesis_instrumentation__.Notify(267483)
		id, ok := schemaexpr.GetSeqIDFromExpr(expr)
		if !ok {
			__antithesis_instrumentation__.Notify(267487)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(267488)
		}
		__antithesis_instrumentation__.Notify(267484)
		annotateTypeExpr, ok := expr.(*tree.AnnotateTypeExpr)
		if !ok {
			__antithesis_instrumentation__.Notify(267489)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(267490)
		}
		__antithesis_instrumentation__.Notify(267485)
		rewrite, ok := rewrites[descpb.ID(id)]
		if !ok {
			__antithesis_instrumentation__.Notify(267491)
			return true, expr, nil
		} else {
			__antithesis_instrumentation__.Notify(267492)
		}
		__antithesis_instrumentation__.Notify(267486)
		annotateTypeExpr.Expr = tree.NewNumVal(
			constant.MakeInt64(int64(rewrite.ID)),
			strconv.Itoa(int(rewrite.ID)),
			false,
		)
		return false, annotateTypeExpr, nil
	}
	__antithesis_instrumentation__.Notify(267480)

	stmt, err := parser.ParseOne(viewQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(267493)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(267494)
	}
	__antithesis_instrumentation__.Notify(267481)
	newStmt, err := tree.SimpleStmtVisit(stmt.AST, rewriteFunc)
	if err != nil {
		__antithesis_instrumentation__.Notify(267495)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(267496)
	}
	__antithesis_instrumentation__.Notify(267482)
	return newStmt.String(), nil
}

func rewriteIDsInTypesT(typ *types.T, descriptorRewrites jobspb.DescRewriteMap) error {
	__antithesis_instrumentation__.Notify(267497)
	if !typ.UserDefined() {
		__antithesis_instrumentation__.Notify(267503)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(267504)
	}
	__antithesis_instrumentation__.Notify(267498)
	tid, err := typedesc.GetUserDefinedTypeDescID(typ)
	if err != nil {
		__antithesis_instrumentation__.Notify(267505)
		return err
	} else {
		__antithesis_instrumentation__.Notify(267506)
	}
	__antithesis_instrumentation__.Notify(267499)

	var newOID, newArrayOID oid.Oid
	if rw, ok := descriptorRewrites[tid]; ok {
		__antithesis_instrumentation__.Notify(267507)
		newOID = typedesc.TypeIDToOID(rw.ID)
	} else {
		__antithesis_instrumentation__.Notify(267508)
	}
	__antithesis_instrumentation__.Notify(267500)
	if typ.Family() != types.ArrayFamily {
		__antithesis_instrumentation__.Notify(267509)
		tid, err = typedesc.GetUserDefinedArrayTypeDescID(typ)
		if err != nil {
			__antithesis_instrumentation__.Notify(267511)
			return err
		} else {
			__antithesis_instrumentation__.Notify(267512)
		}
		__antithesis_instrumentation__.Notify(267510)
		if rw, ok := descriptorRewrites[tid]; ok {
			__antithesis_instrumentation__.Notify(267513)
			newArrayOID = typedesc.TypeIDToOID(rw.ID)
		} else {
			__antithesis_instrumentation__.Notify(267514)
		}
	} else {
		__antithesis_instrumentation__.Notify(267515)
	}
	__antithesis_instrumentation__.Notify(267501)
	types.RemapUserDefinedTypeOIDs(typ, newOID, newArrayOID)

	if typ.Family() == types.ArrayFamily {
		__antithesis_instrumentation__.Notify(267516)
		if err := rewriteIDsInTypesT(typ.ArrayContents(), descriptorRewrites); err != nil {
			__antithesis_instrumentation__.Notify(267517)
			return err
		} else {
			__antithesis_instrumentation__.Notify(267518)
		}
	} else {
		__antithesis_instrumentation__.Notify(267519)
	}
	__antithesis_instrumentation__.Notify(267502)

	return nil
}

func TypeDescs(types []*typedesc.Mutable, descriptorRewrites jobspb.DescRewriteMap) error {
	__antithesis_instrumentation__.Notify(267520)
	for _, typ := range types {
		__antithesis_instrumentation__.Notify(267522)
		rewrite, ok := descriptorRewrites[typ.ID]
		if !ok {
			__antithesis_instrumentation__.Notify(267526)
			return errors.Errorf("missing rewrite for type %d", typ.ID)
		} else {
			__antithesis_instrumentation__.Notify(267527)
		}
		__antithesis_instrumentation__.Notify(267523)

		typ.Version = 1
		typ.ModificationTime = hlc.Timestamp{}

		if err := rewriteSchemaChangerState(typ, descriptorRewrites); err != nil {
			__antithesis_instrumentation__.Notify(267528)
			return err
		} else {
			__antithesis_instrumentation__.Notify(267529)
		}
		__antithesis_instrumentation__.Notify(267524)

		typ.ID = rewrite.ID
		typ.ParentSchemaID = rewrite.ParentSchemaID
		typ.ParentID = rewrite.ParentID
		for i := range typ.ReferencingDescriptorIDs {
			__antithesis_instrumentation__.Notify(267530)
			id := typ.ReferencingDescriptorIDs[i]
			if rw, ok := descriptorRewrites[id]; ok {
				__antithesis_instrumentation__.Notify(267531)
				typ.ReferencingDescriptorIDs[i] = rw.ID
			} else {
				__antithesis_instrumentation__.Notify(267532)
			}
		}
		__antithesis_instrumentation__.Notify(267525)
		switch t := typ.Kind; t {
		case descpb.TypeDescriptor_ENUM, descpb.TypeDescriptor_MULTIREGION_ENUM:
			__antithesis_instrumentation__.Notify(267533)
			if rw, ok := descriptorRewrites[typ.ArrayTypeID]; ok {
				__antithesis_instrumentation__.Notify(267536)
				typ.ArrayTypeID = rw.ID
			} else {
				__antithesis_instrumentation__.Notify(267537)
			}
		case descpb.TypeDescriptor_ALIAS:
			__antithesis_instrumentation__.Notify(267534)

			if err := rewriteIDsInTypesT(typ.Alias, descriptorRewrites); err != nil {
				__antithesis_instrumentation__.Notify(267538)
				return err
			} else {
				__antithesis_instrumentation__.Notify(267539)
			}
		default:
			__antithesis_instrumentation__.Notify(267535)
			return errors.AssertionFailedf("unknown type kind %s", t.String())
		}
	}
	__antithesis_instrumentation__.Notify(267521)
	return nil
}

func SchemaDescs(schemas []*schemadesc.Mutable, descriptorRewrites jobspb.DescRewriteMap) error {
	__antithesis_instrumentation__.Notify(267540)
	for _, sc := range schemas {
		__antithesis_instrumentation__.Notify(267542)
		rewrite, ok := descriptorRewrites[sc.ID]
		if !ok {
			__antithesis_instrumentation__.Notify(267544)
			return errors.Errorf("missing rewrite for schema %d", sc.ID)
		} else {
			__antithesis_instrumentation__.Notify(267545)
		}
		__antithesis_instrumentation__.Notify(267543)

		sc.Version = 1
		sc.ModificationTime = hlc.Timestamp{}

		sc.ID = rewrite.ID
		sc.ParentID = rewrite.ParentID

		if err := rewriteSchemaChangerState(sc, descriptorRewrites); err != nil {
			__antithesis_instrumentation__.Notify(267546)
			return err
		} else {
			__antithesis_instrumentation__.Notify(267547)
		}
	}
	__antithesis_instrumentation__.Notify(267541)
	return nil
}

func rewriteSchemaChangerState(
	d catalog.MutableDescriptor, descriptorRewrites jobspb.DescRewriteMap,
) (err error) {
	__antithesis_instrumentation__.Notify(267548)
	state := d.GetDeclarativeSchemaChangerState()
	if state == nil {
		__antithesis_instrumentation__.Notify(267553)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(267554)
	}
	__antithesis_instrumentation__.Notify(267549)
	defer func() {
		__antithesis_instrumentation__.Notify(267555)
		if err != nil {
			__antithesis_instrumentation__.Notify(267556)
			err = errors.Wrap(err, "rewriting declarative schema changer state")
		} else {
			__antithesis_instrumentation__.Notify(267557)
		}
	}()
	__antithesis_instrumentation__.Notify(267550)
	for i := 0; i < len(state.Targets); i++ {
		__antithesis_instrumentation__.Notify(267558)
		t := &state.Targets[i]
		if err := screl.WalkDescIDs(t.Element(), func(id *descpb.ID) error {
			__antithesis_instrumentation__.Notify(267561)
			rewrite, ok := descriptorRewrites[*id]
			if !ok {
				__antithesis_instrumentation__.Notify(267563)
				return errors.Errorf("missing rewrite for id %d in %T", *id, t)
			} else {
				__antithesis_instrumentation__.Notify(267564)
			}
			__antithesis_instrumentation__.Notify(267562)
			*id = rewrite.ID
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(267565)

			switch el := t.Element().(type) {
			case *scpb.SchemaParent:
				__antithesis_instrumentation__.Notify(267567)
				_, scExists := descriptorRewrites[el.SchemaID]
				if !scExists && func() bool {
					__antithesis_instrumentation__.Notify(267568)
					return state.CurrentStatuses[i] == scpb.Status_ABSENT == true
				}() == true {
					__antithesis_instrumentation__.Notify(267569)
					state.Targets = append(state.Targets[:i], state.Targets[i+1:]...)
					state.CurrentStatuses = append(state.CurrentStatuses[:i], state.CurrentStatuses[i+1:]...)
					state.TargetRanks = append(state.TargetRanks[:i], state.TargetRanks[i+1:]...)
					i--
					continue
				} else {
					__antithesis_instrumentation__.Notify(267570)
				}
			}
			__antithesis_instrumentation__.Notify(267566)
			return errors.Wrap(err, "rewriting descriptor ids")
		} else {
			__antithesis_instrumentation__.Notify(267571)
		}
		__antithesis_instrumentation__.Notify(267559)

		if err := screl.WalkExpressions(t.Element(), func(expr *catpb.Expression) error {
			__antithesis_instrumentation__.Notify(267572)
			if *expr == "" {
				__antithesis_instrumentation__.Notify(267576)
				return nil
			} else {
				__antithesis_instrumentation__.Notify(267577)
			}
			__antithesis_instrumentation__.Notify(267573)
			newExpr, err := rewriteTypesInExpr(string(*expr), descriptorRewrites)
			if err != nil {
				__antithesis_instrumentation__.Notify(267578)
				return errors.Wrapf(err, "rewriting expression type references: %q", *expr)
			} else {
				__antithesis_instrumentation__.Notify(267579)
			}
			__antithesis_instrumentation__.Notify(267574)
			newExpr, err = rewriteSequencesInExpr(newExpr, descriptorRewrites)
			if err != nil {
				__antithesis_instrumentation__.Notify(267580)
				return errors.Wrapf(err, "rewriting expression sequence references: %q", newExpr)
			} else {
				__antithesis_instrumentation__.Notify(267581)
			}
			__antithesis_instrumentation__.Notify(267575)
			*expr = catpb.Expression(newExpr)
			return nil
		}); err != nil {
			__antithesis_instrumentation__.Notify(267582)
			return err
		} else {
			__antithesis_instrumentation__.Notify(267583)
		}
		__antithesis_instrumentation__.Notify(267560)
		if err := screl.WalkTypes(t.Element(), func(t *types.T) error {
			__antithesis_instrumentation__.Notify(267584)
			return rewriteIDsInTypesT(t, descriptorRewrites)
		}); err != nil {
			__antithesis_instrumentation__.Notify(267585)
			return errors.Wrap(err, "rewriting user-defined type references")
		} else {
			__antithesis_instrumentation__.Notify(267586)
		}

	}
	__antithesis_instrumentation__.Notify(267551)
	if len(state.Targets) == 0 {
		__antithesis_instrumentation__.Notify(267587)
		d.SetDeclarativeSchemaChangerState(nil)
	} else {
		__antithesis_instrumentation__.Notify(267588)
	}
	__antithesis_instrumentation__.Notify(267552)
	return nil
}

func DatabaseDescs(databases []*dbdesc.Mutable, descriptorRewrites jobspb.DescRewriteMap) error {
	__antithesis_instrumentation__.Notify(267589)
	for _, db := range databases {
		__antithesis_instrumentation__.Notify(267591)
		rewrite, ok := descriptorRewrites[db.ID]
		if !ok {
			__antithesis_instrumentation__.Notify(267597)
			return errors.Errorf("missing rewrite for database %d", db.ID)
		} else {
			__antithesis_instrumentation__.Notify(267598)
		}
		__antithesis_instrumentation__.Notify(267592)
		db.ID = rewrite.ID

		if rewrite.NewDBName != "" {
			__antithesis_instrumentation__.Notify(267599)
			db.Name = rewrite.NewDBName
		} else {
			__antithesis_instrumentation__.Notify(267600)
		}
		__antithesis_instrumentation__.Notify(267593)

		db.Version = 1
		db.ModificationTime = hlc.Timestamp{}

		if err := rewriteSchemaChangerState(db, descriptorRewrites); err != nil {
			__antithesis_instrumentation__.Notify(267601)
			return err
		} else {
			__antithesis_instrumentation__.Notify(267602)
		}
		__antithesis_instrumentation__.Notify(267594)

		newSchemas := make(map[string]descpb.DatabaseDescriptor_SchemaInfo)
		err := db.ForEachNonDroppedSchema(func(id descpb.ID, name string) error {
			__antithesis_instrumentation__.Notify(267603)
			rewrite, ok := descriptorRewrites[id]
			if !ok {
				__antithesis_instrumentation__.Notify(267605)
				return errors.Errorf("missing rewrite for schema %d", db.ID)
			} else {
				__antithesis_instrumentation__.Notify(267606)
			}
			__antithesis_instrumentation__.Notify(267604)
			newSchemas[name] = descpb.DatabaseDescriptor_SchemaInfo{ID: rewrite.ID}
			return nil
		})
		__antithesis_instrumentation__.Notify(267595)
		if err != nil {
			__antithesis_instrumentation__.Notify(267607)
			return err
		} else {
			__antithesis_instrumentation__.Notify(267608)
		}
		__antithesis_instrumentation__.Notify(267596)
		db.Schemas = newSchemas
	}
	__antithesis_instrumentation__.Notify(267590)
	return nil
}
