package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/base"
	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/cockroach/pkg/util/log/channel"
	"github.com/cockroachdb/cockroach/pkg/util/log/logconfig"
	"github.com/cockroachdb/cockroach/pkg/util/log/severity"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

func setupLogging(ctx context.Context, cmd *cobra.Command, isServerCmd, applyConfig bool) error {
	__antithesis_instrumentation__.Notify(33257)

	if cliCtx.deprecatedLogOverrides.anySet() && func() bool {
		__antithesis_instrumentation__.Notify(33276)
		return cliCtx.logConfigInput.isSet == true
	}() == true {
		__antithesis_instrumentation__.Notify(33277)
		return errors.Newf("--%s is incompatible with legacy discrete logging flags", cliflags.Log.Name)
	} else {
		__antithesis_instrumentation__.Notify(33278)
	}
	__antithesis_instrumentation__.Notify(33258)

	if active, firstUse := log.IsActive(); active {
		__antithesis_instrumentation__.Notify(33279)
		panic(errors.Newf("logging already active; first used at:\n%s", firstUse))
	} else {
		__antithesis_instrumentation__.Notify(33280)
	}
	__antithesis_instrumentation__.Notify(33259)

	var firstStoreDir *string
	var ambiguousLogDirs bool
	if isServerCmd {
		__antithesis_instrumentation__.Notify(33281)
		firstStoreDir, ambiguousLogDirs = getDefaultLogDirFromStores()
	} else {
		__antithesis_instrumentation__.Notify(33282)
	}
	__antithesis_instrumentation__.Notify(33260)
	defaultLogDir := firstStoreDir

	forceDisableLogDir := cliCtx.deprecatedLogOverrides.logDir.isSet && func() bool {
		__antithesis_instrumentation__.Notify(33283)
		return cliCtx.deprecatedLogOverrides.logDir.s == "" == true
	}() == true
	if forceDisableLogDir {
		__antithesis_instrumentation__.Notify(33284)
		defaultLogDir = nil
	} else {
		__antithesis_instrumentation__.Notify(33285)
	}
	__antithesis_instrumentation__.Notify(33261)
	forceSetLogDir := cliCtx.deprecatedLogOverrides.logDir.isSet && func() bool {
		__antithesis_instrumentation__.Notify(33286)
		return cliCtx.deprecatedLogOverrides.logDir.s != "" == true
	}() == true
	if forceSetLogDir {
		__antithesis_instrumentation__.Notify(33287)
		ambiguousLogDirs = false
		defaultLogDir = &cliCtx.deprecatedLogOverrides.logDir.s
	} else {
		__antithesis_instrumentation__.Notify(33288)
	}
	__antithesis_instrumentation__.Notify(33262)

	h := logconfig.Holder{Config: logconfig.DefaultConfig()}
	if !isServerCmd || func() bool {
		__antithesis_instrumentation__.Notify(33289)
		return defaultLogDir == nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(33290)

		h.Config = logconfig.DefaultStderrConfig()
	} else {
		__antithesis_instrumentation__.Notify(33291)
	}
	__antithesis_instrumentation__.Notify(33263)

	commandSpecificDefaultLegacyStderrOverride := severity.INFO

	if isDemoCmd(cmd) {
		__antithesis_instrumentation__.Notify(33292)

		h.Config.Sinks.Stderr.Filter = severity.NONE
	} else {
		__antithesis_instrumentation__.Notify(33293)
		if !isServerCmd && func() bool {
			__antithesis_instrumentation__.Notify(33294)
			return !isWorkloadCmd(cmd) == true
		}() == true {
			__antithesis_instrumentation__.Notify(33295)

			h.Config.Sinks.Stderr.Filter = severity.WARNING

			commandSpecificDefaultLegacyStderrOverride = severity.WARNING
		} else {
			__antithesis_instrumentation__.Notify(33296)
		}
	}
	__antithesis_instrumentation__.Notify(33264)

	cliCtx.deprecatedLogOverrides.propagate(&h.Config, commandSpecificDefaultLegacyStderrOverride)

	if cliCtx.logConfigInput.isSet {
		__antithesis_instrumentation__.Notify(33297)
		if err := h.Set(cliCtx.logConfigInput.s); err != nil {
			__antithesis_instrumentation__.Notify(33299)
			return err
		} else {
			__antithesis_instrumentation__.Notify(33300)
		}
		__antithesis_instrumentation__.Notify(33298)
		if h.Config.FileDefaults.Dir != nil {
			__antithesis_instrumentation__.Notify(33301)
			ambiguousLogDirs = false
		} else {
			__antithesis_instrumentation__.Notify(33302)
		}
	} else {
		__antithesis_instrumentation__.Notify(33303)
	}
	__antithesis_instrumentation__.Notify(33265)

	if isServerCmd && func() bool {
		__antithesis_instrumentation__.Notify(33304)
		return len(h.Config.Sinks.FileGroups) == 0 == true
	}() == true {
		__antithesis_instrumentation__.Notify(33305)
		addPredefinedLogFiles(&h.Config)
	} else {
		__antithesis_instrumentation__.Notify(33306)
	}
	__antithesis_instrumentation__.Notify(33266)

	if err := h.Config.Validate(defaultLogDir); err != nil {
		__antithesis_instrumentation__.Notify(33307)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33308)
	}
	__antithesis_instrumentation__.Notify(33267)

	cliCtx.logConfig = h.Config

	if ambiguousLogDirs && func() bool {
		__antithesis_instrumentation__.Notify(33309)
		return firstStoreDir != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(33310)
		firstStoreDirUsed := false
		if firstStoreAbs, err := filepath.Abs(*firstStoreDir); err == nil {
			__antithesis_instrumentation__.Notify(33312)
			_ = h.Config.IterateDirectories(func(logDir string) error {
				__antithesis_instrumentation__.Notify(33313)
				firstStoreDirUsed = firstStoreDirUsed || func() bool {
					__antithesis_instrumentation__.Notify(33314)
					return logDir == firstStoreAbs == true
				}() == true
				return nil
			})
		} else {
			__antithesis_instrumentation__.Notify(33315)
		}
		__antithesis_instrumentation__.Notify(33311)
		if firstStoreDirUsed {
			__antithesis_instrumentation__.Notify(33316)
			cliCtx.ambiguousLogDir = true
		} else {
			__antithesis_instrumentation__.Notify(33317)
		}
	} else {
		__antithesis_instrumentation__.Notify(33318)
	}
	__antithesis_instrumentation__.Notify(33268)

	if !applyConfig {
		__antithesis_instrumentation__.Notify(33319)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(33320)
	}
	__antithesis_instrumentation__.Notify(33269)

	if err := h.Config.IterateDirectories(func(logDir string) error {
		__antithesis_instrumentation__.Notify(33321)
		return os.MkdirAll(logDir, 0755)
	}); err != nil {
		__antithesis_instrumentation__.Notify(33322)
		return errors.Wrap(err, "unable to create log directory")
	} else {
		__antithesis_instrumentation__.Notify(33323)
	}
	__antithesis_instrumentation__.Notify(33270)

	if _, err := log.ApplyConfig(h.Config); err != nil {
		__antithesis_instrumentation__.Notify(33324)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33325)
	}
	__antithesis_instrumentation__.Notify(33271)

	if cliCtx.logConfigInput.isSet {
		__antithesis_instrumentation__.Notify(33326)
		log.Ops.Infof(ctx, "using explicit logging configuration:\n%s", cliCtx.logConfigInput.s)
	} else {
		__antithesis_instrumentation__.Notify(33327)
	}
	__antithesis_instrumentation__.Notify(33272)

	if cliCtx.ambiguousLogDir {
		__antithesis_instrumentation__.Notify(33328)

		log.Ops.Shout(ctx, severity.WARNING,
			"multiple stores configured, "+
				"you may want to specify --log='file-defaults: {dir: ...}' to disambiguate.")
	} else {
		__antithesis_instrumentation__.Notify(33329)
	}
	__antithesis_instrumentation__.Notify(33273)

	outputDirectory := "."
	if firstStoreDir != nil {
		__antithesis_instrumentation__.Notify(33330)
		outputDirectory = *firstStoreDir
	} else {
		__antithesis_instrumentation__.Notify(33331)
	}
	__antithesis_instrumentation__.Notify(33274)
	for _, fc := range h.Config.Sinks.FileGroups {
		__antithesis_instrumentation__.Notify(33332)
		if fc.Channels.AllChannels.HasChannel(channel.DEV) && func() bool {
			__antithesis_instrumentation__.Notify(33333)
			return fc.Dir != nil == true
		}() == true && func() bool {
			__antithesis_instrumentation__.Notify(33334)
			return *fc.Dir != "" == true
		}() == true {
			__antithesis_instrumentation__.Notify(33335)
			outputDirectory = *fc.Dir
			break
		} else {
			__antithesis_instrumentation__.Notify(33336)
		}
	}
	__antithesis_instrumentation__.Notify(33275)
	serverCfg.GoroutineDumpDirName = filepath.Join(outputDirectory, base.GoroutineDumpDir)
	serverCfg.HeapProfileDirName = filepath.Join(outputDirectory, base.HeapProfileDir)
	serverCfg.CPUProfileDirName = filepath.Join(outputDirectory, base.CPUProfileDir)
	serverCfg.InflightTraceDirName = filepath.Join(outputDirectory, base.InflightTraceDir)

	return nil
}

