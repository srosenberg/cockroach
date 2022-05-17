package tree

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"github.com/cockroachdb/cockroach/pkg/sql/sessiondata"
	"github.com/cockroachdb/cockroach/pkg/sql/types"
)

func GetRenderColName(searchPath sessiondata.SearchPath, target SelectExpr) (string, error) {
	__antithesis_instrumentation__.Notify(604328)
	if target.As != "" {
		__antithesis_instrumentation__.Notify(604332)
		return string(target.As), nil
	} else {
		__antithesis_instrumentation__.Notify(604333)
	}
	__antithesis_instrumentation__.Notify(604329)

	_, s, err := ComputeColNameInternal(searchPath, target.Expr)
	if err != nil {
		__antithesis_instrumentation__.Notify(604334)
		return s, err
	} else {
		__antithesis_instrumentation__.Notify(604335)
	}
	__antithesis_instrumentation__.Notify(604330)
	if len(s) == 0 {
		__antithesis_instrumentation__.Notify(604336)
		s = "?column?"
	} else {
		__antithesis_instrumentation__.Notify(604337)
	}
	__antithesis_instrumentation__.Notify(604331)
	return s, nil
}

func ComputeColNameInternal(sp sessiondata.SearchPath, target Expr) (int, string, error) {
	__antithesis_instrumentation__.Notify(604338)

	switch e := target.(type) {
	case *UnresolvedName:
		__antithesis_instrumentation__.Notify(604340)
		if e.Star {
			__antithesis_instrumentation__.Notify(604370)
			return 0, "", nil
		} else {
			__antithesis_instrumentation__.Notify(604371)
		}
		__antithesis_instrumentation__.Notify(604341)
		return 2, e.Parts[0], nil

	case *ColumnItem:
		__antithesis_instrumentation__.Notify(604342)
		return 2, e.Column(), nil

	case *IndirectionExpr:
		__antithesis_instrumentation__.Notify(604343)
		return ComputeColNameInternal(sp, e.Expr)

	case *FuncExpr:
		__antithesis_instrumentation__.Notify(604344)
		fd, err := e.Func.Resolve(sp)
		if err != nil {
			__antithesis_instrumentation__.Notify(604372)
			return 0, "", err
		} else {
			__antithesis_instrumentation__.Notify(604373)
		}
		__antithesis_instrumentation__.Notify(604345)
		return 2, fd.Name, nil

	case *NullIfExpr:
		__antithesis_instrumentation__.Notify(604346)
		return 2, "nullif", nil

	case *IfExpr:
		__antithesis_instrumentation__.Notify(604347)
		return 2, "if", nil

	case *ParenExpr:
		__antithesis_instrumentation__.Notify(604348)
		return ComputeColNameInternal(sp, e.Expr)

	case *CastExpr:
		__antithesis_instrumentation__.Notify(604349)
		strength, s, err := ComputeColNameInternal(sp, e.Expr)
		if err != nil {
			__antithesis_instrumentation__.Notify(604374)
			return 0, "", err
		} else {
			__antithesis_instrumentation__.Notify(604375)
		}
		__antithesis_instrumentation__.Notify(604350)
		if strength <= 1 {
			__antithesis_instrumentation__.Notify(604376)
			if typ, ok := GetStaticallyKnownType(e.Type); ok {
				__antithesis_instrumentation__.Notify(604378)
				return 0, computeCastName(typ), nil
			} else {
				__antithesis_instrumentation__.Notify(604379)
			}
			__antithesis_instrumentation__.Notify(604377)
			return 1, e.Type.SQLString(), nil
		} else {
			__antithesis_instrumentation__.Notify(604380)
		}
		__antithesis_instrumentation__.Notify(604351)
		return strength, s, nil

	case *AnnotateTypeExpr:
		__antithesis_instrumentation__.Notify(604352)

		strength, s, err := ComputeColNameInternal(sp, e.Expr)
		if err != nil {
			__antithesis_instrumentation__.Notify(604381)
			return 0, "", err
		} else {
			__antithesis_instrumentation__.Notify(604382)
		}
		__antithesis_instrumentation__.Notify(604353)
		if strength <= 1 {
			__antithesis_instrumentation__.Notify(604383)
			if typ, ok := GetStaticallyKnownType(e.Type); ok {
				__antithesis_instrumentation__.Notify(604385)
				return 0, computeCastName(typ), nil
			} else {
				__antithesis_instrumentation__.Notify(604386)
			}
			__antithesis_instrumentation__.Notify(604384)
			return 1, e.Type.SQLString(), nil
		} else {
			__antithesis_instrumentation__.Notify(604387)
		}
		__antithesis_instrumentation__.Notify(604354)
		return strength, s, nil

	case *CollateExpr:
		__antithesis_instrumentation__.Notify(604355)
		return ComputeColNameInternal(sp, e.Expr)

	case *ArrayFlatten:
		__antithesis_instrumentation__.Notify(604356)
		return 2, "array", nil

	case *Subquery:
		__antithesis_instrumentation__.Notify(604357)
		if e.Exists {
			__antithesis_instrumentation__.Notify(604388)
			return 2, "exists", nil
		} else {
			__antithesis_instrumentation__.Notify(604389)
		}
		__antithesis_instrumentation__.Notify(604358)
		return computeColNameInternalSubquery(sp, e.Select)

	case *CaseExpr:
		__antithesis_instrumentation__.Notify(604359)
		strength, s, err := 0, "", error(nil)
		if e.Else != nil {
			__antithesis_instrumentation__.Notify(604390)
			strength, s, err = ComputeColNameInternal(sp, e.Else)
		} else {
			__antithesis_instrumentation__.Notify(604391)
		}
		__antithesis_instrumentation__.Notify(604360)
		if strength <= 1 {
			__antithesis_instrumentation__.Notify(604392)
			s = "case"
			strength = 1
		} else {
			__antithesis_instrumentation__.Notify(604393)
		}
		__antithesis_instrumentation__.Notify(604361)
		return strength, s, err

	case *Array:
		__antithesis_instrumentation__.Notify(604362)
		return 2, "array", nil

	case *Tuple:
		__antithesis_instrumentation__.Notify(604363)
		if e.Row {
			__antithesis_instrumentation__.Notify(604394)
			return 2, "row", nil
		} else {
			__antithesis_instrumentation__.Notify(604395)
		}
		__antithesis_instrumentation__.Notify(604364)
		if len(e.Exprs) == 1 {
			__antithesis_instrumentation__.Notify(604396)
			if len(e.Labels) > 0 {
				__antithesis_instrumentation__.Notify(604398)
				return 2, e.Labels[0], nil
			} else {
				__antithesis_instrumentation__.Notify(604399)
			}
			__antithesis_instrumentation__.Notify(604397)
			return ComputeColNameInternal(sp, e.Exprs[0])
		} else {
			__antithesis_instrumentation__.Notify(604400)
		}

	case *CoalesceExpr:
		__antithesis_instrumentation__.Notify(604365)
		return 2, "coalesce", nil

	case *IfErrExpr:
		__antithesis_instrumentation__.Notify(604366)
		if e.Else == nil {
			__antithesis_instrumentation__.Notify(604401)
			return 2, "iserror", nil
		} else {
			__antithesis_instrumentation__.Notify(604402)
		}
		__antithesis_instrumentation__.Notify(604367)
		return 2, "iferror", nil

	case *ColumnAccessExpr:
		__antithesis_instrumentation__.Notify(604368)
		return 2, string(e.ColName), nil

	case *DBool:
		__antithesis_instrumentation__.Notify(604369)

		return 1, "bool", nil
	}
	__antithesis_instrumentation__.Notify(604339)

	return 0, "", nil
}

