package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"net/url"

	"github.com/cockroachdb/cockroach/pkg/ccl/backupccl/backupresolver"
	"github.com/cockroachdb/cockroach/pkg/ccl/changefeedccl/changefeedbase"
	"github.com/cockroachdb/cockroach/pkg/jobs"
	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/colinfo"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/parser"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
	"github.com/cockroachdb/cockroach/pkg/util/hlc"
	"github.com/cockroachdb/errors"
)

func init() {
	sql.AddPlanHook("alter changefeed", alterChangefeedPlanHook)
}

func alterChangefeedPlanHook(
	ctx context.Context, stmt tree.Statement, p sql.PlanHookState,
) (sql.PlanHookRowFn, colinfo.ResultColumns, []sql.PlanNode, bool, error) {
	__antithesis_instrumentation__.Notify(14092)
	alterChangefeedStmt, ok := stmt.(*tree.AlterChangefeed)
	if !ok {
		__antithesis_instrumentation__.Notify(14095)
		return nil, nil, nil, false, nil
	} else {
		__antithesis_instrumentation__.Notify(14096)
	}
	__antithesis_instrumentation__.Notify(14093)

	header := colinfo.ResultColumns{
		{Name: "job_id", Typ: types.Int},
		{Name: "job_description", Typ: types.String},
	}
	lockForUpdate := false

	fn := func(ctx context.Context, _ []sql.PlanNode, resultsCh chan<- tree.Datums) error {
		__antithesis_instrumentation__.Notify(14097)
		if err := validateSettings(ctx, p); err != nil {
			__antithesis_instrumentation__.Notify(14110)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14111)
		}
		__antithesis_instrumentation__.Notify(14098)

		typedExpr, err := alterChangefeedStmt.Jobs.TypeCheck(ctx, p.SemaCtx(), types.Int)
		if err != nil {
			__antithesis_instrumentation__.Notify(14112)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14113)
		}
		__antithesis_instrumentation__.Notify(14099)
		jobID := jobspb.JobID(tree.MustBeDInt(typedExpr))

		job, err := p.ExecCfg().JobRegistry.LoadJobWithTxn(ctx, jobID, p.ExtendedEvalContext().Txn)
		if err != nil {
			__antithesis_instrumentation__.Notify(14114)
			err = errors.Wrapf(err, `could not load job with job id %d`, jobID)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14115)
		}
		__antithesis_instrumentation__.Notify(14100)

		prevDetails, ok := job.Details().(jobspb.ChangefeedDetails)
		if !ok {
			__antithesis_instrumentation__.Notify(14116)
			return errors.Errorf(`job %d is not changefeed job`, jobID)
		} else {
			__antithesis_instrumentation__.Notify(14117)
		}
		__antithesis_instrumentation__.Notify(14101)

		if job.Status() != jobs.StatusPaused {
			__antithesis_instrumentation__.Notify(14118)
			return errors.Errorf(`job %d is not paused`, jobID)
		} else {
			__antithesis_instrumentation__.Notify(14119)
		}
		__antithesis_instrumentation__.Notify(14102)

		newChangefeedStmt := &tree.CreateChangefeed{}

		prevOpts, err := getPrevOpts(job.Payload().Description, prevDetails.Opts)
		if err != nil {
			__antithesis_instrumentation__.Notify(14120)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14121)
		}
		__antithesis_instrumentation__.Notify(14103)
		newOptions, newSinkURI, err := generateNewOpts(ctx, p, alterChangefeedStmt.Cmds, prevOpts, prevDetails.SinkURI)
		if err != nil {
			__antithesis_instrumentation__.Notify(14122)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14123)
		}
		__antithesis_instrumentation__.Notify(14104)

		newTargets, newProgress, newStatementTime, originalSpecs, err := generateNewTargets(ctx,
			p,
			alterChangefeedStmt.Cmds,
			newOptions,
			prevDetails,
			job.Progress(),
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(14124)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14125)
		}
		__antithesis_instrumentation__.Notify(14105)
		newChangefeedStmt.Targets = newTargets

		for key, value := range newOptions {
			__antithesis_instrumentation__.Notify(14126)
			opt := tree.KVOption{Key: tree.Name(key)}
			if len(value) > 0 {
				__antithesis_instrumentation__.Notify(14128)
				opt.Value = tree.NewDString(value)
			} else {
				__antithesis_instrumentation__.Notify(14129)
			}
			__antithesis_instrumentation__.Notify(14127)
			newChangefeedStmt.Options = append(newChangefeedStmt.Options, opt)
		}
		__antithesis_instrumentation__.Notify(14106)
		newChangefeedStmt.SinkURI = tree.NewDString(newSinkURI)

		annotatedStmt := &annotatedChangefeedStatement{
			CreateChangefeed: newChangefeedStmt,
			originalSpecs:    originalSpecs,
		}

		jobRecord, err := createChangefeedJobRecord(
			ctx,
			p,
			annotatedStmt,
			newSinkURI,
			newOptions,
			jobID,
			``,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(14130)
			return errors.Wrap(err, `failed to alter changefeed`)
		} else {
			__antithesis_instrumentation__.Notify(14131)
		}
		__antithesis_instrumentation__.Notify(14107)

		newDetails := jobRecord.Details.(jobspb.ChangefeedDetails)
		newDetails.Opts[changefeedbase.OptInitialScan] = ``

		newDetails.StatementTime = newStatementTime

		newPayload := job.Payload()
		newPayload.Details = jobspb.WrapPayloadDetails(newDetails)
		newPayload.Description = jobRecord.Description
		newPayload.DescriptorIDs = jobRecord.DescriptorIDs

		err = p.ExecCfg().JobRegistry.UpdateJobWithTxn(ctx, jobID, p.ExtendedEvalContext().Txn, lockForUpdate, func(
			txn *kv.Txn, md jobs.JobMetadata, ju *jobs.JobUpdater,
		) error {
			__antithesis_instrumentation__.Notify(14132)
			ju.UpdatePayload(&newPayload)
			if newProgress != nil {
				__antithesis_instrumentation__.Notify(14134)
				ju.UpdateProgress(newProgress)
			} else {
				__antithesis_instrumentation__.Notify(14135)
			}
			__antithesis_instrumentation__.Notify(14133)
			return nil
		})
		__antithesis_instrumentation__.Notify(14108)

		if err != nil {
			__antithesis_instrumentation__.Notify(14136)
			return err
		} else {
			__antithesis_instrumentation__.Notify(14137)
		}
		__antithesis_instrumentation__.Notify(14109)

		select {
		case <-ctx.Done():
			__antithesis_instrumentation__.Notify(14138)
			return ctx.Err()
		case resultsCh <- tree.Datums{
			tree.NewDInt(tree.DInt(jobID)),
			tree.NewDString(jobRecord.Description),
		}:
			__antithesis_instrumentation__.Notify(14139)
			return nil
		}
	}
	__antithesis_instrumentation__.Notify(14094)

	return fn, header, nil, false, nil
}