func getDefaultLogDirFromStores() (dir *string, ambiguousLogDirs bool) {
	__antithesis_instrumentation__.Notify(33337)

	for _, spec := range serverCfg.Stores.Specs {
		__antithesis_instrumentation__.Notify(33339)
		if spec.InMemory {
			__antithesis_instrumentation__.Notify(33342)
			continue
		} else {
			__antithesis_instrumentation__.Notify(33343)
		}
		__antithesis_instrumentation__.Notify(33340)
		if dir != nil {
			__antithesis_instrumentation__.Notify(33344)
			ambiguousLogDirs = true
			break
		} else {
			__antithesis_instrumentation__.Notify(33345)
		}
		__antithesis_instrumentation__.Notify(33341)
		s := filepath.Join(spec.Path, "logs")
		dir = &s
	}
	__antithesis_instrumentation__.Notify(33338)

	return
}

type logConfigFlags struct {
	logDir settableString

	sqlAuditLogDir settableString

	fileMaxSize    int64
	fileMaxSizeVal *humanizeutil.BytesValue

	maxGroupSize    int64
	maxGroupSizeVal *humanizeutil.BytesValue

	fileThreshold log.Severity

	stderrThreshold log.Severity

	stderrNoColor settableBool

	redactableLogs settableBool
}

