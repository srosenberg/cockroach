package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
)

type IsolationLevel int

const (
	UnspecifiedIsolation IsolationLevel = iota
	SerializableIsolation
)

var isolationLevelNames = [...]string{
	UnspecifiedIsolation:  "UNSPECIFIED",
	SerializableIsolation: "SERIALIZABLE",
}

var IsolationLevelMap = map[string]IsolationLevel{
	"serializable": SerializableIsolation,
}

func (i IsolationLevel) String() string {
	__antithesis_instrumentation__.Notify(614655)
	if i < 0 || func() bool {
		__antithesis_instrumentation__.Notify(614657)
		return i > IsolationLevel(len(isolationLevelNames)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(614658)
		return fmt.Sprintf("IsolationLevel(%d)", i)
	} else {
		__antithesis_instrumentation__.Notify(614659)
	}
	__antithesis_instrumentation__.Notify(614656)
	return isolationLevelNames[i]
}

type UserPriority int

const (
	UnspecifiedUserPriority UserPriority = iota
	Low
	Normal
	High
)

var userPriorityNames = [...]string{
	UnspecifiedUserPriority: "UNSPECIFIED",
	Low:                     "LOW",
	Normal:                  "NORMAL",
	High:                    "HIGH",
}

func (up UserPriority) String() string {
	__antithesis_instrumentation__.Notify(614660)
	if up < 0 || func() bool {
		__antithesis_instrumentation__.Notify(614662)
		return up > UserPriority(len(userPriorityNames)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(614663)
		return fmt.Sprintf("UserPriority(%d)", up)
	} else {
		__antithesis_instrumentation__.Notify(614664)
	}
	__antithesis_instrumentation__.Notify(614661)
	return userPriorityNames[up]
}

func UserPriorityFromString(val string) (_ UserPriority, ok bool) {
	__antithesis_instrumentation__.Notify(614665)
	switch strings.ToUpper(val) {
	case "LOW":
		__antithesis_instrumentation__.Notify(614666)
		return Low, true
	case "NORMAL":
		__antithesis_instrumentation__.Notify(614667)
		return Normal, true
	case "HIGH":
		__antithesis_instrumentation__.Notify(614668)
		return High, true
	default:
		__antithesis_instrumentation__.Notify(614669)
		return 0, false
	}
}

type ReadWriteMode int

const (
	UnspecifiedReadWriteMode ReadWriteMode = iota
	ReadOnly
	ReadWrite
)

var readWriteModeNames = [...]string{
	UnspecifiedReadWriteMode: "UNSPECIFIED",
	ReadOnly:                 "ONLY",
	ReadWrite:                "WRITE",
}

func (ro ReadWriteMode) String() string {
	__antithesis_instrumentation__.Notify(614670)
	if ro < 0 || func() bool {
		__antithesis_instrumentation__.Notify(614672)
		return ro > ReadWriteMode(len(readWriteModeNames)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(614673)
		return fmt.Sprintf("ReadWriteMode(%d)", ro)
	} else {
		__antithesis_instrumentation__.Notify(614674)
	}
	__antithesis_instrumentation__.Notify(614671)
	return readWriteModeNames[ro]
}

type DeferrableMode int

const (
	UnspecifiedDeferrableMode DeferrableMode = iota
	Deferrable
	NotDeferrable
)

var deferrableModeNames = [...]string{
	UnspecifiedDeferrableMode: "UNSPECIFIED",
	Deferrable:                "DEFERRABLE",
	NotDeferrable:             "NOT DEFERRABLE",
}

func (d DeferrableMode) String() string {
	__antithesis_instrumentation__.Notify(614675)
	if d < 0 || func() bool {
		__antithesis_instrumentation__.Notify(614677)
		return d > DeferrableMode(len(deferrableModeNames)-1) == true
	}() == true {
		__antithesis_instrumentation__.Notify(614678)
		return fmt.Sprintf("DeferrableMode(%d)", d)
	} else {
		__antithesis_instrumentation__.Notify(614679)
	}
	__antithesis_instrumentation__.Notify(614676)
	return deferrableModeNames[d]
}

type TransactionModes struct {
	Isolation     IsolationLevel
	UserPriority  UserPriority
	ReadWriteMode ReadWriteMode
	AsOf          AsOfClause
	Deferrable    DeferrableMode
}

func (node *TransactionModes) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614680)
	var sep string
	if node.Isolation != UnspecifiedIsolation {
		__antithesis_instrumentation__.Notify(614685)
		ctx.Printf(" ISOLATION LEVEL %s", node.Isolation)
		sep = ","
	} else {
		__antithesis_instrumentation__.Notify(614686)
	}
	__antithesis_instrumentation__.Notify(614681)
	if node.UserPriority != UnspecifiedUserPriority {
		__antithesis_instrumentation__.Notify(614687)
		ctx.Printf("%s PRIORITY %s", sep, node.UserPriority)
		sep = ","
	} else {
		__antithesis_instrumentation__.Notify(614688)
	}
	__antithesis_instrumentation__.Notify(614682)
	if node.ReadWriteMode != UnspecifiedReadWriteMode {
		__antithesis_instrumentation__.Notify(614689)
		ctx.Printf("%s READ %s", sep, node.ReadWriteMode)
		sep = ","
	} else {
		__antithesis_instrumentation__.Notify(614690)
	}
	__antithesis_instrumentation__.Notify(614683)
	if node.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(614691)
		ctx.WriteString(sep)
		ctx.WriteString(" ")
		ctx.FormatNode(&node.AsOf)
		sep = ","
	} else {
		__antithesis_instrumentation__.Notify(614692)
	}
	__antithesis_instrumentation__.Notify(614684)
	if node.Deferrable != UnspecifiedDeferrableMode {
		__antithesis_instrumentation__.Notify(614693)
		ctx.Printf("%s %s", sep, node.Deferrable)
	} else {
		__antithesis_instrumentation__.Notify(614694)
	}
}

