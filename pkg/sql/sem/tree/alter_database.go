package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type AlterDatabaseOwner struct {
	Name  Name
	Owner RoleSpec
}

func (node *AlterDatabaseOwner) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602948)
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" OWNER TO ")
	ctx.FormatNode(&node.Owner)
}

type AlterDatabaseAddRegion struct {
	Name        Name
	Region      Name
	IfNotExists bool
}

var _ Statement = &AlterDatabaseAddRegion{}

func (node *AlterDatabaseAddRegion) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602949)
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" ADD REGION ")
	if node.IfNotExists {
		__antithesis_instrumentation__.Notify(602951)
		ctx.WriteString("IF NOT EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(602952)
	}
	__antithesis_instrumentation__.Notify(602950)
	ctx.FormatNode(&node.Region)
}

type AlterDatabaseDropRegion struct {
	Name     Name
	Region   Name
	IfExists bool
}

var _ Statement = &AlterDatabaseDropRegion{}

func (node *AlterDatabaseDropRegion) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602953)
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" DROP REGION ")
	if node.IfExists {
		__antithesis_instrumentation__.Notify(602955)
		ctx.WriteString("IF EXISTS ")
	} else {
		__antithesis_instrumentation__.Notify(602956)
	}
	__antithesis_instrumentation__.Notify(602954)
	ctx.FormatNode(&node.Region)
}

type AlterDatabasePrimaryRegion struct {
	Name          Name
	PrimaryRegion Name
}

var _ Statement = &AlterDatabasePrimaryRegion{}

func (node *AlterDatabasePrimaryRegion) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602957)
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" PRIMARY REGION ")
	node.PrimaryRegion.Format(ctx)
}

type AlterDatabaseSurvivalGoal struct {
	Name         Name
	SurvivalGoal SurvivalGoal
}

var _ Statement = &AlterDatabaseSurvivalGoal{}

func (node *AlterDatabaseSurvivalGoal) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602958)
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" ")
	node.SurvivalGoal.Format(ctx)
}

type AlterDatabasePlacement struct {
	Name      Name
	Placement DataPlacement
}

var _ Statement = &AlterDatabasePlacement{}

func (node *AlterDatabasePlacement) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602959)
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.Name)
	ctx.WriteString(" ")
	node.Placement.Format(ctx)
}

type AlterDatabaseAddSuperRegion struct {
	DatabaseName    Name
	SuperRegionName Name
	Regions         []Name
}

var _ Statement = &AlterDatabaseAddSuperRegion{}

func (node *AlterDatabaseAddSuperRegion) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602960)
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.DatabaseName)
	ctx.WriteString(" ADD SUPER REGION ")
	ctx.FormatNode(&node.SuperRegionName)
	ctx.WriteString(" VALUES ")
	for i, region := range node.Regions {
		__antithesis_instrumentation__.Notify(602961)
		if i != 0 {
			__antithesis_instrumentation__.Notify(602963)
			ctx.WriteString(",")
		} else {
			__antithesis_instrumentation__.Notify(602964)
		}
		__antithesis_instrumentation__.Notify(602962)
		ctx.FormatNode(&region)
	}
}

type AlterDatabaseDropSuperRegion struct {
	DatabaseName    Name
	SuperRegionName Name
}

var _ Statement = &AlterDatabaseDropSuperRegion{}

func (node *AlterDatabaseDropSuperRegion) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602965)
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.DatabaseName)
	ctx.WriteString(" DROP SUPER REGION ")
	ctx.FormatNode(&node.SuperRegionName)
}

type AlterDatabaseAlterSuperRegion struct {
	DatabaseName    Name
	SuperRegionName Name
	Regions         []Name
}

var _ Statement = &AlterDatabaseAlterSuperRegion{}

func (node *AlterDatabaseAlterSuperRegion) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(602966)
	ctx.WriteString("ALTER DATABASE ")
	ctx.FormatNode(&node.DatabaseName)
	ctx.WriteString(" ALTER SUPER REGION ")
	ctx.FormatNode(&node.SuperRegionName)
	ctx.WriteString(" VALUES ")
	for i, region := range node.Regions {
		__antithesis_instrumentation__.Notify(602967)
		if i != 0 {
			__antithesis_instrumentation__.Notify(602969)
			ctx.WriteString(",")
		} else {
			__antithesis_instrumentation__.Notify(602970)
		}
		__antithesis_instrumentation__.Notify(602968)
		ctx.FormatNode(&region)
	}
}
