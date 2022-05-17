package sql

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/dbdesc"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/seqexpr"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/privilege"
	"github.com/cockroachdb/cockroach/pkg/sql/roleoption"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/sqlerrors"
	"github.com/cockroachdb/cockroach/pkg/util"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/eventpb"
	"github.com/cockroachdb/errors"
)

type renameDatabaseNode struct {
	n       *tree.RenameDatabase
	dbDesc  *dbdesc.Mutable
	newName string
}

func (p *planner) RenameDatabase(ctx context.Context, n *tree.RenameDatabase) (planNode, error) {
	__antithesis_instrumentation__.Notify(565740)
	if err := checkSchemaChangeEnabled(
		ctx,
		p.ExecCfg(),
		"ALTER DATABASE",
	); err != nil {
		__antithesis_instrumentation__.Notify(565748)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565749)
	}
	__antithesis_instrumentation__.Notify(565741)

	if n.Name == "" || func() bool {
		__antithesis_instrumentation__.Notify(565750)
		return n.NewName == "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(565751)
		return nil, errEmptyDatabaseName
	} else {
		__antithesis_instrumentation__.Notify(565752)
	}
	__antithesis_instrumentation__.Notify(565742)

	if string(n.Name) == p.SessionData().Database && func() bool {
		__antithesis_instrumentation__.Notify(565753)
		return p.SessionData().SafeUpdates == true
	}() == true {
		__antithesis_instrumentation__.Notify(565754)
		return nil, pgerror.DangerousStatementf("RENAME DATABASE on current database")
	} else {
		__antithesis_instrumentation__.Notify(565755)
	}
	__antithesis_instrumentation__.Notify(565743)

	dbDesc, err := p.Descriptors().GetMutableDatabaseByName(ctx, p.txn, string(n.Name),
		tree.DatabaseLookupFlags{Required: true})
	if err != nil {
		__antithesis_instrumentation__.Notify(565756)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565757)
	}
	__antithesis_instrumentation__.Notify(565744)

	hasAdmin, err := p.HasAdminRole(ctx)
	if err != nil {
		__antithesis_instrumentation__.Notify(565758)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(565759)
	}
	__antithesis_instrumentation__.Notify(565745)

	if hasAdmin {
		__antithesis_instrumentation__.Notify(565760)

		if err := p.CheckPrivilege(ctx, dbDesc, privilege.DROP); err != nil {
			__antithesis_instrumentation__.Notify(565761)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(565762)
		}
	} else {
		__antithesis_instrumentation__.Notify(565763)

		hasOwnership, err := p.HasOwnership(ctx, dbDesc)
		if err != nil {
			__antithesis_instrumentation__.Notify(565767)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(565768)
		}
		__antithesis_instrumentation__.Notify(565764)
		if !hasOwnership {
			__antithesis_instrumentation__.Notify(565769)
			return nil, pgerror.Newf(
				pgcode.InsufficientPrivilege, "must be owner of database %s", n.Name)
		} else {
			__antithesis_instrumentation__.Notify(565770)
		}
		__antithesis_instrumentation__.Notify(565765)
		hasCreateDB, err := p.HasRoleOption(ctx, roleoption.CREATEDB)
		if err != nil {
			__antithesis_instrumentation__.Notify(565771)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(565772)
		}
		__antithesis_instrumentation__.Notify(565766)
		if !hasCreateDB {
			__antithesis_instrumentation__.Notify(565773)
			return nil, pgerror.New(
				pgcode.InsufficientPrivilege, "permission denied to rename database")
		} else {
			__antithesis_instrumentation__.Notify(565774)
		}
	}
	__antithesis_instrumentation__.Notify(565746)

	if n.Name == n.NewName {
		__antithesis_instrumentation__.Notify(565775)

		return newZeroNode(nil), nil
	} else {
		__antithesis_instrumentation__.Notify(565776)
	}
	__antithesis_instrumentation__.Notify(565747)

	return &renameDatabaseNode{
		n:       n,
		dbDesc:  dbDesc,
		newName: string(n.NewName),
	}, nil
}

func (n *renameDatabaseNode) ReadingOwnWrites() { __antithesis_instrumentation__.Notify(565777) }