var (
	errIsolationLevelSpecifiedMultipleTimes = pgerror.New(pgcode.Syntax, "isolation level specified multiple times")
	errUserPrioritySpecifiedMultipleTimes   = pgerror.New(pgcode.Syntax, "user priority specified multiple times")
	errReadModeSpecifiedMultipleTimes       = pgerror.New(pgcode.Syntax, "read mode specified multiple times")
	errAsOfSpecifiedMultipleTimes           = pgerror.New(pgcode.Syntax, "AS OF SYSTEM TIME specified multiple times")
	errDeferrableSpecifiedMultipleTimes     = pgerror.New(pgcode.Syntax, "deferrable mode specified multiple times")

	ErrAsOfSpecifiedWithReadWrite = pgerror.New(pgcode.Syntax, "AS OF SYSTEM TIME specified with READ WRITE mode")
)

func (node *TransactionModes) Merge(other TransactionModes) error {
	__antithesis_instrumentation__.Notify(614695)
	if other.Isolation != UnspecifiedIsolation {
		__antithesis_instrumentation__.Notify(614702)
		if node.Isolation != UnspecifiedIsolation {
			__antithesis_instrumentation__.Notify(614704)
			return errIsolationLevelSpecifiedMultipleTimes
		} else {
			__antithesis_instrumentation__.Notify(614705)
		}
		__antithesis_instrumentation__.Notify(614703)
		node.Isolation = other.Isolation
	} else {
		__antithesis_instrumentation__.Notify(614706)
	}
	__antithesis_instrumentation__.Notify(614696)
	if other.UserPriority != UnspecifiedUserPriority {
		__antithesis_instrumentation__.Notify(614707)
		if node.UserPriority != UnspecifiedUserPriority {
			__antithesis_instrumentation__.Notify(614709)
			return errUserPrioritySpecifiedMultipleTimes
		} else {
			__antithesis_instrumentation__.Notify(614710)
		}
		__antithesis_instrumentation__.Notify(614708)
		node.UserPriority = other.UserPriority
	} else {
		__antithesis_instrumentation__.Notify(614711)
	}
	__antithesis_instrumentation__.Notify(614697)
	if other.AsOf.Expr != nil {
		__antithesis_instrumentation__.Notify(614712)
		if node.AsOf.Expr != nil {
			__antithesis_instrumentation__.Notify(614714)
			return errAsOfSpecifiedMultipleTimes
		} else {
			__antithesis_instrumentation__.Notify(614715)
		}
		__antithesis_instrumentation__.Notify(614713)
		node.AsOf.Expr = other.AsOf.Expr
	} else {
		__antithesis_instrumentation__.Notify(614716)
	}
	__antithesis_instrumentation__.Notify(614698)
	if other.ReadWriteMode != UnspecifiedReadWriteMode {
		__antithesis_instrumentation__.Notify(614717)
		if node.ReadWriteMode != UnspecifiedReadWriteMode {
			__antithesis_instrumentation__.Notify(614719)
			return errReadModeSpecifiedMultipleTimes
		} else {
			__antithesis_instrumentation__.Notify(614720)
		}
		__antithesis_instrumentation__.Notify(614718)
		node.ReadWriteMode = other.ReadWriteMode
	} else {
		__antithesis_instrumentation__.Notify(614721)
	}
	__antithesis_instrumentation__.Notify(614699)
	if node.ReadWriteMode != UnspecifiedReadWriteMode && func() bool {
		__antithesis_instrumentation__.Notify(614722)
		return node.ReadWriteMode != ReadOnly == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(614723)
		return node.AsOf.Expr != nil == true
	}() == true {
		__antithesis_instrumentation__.Notify(614724)
		return ErrAsOfSpecifiedWithReadWrite
	} else {
		__antithesis_instrumentation__.Notify(614725)
	}
	__antithesis_instrumentation__.Notify(614700)
	if other.Deferrable != UnspecifiedDeferrableMode {
		__antithesis_instrumentation__.Notify(614726)
		if node.Deferrable != UnspecifiedDeferrableMode {
			__antithesis_instrumentation__.Notify(614728)
			return errDeferrableSpecifiedMultipleTimes
		} else {
			__antithesis_instrumentation__.Notify(614729)
		}
		__antithesis_instrumentation__.Notify(614727)
		node.Deferrable = other.Deferrable
	} else {
		__antithesis_instrumentation__.Notify(614730)
	}
	__antithesis_instrumentation__.Notify(614701)
	return nil
}

type BeginTransaction struct {
	Modes TransactionModes
}

func (node *BeginTransaction) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614731)
	ctx.WriteString("BEGIN TRANSACTION")
	ctx.FormatNode(&node.Modes)
}

type CommitTransaction struct{}

func (node *CommitTransaction) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614732)
	ctx.WriteString("COMMIT TRANSACTION")
}

type RollbackTransaction struct{}

func (node *RollbackTransaction) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614733)
	ctx.WriteString("ROLLBACK TRANSACTION")
}

type Savepoint struct {
	Name Name
}

func (node *Savepoint) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614734)
	ctx.WriteString("SAVEPOINT ")
	ctx.FormatNode(&node.Name)
}

type ReleaseSavepoint struct {
	Savepoint Name
}

func (node *ReleaseSavepoint) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614735)
	ctx.WriteString("RELEASE SAVEPOINT ")
	ctx.FormatNode(&node.Savepoint)
}

type RollbackToSavepoint struct {
	Savepoint Name
}

func (node *RollbackToSavepoint) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(614736)
	ctx.WriteString("ROLLBACK TRANSACTION TO SAVEPOINT ")
	ctx.FormatNode(&node.Savepoint)
}
