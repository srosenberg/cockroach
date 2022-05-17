package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/errors"
	"github.com/google/go-cmp/cmp"
)

type DescriptorCoverage int32

const (
	RequestedDescriptors DescriptorCoverage = iota

	AllDescriptors
)

type BackupOptions struct {
	CaptureRevisionHistory bool
	EncryptionPassphrase   Expr
	Detached               bool
	EncryptionKMSURI       StringOrPlaceholderOptList
	IncrementalStorage     StringOrPlaceholderOptList
}

var _ NodeFormatter = &BackupOptions{}

type Backup struct {
	Targets *TargetList

	To StringOrPlaceholderOptList

	IncrementalFrom Exprs

	AsOf    AsOfClause
	Options BackupOptions

	Nested bool

	AppendToLatest bool

	Subdir Expr
}

var _ Statement = &Backup{}

func (node *Backup) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603372)
	ctx.WriteString("BACKUP ")
	if node.Targets != nil {
		__antithesis_instrumentation__.Notify(603377)
		ctx.FormatNode(node.Targets)
		ctx.WriteString(" ")
	} else {
		__antithesis_instrumentation__.Notify(603378)
	}
	__antithesis_instrumentation__.Notify(603373)
	if node.Nested {
		__antithesis_instrumentation__.Notify(603379)
		ctx.WriteString("INTO ")
		if node.Subdir != nil {
			__antithesis_instrumentation__.Notify(603380)
			ctx.FormatNode(node.Subdir)
			ctx.WriteString(" IN ")
		} else {
			__antithesis_instrumentation__.Notify(603381)
			if node.AppendToLatest {
				__antithesis_instrumentation__.Notify(603382)
				ctx.WriteString("LATEST IN ")
			} else {
				__antithesis_instrumentation__.Notify(603383)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(603384)
		ctx.WriteString("TO ")
	}
	__antithesis_instrumentation__.Notify(603374)
	ctx.FormatNode(&node.To)
	if node.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(603385)
		ctx.WriteString(" ")
		ctx.FormatNode(&node.AsOf)
	} else {
		__antithesis_instrumentation__.Notify(603386)
	}
	__antithesis_instrumentation__.Notify(603375)
	if node.IncrementalFrom != nil {
		__antithesis_instrumentation__.Notify(603387)
		ctx.WriteString(" INCREMENTAL FROM ")
		ctx.FormatNode(&node.IncrementalFrom)
	} else {
		__antithesis_instrumentation__.Notify(603388)
	}
	__antithesis_instrumentation__.Notify(603376)

	if !node.Options.IsDefault() {
		__antithesis_instrumentation__.Notify(603389)
		ctx.WriteString(" WITH ")
		ctx.FormatNode(&node.Options)
	} else {
		__antithesis_instrumentation__.Notify(603390)
	}
}

func (node Backup) Coverage() DescriptorCoverage {
	__antithesis_instrumentation__.Notify(603391)
	if node.Targets == nil {
		__antithesis_instrumentation__.Notify(603393)
		return AllDescriptors
	} else {
		__antithesis_instrumentation__.Notify(603394)
	}
	__antithesis_instrumentation__.Notify(603392)
	return RequestedDescriptors
}

type RestoreOptions struct {
	EncryptionPassphrase      Expr
	DecryptionKMSURI          StringOrPlaceholderOptList
	IntoDB                    Expr
	SkipMissingFKs            bool
	SkipMissingSequences      bool
	SkipMissingSequenceOwners bool
	SkipMissingViews          bool
	Detached                  bool
	SkipLocalitiesCheck       bool
	DebugPauseOn              Expr
	NewDBName                 Expr
	IncrementalStorage        StringOrPlaceholderOptList
	AsTenant                  Expr
}

var _ NodeFormatter = &RestoreOptions{}

type Restore struct {
	Targets TargetList

	SystemUsers        bool
	DescriptorCoverage DescriptorCoverage

	From    []StringOrPlaceholderOptList
	AsOf    AsOfClause
	Options RestoreOptions

	Subdir Expr
}