func (n *renameDatabaseNode) startExec(params runParams) error {
	__antithesis_instrumentation__.Notify(565778)
	p := params.p
	ctx := params.ctx
	dbDesc := n.dbDesc

	lookupFlags := p.CommonLookupFlags(true)

	lookupFlags.AvoidLeased = true
	schemas, err := p.Descriptors().GetSchemasForDatabase(ctx, p.txn, dbDesc)
	if err != nil {
		__antithesis_instrumentation__.Notify(565782)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565783)
	}
	__antithesis_instrumentation__.Notify(565779)
	for _, schema := range schemas {
		__antithesis_instrumentation__.Notify(565784)
		tbNames, _, err := p.Descriptors().GetObjectNamesAndIDs(
			ctx,
			p.txn,
			dbDesc,
			schema,
			tree.DatabaseListFlags{
				CommonLookupFlags: lookupFlags,
				ExplicitPrefix:    true,
			},
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(565786)
			return err
		} else {
			__antithesis_instrumentation__.Notify(565787)
		}
		__antithesis_instrumentation__.Notify(565785)
		lookupFlags.Required = false

		for i := range tbNames {
			__antithesis_instrumentation__.Notify(565788)
			found, tbDesc, err := p.Descriptors().GetImmutableTableByName(
				ctx, p.txn, &tbNames[i], tree.ObjectLookupFlags{CommonLookupFlags: lookupFlags},
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(565791)
				return err
			} else {
				__antithesis_instrumentation__.Notify(565792)
			}
			__antithesis_instrumentation__.Notify(565789)
			if !found {
				__antithesis_instrumentation__.Notify(565793)
				continue
			} else {
				__antithesis_instrumentation__.Notify(565794)
			}
			__antithesis_instrumentation__.Notify(565790)

			if err := tbDesc.ForeachDependedOnBy(func(dependedOn *descpb.TableDescriptor_Reference) error {
				__antithesis_instrumentation__.Notify(565795)
				dependentDesc, err := p.Descriptors().Direct().MustGetTableDescByID(ctx, p.txn, dependedOn.ID)
				if err != nil {
					__antithesis_instrumentation__.Notify(565801)
					return err
				} else {
					__antithesis_instrumentation__.Notify(565802)
				}
				__antithesis_instrumentation__.Notify(565796)

				isAllowed, referencedCol, err := isAllowedDependentDescInRenameDatabase(
					ctx,
					dependedOn,
					tbDesc,
					dependentDesc,
					dbDesc.GetName(),
				)
				if err != nil {
					__antithesis_instrumentation__.Notify(565803)
					return err
				} else {
					__antithesis_instrumentation__.Notify(565804)
				}
				__antithesis_instrumentation__.Notify(565797)
				if isAllowed {
					__antithesis_instrumentation__.Notify(565805)
					return nil
				} else {
					__antithesis_instrumentation__.Notify(565806)
				}
				__antithesis_instrumentation__.Notify(565798)

				tbTableName := tree.MakeTableNameWithSchema(
					tree.Name(dbDesc.GetName()),
					tree.Name(schema),
					tree.Name(tbDesc.GetName()),
				)
				var dependentDescQualifiedString string
				if dbDesc.GetID() != dependentDesc.GetParentID() || func() bool {
					__antithesis_instrumentation__.Notify(565807)
					return tbDesc.GetParentSchemaID() != dependentDesc.GetParentSchemaID() == true
				}() == true {
					__antithesis_instrumentation__.Notify(565808)
					descFQName, err := p.getQualifiedTableName(ctx, dependentDesc)
					if err != nil {
						__antithesis_instrumentation__.Notify(565810)
						log.Warningf(
							ctx,
							"unable to retrieve fully-qualified name of %s (id: %d): %v",
							tbTableName.String(),
							dependentDesc.GetID(),
							err,
						)
						return sqlerrors.NewDependentObjectErrorf(
							"cannot rename database because a relation depends on relation %q",
							tbTableName.String())
					} else {
						__antithesis_instrumentation__.Notify(565811)
					}
					__antithesis_instrumentation__.Notify(565809)
					dependentDescQualifiedString = descFQName.FQString()
				} else {
					__antithesis_instrumentation__.Notify(565812)
					dependentDescTableName := tree.MakeTableNameWithSchema(
						tree.Name(dbDesc.GetName()),
						tree.Name(schema),
						tree.Name(dependentDesc.GetName()),
					)
					dependentDescQualifiedString = dependentDescTableName.String()
				}
				__antithesis_instrumentation__.Notify(565799)
				depErr := sqlerrors.NewDependentObjectErrorf(
					"cannot rename database because relation %q depends on relation %q",
					dependentDescQualifiedString,
					tbTableName.String(),
				)

				if tbDesc.IsSequence() {
					__antithesis_instrumentation__.Notify(565813)
					hint := fmt.Sprintf(
						"you can drop the column default %q of %q referencing %q",
						referencedCol,
						tbTableName.String(),
						dependentDescQualifiedString,
					)
					if dependentDesc.GetParentID() == dbDesc.GetID() {
						__antithesis_instrumentation__.Notify(565815)
						hint += fmt.Sprintf(
							" or modify the default to not reference the database name %q",
							dbDesc.GetName(),
						)
					} else {
						__antithesis_instrumentation__.Notify(565816)
					}
					__antithesis_instrumentation__.Notify(565814)
					return errors.WithHint(depErr, hint)
				} else {
					__antithesis_instrumentation__.Notify(565817)
				}
				__antithesis_instrumentation__.Notify(565800)

				return errors.WithHintf(depErr,
					"you can drop %q instead", dependentDescQualifiedString)
			}); err != nil {
				__antithesis_instrumentation__.Notify(565818)
				return err
			} else {
				__antithesis_instrumentation__.Notify(565819)
			}
		}
	}
	__antithesis_instrumentation__.Notify(565780)

	if err := p.renameDatabase(ctx, dbDesc, n.newName, tree.AsStringWithFQNames(n.n, params.Ann())); err != nil {
		__antithesis_instrumentation__.Notify(565820)
		return err
	} else {
		__antithesis_instrumentation__.Notify(565821)
	}
	__antithesis_instrumentation__.Notify(565781)

	return p.logEvent(ctx,
		n.dbDesc.GetID(),
		&eventpb.RenameDatabase{
			DatabaseName:    n.n.Name.String(),
			NewDatabaseName: n.newName,
		})
}