func newLogConfigOverrides() *logConfigFlags {
	__antithesis_instrumentation__.Notify(33346)
	l := &logConfigFlags{}
	l.fileMaxSizeVal = humanizeutil.NewBytesValue(&l.fileMaxSize)
	l.maxGroupSizeVal = humanizeutil.NewBytesValue(&l.maxGroupSize)
	l.reset()
	return l
}

func (l *logConfigFlags) reset() {
	__antithesis_instrumentation__.Notify(33347)
	d := logconfig.DefaultConfig()

	l.logDir = settableString{}
	l.sqlAuditLogDir = settableString{}
	*l.fileMaxSizeVal = *humanizeutil.NewBytesValue(&l.fileMaxSize)
	*l.maxGroupSizeVal = *humanizeutil.NewBytesValue(&l.maxGroupSize)
	l.fileMaxSize = int64(*d.FileDefaults.MaxFileSize)
	l.maxGroupSize = int64(*d.FileDefaults.MaxGroupSize)
	l.fileThreshold = severity.UNKNOWN
	l.stderrThreshold = severity.UNKNOWN
	l.stderrNoColor = settableBool{}
	l.redactableLogs = settableBool{}
}

func (l *logConfigFlags) anySet() bool {
	__antithesis_instrumentation__.Notify(33348)
	return l.logDir.isSet || func() bool {
		__antithesis_instrumentation__.Notify(33349)
		return l.sqlAuditLogDir.isSet == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(33350)
		return l.fileMaxSizeVal.IsSet() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(33351)
		return l.maxGroupSizeVal.IsSet() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(33352)
		return l.fileThreshold.IsSet() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(33353)
		return l.stderrThreshold.IsSet() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(33354)
		return l.stderrNoColor.isSet == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(33355)
		return l.redactableLogs.isSet == true
	}() == true
}

func (l *logConfigFlags) propagate(
	c *logconfig.Config, commandSpecificDefaultLegacyStderrOverride log.Severity,
) {
	__antithesis_instrumentation__.Notify(33356)
	if l.logDir.isSet && func() bool {
		__antithesis_instrumentation__.Notify(33363)
		return l.logDir.s != "" == true
	}() == true {
		__antithesis_instrumentation__.Notify(33364)

		c.FileDefaults.Dir = &l.logDir.s
	} else {
		__antithesis_instrumentation__.Notify(33365)
	}
	__antithesis_instrumentation__.Notify(33357)
	if l.fileMaxSizeVal.IsSet() {
		__antithesis_instrumentation__.Notify(33366)
		s := logconfig.ByteSize(l.fileMaxSize)
		c.FileDefaults.MaxFileSize = &s
	} else {
		__antithesis_instrumentation__.Notify(33367)
	}
	__antithesis_instrumentation__.Notify(33358)
	if l.maxGroupSizeVal.IsSet() {
		__antithesis_instrumentation__.Notify(33368)
		s := logconfig.ByteSize(l.maxGroupSize)
		c.FileDefaults.MaxGroupSize = &s
	} else {
		__antithesis_instrumentation__.Notify(33369)
	}
	__antithesis_instrumentation__.Notify(33359)
	if l.fileThreshold.IsSet() {
		__antithesis_instrumentation__.Notify(33370)
		c.FileDefaults.Filter = l.fileThreshold
	} else {
		__antithesis_instrumentation__.Notify(33371)
	}
	__antithesis_instrumentation__.Notify(33360)
	if l.stderrThreshold.IsSet() {
		__antithesis_instrumentation__.Notify(33372)
		if l.stderrThreshold == severity.DEFAULT {
			__antithesis_instrumentation__.Notify(33373)
			c.Sinks.Stderr.Filter = commandSpecificDefaultLegacyStderrOverride
		} else {
			__antithesis_instrumentation__.Notify(33374)
			c.Sinks.Stderr.Filter = l.stderrThreshold
		}
	} else {
		__antithesis_instrumentation__.Notify(33375)
	}
	__antithesis_instrumentation__.Notify(33361)
	if l.stderrNoColor.isSet {
		__antithesis_instrumentation__.Notify(33376)
		c.Sinks.Stderr.NoColor = l.stderrNoColor.val
	} else {
		__antithesis_instrumentation__.Notify(33377)
	}
	__antithesis_instrumentation__.Notify(33362)
	if l.redactableLogs.isSet {
		__antithesis_instrumentation__.Notify(33378)
		c.FileDefaults.Redactable = &l.redactableLogs.val
		c.Sinks.Stderr.Redactable = &l.redactableLogs.val
	} else {
		__antithesis_instrumentation__.Notify(33379)
	}
}