var _ Statement = &Restore{}

func (node *Restore) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603395)
	ctx.WriteString("RESTORE ")
	if node.DescriptorCoverage == RequestedDescriptors {
		__antithesis_instrumentation__.Notify(603400)
		ctx.FormatNode(&node.Targets)
		ctx.WriteString(" ")
	} else {
		__antithesis_instrumentation__.Notify(603401)
	}
	__antithesis_instrumentation__.Notify(603396)
	ctx.WriteString("FROM ")
	if node.Subdir != nil {
		__antithesis_instrumentation__.Notify(603402)
		ctx.FormatNode(node.Subdir)
		ctx.WriteString(" IN ")
	} else {
		__antithesis_instrumentation__.Notify(603403)
	}
	__antithesis_instrumentation__.Notify(603397)
	for i := range node.From {
		__antithesis_instrumentation__.Notify(603404)
		if i > 0 {
			__antithesis_instrumentation__.Notify(603406)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(603407)
		}
		__antithesis_instrumentation__.Notify(603405)
		ctx.FormatNode(&node.From[i])
	}
	__antithesis_instrumentation__.Notify(603398)
	if node.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(603408)
		ctx.WriteString(" ")
		ctx.FormatNode(&node.AsOf)
	} else {
		__antithesis_instrumentation__.Notify(603409)
	}
	__antithesis_instrumentation__.Notify(603399)
	if !node.Options.IsDefault() {
		__antithesis_instrumentation__.Notify(603410)
		ctx.WriteString(" WITH ")
		ctx.FormatNode(&node.Options)
	} else {
		__antithesis_instrumentation__.Notify(603411)
	}
}

type KVOption struct {
	Key   Name
	Value Expr
}

type KVOptions []KVOption

func (o *KVOptions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603412)
	for i := range *o {
		__antithesis_instrumentation__.Notify(603413)
		n := &(*o)[i]
		if i > 0 {
			__antithesis_instrumentation__.Notify(603416)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(603417)
		}
		__antithesis_instrumentation__.Notify(603414)

		ctx.WithFlags(ctx.flags&^FmtMarkRedactionNode, func() {
			__antithesis_instrumentation__.Notify(603418)
			ctx.FormatNode(&n.Key)
		})
		__antithesis_instrumentation__.Notify(603415)
		if n.Value != nil {
			__antithesis_instrumentation__.Notify(603419)
			ctx.WriteString(` = `)
			ctx.FormatNode(n.Value)
		} else {
			__antithesis_instrumentation__.Notify(603420)
		}
	}
}

type StringOrPlaceholderOptList []Expr

func (node *StringOrPlaceholderOptList) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603421)
	if len(*node) > 1 {
		__antithesis_instrumentation__.Notify(603423)
		ctx.WriteString("(")
	} else {
		__antithesis_instrumentation__.Notify(603424)
	}
	__antithesis_instrumentation__.Notify(603422)
	ctx.FormatNode((*Exprs)(node))
	if len(*node) > 1 {
		__antithesis_instrumentation__.Notify(603425)
		ctx.WriteString(")")
	} else {
		__antithesis_instrumentation__.Notify(603426)
	}
}