func isAllowedDependentDescInRenameDatabase(
	ctx context.Context,
	dependedOn *descpb.TableDescriptor_Reference,
	tbDesc catalog.TableDescriptor,
	dependentDesc catalog.TableDescriptor,
	dbName string,
) (bool, string, error) {
	__antithesis_instrumentation__.Notify(565822)

	if !tbDesc.IsSequence() {
		__antithesis_instrumentation__.Notify(565827)
		return false, "", nil
	} else {
		__antithesis_instrumentation__.Notify(565828)
	}
	__antithesis_instrumentation__.Notify(565823)

	colIDs := util.MakeFastIntSet()
	for _, colID := range dependedOn.ColumnIDs {
		__antithesis_instrumentation__.Notify(565829)
		colIDs.Add(int(colID))
	}
	__antithesis_instrumentation__.Notify(565824)

	for _, column := range dependentDesc.PublicColumns() {
		__antithesis_instrumentation__.Notify(565830)
		if !colIDs.Contains(int(column.GetID())) {
			__antithesis_instrumentation__.Notify(565836)
			continue
		} else {
			__antithesis_instrumentation__.Notify(565837)
		}
		__antithesis_instrumentation__.Notify(565831)
		colIDs.Remove(int(column.GetID()))

		if !column.HasDefault() {
			__antithesis_instrumentation__.Notify(565838)
			return false, "", errors.AssertionFailedf(
				"rename_database: expected column id %d in table id %d to have a default expr",
				dependedOn.ID,
				dependentDesc.GetID(),
			)
		} else {
			__antithesis_instrumentation__.Notify(565839)
		}
		__antithesis_instrumentation__.Notify(565832)

		parsedExpr, err := parser.ParseExpr(column.GetDefaultExpr())
		if err != nil {
			__antithesis_instrumentation__.Notify(565840)
			return false, "", err
		} else {
			__antithesis_instrumentation__.Notify(565841)
		}
		__antithesis_instrumentation__.Notify(565833)
		typedExpr, err := tree.TypeCheck(ctx, parsedExpr, nil, column.GetType())
		if err != nil {
			__antithesis_instrumentation__.Notify(565842)
			return false, "", err
		} else {
			__antithesis_instrumentation__.Notify(565843)
		}
		__antithesis_instrumentation__.Notify(565834)
		seqIdentifiers, err := seqexpr.GetUsedSequences(typedExpr)
		if err != nil {
			__antithesis_instrumentation__.Notify(565844)
			return false, "", err
		} else {
			__antithesis_instrumentation__.Notify(565845)
		}
		__antithesis_instrumentation__.Notify(565835)
		for _, seqIdentifier := range seqIdentifiers {
			__antithesis_instrumentation__.Notify(565846)
			if seqIdentifier.IsByID() {
				__antithesis_instrumentation__.Notify(565849)
				continue
			} else {
				__antithesis_instrumentation__.Notify(565850)
			}
			__antithesis_instrumentation__.Notify(565847)
			parsedSeqName, err := parser.ParseTableName(seqIdentifier.SeqName)
			if err != nil {
				__antithesis_instrumentation__.Notify(565851)
				return false, "", err
			} else {
				__antithesis_instrumentation__.Notify(565852)
			}
			__antithesis_instrumentation__.Notify(565848)

			if parsedSeqName.NumParts >= 2 {
				__antithesis_instrumentation__.Notify(565853)

				if tree.Name(parsedSeqName.Parts[parsedSeqName.NumParts-1]).Normalize() == tree.Name(dbName).Normalize() {
					__antithesis_instrumentation__.Notify(565854)
					return false, column.GetName(), nil
				} else {
					__antithesis_instrumentation__.Notify(565855)
				}
			} else {
				__antithesis_instrumentation__.Notify(565856)
			}
		}
	}
	__antithesis_instrumentation__.Notify(565825)
	if colIDs.Len() > 0 {
		__antithesis_instrumentation__.Notify(565857)
		return false, "", errors.AssertionFailedf(
			"expected to find column ids %s in table id %d",
			colIDs.String(),
			dependentDesc.GetID(),
		)
	} else {
		__antithesis_instrumentation__.Notify(565858)
	}
	__antithesis_instrumentation__.Notify(565826)
	return true, "", nil
}

func (n *renameDatabaseNode) Next(runParams) (bool, error) {
	__antithesis_instrumentation__.Notify(565859)
	return false, nil
}
func (n *renameDatabaseNode) Values() tree.Datums {
	__antithesis_instrumentation__.Notify(565860)
	return tree.Datums{}
}
func (n *renameDatabaseNode) Close(context.Context) { __antithesis_instrumentation__.Notify(565861) }