type settableString struct {
	s     string
	isSet bool
}

type stringValue struct {
	*settableString
}

var _ flag.Value = (*stringValue)(nil)
var _ flag.Value = (*fileContentsValue)(nil)

func (l *stringValue) Set(s string) error {
	__antithesis_instrumentation__.Notify(33380)
	l.s = s
	l.isSet = true
	return nil
}

func (l stringValue) Type() string { __antithesis_instrumentation__.Notify(33381); return "<string>" }

func (l stringValue) String() string { __antithesis_instrumentation__.Notify(33382); return l.s }

type fileContentsValue struct {
	*settableString
	fileName string
}

func (l *fileContentsValue) Set(s string) error {
	__antithesis_instrumentation__.Notify(33383)
	l.fileName = s
	b, err := ioutil.ReadFile(s)
	if err != nil {
		__antithesis_instrumentation__.Notify(33385)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33386)
	}
	__antithesis_instrumentation__.Notify(33384)
	l.s = string(b)
	l.isSet = true
	return nil
}

func (l fileContentsValue) Type() string {
	__antithesis_instrumentation__.Notify(33387)
	return "<file>"
}

func (l fileContentsValue) String() string {
	__antithesis_instrumentation__.Notify(33388)
	return l.fileName
}

type settableBool struct {
	val   bool
	isSet bool
}

func (s settableBool) String() string {
	__antithesis_instrumentation__.Notify(33389)
	return strconv.FormatBool(s.val)
}

func (s settableBool) Type() string { __antithesis_instrumentation__.Notify(33390); return "bool" }

func (s *settableBool) Set(v string) error {
	__antithesis_instrumentation__.Notify(33391)
	b, err := strconv.ParseBool(v)
	if err != nil {
		__antithesis_instrumentation__.Notify(33393)
		return err
	} else {
		__antithesis_instrumentation__.Notify(33394)
	}
	__antithesis_instrumentation__.Notify(33392)
	s.val = b
	s.isSet = true
	return nil
}

func addPredefinedLogFiles(c *logconfig.Config) {
	__antithesis_instrumentation__.Notify(33395)
	h := logconfig.Holder{Config: *c}
	if err := h.Set(predefinedLogFiles); err != nil {
		__antithesis_instrumentation__.Notify(33397)
		panic(errors.NewAssertionErrorWithWrappedErrf(err, "programming error: incorrect config"))
	} else {
		__antithesis_instrumentation__.Notify(33398)
	}
	__antithesis_instrumentation__.Notify(33396)
	*c = h.Config
	if cliCtx.deprecatedLogOverrides.sqlAuditLogDir.isSet {
		__antithesis_instrumentation__.Notify(33399)
		c.Sinks.FileGroups["sql-audit"].Dir = &cliCtx.deprecatedLogOverrides.sqlAuditLogDir.s
	} else {
		__antithesis_instrumentation__.Notify(33400)
	}
}

const predefinedLogFiles = `
sinks:
 file-groups:
  default:
    channels:
      INFO: [DEV, OPS]
      WARNING: all except [DEV, OPS]
  health:                 { channels: HEALTH  }
  pebble:                 { channels: STORAGE }
  security:               { channels: [PRIVILEGES, USER_ADMIN], auditable: true  }
  sql-auth:               { channels: SESSIONS, auditable: true }
  sql-audit:              { channels: SENSITIVE_ACCESS, auditable: true }
  sql-exec:               { channels: SQL_EXEC }
  sql-schema:             { channels: SQL_SCHEMA }
  sql-slow:               { channels: SQL_PERF }
  sql-slow-internal-only: { channels: SQL_INTERNAL_PERF }
  telemetry:
    channels: TELEMETRY
    max-file-size: 102400
    max-group-size: 1048576
`