func (o *BackupOptions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603427)
	var addSep bool
	maybeAddSep := func() {
		__antithesis_instrumentation__.Notify(603433)
		if addSep {
			__antithesis_instrumentation__.Notify(603435)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(603436)
		}
		__antithesis_instrumentation__.Notify(603434)
		addSep = true
	}
	__antithesis_instrumentation__.Notify(603428)
	if o.CaptureRevisionHistory {
		__antithesis_instrumentation__.Notify(603437)
		ctx.WriteString("revision_history")
		addSep = true
	} else {
		__antithesis_instrumentation__.Notify(603438)
	}
	__antithesis_instrumentation__.Notify(603429)

	if o.EncryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(603439)
		maybeAddSep()
		ctx.WriteString("encryption_passphrase = ")
		if ctx.flags.HasFlags(FmtShowPasswords) {
			__antithesis_instrumentation__.Notify(603440)
			ctx.FormatNode(o.EncryptionPassphrase)
		} else {
			__antithesis_instrumentation__.Notify(603441)
			ctx.WriteString(PasswordSubstitution)
		}
	} else {
		__antithesis_instrumentation__.Notify(603442)
	}
	__antithesis_instrumentation__.Notify(603430)

	if o.Detached {
		__antithesis_instrumentation__.Notify(603443)
		maybeAddSep()
		ctx.WriteString("detached")
	} else {
		__antithesis_instrumentation__.Notify(603444)
	}
	__antithesis_instrumentation__.Notify(603431)

	if o.EncryptionKMSURI != nil {
		__antithesis_instrumentation__.Notify(603445)
		maybeAddSep()
		ctx.WriteString("kms = ")
		ctx.FormatNode(&o.EncryptionKMSURI)
	} else {
		__antithesis_instrumentation__.Notify(603446)
	}
	__antithesis_instrumentation__.Notify(603432)

	if o.IncrementalStorage != nil {
		__antithesis_instrumentation__.Notify(603447)
		maybeAddSep()
		ctx.WriteString("incremental_location = ")
		ctx.FormatNode(&o.IncrementalStorage)
	} else {
		__antithesis_instrumentation__.Notify(603448)
	}
}

