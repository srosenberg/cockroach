package pgcode

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Code struct {
	code string
}

func MakeCode(s string) Code {
	__antithesis_instrumentation__.Notify(560259)
	return Code{code: s}
}

func (c Code) String() string {
	__antithesis_instrumentation__.Notify(560260)
	return c.code
}

var (
	SuccessfulCompletion = MakeCode("00000")

	Warning                                 = MakeCode("01000")
	WarningDynamicResultSetsReturned        = MakeCode("0100C")
	WarningImplicitZeroBitPadding           = MakeCode("01008")
	WarningNullValueEliminatedInSetFunction = MakeCode("01003")
	WarningPrivilegeNotGranted              = MakeCode("01007")
	WarningPrivilegeNotRevoked              = MakeCode("01006")
	WarningStringDataRightTruncation        = MakeCode("01004")
	WarningDeprecatedFeature                = MakeCode("01P01")

	NoData                                = MakeCode("02000")
	NoAdditionalDynamicResultSetsReturned = MakeCode("02001")

	SQLStatementNotYetComplete = MakeCode("03000")

	ConnectionException                           = MakeCode("08000")
	ConnectionDoesNotExist                        = MakeCode("08003")
	ConnectionFailure                             = MakeCode("08006")
	SQLclientUnableToEstablishSQLconnection       = MakeCode("08001")
	SQLserverRejectedEstablishmentOfSQLconnection = MakeCode("08004")
	TransactionResolutionUnknown                  = MakeCode("08007")
	ProtocolViolation                             = MakeCode("08P01")

	TriggeredActionException = MakeCode("09000")

	FeatureNotSupported = MakeCode("0A000")

	InvalidTransactionInitiation = MakeCode("0B000")

	LocatorException            = MakeCode("0F000")
	InvalidLocatorSpecification = MakeCode("0F001")

	InvalidGrantor        = MakeCode("0L000")
	InvalidGrantOperation = MakeCode("0LP01")

	InvalidRoleSpecification = MakeCode("0P000")

	DiagnosticsException                           = MakeCode("0Z000")
	StackedDiagnosticsAccessedWithoutActiveHandler = MakeCode("0Z002")

	CaseNotFound = MakeCode("20000")

	CardinalityViolation = MakeCode("21000")

	DataException                         = MakeCode("22000")
	ArraySubscript                        = MakeCode("2202E")
	CharacterNotInRepertoire              = MakeCode("22021")
	DatetimeFieldOverflow                 = MakeCode("22008")
	DivisionByZero                        = MakeCode("22012")
	InvalidWindowFrameOffset              = MakeCode("22013")
	ErrorInAssignment                     = MakeCode("22005")
	EscapeCharacterConflict               = MakeCode("2200B")
	IndicatorOverflow                     = MakeCode("22022")
	IntervalFieldOverflow                 = MakeCode("22015")
	InvalidArgumentForLogarithm           = MakeCode("2201E")
	InvalidArgumentForNtileFunction       = MakeCode("22014")
	InvalidArgumentForNthValueFunction    = MakeCode("22016")
	InvalidArgumentForPowerFunction       = MakeCode("2201F")
	InvalidArgumentForWidthBucketFunction = MakeCode("2201G")
	InvalidCharacterValueForCast          = MakeCode("22018")
	InvalidDatetimeFormat                 = MakeCode("22007")
	InvalidEscapeCharacter                = MakeCode("22019")
	InvalidEscapeOctet                    = MakeCode("2200D")
	InvalidEscapeSequence                 = MakeCode("22025")
	NonstandardUseOfEscapeCharacter       = MakeCode("22P06")
	InvalidIndicatorParameterValue        = MakeCode("22010")
	InvalidParameterValue                 = MakeCode("22023")
	InvalidRegularExpression              = MakeCode("2201B")
	InvalidRowCountInLimitClause          = MakeCode("2201W")
	InvalidRowCountInResultOffsetClause   = MakeCode("2201X")
	InvalidTimeZoneDisplacementValue      = MakeCode("22009")
	InvalidUseOfEscapeCharacter           = MakeCode("2200C")
	MostSpecificTypeMismatch              = MakeCode("2200G")
	NullValueNotAllowed                   = MakeCode("22004")
	NullValueNoIndicatorParameter         = MakeCode("22002")
	NumericValueOutOfRange                = MakeCode("22003")
	SequenceGeneratorLimitExceeded        = MakeCode("2200H")
	StringDataLengthMismatch              = MakeCode("22026")
	StringDataRightTruncation             = MakeCode("22001")
	Substring                             = MakeCode("22011")
	Trim                                  = MakeCode("22027")
	UnterminatedCString                   = MakeCode("22024")
	ZeroLengthCharacterString             = MakeCode("2200F")
	FloatingPointException                = MakeCode("22P01")
	InvalidTextRepresentation             = MakeCode("22P02")
	InvalidBinaryRepresentation           = MakeCode("22P03")
	BadCopyFileFormat                     = MakeCode("22P04")
	UntranslatableCharacter               = MakeCode("22P05")
	NotAnXMLDocument                      = MakeCode("2200L")
	InvalidXMLDocument                    = MakeCode("2200M")
	InvalidXMLContent                     = MakeCode("2200N")
	InvalidXMLComment                     = MakeCode("2200S")
	InvalidXMLProcessingInstruction       = MakeCode("2200T")

	IntegrityConstraintViolation = MakeCode("23000")
	RestrictViolation            = MakeCode("23001")
	NotNullViolation             = MakeCode("23502")
	ForeignKeyViolation          = MakeCode("23503")
	UniqueViolation              = MakeCode("23505")
	CheckViolation               = MakeCode("23514")
	ExclusionViolation           = MakeCode("23P01")

	InvalidCursorState = MakeCode("24000")

	InvalidTransactionState                         = MakeCode("25000")
	ActiveSQLTransaction                            = MakeCode("25001")
	BranchTransactionAlreadyActive                  = MakeCode("25002")
	HeldCursorRequiresSameIsolationLevel            = MakeCode("25008")
	InappropriateAccessModeForBranchTransaction     = MakeCode("25003")
	InappropriateIsolationLevelForBranchTransaction = MakeCode("25004")
	NoActiveSQLTransactionForBranchTransaction      = MakeCode("25005")
	ReadOnlySQLTransaction                          = MakeCode("25006")
	SchemaAndDataStatementMixingNotSupported        = MakeCode("25007")
	NoActiveSQLTransaction                          = MakeCode("25P01")
	InFailedSQLTransaction                          = MakeCode("25P02")

	InvalidSQLStatementName = MakeCode("26000")

	TriggeredDataChangeViolation = MakeCode("27000")

	InvalidAuthorizationSpecification = MakeCode("28000")
	InvalidPassword                   = MakeCode("28P01")

	DependentPrivilegeDescriptorsStillExist = MakeCode("2B000")
	DependentObjectsStillExist              = MakeCode("2BP01")

	InvalidTransactionTermination = MakeCode("2D000")

	RoutineExceptionFunctionExecutedNoReturnStatement = MakeCode("2F005")
	RoutineExceptionModifyingSQLDataNotPermitted      = MakeCode("2F002")
	RoutineExceptionProhibitedSQLStatementAttempted   = MakeCode("2F003")
	RoutineExceptionReadingSQLDataNotPermitted        = MakeCode("2F004")

	InvalidCursorName = MakeCode("34000")

	ExternalRoutineException                       = MakeCode("38000")
	ExternalRoutineContainingSQLNotPermitted       = MakeCode("38001")
	ExternalRoutineModifyingSQLDataNotPermitted    = MakeCode("38002")
	ExternalRoutineProhibitedSQLStatementAttempted = MakeCode("38003")
	ExternalRoutineReadingSQLDataNotPermitted      = MakeCode("38004")

	ExternalRoutineInvocationException     = MakeCode("39000")
	ExternalRoutineInvalidSQLstateReturned = MakeCode("39001")
	ExternalRoutineNullValueNotAllowed     = MakeCode("39004")
	ExternalRoutineTriggerProtocolViolated = MakeCode("39P01")
	ExternalRoutineSrfProtocolViolated     = MakeCode("39P02")

	SavepointException            = MakeCode("3B000")
	InvalidSavepointSpecification = MakeCode("3B001")

	InvalidCatalogName = MakeCode("3D000")

	InvalidSchemaName = MakeCode("3F000")

	TransactionRollback                     = MakeCode("40000")
	TransactionIntegrityConstraintViolation = MakeCode("40002")
	SerializationFailure                    = MakeCode("40001")
	StatementCompletionUnknown              = MakeCode("40003")
	DeadlockDetected                        = MakeCode("40P01")

	SyntaxErrorOrAccessRuleViolation   = MakeCode("42000")
	Syntax                             = MakeCode("42601")
	InsufficientPrivilege              = MakeCode("42501")
	CannotCoerce                       = MakeCode("42846")
	Grouping                           = MakeCode("42803")
	Windowing                          = MakeCode("42P20")
	InvalidRecursion                   = MakeCode("42P19")
	InvalidForeignKey                  = MakeCode("42830")
	InvalidName                        = MakeCode("42602")
	NameTooLong                        = MakeCode("42622")
	ReservedName                       = MakeCode("42939")
	DatatypeMismatch                   = MakeCode("42804")
	IndeterminateDatatype              = MakeCode("42P18")
	CollationMismatch                  = MakeCode("42P21")
	IndeterminateCollation             = MakeCode("42P22")
	WrongObjectType                    = MakeCode("42809")
	GeneratedAlways                    = MakeCode("428C9")
	UndefinedColumn                    = MakeCode("42703")
	UndefinedCursor                    = MakeCode("34000")
	UndefinedDatabase                  = MakeCode("3D000")
	UndefinedFunction                  = MakeCode("42883")
	UndefinedPreparedStatement         = MakeCode("26000")
	UndefinedSchema                    = MakeCode("3F000")
	UndefinedTable                     = MakeCode("42P01")
	UndefinedParameter                 = MakeCode("42P02")
	UndefinedObject                    = MakeCode("42704")
	DuplicateColumn                    = MakeCode("42701")
	DuplicateCursor                    = MakeCode("42P03")
	DuplicateDatabase                  = MakeCode("42P04")
	DuplicateFunction                  = MakeCode("42723")
	DuplicatePreparedStatement         = MakeCode("42P05")
	DuplicateSchema                    = MakeCode("42P06")
	DuplicateRelation                  = MakeCode("42P07")
	DuplicateAlias                     = MakeCode("42712")
	DuplicateObject                    = MakeCode("42710")
	AmbiguousColumn                    = MakeCode("42702")
	AmbiguousFunction                  = MakeCode("42725")
	AmbiguousParameter                 = MakeCode("42P08")
	AmbiguousAlias                     = MakeCode("42P09")
	InvalidColumnReference             = MakeCode("42P10")
	InvalidColumnDefinition            = MakeCode("42611")
	InvalidCursorDefinition            = MakeCode("42P11")
	InvalidDatabaseDefinition          = MakeCode("42P12")
	InvalidFunctionDefinition          = MakeCode("42P13")
	InvalidPreparedStatementDefinition = MakeCode("42P14")
	InvalidSchemaDefinition            = MakeCode("42P15")
	InvalidTableDefinition             = MakeCode("42P16")
	InvalidObjectDefinition            = MakeCode("42P17")
	FileAlreadyExists                  = MakeCode("42C01")

	WithCheckOptionViolation = MakeCode("44000")

	InsufficientResources      = MakeCode("53000")
	DiskFull                   = MakeCode("53100")
	OutOfMemory                = MakeCode("53200")
	TooManyConnections         = MakeCode("53300")
	ConfigurationLimitExceeded = MakeCode("53400")

	ProgramLimitExceeded = MakeCode("54000")
	StatementTooComplex  = MakeCode("54001")
	TooManyColumns       = MakeCode("54011")
	TooManyArguments     = MakeCode("54023")

	ObjectNotInPrerequisiteState = MakeCode("55000")
	ObjectInUse                  = MakeCode("55006")
	CantChangeRuntimeParam       = MakeCode("55P02")
	LockNotAvailable             = MakeCode("55P03")

	OperatorIntervention = MakeCode("57000")
	QueryCanceled        = MakeCode("57014")
	AdminShutdown        = MakeCode("57P01")
	CrashShutdown        = MakeCode("57P02")
	CannotConnectNow     = MakeCode("57P03")
	DatabaseDropped      = MakeCode("57P04")

	System        = MakeCode("58000")
	Io            = MakeCode("58030")
	UndefinedFile = MakeCode("58P01")
	DuplicateFile = MakeCode("58P02")

	ConfigFile     = MakeCode("F0000")
	LockFileExists = MakeCode("F0001")

	FdwError                             = MakeCode("HV000")
	FdwColumnNameNotFound                = MakeCode("HV005")
	FdwDynamicParameterValueNeeded       = MakeCode("HV002")
	FdwFunctionSequenceError             = MakeCode("HV010")
	FdwInconsistentDescriptorInformation = MakeCode("HV021")
	FdwInvalidAttributeValue             = MakeCode("HV024")
	FdwInvalidColumnName                 = MakeCode("HV007")
	FdwInvalidColumnNumber               = MakeCode("HV008")
	FdwInvalidDataType                   = MakeCode("HV004")
	FdwInvalidDataTypeDescriptors        = MakeCode("HV006")
	FdwInvalidDescriptorFieldIdentifier  = MakeCode("HV091")
	FdwInvalidHandle                     = MakeCode("HV00B")
	FdwInvalidOptionIndex                = MakeCode("HV00C")
	FdwInvalidOptionName                 = MakeCode("HV00D")
	FdwInvalidStringLengthOrBufferLength = MakeCode("HV090")
	FdwInvalidStringFormat               = MakeCode("HV00A")
	FdwInvalidUseOfNullPointer           = MakeCode("HV009")
	FdwTooManyHandles                    = MakeCode("HV014")
	FdwOutOfMemory                       = MakeCode("HV001")
	FdwNoSchemas                         = MakeCode("HV00P")
	FdwOptionNameNotFound                = MakeCode("HV00J")
	FdwReplyHandle                       = MakeCode("HV00K")
	FdwSchemaNotFound                    = MakeCode("HV00Q")
	FdwTableNotFound                     = MakeCode("HV00R")
	FdwUnableToCreateExecution           = MakeCode("HV00L")
	FdwUnableToCreateReply               = MakeCode("HV00M")
	FdwUnableToEstablishConnection       = MakeCode("HV00N")

	PLpgSQL        = MakeCode("P0000")
	RaiseException = MakeCode("P0001")
	NoDataFound    = MakeCode("P0002")
	TooManyRows    = MakeCode("P0003")
	AssertFailure  = MakeCode("P0004")

	Internal       = MakeCode("XX000")
	DataCorrupted  = MakeCode("XX001")
	IndexCorrupted = MakeCode("XX002")
)

var (
	Uncategorized = MakeCode("XXUUU")

	CCLRequired = MakeCode("XXC01")

	CCLValidLicenseRequired = MakeCode("XXC02")

	TransactionCommittedWithSchemaChangeFailure = MakeCode("XXA00")

	ScalarOperationCannotRunWithoutFullSessionContext = MakeCode("22C01")

	SchemaChangeOccurred = MakeCode("55C01")

	NoPrimaryKey = MakeCode("55C02")

	RangeUnavailable = MakeCode("58C00")

	InternalConnectionFailure = MakeCode("58C01")

	UnsatisfiableBoundedStaleness = MakeCode("XCUBS")
)
