package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type ZoneSpecifier struct {
	NamedZone UnrestrictedName
	Database  Name

	TableOrIndex TableIndexName

	Partition Name
}

func (node ZoneSpecifier) TelemetryName() string {
	__antithesis_instrumentation__.Notify(617159)
	if node.NamedZone != "" {
		__antithesis_instrumentation__.Notify(617164)
		return "range"
	} else {
		__antithesis_instrumentation__.Notify(617165)
	}
	__antithesis_instrumentation__.Notify(617160)
	if node.Database != "" {
		__antithesis_instrumentation__.Notify(617166)
		return "database"
	} else {
		__antithesis_instrumentation__.Notify(617167)
	}
	__antithesis_instrumentation__.Notify(617161)
	str := ""
	if node.Partition != "" {
		__antithesis_instrumentation__.Notify(617168)
		str = "partition."
	} else {
		__antithesis_instrumentation__.Notify(617169)
	}
	__antithesis_instrumentation__.Notify(617162)
	if node.TargetsIndex() {
		__antithesis_instrumentation__.Notify(617170)
		str += "index"
	} else {
		__antithesis_instrumentation__.Notify(617171)
		str += "table"
	}
	__antithesis_instrumentation__.Notify(617163)
	return str
}

func (node ZoneSpecifier) TargetsTable() bool {
	__antithesis_instrumentation__.Notify(617172)
	return node.NamedZone == "" && func() bool {
		__antithesis_instrumentation__.Notify(617173)
		return node.Database == "" == true
	}() == true
}

func (node ZoneSpecifier) TargetsIndex() bool {
	__antithesis_instrumentation__.Notify(617174)
	return node.TargetsTable() && func() bool {
		__antithesis_instrumentation__.Notify(617175)
		return node.TableOrIndex.Index != "" == true
	}() == true
}

func (node ZoneSpecifier) TargetsPartition() bool {
	__antithesis_instrumentation__.Notify(617176)
	return node.TargetsTable() && func() bool {
		__antithesis_instrumentation__.Notify(617177)
		return node.Partition != "" == true
	}() == true
}

func (node *ZoneSpecifier) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(617178)
	if node.NamedZone != "" {
		__antithesis_instrumentation__.Notify(617179)
		ctx.WriteString("RANGE ")
		ctx.FormatNode(&node.NamedZone)
	} else {
		__antithesis_instrumentation__.Notify(617180)
		if node.Database != "" {
			__antithesis_instrumentation__.Notify(617181)
			ctx.WriteString("DATABASE ")
			ctx.FormatNode(&node.Database)
		} else {
			__antithesis_instrumentation__.Notify(617182)
			if node.Partition != "" {
				__antithesis_instrumentation__.Notify(617185)
				ctx.WriteString("PARTITION ")
				ctx.FormatNode(&node.Partition)
				ctx.WriteString(" OF ")
			} else {
				__antithesis_instrumentation__.Notify(617186)
			}
			__antithesis_instrumentation__.Notify(617183)
			if node.TargetsIndex() {
				__antithesis_instrumentation__.Notify(617187)
				ctx.WriteString("INDEX ")
			} else {
				__antithesis_instrumentation__.Notify(617188)
				ctx.WriteString("TABLE ")
			}
			__antithesis_instrumentation__.Notify(617184)
			ctx.FormatNode(&node.TableOrIndex)
		}
	}
}

func (node *ZoneSpecifier) String() string {
	__antithesis_instrumentation__.Notify(617189)
	return AsString(node)
}

type ShowZoneConfig struct {
	ZoneSpecifier
}

func (node *ShowZoneConfig) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(617190)
	if node.ZoneSpecifier == (ZoneSpecifier{}) {
		__antithesis_instrumentation__.Notify(617191)
		ctx.WriteString("SHOW ZONE CONFIGURATIONS")
	} else {
		__antithesis_instrumentation__.Notify(617192)
		ctx.WriteString("SHOW ZONE CONFIGURATION FROM ")
		ctx.FormatNode(&node.ZoneSpecifier)
	}
}

type SetZoneConfig struct {
	ZoneSpecifier

	AllIndexes bool
	SetDefault bool
	YAMLConfig Expr
	Options    KVOptions
}

func (node *SetZoneConfig) Format(ctx *FmtCtx) {
	__antithesis_instrumentation__.Notify(617193)
	ctx.WriteString("ALTER ")
	ctx.FormatNode(&node.ZoneSpecifier)
	ctx.WriteString(" CONFIGURE ZONE ")
	if node.SetDefault {
		__antithesis_instrumentation__.Notify(617194)
		ctx.WriteString("USING DEFAULT")
	} else {
		__antithesis_instrumentation__.Notify(617195)
		if node.YAMLConfig != nil {
			__antithesis_instrumentation__.Notify(617196)
			if node.YAMLConfig == DNull {
				__antithesis_instrumentation__.Notify(617197)
				ctx.WriteString("DISCARD")
			} else {
				__antithesis_instrumentation__.Notify(617198)
				ctx.WriteString("= ")
				ctx.FormatNode(node.YAMLConfig)
			}
		} else {
			__antithesis_instrumentation__.Notify(617199)
			ctx.WriteString("USING ")
			kvOptions := node.Options
			comma := ""
			for _, kv := range kvOptions {
				__antithesis_instrumentation__.Notify(617200)
				ctx.WriteString(comma)
				comma = ", "
				ctx.FormatNode(&kv.Key)
				if kv.Value != nil {
					__antithesis_instrumentation__.Notify(617201)
					ctx.WriteString(` = `)
					ctx.FormatNode(kv.Value)
				} else {
					__antithesis_instrumentation__.Notify(617202)
					ctx.WriteString(` = COPY FROM PARENT`)
				}
			}
		}
	}
}