func getTargetDesc(
	ctx context.Context,
	p sql.PlanHookState,
	descResolver *backupresolver.DescriptorResolver,
	targetPattern tree.TablePattern,
) (catalog.Descriptor, bool, error) {
	__antithesis_instrumentation__.Notify(14140)
	pattern, err := targetPattern.NormalizeTablePattern()
	if err != nil {
		__antithesis_instrumentation__.Notify(14144)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(14145)
	}
	__antithesis_instrumentation__.Notify(14141)
	targetName, ok := pattern.(*tree.TableName)
	if !ok {
		__antithesis_instrumentation__.Notify(14146)
		return nil, false, errors.Errorf(`CHANGEFEED cannot target %q`, tree.AsString(targetPattern))
	} else {
		__antithesis_instrumentation__.Notify(14147)
	}
	__antithesis_instrumentation__.Notify(14142)

	found, _, desc, err := resolver.ResolveExisting(
		ctx,
		targetName.ToUnresolvedObjectName(),
		descResolver,
		tree.ObjectLookupFlags{},
		p.CurrentDatabase(),
		p.CurrentSearchPath(),
	)
	if err != nil {
		__antithesis_instrumentation__.Notify(14148)
		return nil, false, err
	} else {
		__antithesis_instrumentation__.Notify(14149)
	}
	__antithesis_instrumentation__.Notify(14143)

	return desc, found, nil
}

