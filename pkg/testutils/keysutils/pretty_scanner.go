package keysutils

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/cockroach/pkg/keys"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/sql/rowenc/keyside"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/encoding"
	"github.com/cockroachdb/cockroach/pkg/util/keysutil"
)

func MakePrettyScannerForNamedTables(
	tableNameToID map[string]int, idxNameToID map[string]int,
) keysutil.PrettyScanner {
	__antithesis_instrumentation__.Notify(644381)
	return keysutil.MakePrettyScanner(func(input string) (string, roachpb.Key) {
		__antithesis_instrumentation__.Notify(644382)
		remainder, k := parseTableKeysAsAscendingInts(input, tableNameToID, idxNameToID)
		return remainder, k
	})
}

func parseTableKeysAsAscendingInts(
	input string, tableNameToID map[string]int, idxNameToID map[string]int,
) (string, roachpb.Key) {
	__antithesis_instrumentation__.Notify(644383)

	input = mustShiftSlash(input)
	slashPos := strings.Index(input, "/")
	if slashPos < 0 {
		__antithesis_instrumentation__.Notify(644390)
		slashPos = len(input)
	} else {
		__antithesis_instrumentation__.Notify(644391)
	}
	__antithesis_instrumentation__.Notify(644384)
	remainder := input[slashPos:]
	tableName := input[:slashPos]
	tableID, ok := tableNameToID[tableName]
	if !ok {
		__antithesis_instrumentation__.Notify(644392)
		panic(fmt.Sprintf("unknown table: %s", tableName))
	} else {
		__antithesis_instrumentation__.Notify(644393)
	}
	__antithesis_instrumentation__.Notify(644385)
	output := keys.TODOSQLCodec.TablePrefix(uint32(tableID))
	if remainder == "" {
		__antithesis_instrumentation__.Notify(644394)
		return "", output
	} else {
		__antithesis_instrumentation__.Notify(644395)
	}
	__antithesis_instrumentation__.Notify(644386)
	input = remainder

	input = mustShiftSlash(input)
	slashPos = strings.Index(input, "/")
	if slashPos < 0 {
		__antithesis_instrumentation__.Notify(644396)

		slashPos = len(input)
	} else {
		__antithesis_instrumentation__.Notify(644397)
	}
	__antithesis_instrumentation__.Notify(644387)
	remainder = input[slashPos:]
	idxName := input[:slashPos]
	var idxID int

	if idxName == "pk" {
		__antithesis_instrumentation__.Notify(644398)
		idxID = 1
	} else {
		__antithesis_instrumentation__.Notify(644399)
		idxID, ok = idxNameToID[fmt.Sprintf("%s.%s", tableName, idxName)]
		if !ok {
			__antithesis_instrumentation__.Notify(644400)
			panic(fmt.Sprintf("unknown index: %s", idxName))
		} else {
			__antithesis_instrumentation__.Notify(644401)
		}
	}
	__antithesis_instrumentation__.Notify(644388)
	output = encoding.EncodeUvarintAscending(output, uint64(idxID))
	if remainder == "" {
		__antithesis_instrumentation__.Notify(644402)
		return "", output
	} else {
		__antithesis_instrumentation__.Notify(644403)
	}
	__antithesis_instrumentation__.Notify(644389)

	input = remainder
	remainder, moreOutput := parseAscendingIntIndexKeys(input)
	output = append(output, moreOutput...)
	return remainder, output
}

func mustShiftSlash(in string) string {
	__antithesis_instrumentation__.Notify(644404)
	slash, out := mustShift(in)
	if slash != "/" {
		__antithesis_instrumentation__.Notify(644406)
		panic("expected /: " + in)
	} else {
		__antithesis_instrumentation__.Notify(644407)
	}
	__antithesis_instrumentation__.Notify(644405)
	return out
}

func mustShift(in string) (first, remainder string) {
	__antithesis_instrumentation__.Notify(644408)
	if len(in) == 0 {
		__antithesis_instrumentation__.Notify(644410)
		panic("premature end of string")
	} else {
		__antithesis_instrumentation__.Notify(644411)
	}
	__antithesis_instrumentation__.Notify(644409)
	return in[:1], in[1:]
}

func parseAscendingIntIndexKeys(input string) (string, roachpb.Key) {
	__antithesis_instrumentation__.Notify(644412)
	var key roachpb.Key
	for {
		__antithesis_instrumentation__.Notify(644413)
		remainder, k := parseAscendingIntIndexKey(input)
		if k == nil {
			__antithesis_instrumentation__.Notify(644415)

			return remainder, key
		} else {
			__antithesis_instrumentation__.Notify(644416)
		}
		__antithesis_instrumentation__.Notify(644414)
		key = append(key, k...)
		input = remainder
		if remainder == "" {
			__antithesis_instrumentation__.Notify(644417)

			return "", key
		} else {
			__antithesis_instrumentation__.Notify(644418)
		}
	}
}

func parseAscendingIntIndexKey(input string) (string, roachpb.Key) {
	__antithesis_instrumentation__.Notify(644419)
	origInput := input
	input = mustShiftSlash(input)
	slashPos := strings.Index(input, "/")
	if slashPos < 0 {
		__antithesis_instrumentation__.Notify(644423)

		slashPos = len(input)
	} else {
		__antithesis_instrumentation__.Notify(644424)
	}
	__antithesis_instrumentation__.Notify(644420)
	indexValStr := input[:slashPos]
	datum, err := tree.ParseDInt(indexValStr)
	if err != nil {
		__antithesis_instrumentation__.Notify(644425)

		return origInput, nil
	} else {
		__antithesis_instrumentation__.Notify(644426)
	}
	__antithesis_instrumentation__.Notify(644421)
	remainder := input[slashPos:]
	key, err := keyside.Encode(nil, datum, encoding.Ascending)
	if err != nil {
		__antithesis_instrumentation__.Notify(644427)
		panic(err)
	} else {
		__antithesis_instrumentation__.Notify(644428)
	}
	__antithesis_instrumentation__.Notify(644422)
	return remainder, key
}