func (o *BackupOptions) CombineWith(other *BackupOptions) error {
	__antithesis_instrumentation__.Notify(603449)
	if o.CaptureRevisionHistory {
		__antithesis_instrumentation__.Notify(603455)
		if other.CaptureRevisionHistory {
			__antithesis_instrumentation__.Notify(603456)
			return errors.New("revision_history option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603457)
		}
	} else {
		__antithesis_instrumentation__.Notify(603458)
		o.CaptureRevisionHistory = other.CaptureRevisionHistory
	}
	__antithesis_instrumentation__.Notify(603450)

	if o.EncryptionPassphrase == nil {
		__antithesis_instrumentation__.Notify(603459)
		o.EncryptionPassphrase = other.EncryptionPassphrase
	} else {
		__antithesis_instrumentation__.Notify(603460)
		if other.EncryptionPassphrase != nil {
			__antithesis_instrumentation__.Notify(603461)
			return errors.New("encryption_passphrase specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603462)
		}
	}
	__antithesis_instrumentation__.Notify(603451)

	if o.Detached {
		__antithesis_instrumentation__.Notify(603463)
		if other.Detached {
			__antithesis_instrumentation__.Notify(603464)
			return errors.New("detached option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603465)
		}
	} else {
		__antithesis_instrumentation__.Notify(603466)
		o.Detached = other.Detached
	}
	__antithesis_instrumentation__.Notify(603452)

	if o.EncryptionKMSURI == nil {
		__antithesis_instrumentation__.Notify(603467)
		o.EncryptionKMSURI = other.EncryptionKMSURI
	} else {
		__antithesis_instrumentation__.Notify(603468)
		if other.EncryptionKMSURI != nil {
			__antithesis_instrumentation__.Notify(603469)
			return errors.New("kms specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603470)
		}
	}
	__antithesis_instrumentation__.Notify(603453)

	if o.IncrementalStorage == nil {
		__antithesis_instrumentation__.Notify(603471)
		o.IncrementalStorage = other.IncrementalStorage
	} else {
		__antithesis_instrumentation__.Notify(603472)
		if other.IncrementalStorage != nil {
			__antithesis_instrumentation__.Notify(603473)
			return errors.New("incremental_location option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603474)
		}
	}
	__antithesis_instrumentation__.Notify(603454)

	return nil
}

func (o BackupOptions) IsDefault() bool {
	__antithesis_instrumentation__.Notify(603475)
	options := BackupOptions{}
	return o.CaptureRevisionHistory == options.CaptureRevisionHistory && func() bool {
		__antithesis_instrumentation__.Notify(603476)
		return o.Detached == options.Detached == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603477)
		return cmp.Equal(o.EncryptionKMSURI, options.EncryptionKMSURI) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603478)
		return o.EncryptionPassphrase == options.EncryptionPassphrase == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603479)
		return cmp.Equal(o.IncrementalStorage, options.IncrementalStorage) == true
	}() == true
}

func (o *RestoreOptions) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(603480)
	var addSep bool
	maybeAddSep := func() {
		__antithesis_instrumentation__.Notify(603494)
		if addSep {
			__antithesis_instrumentation__.Notify(603496)
			ctx.WriteString(", ")
		} else {
			__antithesis_instrumentation__.Notify(603497)
		}
		__antithesis_instrumentation__.Notify(603495)
		addSep = true
	}
	__antithesis_instrumentation__.Notify(603481)
	if o.EncryptionPassphrase != nil {
		__antithesis_instrumentation__.Notify(603498)
		addSep = true
		ctx.WriteString("encryption_passphrase = ")
		ctx.FormatNode(o.EncryptionPassphrase)
	} else {
		__antithesis_instrumentation__.Notify(603499)
	}
	__antithesis_instrumentation__.Notify(603482)

	if o.DecryptionKMSURI != nil {
		__antithesis_instrumentation__.Notify(603500)
		maybeAddSep()
		ctx.WriteString("kms = ")
		ctx.FormatNode(&o.DecryptionKMSURI)
	} else {
		__antithesis_instrumentation__.Notify(603501)
	}
	__antithesis_instrumentation__.Notify(603483)

	if o.IntoDB != nil {
		__antithesis_instrumentation__.Notify(603502)
		maybeAddSep()
		ctx.WriteString("into_db = ")
		ctx.FormatNode(o.IntoDB)
	} else {
		__antithesis_instrumentation__.Notify(603503)
	}
	__antithesis_instrumentation__.Notify(603484)

	if o.DebugPauseOn != nil {
		__antithesis_instrumentation__.Notify(603504)
		maybeAddSep()
		ctx.WriteString("debug_pause_on = ")
		ctx.FormatNode(o.DebugPauseOn)
	} else {
		__antithesis_instrumentation__.Notify(603505)
	}
	__antithesis_instrumentation__.Notify(603485)

	if o.SkipMissingFKs {
		__antithesis_instrumentation__.Notify(603506)
		maybeAddSep()
		ctx.WriteString("skip_missing_foreign_keys")
	} else {
		__antithesis_instrumentation__.Notify(603507)
	}
	__antithesis_instrumentation__.Notify(603486)

	if o.SkipMissingSequenceOwners {
		__antithesis_instrumentation__.Notify(603508)
		maybeAddSep()
		ctx.WriteString("skip_missing_sequence_owners")
	} else {
		__antithesis_instrumentation__.Notify(603509)
	}
	__antithesis_instrumentation__.Notify(603487)

	if o.SkipMissingSequences {
		__antithesis_instrumentation__.Notify(603510)
		maybeAddSep()
		ctx.WriteString("skip_missing_sequences")
	} else {
		__antithesis_instrumentation__.Notify(603511)
	}
	__antithesis_instrumentation__.Notify(603488)

	if o.SkipMissingViews {
		__antithesis_instrumentation__.Notify(603512)
		maybeAddSep()
		ctx.WriteString("skip_missing_views")
	} else {
		__antithesis_instrumentation__.Notify(603513)
	}
	__antithesis_instrumentation__.Notify(603489)

	if o.Detached {
		__antithesis_instrumentation__.Notify(603514)
		maybeAddSep()
		ctx.WriteString("detached")
	} else {
		__antithesis_instrumentation__.Notify(603515)
	}
	__antithesis_instrumentation__.Notify(603490)

	if o.SkipLocalitiesCheck {
		__antithesis_instrumentation__.Notify(603516)
		maybeAddSep()
		ctx.WriteString("skip_localities_check")
	} else {
		__antithesis_instrumentation__.Notify(603517)
	}
	__antithesis_instrumentation__.Notify(603491)

	if o.NewDBName != nil {
		__antithesis_instrumentation__.Notify(603518)
		maybeAddSep()
		ctx.WriteString("new_db_name = ")
		ctx.FormatNode(o.NewDBName)
	} else {
		__antithesis_instrumentation__.Notify(603519)
	}
	__antithesis_instrumentation__.Notify(603492)

	if o.IncrementalStorage != nil {
		__antithesis_instrumentation__.Notify(603520)
		maybeAddSep()
		ctx.WriteString("incremental_location = ")
		ctx.FormatNode(&o.IncrementalStorage)
	} else {
		__antithesis_instrumentation__.Notify(603521)
	}
	__antithesis_instrumentation__.Notify(603493)

	if o.AsTenant != nil {
		__antithesis_instrumentation__.Notify(603522)
		maybeAddSep()
		ctx.WriteString("tenant = ")
		ctx.FormatNode(o.AsTenant)
	} else {
		__antithesis_instrumentation__.Notify(603523)
	}
}

func (o *RestoreOptions) CombineWith(other *RestoreOptions) error {
	__antithesis_instrumentation__.Notify(603524)
	if o.EncryptionPassphrase == nil {
		__antithesis_instrumentation__.Notify(603538)
		o.EncryptionPassphrase = other.EncryptionPassphrase
	} else {
		__antithesis_instrumentation__.Notify(603539)
		if other.EncryptionPassphrase != nil {
			__antithesis_instrumentation__.Notify(603540)
			return errors.New("encryption_passphrase specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603541)
		}
	}
	__antithesis_instrumentation__.Notify(603525)

	if o.DecryptionKMSURI == nil {
		__antithesis_instrumentation__.Notify(603542)
		o.DecryptionKMSURI = other.DecryptionKMSURI
	} else {
		__antithesis_instrumentation__.Notify(603543)
		if other.DecryptionKMSURI != nil {
			__antithesis_instrumentation__.Notify(603544)
			return errors.New("kms specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603545)
		}
	}
	__antithesis_instrumentation__.Notify(603526)

	if o.IntoDB == nil {
		__antithesis_instrumentation__.Notify(603546)
		o.IntoDB = other.IntoDB
	} else {
		__antithesis_instrumentation__.Notify(603547)
		if other.IntoDB != nil {
			__antithesis_instrumentation__.Notify(603548)
			return errors.New("into_db specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603549)
		}
	}
	__antithesis_instrumentation__.Notify(603527)

	if o.SkipMissingFKs {
		__antithesis_instrumentation__.Notify(603550)
		if other.SkipMissingFKs {
			__antithesis_instrumentation__.Notify(603551)
			return errors.New("skip_missing_foreign_keys specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603552)
		}
	} else {
		__antithesis_instrumentation__.Notify(603553)
		o.SkipMissingFKs = other.SkipMissingFKs
	}
	__antithesis_instrumentation__.Notify(603528)

	if o.SkipMissingSequences {
		__antithesis_instrumentation__.Notify(603554)
		if other.SkipMissingSequences {
			__antithesis_instrumentation__.Notify(603555)
			return errors.New("skip_missing_sequences specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603556)
		}
	} else {
		__antithesis_instrumentation__.Notify(603557)
		o.SkipMissingSequences = other.SkipMissingSequences
	}
	__antithesis_instrumentation__.Notify(603529)

	if o.SkipMissingSequenceOwners {
		__antithesis_instrumentation__.Notify(603558)
		if other.SkipMissingSequenceOwners {
			__antithesis_instrumentation__.Notify(603559)
			return errors.New("skip_missing_sequence_owners specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603560)
		}
	} else {
		__antithesis_instrumentation__.Notify(603561)
		o.SkipMissingSequenceOwners = other.SkipMissingSequenceOwners
	}
	__antithesis_instrumentation__.Notify(603530)

	if o.SkipMissingViews {
		__antithesis_instrumentation__.Notify(603562)
		if other.SkipMissingViews {
			__antithesis_instrumentation__.Notify(603563)
			return errors.New("skip_missing_views specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603564)
		}
	} else {
		__antithesis_instrumentation__.Notify(603565)
		o.SkipMissingViews = other.SkipMissingViews
	}
	__antithesis_instrumentation__.Notify(603531)

	if o.Detached {
		__antithesis_instrumentation__.Notify(603566)
		if other.Detached {
			__antithesis_instrumentation__.Notify(603567)
			return errors.New("detached option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603568)
		}
	} else {
		__antithesis_instrumentation__.Notify(603569)
		o.Detached = other.Detached
	}
	__antithesis_instrumentation__.Notify(603532)

	if o.SkipLocalitiesCheck {
		__antithesis_instrumentation__.Notify(603570)
		if other.SkipLocalitiesCheck {
			__antithesis_instrumentation__.Notify(603571)
			return errors.New("skip_localities_check specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603572)
		}
	} else {
		__antithesis_instrumentation__.Notify(603573)
		o.SkipLocalitiesCheck = other.SkipLocalitiesCheck
	}
	__antithesis_instrumentation__.Notify(603533)

	if o.DebugPauseOn == nil {
		__antithesis_instrumentation__.Notify(603574)
		o.DebugPauseOn = other.DebugPauseOn
	} else {
		__antithesis_instrumentation__.Notify(603575)
		if other.DebugPauseOn != nil {
			__antithesis_instrumentation__.Notify(603576)
			return errors.New("debug_pause_on specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603577)
		}
	}
	__antithesis_instrumentation__.Notify(603534)

	if o.NewDBName == nil {
		__antithesis_instrumentation__.Notify(603578)
		o.NewDBName = other.NewDBName
	} else {
		__antithesis_instrumentation__.Notify(603579)
		if other.NewDBName != nil {
			__antithesis_instrumentation__.Notify(603580)
			return errors.New("new_db_name specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603581)
		}
	}
	__antithesis_instrumentation__.Notify(603535)

	if o.IncrementalStorage == nil {
		__antithesis_instrumentation__.Notify(603582)
		o.IncrementalStorage = other.IncrementalStorage
	} else {
		__antithesis_instrumentation__.Notify(603583)
		if other.IncrementalStorage != nil {
			__antithesis_instrumentation__.Notify(603584)
			return errors.New("incremental_location option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603585)
		}
	}
	__antithesis_instrumentation__.Notify(603536)

	if o.AsTenant == nil {
		__antithesis_instrumentation__.Notify(603586)
		o.AsTenant = other.AsTenant
	} else {
		__antithesis_instrumentation__.Notify(603587)
		if other.AsTenant != nil {
			__antithesis_instrumentation__.Notify(603588)
			return errors.New("tenant option specified multiple times")
		} else {
			__antithesis_instrumentation__.Notify(603589)
		}
	}
	__antithesis_instrumentation__.Notify(603537)

	return nil
}

func (o RestoreOptions) IsDefault() bool {
	__antithesis_instrumentation__.Notify(603590)
	options := RestoreOptions{}
	return o.SkipMissingFKs == options.SkipMissingFKs && func() bool {
		__antithesis_instrumentation__.Notify(603591)
		return o.SkipMissingSequences == options.SkipMissingSequences == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603592)
		return o.SkipMissingSequenceOwners == options.SkipMissingSequenceOwners == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603593)
		return o.SkipMissingViews == options.SkipMissingViews == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603594)
		return cmp.Equal(o.DecryptionKMSURI, options.DecryptionKMSURI) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603595)
		return o.EncryptionPassphrase == options.EncryptionPassphrase == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603596)
		return o.IntoDB == options.IntoDB == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603597)
		return o.Detached == options.Detached == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603598)
		return o.SkipLocalitiesCheck == options.SkipLocalitiesCheck == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603599)
		return o.DebugPauseOn == options.DebugPauseOn == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603600)
		return o.NewDBName == options.NewDBName == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603601)
		return cmp.Equal(o.IncrementalStorage, options.IncrementalStorage) == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(603602)
		return o.AsTenant == options.AsTenant == true
	}() == true
}