func generateNewOpts(
	ctx context.Context,
	p sql.PlanHookState,
	alterCmds tree.AlterChangefeedCmds,
	prevOpts map[string]string,
	prevSinkURI string,
) (map[string]string, string, error) {
	__antithesis_instrumentation__.Notify(14150)
	sinkURI := prevSinkURI
	newOptions := prevOpts

	for _, cmd := range alterCmds {
		__antithesis_instrumentation__.Notify(14152)
		switch v := cmd.(type) {
		case *tree.AlterChangefeedSetOptions:
			__antithesis_instrumentation__.Notify(14153)
			optsFn, err := p.TypeAsStringOpts(ctx, v.Options, changefeedbase.AlterChangefeedOptionExpectValues)
			if err != nil {
				__antithesis_instrumentation__.Notify(14157)
				return nil, ``, err
			} else {
				__antithesis_instrumentation__.Notify(14158)
			}
			__antithesis_instrumentation__.Notify(14154)

			opts, err := optsFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(14159)
				return nil, ``, err
			} else {
				__antithesis_instrumentation__.Notify(14160)
			}
			__antithesis_instrumentation__.Notify(14155)

			for key, value := range opts {
				__antithesis_instrumentation__.Notify(14161)
				if _, ok := changefeedbase.AlterChangefeedUnsupportedOptions[key]; ok {
					__antithesis_instrumentation__.Notify(14163)
					return nil, ``, pgerror.Newf(pgcode.InvalidParameterValue, `cannot alter option %q`, key)
				} else {
					__antithesis_instrumentation__.Notify(14164)
				}
				__antithesis_instrumentation__.Notify(14162)
				if key == changefeedbase.OptSink {
					__antithesis_instrumentation__.Notify(14165)
					newSinkURI, err := url.Parse(value)
					if err != nil {
						__antithesis_instrumentation__.Notify(14169)
						return nil, ``, err
					} else {
						__antithesis_instrumentation__.Notify(14170)
					}
					__antithesis_instrumentation__.Notify(14166)

					prevSinkURI, err := url.Parse(sinkURI)
					if err != nil {
						__antithesis_instrumentation__.Notify(14171)
						return nil, ``, err
					} else {
						__antithesis_instrumentation__.Notify(14172)
					}
					__antithesis_instrumentation__.Notify(14167)

					if newSinkURI.Scheme != prevSinkURI.Scheme {
						__antithesis_instrumentation__.Notify(14173)
						return nil, ``, pgerror.Newf(
							pgcode.InvalidParameterValue,
							`New sink type %q does not match original sink type %q. `+
								`Altering the sink type of a changefeed is disallowed, consider creating a new changefeed instead.`,
							newSinkURI.Scheme,
							prevSinkURI.Scheme,
						)
					} else {
						__antithesis_instrumentation__.Notify(14174)
					}
					__antithesis_instrumentation__.Notify(14168)

					sinkURI = value
				} else {
					__antithesis_instrumentation__.Notify(14175)
					newOptions[key] = value
				}
			}
		case *tree.AlterChangefeedUnsetOptions:
			__antithesis_instrumentation__.Notify(14156)
			optKeys := v.Options.ToStrings()
			for _, key := range optKeys {
				__antithesis_instrumentation__.Notify(14176)
				if key == changefeedbase.OptSink {
					__antithesis_instrumentation__.Notify(14180)
					return nil, ``, pgerror.Newf(pgcode.InvalidParameterValue, `cannot unset option %q`, key)
				} else {
					__antithesis_instrumentation__.Notify(14181)
				}
				__antithesis_instrumentation__.Notify(14177)
				if _, ok := changefeedbase.ChangefeedOptionExpectValues[key]; !ok {
					__antithesis_instrumentation__.Notify(14182)
					return nil, ``, pgerror.Newf(pgcode.InvalidParameterValue, `invalid option %q`, key)
				} else {
					__antithesis_instrumentation__.Notify(14183)
				}
				__antithesis_instrumentation__.Notify(14178)
				if _, ok := changefeedbase.AlterChangefeedUnsupportedOptions[key]; ok {
					__antithesis_instrumentation__.Notify(14184)
					return nil, ``, pgerror.Newf(pgcode.InvalidParameterValue, `cannot alter option %q`, key)
				} else {
					__antithesis_instrumentation__.Notify(14185)
				}
				__antithesis_instrumentation__.Notify(14179)
				delete(newOptions, key)
			}
		}
	}
	__antithesis_instrumentation__.Notify(14151)

	return newOptions, sinkURI, nil
}

