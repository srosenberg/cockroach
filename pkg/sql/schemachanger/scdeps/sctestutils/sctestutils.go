package sctestutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	gojson "encoding/json"
	"strings"
	"testing"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descs"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/resolver"
	"github.com/cockroachdb/cockroach/pkg/sql/protoreflect"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scbuild"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scdeps"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scop"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scpb"
	"github.com/cockroachdb/cockroach/pkg/sql/schemachanger/scplan"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondatapb"
	"github.com/cockroachdb/cockroach/pkg/testutils/serverutils"
	jsonb "github.com/cockroachdb/cockroach/pkg/util/json"
	"github.com/cockroachdb/cockroach/pkg/util/protoutil"
	"github.com/kylelemons/godebug/diff"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

func WithBuilderDependenciesFromTestServer(
	s serverutils.TestServerInterface, fn func(scbuild.Dependencies),
) {
	__antithesis_instrumentation__.Notify(581256)
	execCfg := s.ExecutorConfig().(sql.ExecutorConfig)
	ip, cleanup := sql.NewInternalPlanner(
		"test",
		kv.NewTxn(context.Background(), s.DB(), s.NodeID()),
		security.RootUserName(),
		&sql.MemoryMetrics{},
		&execCfg,

		sessiondatapb.SessionData{},
	)
	defer cleanup()
	planner := ip.(interface {
		Txn() *kv.Txn
		Descriptors() *descs.Collection
		SessionData() *sessiondata.SessionData
		resolver.SchemaResolver
		scbuild.AuthorizationAccessor
		scbuild.AstFormatter
		scbuild.FeatureChecker
	})

	planner.SessionData().NewSchemaChangerMode = sessiondatapb.UseNewSchemaChangerUnsafe
	fn(scdeps.NewBuilderDependencies(
		execCfg.LogicalClusterID(),
		execCfg.Codec,
		planner.Txn(),
		planner.Descriptors(),
		planner,
		planner,
		planner,
		planner,
		planner.SessionData(),
		execCfg.Settings,
		nil,
	))
}

func ProtoToYAML(m protoutil.Message) (string, error) {
	__antithesis_instrumentation__.Notify(581257)
	js, err := protoreflect.MessageToJSON(m, protoreflect.FmtFlags{})
	if err != nil {
		__antithesis_instrumentation__.Notify(581262)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(581263)
	}
	__antithesis_instrumentation__.Notify(581258)
	str, err := jsonb.Pretty(js)
	if err != nil {
		__antithesis_instrumentation__.Notify(581264)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(581265)
	}
	__antithesis_instrumentation__.Notify(581259)
	var buf bytes.Buffer
	buf.WriteString(str)
	target := make(map[string]interface{})
	err = gojson.Unmarshal(buf.Bytes(), &target)
	if err != nil {
		__antithesis_instrumentation__.Notify(581266)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(581267)
	}
	__antithesis_instrumentation__.Notify(581260)
	out, err := yaml.Marshal(target)
	if err != nil {
		__antithesis_instrumentation__.Notify(581268)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(581269)
	}
	__antithesis_instrumentation__.Notify(581261)
	return string(out), nil
}

type DiffArgs struct {
	Indent       string
	CompactLevel uint
}

func Diff(a, b string, args DiffArgs) string {
	__antithesis_instrumentation__.Notify(581270)
	d := diff.Diff(a, b)
	lines := strings.Split(d, "\n")

	visible := make(map[int]struct{})
	if args.CompactLevel > 0 {
		__antithesis_instrumentation__.Notify(581273)
		n := int(args.CompactLevel) - 1
		for lineno, line := range lines {
			__antithesis_instrumentation__.Notify(581274)
			if strings.HasPrefix(line, "+") || func() bool {
				__antithesis_instrumentation__.Notify(581275)
				return strings.HasPrefix(line, "-") == true
			}() == true {
				__antithesis_instrumentation__.Notify(581276)
				for i := lineno - n; i <= lineno+n; i++ {
					__antithesis_instrumentation__.Notify(581277)
					visible[i] = struct{}{}
				}
			} else {
				__antithesis_instrumentation__.Notify(581278)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(581279)
	}
	__antithesis_instrumentation__.Notify(581271)

	result := make([]string, 0, len(lines))
	skipping := false
	for lineno, line := range lines {
		__antithesis_instrumentation__.Notify(581280)
		if _, found := visible[lineno]; found || func() bool {
			__antithesis_instrumentation__.Notify(581281)
			return args.CompactLevel == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(581282)
			skipping = false
			result = append(result, args.Indent+line)
		} else {
			__antithesis_instrumentation__.Notify(581283)
			if !skipping {
				__antithesis_instrumentation__.Notify(581284)
				skipping = true
				result = append(result, args.Indent+"...")
			} else {
				__antithesis_instrumentation__.Notify(581285)
			}
		}
	}
	__antithesis_instrumentation__.Notify(581272)
	return strings.Join(result, "\n")
}

func ProtoDiff(a, b protoutil.Message, args DiffArgs) string {
	__antithesis_instrumentation__.Notify(581286)
	toYAML := func(m protoutil.Message) string {
		__antithesis_instrumentation__.Notify(581288)
		if m == nil {
			__antithesis_instrumentation__.Notify(581291)
			return ""
		} else {
			__antithesis_instrumentation__.Notify(581292)
		}
		__antithesis_instrumentation__.Notify(581289)
		str, err := ProtoToYAML(m)
		if err != nil {
			__antithesis_instrumentation__.Notify(581293)
			panic(err)
		} else {
			__antithesis_instrumentation__.Notify(581294)
		}
		__antithesis_instrumentation__.Notify(581290)
		return strings.TrimSpace(str)
	}
	__antithesis_instrumentation__.Notify(581287)

	return Diff(toYAML(a), toYAML(b), args)
}

func MakePlan(t *testing.T, state scpb.CurrentState, phase scop.Phase) scplan.Plan {
	__antithesis_instrumentation__.Notify(581295)
	plan, err := scplan.MakePlan(state, scplan.Params{
		ExecutionPhase:             phase,
		SchemaChangerJobIDSupplier: func() jobspb.JobID { __antithesis_instrumentation__.Notify(581297); return 1 },
	})
	__antithesis_instrumentation__.Notify(581296)
	require.NoError(t, err)
	return plan
}

func TruncateJobOps(plan *scplan.Plan) {
	__antithesis_instrumentation__.Notify(581298)
	for _, s := range plan.Stages {
		__antithesis_instrumentation__.Notify(581299)
		for _, o := range s.ExtraOps {
			__antithesis_instrumentation__.Notify(581300)
			switch op := o.(type) {
			case *scop.SetJobStateOnDescriptor:
				__antithesis_instrumentation__.Notify(581301)
				op.State = scpb.DescriptorState{
					JobID: op.State.JobID,
				}
			case *scop.UpdateSchemaChangerJob:
				__antithesis_instrumentation__.Notify(581302)
				op.RunningStatus = ""
			}
		}
	}
}