func computeColNameInternalSubquery(
	sp sessiondata.SearchPath, s SelectStatement,
) (int, string, error) {
	__antithesis_instrumentation__.Notify(604403)
	switch e := s.(type) {
	case *ParenSelect:
		__antithesis_instrumentation__.Notify(604405)
		return computeColNameInternalSubquery(sp, e.Select.Select)
	case *ValuesClause:
		__antithesis_instrumentation__.Notify(604406)
		if len(e.Rows) > 0 && func() bool {
			__antithesis_instrumentation__.Notify(604408)
			return len(e.Rows[0]) == 1 == true
		}() == true {
			__antithesis_instrumentation__.Notify(604409)
			return 2, "column1", nil
		} else {
			__antithesis_instrumentation__.Notify(604410)
		}
	case *SelectClause:
		__antithesis_instrumentation__.Notify(604407)
		if len(e.Exprs) == 1 {
			__antithesis_instrumentation__.Notify(604411)
			if len(e.Exprs[0].As) > 0 {
				__antithesis_instrumentation__.Notify(604413)
				return 2, string(e.Exprs[0].As), nil
			} else {
				__antithesis_instrumentation__.Notify(604414)
			}
			__antithesis_instrumentation__.Notify(604412)
			return ComputeColNameInternal(sp, e.Exprs[0].Expr)
		} else {
			__antithesis_instrumentation__.Notify(604415)
		}
	}
	__antithesis_instrumentation__.Notify(604404)
	return 0, "", nil
}

func computeCastName(typ *types.T) string {
	__antithesis_instrumentation__.Notify(604416)

	if typ.Family() == types.ArrayFamily {
		__antithesis_instrumentation__.Notify(604418)
		typ = typ.ArrayContents()
	} else {
		__antithesis_instrumentation__.Notify(604419)
	}
	__antithesis_instrumentation__.Notify(604417)
	return typ.PGName()

}