func generateNewTargets(
	ctx context.Context,
	p sql.PlanHookState,
	alterCmds tree.AlterChangefeedCmds,
	opts map[string]string,
	prevDetails jobspb.ChangefeedDetails,
	prevProgress jobspb.Progress,
) (
	tree.ChangefeedTargets,
	*jobspb.Progress,
	hlc.Timestamp,
	map[tree.ChangefeedTarget]jobspb.ChangefeedTargetSpecification,
	error,
) {
	__antithesis_instrumentation__.Notify(14186)

	type targetKey struct {
		TableID    descpb.ID
		FamilyName tree.Name
	}
	newTargets := make(map[targetKey]tree.ChangefeedTarget)
	droppedTargets := make(map[targetKey]tree.ChangefeedTarget)
	newTableDescs := make(map[descpb.ID]catalog.Descriptor)

	originalSpecs := make(map[tree.ChangefeedTarget]jobspb.ChangefeedTargetSpecification)

	delete(opts, changefeedbase.OptNoInitialScan)
	delete(opts, changefeedbase.OptInitialScan)

	newJobProgress := prevProgress
	newJobStatementTime := prevDetails.StatementTime

	statementTime := hlc.Timestamp{
		WallTime: p.ExtendedEvalContext().GetStmtTimestamp().UnixNano(),
	}

	allDescs, err := backupresolver.LoadAllDescs(ctx, p.ExecCfg(), statementTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(14194)
		return nil, nil, hlc.Timestamp{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(14195)
	}
	__antithesis_instrumentation__.Notify(14187)
	descResolver, err := backupresolver.NewDescriptorResolver(allDescs)
	if err != nil {
		__antithesis_instrumentation__.Notify(14196)
		return nil, nil, hlc.Timestamp{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(14197)
	}
	__antithesis_instrumentation__.Notify(14188)

	for _, targetSpec := range AllTargets(prevDetails) {
		__antithesis_instrumentation__.Notify(14198)
		k := targetKey{TableID: targetSpec.TableID, FamilyName: tree.Name(targetSpec.FamilyName)}
		desc := descResolver.DescByID[targetSpec.TableID].(catalog.TableDescriptor)

		tbName, err := getQualifiedTableNameObj(ctx, p.ExecCfg(), p.ExtendedEvalContext().Txn, desc)
		if err != nil {
			__antithesis_instrumentation__.Notify(14201)
			return nil, nil, hlc.Timestamp{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(14202)
		}
		__antithesis_instrumentation__.Notify(14199)

		tablePattern, err := tbName.NormalizeTablePattern()
		if err != nil {
			__antithesis_instrumentation__.Notify(14203)
			return nil, nil, hlc.Timestamp{}, nil, err
		} else {
			__antithesis_instrumentation__.Notify(14204)
		}
		__antithesis_instrumentation__.Notify(14200)

		newTarget := tree.ChangefeedTarget{
			TableName:  tablePattern,
			FamilyName: tree.Name(targetSpec.FamilyName),
		}
		newTargets[k] = newTarget
		newTableDescs[targetSpec.TableID] = descResolver.DescByID[targetSpec.TableID]

		originalSpecs[newTarget] = targetSpec
	}
	__antithesis_instrumentation__.Notify(14189)

	for _, cmd := range alterCmds {
		__antithesis_instrumentation__.Notify(14205)
		switch v := cmd.(type) {
		case *tree.AlterChangefeedAddTarget:
			__antithesis_instrumentation__.Notify(14206)
			targetOptsFn, err := p.TypeAsStringOpts(ctx, v.Options, changefeedbase.AlterChangefeedTargetOptions)
			if err != nil {
				__antithesis_instrumentation__.Notify(14215)
				return nil, nil, hlc.Timestamp{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(14216)
			}
			__antithesis_instrumentation__.Notify(14207)
			targetOpts, err := targetOptsFn()
			if err != nil {
				__antithesis_instrumentation__.Notify(14217)
				return nil, nil, hlc.Timestamp{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(14218)
			}
			__antithesis_instrumentation__.Notify(14208)

			_, withInitialScan := targetOpts[changefeedbase.OptInitialScan]
			_, noInitialScan := targetOpts[changefeedbase.OptNoInitialScan]
			if withInitialScan && func() bool {
				__antithesis_instrumentation__.Notify(14219)
				return noInitialScan == true
			}() == true {
				__antithesis_instrumentation__.Notify(14220)
				return nil, nil, hlc.Timestamp{}, nil, pgerror.Newf(
					pgcode.InvalidParameterValue,
					`cannot specify both %q and %q`, changefeedbase.OptInitialScan,
					changefeedbase.OptNoInitialScan,
				)
			} else {
				__antithesis_instrumentation__.Notify(14221)
			}
			__antithesis_instrumentation__.Notify(14209)

			var existingTargetDescs []catalog.Descriptor
			for _, targetDesc := range newTableDescs {
				__antithesis_instrumentation__.Notify(14222)
				existingTargetDescs = append(existingTargetDescs, targetDesc)
			}
			__antithesis_instrumentation__.Notify(14210)
			existingTargetSpans, err := fetchSpansForDescs(ctx, p, statementTime, existingTargetDescs)
			if err != nil {
				__antithesis_instrumentation__.Notify(14223)
				return nil, nil, hlc.Timestamp{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(14224)
			}
			__antithesis_instrumentation__.Notify(14211)

			var newTargetDescs []catalog.Descriptor
			for _, target := range v.Targets {
				__antithesis_instrumentation__.Notify(14225)
				desc, found, err := getTargetDesc(ctx, p, descResolver, target.TableName)
				if err != nil {
					__antithesis_instrumentation__.Notify(14228)
					return nil, nil, hlc.Timestamp{}, nil, err
				} else {
					__antithesis_instrumentation__.Notify(14229)
				}
				__antithesis_instrumentation__.Notify(14226)
				if !found {
					__antithesis_instrumentation__.Notify(14230)
					return nil, nil, hlc.Timestamp{}, nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						`target %q does not exist`,
						tree.ErrString(&target),
					)
				} else {
					__antithesis_instrumentation__.Notify(14231)
				}
				__antithesis_instrumentation__.Notify(14227)
				k := targetKey{TableID: desc.GetID(), FamilyName: target.FamilyName}
				newTargets[k] = target
				newTableDescs[desc.GetID()] = desc
				newTargetDescs = append(newTargetDescs, desc)
			}
			__antithesis_instrumentation__.Notify(14212)

			addedTargetSpans, err := fetchSpansForDescs(ctx, p, statementTime, newTargetDescs)
			if err != nil {
				__antithesis_instrumentation__.Notify(14232)
				return nil, nil, hlc.Timestamp{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(14233)
			}
			__antithesis_instrumentation__.Notify(14213)

			newJobProgress, newJobStatementTime, err = generateNewProgress(
				newJobProgress,
				newJobStatementTime,
				existingTargetSpans,
				addedTargetSpans,
				withInitialScan,
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(14234)
				return nil, nil, hlc.Timestamp{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(14235)
			}
		case *tree.AlterChangefeedDropTarget:
			__antithesis_instrumentation__.Notify(14214)
			for _, target := range v.Targets {
				__antithesis_instrumentation__.Notify(14236)
				desc, found, err := getTargetDesc(ctx, p, descResolver, target.TableName)
				if err != nil {
					__antithesis_instrumentation__.Notify(14240)
					return nil, nil, hlc.Timestamp{}, nil, err
				} else {
					__antithesis_instrumentation__.Notify(14241)
				}
				__antithesis_instrumentation__.Notify(14237)
				if !found {
					__antithesis_instrumentation__.Notify(14242)
					return nil, nil, hlc.Timestamp{}, nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						`target %q does not exist`,
						tree.ErrString(&target),
					)
				} else {
					__antithesis_instrumentation__.Notify(14243)
				}
				__antithesis_instrumentation__.Notify(14238)
				k := targetKey{TableID: desc.GetID(), FamilyName: target.FamilyName}
				droppedTargets[k] = target
				_, recognized := newTargets[k]
				if !recognized {
					__antithesis_instrumentation__.Notify(14244)
					return nil, nil, hlc.Timestamp{}, nil, pgerror.Newf(
						pgcode.InvalidParameterValue,
						`target %q already not watched by changefeed`,
						tree.ErrString(&target),
					)
				} else {
					__antithesis_instrumentation__.Notify(14245)
				}
				__antithesis_instrumentation__.Notify(14239)
				delete(newTargets, k)
			}
		}
	}
	__antithesis_instrumentation__.Notify(14190)

	if len(droppedTargets) > 0 {
		__antithesis_instrumentation__.Notify(14246)
		stillThere := make(map[descpb.ID]bool)
		for k := range newTargets {
			__antithesis_instrumentation__.Notify(14250)
			stillThere[k.TableID] = true
		}
		__antithesis_instrumentation__.Notify(14247)
		for k := range droppedTargets {
			__antithesis_instrumentation__.Notify(14251)
			if !stillThere[k.TableID] {
				__antithesis_instrumentation__.Notify(14252)
				stillThere[k.TableID] = false
			} else {
				__antithesis_instrumentation__.Notify(14253)
			}
		}
		__antithesis_instrumentation__.Notify(14248)
		var droppedTargetDescs []catalog.Descriptor
		for id, there := range stillThere {
			__antithesis_instrumentation__.Notify(14254)
			if !there {
				__antithesis_instrumentation__.Notify(14255)
				droppedTargetDescs = append(droppedTargetDescs, descResolver.DescByID[id])
			} else {
				__antithesis_instrumentation__.Notify(14256)
			}
		}
		__antithesis_instrumentation__.Notify(14249)
		if len(droppedTargetDescs) > 0 {
			__antithesis_instrumentation__.Notify(14257)
			droppedTargetSpans, err := fetchSpansForDescs(ctx, p, statementTime, droppedTargetDescs)
			if err != nil {
				__antithesis_instrumentation__.Notify(14259)
				return nil, nil, hlc.Timestamp{}, nil, err
			} else {
				__antithesis_instrumentation__.Notify(14260)
			}
			__antithesis_instrumentation__.Notify(14258)
			removeSpansFromProgress(newJobProgress, droppedTargetSpans)
		} else {
			__antithesis_instrumentation__.Notify(14261)
		}
	} else {
		__antithesis_instrumentation__.Notify(14262)
	}
	__antithesis_instrumentation__.Notify(14191)

	newTargetList := tree.ChangefeedTargets{}

	for _, target := range newTargets {
		__antithesis_instrumentation__.Notify(14263)
		newTargetList = append(newTargetList, target)
	}
	__antithesis_instrumentation__.Notify(14192)

	if err := validateNewTargets(ctx, p, newTargetList, newJobProgress, newJobStatementTime); err != nil {
		__antithesis_instrumentation__.Notify(14264)
		return nil, nil, hlc.Timestamp{}, nil, err
	} else {
		__antithesis_instrumentation__.Notify(14265)
	}
	__antithesis_instrumentation__.Notify(14193)

	return newTargetList, &newJobProgress, newJobStatementTime, originalSpecs, nil
}

func validateNewTargets(
	ctx context.Context,
	p sql.PlanHookState,
	newTargets tree.ChangefeedTargets,
	jobProgress jobspb.Progress,
	jobStatementTime hlc.Timestamp,
) error {
	__antithesis_instrumentation__.Notify(14266)
	if len(newTargets) == 0 {
		__antithesis_instrumentation__.Notify(14272)
		return pgerror.New(pgcode.InvalidParameterValue, "cannot drop all targets")
	} else {
		__antithesis_instrumentation__.Notify(14273)
	}
	__antithesis_instrumentation__.Notify(14267)

	var resolveTime hlc.Timestamp
	highWater := jobProgress.GetHighWater()
	if highWater != nil && func() bool {
		__antithesis_instrumentation__.Notify(14274)
		return !highWater.IsEmpty() == true
	}() == true {
		__antithesis_instrumentation__.Notify(14275)
		resolveTime = *highWater
	} else {
		__antithesis_instrumentation__.Notify(14276)
		resolveTime = jobStatementTime
	}
	__antithesis_instrumentation__.Notify(14268)

	allDescs, err := backupresolver.LoadAllDescs(ctx, p.ExecCfg(), resolveTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(14277)
		return errors.Wrap(err, `error while validating new targets`)
	} else {
		__antithesis_instrumentation__.Notify(14278)
	}
	__antithesis_instrumentation__.Notify(14269)
	descResolver, err := backupresolver.NewDescriptorResolver(allDescs)
	if err != nil {
		__antithesis_instrumentation__.Notify(14279)
		return errors.Wrap(err, `error while validating new targets`)
	} else {
		__antithesis_instrumentation__.Notify(14280)
	}
	__antithesis_instrumentation__.Notify(14270)

	for _, target := range newTargets {
		__antithesis_instrumentation__.Notify(14281)
		targetName := target.TableName
		_, found, err := getTargetDesc(ctx, p, descResolver, targetName)
		if err != nil {
			__antithesis_instrumentation__.Notify(14283)
			return errors.Wrap(err, `error while validating new targets`)
		} else {
			__antithesis_instrumentation__.Notify(14284)
		}
		__antithesis_instrumentation__.Notify(14282)
		if !found {
			__antithesis_instrumentation__.Notify(14285)
			if highWater != nil && func() bool {
				__antithesis_instrumentation__.Notify(14287)
				return !highWater.IsEmpty() == true
			}() == true {
				__antithesis_instrumentation__.Notify(14288)
				return errors.Errorf(`target %q cannot be resolved as of the high water mark. `+
					`Please wait until the high water mark progresses past the creation time of this target in order to add it to the changefeed.`,
					tree.ErrString(targetName),
				)
			} else {
				__antithesis_instrumentation__.Notify(14289)
			}
			__antithesis_instrumentation__.Notify(14286)
			return errors.Errorf(`target %q cannot be resolved as of the creation time of the changefeed. `+
				`Please wait until the high water mark progresses past the creation time of this target in order to add it to the changefeed.`,
				tree.ErrString(targetName),
			)
		} else {
			__antithesis_instrumentation__.Notify(14290)
		}
	}
	__antithesis_instrumentation__.Notify(14271)

	return nil
}

func generateNewProgress(
	prevProgress jobspb.Progress,
	prevStatementTime hlc.Timestamp,
	existingTargetSpans []roachpb.Span,
	newSpans []roachpb.Span,
	withInitialScan bool,
) (jobspb.Progress, hlc.Timestamp, error) {
	__antithesis_instrumentation__.Notify(14291)
	prevHighWater := prevProgress.GetHighWater()
	changefeedProgress := prevProgress.GetChangefeed()

	haveHighwater := !(prevHighWater == nil || func() bool {
		__antithesis_instrumentation__.Notify(14296)
		return prevHighWater.IsEmpty() == true
	}() == true)
	haveCheckpoint := changefeedProgress != nil && func() bool {
		__antithesis_instrumentation__.Notify(14297)
		return changefeedProgress.Checkpoint != nil == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(14298)
		return len(changefeedProgress.Checkpoint.Spans) != 0 == true
	}() == true

	if (!haveHighwater && func() bool {
		__antithesis_instrumentation__.Notify(14299)
		return withInitialScan == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(14300)
		return (haveHighwater && func() bool {
			__antithesis_instrumentation__.Notify(14301)
			return !haveCheckpoint == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(14302)
			return !withInitialScan == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(14303)
		return prevProgress, prevStatementTime, nil
	} else {
		__antithesis_instrumentation__.Notify(14304)
	}
	__antithesis_instrumentation__.Notify(14292)

	if haveHighwater && func() bool {
		__antithesis_instrumentation__.Notify(14305)
		return haveCheckpoint == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(14306)
		return withInitialScan == true
	}() == true {
		__antithesis_instrumentation__.Notify(14307)
		return prevProgress, prevStatementTime, errors.Errorf(
			`cannot perform initial scan on newly added targets while the checkpoint is non-empty, `+
				`please unpause the changefeed and wait until the high watermark progresses past the current value %s to add these targets.`,
			tree.TimestampToDecimalDatum(*prevHighWater).Decimal.String(),
		)
	} else {
		__antithesis_instrumentation__.Notify(14308)
	}
	__antithesis_instrumentation__.Notify(14293)

	if haveHighwater && func() bool {
		__antithesis_instrumentation__.Notify(14309)
		return !haveCheckpoint == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(14310)
		return withInitialScan == true
	}() == true {
		__antithesis_instrumentation__.Notify(14311)

		newStatementTime := *prevHighWater

		newProgress := jobspb.Progress{
			Progress: &jobspb.Progress_HighWater{},
			Details: &jobspb.Progress_Changefeed{
				Changefeed: &jobspb.ChangefeedProgress{
					Checkpoint: &jobspb.ChangefeedProgress_Checkpoint{
						Spans: existingTargetSpans,
					},
				},
			},
		}

		return newProgress, newStatementTime, nil
	} else {
		__antithesis_instrumentation__.Notify(14312)
	}
	__antithesis_instrumentation__.Notify(14294)

	var mergedSpanGroup roachpb.SpanGroup
	if haveCheckpoint {
		__antithesis_instrumentation__.Notify(14313)
		mergedSpanGroup.Add(changefeedProgress.Checkpoint.Spans...)
	} else {
		__antithesis_instrumentation__.Notify(14314)
	}
	__antithesis_instrumentation__.Notify(14295)
	mergedSpanGroup.Add(newSpans...)

	newProgress := jobspb.Progress{
		Progress: &jobspb.Progress_HighWater{},
		Details: &jobspb.Progress_Changefeed{
			Changefeed: &jobspb.ChangefeedProgress{
				Checkpoint: &jobspb.ChangefeedProgress_Checkpoint{
					Spans: mergedSpanGroup.Slice(),
				},
			},
		},
	}
	return newProgress, prevStatementTime, nil
}

func removeSpansFromProgress(prevProgress jobspb.Progress, spansToRemove []roachpb.Span) {
	__antithesis_instrumentation__.Notify(14315)
	changefeedProgress := prevProgress.GetChangefeed()
	if changefeedProgress == nil {
		__antithesis_instrumentation__.Notify(14318)
		return
	} else {
		__antithesis_instrumentation__.Notify(14319)
	}
	__antithesis_instrumentation__.Notify(14316)
	changefeedCheckpoint := changefeedProgress.Checkpoint
	if changefeedCheckpoint == nil {
		__antithesis_instrumentation__.Notify(14320)
		return
	} else {
		__antithesis_instrumentation__.Notify(14321)
	}
	__antithesis_instrumentation__.Notify(14317)
	prevSpans := changefeedCheckpoint.Spans

	var spanGroup roachpb.SpanGroup
	spanGroup.Add(prevSpans...)
	spanGroup.Sub(spansToRemove...)
	changefeedProgress.Checkpoint.Spans = spanGroup.Slice()
}

func fetchSpansForDescs(
	ctx context.Context, p sql.PlanHookState, statementTime hlc.Timestamp, descs []catalog.Descriptor,
) ([]roachpb.Span, error) {
	__antithesis_instrumentation__.Notify(14322)
	targets := make([]jobspb.ChangefeedTargetSpecification, len(descs))
	for i, d := range descs {
		__antithesis_instrumentation__.Notify(14324)
		targets[i] = jobspb.ChangefeedTargetSpecification{TableID: d.GetID()}
	}
	__antithesis_instrumentation__.Notify(14323)

	spans, err := fetchSpansForTargets(ctx, p.ExecCfg(), targets, statementTime)

	return spans, err
}

func getPrevOpts(prevDescription string, opts map[string]string) (map[string]string, error) {
	__antithesis_instrumentation__.Notify(14325)
	prevStmt, err := parser.ParseOne(prevDescription)
	if err != nil {
		__antithesis_instrumentation__.Notify(14329)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(14330)
	}
	__antithesis_instrumentation__.Notify(14326)

	prevChangefeedStmt, ok := prevStmt.AST.(*tree.CreateChangefeed)
	if !ok {
		__antithesis_instrumentation__.Notify(14331)
		return nil, errors.Errorf(`could not parse job description`)
	} else {
		__antithesis_instrumentation__.Notify(14332)
	}
	__antithesis_instrumentation__.Notify(14327)

	prevOpts := make(map[string]string, len(prevChangefeedStmt.Options))
	for _, opt := range prevChangefeedStmt.Options {
		__antithesis_instrumentation__.Notify(14333)
		prevOpts[opt.Key.String()] = opts[opt.Key.String()]
	}
	__antithesis_instrumentation__.Notify(14328)

	return prevOpts, nil
}
