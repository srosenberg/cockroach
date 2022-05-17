package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	gosql "database/sql"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"golang.org/x/exp/rand"
)

type partitionStrategy int

const (
	partitionedReplication partitionStrategy = iota

	partitionedLeases
)

func (ps partitionStrategy) String() string {
	__antithesis_instrumentation__.Notify(698003)
	switch ps {
	case partitionedReplication:
		__antithesis_instrumentation__.Notify(698005)
		return "replication"
	case partitionedLeases:
		__antithesis_instrumentation__.Notify(698006)
		return "leases"
	default:
		__antithesis_instrumentation__.Notify(698007)
	}
	__antithesis_instrumentation__.Notify(698004)
	panic("unexpected")
}

func (ps *partitionStrategy) Set(value string) error {
	__antithesis_instrumentation__.Notify(698008)
	switch value {
	case "replication":
		__antithesis_instrumentation__.Notify(698010)
		*ps = partitionedReplication
		return nil
	case "leases":
		__antithesis_instrumentation__.Notify(698011)
		*ps = partitionedLeases
		return nil
	default:
		__antithesis_instrumentation__.Notify(698012)
	}
	__antithesis_instrumentation__.Notify(698009)
	return errors.Errorf("unknown partition strategy %q", value)
}

func (ps partitionStrategy) Type() string {
	__antithesis_instrumentation__.Notify(698013)
	return "partitionStrategy"
}

type zoneConfig struct {
	zones    []string
	strategy partitionStrategy
}

type survivalGoal int

const (
	survivalGoalZone survivalGoal = iota
	survivalGoalRegion
)

func (s survivalGoal) String() string {
	__antithesis_instrumentation__.Notify(698014)
	switch s {
	case survivalGoalZone:
		__antithesis_instrumentation__.Notify(698016)
		return "zone"
	case survivalGoalRegion:
		__antithesis_instrumentation__.Notify(698017)
		return "region"
	default:
		__antithesis_instrumentation__.Notify(698018)
	}
	__antithesis_instrumentation__.Notify(698015)
	panic("unexpected")
}

func (s *survivalGoal) Set(value string) error {
	__antithesis_instrumentation__.Notify(698019)
	switch value {
	case "zone":
		__antithesis_instrumentation__.Notify(698021)
		*s = survivalGoalZone
	case "region":
		__antithesis_instrumentation__.Notify(698022)
		*s = survivalGoalRegion
	default:
		__antithesis_instrumentation__.Notify(698023)
		return errors.Errorf("unknown survival goal %q", value)
	}
	__antithesis_instrumentation__.Notify(698020)
	return nil
}

func (s survivalGoal) Type() string {
	__antithesis_instrumentation__.Notify(698024)
	return "survival_goal"
}

type multiRegionConfig struct {
	regions      []string
	survivalGoal survivalGoal
}

type partitioner struct {
	total  int
	active int
	parts  int

	partBounds   []int
	partElems    [][]int
	partElemsMap map[int]int
	totalElems   []int
}

func makeMRPartitioner(total, active, parts int) (*partitioner, error) {
	__antithesis_instrumentation__.Notify(698025)
	if total <= 0 {
		__antithesis_instrumentation__.Notify(698035)
		return nil, errors.Errorf("total must be positive; %d", total)
	} else {
		__antithesis_instrumentation__.Notify(698036)
	}
	__antithesis_instrumentation__.Notify(698026)
	if active <= 0 {
		__antithesis_instrumentation__.Notify(698037)
		return nil, errors.Errorf("active must be positive; %d", active)
	} else {
		__antithesis_instrumentation__.Notify(698038)
	}
	__antithesis_instrumentation__.Notify(698027)
	if parts <= 0 {
		__antithesis_instrumentation__.Notify(698039)
		return nil, errors.Errorf("parts must be positive; %d", parts)
	} else {
		__antithesis_instrumentation__.Notify(698040)
	}
	__antithesis_instrumentation__.Notify(698028)
	if active > total {
		__antithesis_instrumentation__.Notify(698041)
		return nil, errors.Errorf("active > total; %d > %d", active, total)
	} else {
		__antithesis_instrumentation__.Notify(698042)
	}
	__antithesis_instrumentation__.Notify(698029)
	if parts > total {
		__antithesis_instrumentation__.Notify(698043)
		return nil, errors.Errorf("parts > total; %d > %d", parts, total)
	} else {
		__antithesis_instrumentation__.Notify(698044)
	}
	__antithesis_instrumentation__.Notify(698030)

	sizes := make([]int, parts)
	adjustment := active % parts
	for i := range sizes {
		__antithesis_instrumentation__.Notify(698045)

		base := active / parts
		if adjustment > 0 {
			__antithesis_instrumentation__.Notify(698047)
			base++
			adjustment--
		} else {
			__antithesis_instrumentation__.Notify(698048)
		}
		__antithesis_instrumentation__.Notify(698046)
		sizes[i] = base
	}
	__antithesis_instrumentation__.Notify(698031)

	partElems := make([][]int, parts)
	for i := 0; i < active; i++ {
		__antithesis_instrumentation__.Notify(698049)
		if i < parts {
			__antithesis_instrumentation__.Notify(698051)
			partAct := make([]int, sizes[i])
			partElems[i] = partAct
		} else {
			__antithesis_instrumentation__.Notify(698052)
		}
		__antithesis_instrumentation__.Notify(698050)
		partElems[i%parts][i/parts] = i
	}
	__antithesis_instrumentation__.Notify(698032)

	partElemsMap := make(map[int]int)
	for p, elems := range partElems {
		__antithesis_instrumentation__.Notify(698053)
		for _, elem := range elems {
			__antithesis_instrumentation__.Notify(698054)
			partElemsMap[elem] = p
		}
	}
	__antithesis_instrumentation__.Notify(698033)

	var totalElems []int
	for _, elems := range partElems {
		__antithesis_instrumentation__.Notify(698055)
		totalElems = append(totalElems, elems...)
	}
	__antithesis_instrumentation__.Notify(698034)

	return &partitioner{
		total:  total,
		active: active,
		parts:  parts,

		partElems:    partElems,
		partElemsMap: partElemsMap,
		totalElems:   totalElems,
	}, nil
}

func makePartitioner(total, active, parts int) (*partitioner, error) {
	__antithesis_instrumentation__.Notify(698056)
	if total <= 0 {
		__antithesis_instrumentation__.Notify(698067)
		return nil, errors.Errorf("total must be positive; %d", total)
	} else {
		__antithesis_instrumentation__.Notify(698068)
	}
	__antithesis_instrumentation__.Notify(698057)
	if active <= 0 {
		__antithesis_instrumentation__.Notify(698069)
		return nil, errors.Errorf("active must be positive; %d", active)
	} else {
		__antithesis_instrumentation__.Notify(698070)
	}
	__antithesis_instrumentation__.Notify(698058)
	if parts <= 0 {
		__antithesis_instrumentation__.Notify(698071)
		return nil, errors.Errorf("parts must be positive; %d", parts)
	} else {
		__antithesis_instrumentation__.Notify(698072)
	}
	__antithesis_instrumentation__.Notify(698059)
	if active > total {
		__antithesis_instrumentation__.Notify(698073)
		return nil, errors.Errorf("active > total; %d > %d", active, total)
	} else {
		__antithesis_instrumentation__.Notify(698074)
	}
	__antithesis_instrumentation__.Notify(698060)
	if parts > total {
		__antithesis_instrumentation__.Notify(698075)
		return nil, errors.Errorf("parts > total; %d > %d", parts, total)
	} else {
		__antithesis_instrumentation__.Notify(698076)
	}
	__antithesis_instrumentation__.Notify(698061)

	bounds := make([]int, parts+1)
	for i := range bounds {
		__antithesis_instrumentation__.Notify(698077)
		bounds[i] = (i * total) / parts
	}
	__antithesis_instrumentation__.Notify(698062)

	sizes := make([]int, parts)
	for i := range sizes {
		__antithesis_instrumentation__.Notify(698078)
		s := (i * active) / parts
		e := ((i + 1) * active) / parts
		sizes[i] = e - s
	}
	__antithesis_instrumentation__.Notify(698063)

	partElems := make([][]int, parts)
	for i := range partElems {
		__antithesis_instrumentation__.Notify(698079)
		partAct := make([]int, sizes[i])
		for j := range partAct {
			__antithesis_instrumentation__.Notify(698081)
			partAct[j] = bounds[i] + j
		}
		__antithesis_instrumentation__.Notify(698080)
		partElems[i] = partAct
	}
	__antithesis_instrumentation__.Notify(698064)

	partElemsMap := make(map[int]int)
	for p, elems := range partElems {
		__antithesis_instrumentation__.Notify(698082)
		for _, elem := range elems {
			__antithesis_instrumentation__.Notify(698083)
			partElemsMap[elem] = p
		}
	}
	__antithesis_instrumentation__.Notify(698065)

	var totalElems []int
	for _, elems := range partElems {
		__antithesis_instrumentation__.Notify(698084)
		totalElems = append(totalElems, elems...)
	}
	__antithesis_instrumentation__.Notify(698066)

	return &partitioner{
		total:  total,
		active: active,
		parts:  parts,

		partBounds:   bounds,
		partElems:    partElems,
		partElemsMap: partElemsMap,
		totalElems:   totalElems,
	}, nil
}

func (p *partitioner) randActive(rng *rand.Rand) int {
	__antithesis_instrumentation__.Notify(698085)
	return p.totalElems[rng.Intn(len(p.totalElems))]
}

func configureZone(db *gosql.DB, cfg zoneConfig, table, partition string, partIdx int) error {
	__antithesis_instrumentation__.Notify(698086)
	var kv string
	if len(cfg.zones) > 0 {
		__antithesis_instrumentation__.Notify(698090)
		kv = fmt.Sprintf("zone=%s", cfg.zones[partIdx])
	} else {
		__antithesis_instrumentation__.Notify(698091)
		kv = fmt.Sprintf("rack=%d", partIdx)
	}
	__antithesis_instrumentation__.Notify(698087)

	var opts string
	switch cfg.strategy {
	case partitionedReplication:
		__antithesis_instrumentation__.Notify(698092)

		opts = fmt.Sprintf(`constraints = '[+%s]'`, kv)
	case partitionedLeases:
		__antithesis_instrumentation__.Notify(698093)

		opts = fmt.Sprintf(`num_replicas = COPY FROM PARENT, constraints = '{"+%s":1}', lease_preferences = '[[+%s]]'`, kv, kv)
	default:
		__antithesis_instrumentation__.Notify(698094)
		panic("unexpected")
	}
	__antithesis_instrumentation__.Notify(698088)

	sql := fmt.Sprintf(`ALTER PARTITION %s OF TABLE %s CONFIGURE ZONE USING %s`,
		partition, table, opts)
	if _, err := db.Exec(sql); err != nil {
		__antithesis_instrumentation__.Notify(698095)
		return errors.Wrapf(err, "Couldn't exec %q", sql)
	} else {
		__antithesis_instrumentation__.Notify(698096)
	}
	__antithesis_instrumentation__.Notify(698089)
	return nil
}

func partitionObject(
	db *gosql.DB, cfg zoneConfig, p *partitioner, obj, name, col, table string, idx int,
) error {
	__antithesis_instrumentation__.Notify(698097)
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "ALTER %s %s PARTITION BY RANGE (%s) (\n", obj, name, col)
	for i := 0; i < p.parts; i++ {
		__antithesis_instrumentation__.Notify(698101)
		fmt.Fprintf(&buf, "  PARTITION p%d_%d VALUES FROM (%d) to (%d)",
			idx, i, p.partBounds[i], p.partBounds[i+1])
		if i+1 < p.parts {
			__antithesis_instrumentation__.Notify(698103)
			buf.WriteString(",")
		} else {
			__antithesis_instrumentation__.Notify(698104)
		}
		__antithesis_instrumentation__.Notify(698102)
		buf.WriteString("\n")
	}
	__antithesis_instrumentation__.Notify(698098)
	buf.WriteString(")\n")
	if _, err := db.Exec(buf.String()); err != nil {
		__antithesis_instrumentation__.Notify(698105)
		return errors.Wrapf(err, "Couldn't exec %q", buf.String())
	} else {
		__antithesis_instrumentation__.Notify(698106)
	}
	__antithesis_instrumentation__.Notify(698099)

	for i := 0; i < p.parts; i++ {
		__antithesis_instrumentation__.Notify(698107)
		if err := configureZone(db, cfg, table, fmt.Sprintf("p%d_%d", idx, i), i); err != nil {
			__antithesis_instrumentation__.Notify(698108)
			return err
		} else {
			__antithesis_instrumentation__.Notify(698109)
		}
	}
	__antithesis_instrumentation__.Notify(698100)
	return nil
}

func partitionTable(
	db *gosql.DB, cfg zoneConfig, p *partitioner, table, col string, idx int,
) error {
	__antithesis_instrumentation__.Notify(698110)
	return partitionObject(db, cfg, p, "TABLE", table, col, table, idx)
}

func partitionIndex(
	db *gosql.DB, cfg zoneConfig, p *partitioner, table, index, col string, idx int,
) error {
	__antithesis_instrumentation__.Notify(698111)
	indexStr := fmt.Sprintf("%s@%s", table, index)
	if exists, err := indexExists(db, table, index); err != nil {
		__antithesis_instrumentation__.Notify(698113)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698114)
		if !exists {
			__antithesis_instrumentation__.Notify(698115)
			return errors.Errorf("could not find index %q", indexStr)
		} else {
			__antithesis_instrumentation__.Notify(698116)
		}
	}
	__antithesis_instrumentation__.Notify(698112)
	return partitionObject(db, cfg, p, "INDEX", indexStr, col, table, idx)
}

func partitionWarehouse(db *gosql.DB, cfg zoneConfig, wPart *partitioner) error {
	__antithesis_instrumentation__.Notify(698117)
	return partitionTable(db, cfg, wPart, "warehouse", "w_id", 0)
}

func partitionDistrict(db *gosql.DB, cfg zoneConfig, wPart *partitioner) error {
	__antithesis_instrumentation__.Notify(698118)
	return partitionTable(db, cfg, wPart, "district", "d_w_id", 0)
}

func partitionNewOrder(db *gosql.DB, cfg zoneConfig, wPart *partitioner) error {
	__antithesis_instrumentation__.Notify(698119)
	return partitionTable(db, cfg, wPart, "new_order", "no_w_id", 0)
}

func partitionOrder(db *gosql.DB, cfg zoneConfig, wPart *partitioner) error {
	__antithesis_instrumentation__.Notify(698120)
	if err := partitionTable(db, cfg, wPart, `"order"`, "o_w_id", 0); err != nil {
		__antithesis_instrumentation__.Notify(698122)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698123)
	}
	__antithesis_instrumentation__.Notify(698121)
	return partitionIndex(db, cfg, wPart, `"order"`, "order_idx", "o_w_id", 1)
}

func partitionOrderLine(db *gosql.DB, cfg zoneConfig, wPart *partitioner) error {
	__antithesis_instrumentation__.Notify(698124)
	return partitionTable(db, cfg, wPart, "order_line", "ol_w_id", 0)
}

func partitionStock(db *gosql.DB, cfg zoneConfig, wPart *partitioner) error {
	__antithesis_instrumentation__.Notify(698125)
	return partitionTable(db, cfg, wPart, "stock", "s_w_id", 0)
}

func partitionCustomer(db *gosql.DB, cfg zoneConfig, wPart *partitioner) error {
	__antithesis_instrumentation__.Notify(698126)
	if err := partitionTable(db, cfg, wPart, "customer", "c_w_id", 0); err != nil {
		__antithesis_instrumentation__.Notify(698128)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698129)
	}
	__antithesis_instrumentation__.Notify(698127)
	return partitionIndex(db, cfg, wPart, "customer", "customer_idx", "c_w_id", 1)
}

func partitionHistory(db *gosql.DB, cfg zoneConfig, wPart *partitioner) error {
	__antithesis_instrumentation__.Notify(698130)
	return partitionTable(db, cfg, wPart, "history", "h_w_id", 0)
}

func replicateColumns(
	db *gosql.DB,
	cfg zoneConfig,
	wPart *partitioner,
	name string,
	pkColumns []string,
	storedColumns []string,
) error {
	__antithesis_instrumentation__.Notify(698131)
	constraints := synthesizeConstraints(cfg, wPart)
	for i, constraint := range constraints {
		__antithesis_instrumentation__.Notify(698133)
		if _, err := db.Exec(
			fmt.Sprintf(`CREATE UNIQUE INDEX %[1]s_idx_%[2]d ON %[1]s (%[3]s) STORING (%[4]s)`,
				name, i, strings.Join(pkColumns, ","), strings.Join(storedColumns, ",")),
		); err != nil {
			__antithesis_instrumentation__.Notify(698135)
			return err
		} else {
			__antithesis_instrumentation__.Notify(698136)
		}
		__antithesis_instrumentation__.Notify(698134)
		if _, err := db.Exec(fmt.Sprintf(
			`ALTER INDEX %[1]s@%[1]s_idx_%[2]d
CONFIGURE ZONE USING num_replicas = COPY FROM PARENT, constraints='{"%[3]s": 1}', lease_preferences='[[%[3]s]]'`,
			name, i, constraint)); err != nil {
			__antithesis_instrumentation__.Notify(698137)
			return err
		} else {
			__antithesis_instrumentation__.Notify(698138)
		}
	}
	__antithesis_instrumentation__.Notify(698132)
	return nil
}

func replicateWarehouse(db *gosql.DB, cfg zoneConfig, wPart *partitioner) error {
	__antithesis_instrumentation__.Notify(698139)
	return replicateColumns(db, cfg, wPart, "warehouse", []string{"w_id"}, []string{"w_tax"})
}

func replicateDistrict(db *gosql.DB, cfg zoneConfig, wPart *partitioner) error {
	__antithesis_instrumentation__.Notify(698140)
	return replicateColumns(db, cfg, wPart, "district", []string{"d_w_id", "d_id"},
		[]string{"d_name", "d_street_1", "d_street_2", "d_city", "d_state", "d_zip"})
}

func replicateItem(db *gosql.DB, cfg zoneConfig, wPart *partitioner) error {
	__antithesis_instrumentation__.Notify(698141)
	return replicateColumns(db, cfg, wPart, "item", []string{"i_id"},
		[]string{"i_im_id", "i_name", "i_price", "i_data"})
}

func synthesizeConstraints(cfg zoneConfig, wPart *partitioner) []string {
	__antithesis_instrumentation__.Notify(698142)
	var constraints []string
	if len(cfg.zones) > 0 {
		__antithesis_instrumentation__.Notify(698144)
		for _, zone := range cfg.zones {
			__antithesis_instrumentation__.Notify(698145)
			constraints = append(constraints, "+zone="+zone)
		}
	} else {
		__antithesis_instrumentation__.Notify(698146)

		for i := 0; i < wPart.parts; i++ {
			__antithesis_instrumentation__.Notify(698147)
			constraints = append(constraints, fmt.Sprintf("+rack=%d", i))
		}
	}
	__antithesis_instrumentation__.Notify(698143)
	return constraints
}

func partitionTables(
	db *gosql.DB, cfg zoneConfig, wPart *partitioner, replicateStaticColumns bool,
) error {
	__antithesis_instrumentation__.Notify(698148)
	if err := partitionWarehouse(db, cfg, wPart); err != nil {
		__antithesis_instrumentation__.Notify(698158)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698159)
	}
	__antithesis_instrumentation__.Notify(698149)
	if err := partitionDistrict(db, cfg, wPart); err != nil {
		__antithesis_instrumentation__.Notify(698160)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698161)
	}
	__antithesis_instrumentation__.Notify(698150)
	if err := partitionNewOrder(db, cfg, wPart); err != nil {
		__antithesis_instrumentation__.Notify(698162)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698163)
	}
	__antithesis_instrumentation__.Notify(698151)
	if err := partitionOrder(db, cfg, wPart); err != nil {
		__antithesis_instrumentation__.Notify(698164)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698165)
	}
	__antithesis_instrumentation__.Notify(698152)
	if err := partitionOrderLine(db, cfg, wPart); err != nil {
		__antithesis_instrumentation__.Notify(698166)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698167)
	}
	__antithesis_instrumentation__.Notify(698153)
	if err := partitionStock(db, cfg, wPart); err != nil {
		__antithesis_instrumentation__.Notify(698168)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698169)
	}
	__antithesis_instrumentation__.Notify(698154)
	if err := partitionCustomer(db, cfg, wPart); err != nil {
		__antithesis_instrumentation__.Notify(698170)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698171)
	}
	__antithesis_instrumentation__.Notify(698155)
	if err := partitionHistory(db, cfg, wPart); err != nil {
		__antithesis_instrumentation__.Notify(698172)
		return err
	} else {
		__antithesis_instrumentation__.Notify(698173)
	}
	__antithesis_instrumentation__.Notify(698156)
	if replicateStaticColumns {
		__antithesis_instrumentation__.Notify(698174)
		if err := replicateDistrict(db, cfg, wPart); err != nil {
			__antithesis_instrumentation__.Notify(698176)
			return err
		} else {
			__antithesis_instrumentation__.Notify(698177)
		}
		__antithesis_instrumentation__.Notify(698175)
		if err := replicateWarehouse(db, cfg, wPart); err != nil {
			__antithesis_instrumentation__.Notify(698178)
			return err
		} else {
			__antithesis_instrumentation__.Notify(698179)
		}
	} else {
		__antithesis_instrumentation__.Notify(698180)
	}
	__antithesis_instrumentation__.Notify(698157)
	return replicateItem(db, cfg, wPart)
}

func partitionCount(db *gosql.DB) (int, error) {
	__antithesis_instrumentation__.Notify(698181)
	var count int
	if err := db.QueryRow(`
		SELECT count(*)
		FROM crdb_internal.tables t
		JOIN crdb_internal.partitions p
		USING (table_id)
		WHERE t.name = 'warehouse'
		AND p.name ~ 'p0_\d+'
	`).Scan(&count); err != nil {
		__antithesis_instrumentation__.Notify(698183)
		return 0, err
	} else {
		__antithesis_instrumentation__.Notify(698184)
	}
	__antithesis_instrumentation__.Notify(698182)
	return count, nil
}

func indexExists(db *gosql.DB, table, index string) (bool, error) {
	__antithesis_instrumentation__.Notify(698185)

	table = strings.ReplaceAll(table, `"`, ``)

	var exists bool
	if err := db.QueryRow(`
		SELECT count(*) > 0
		FROM information_schema.statistics
		WHERE table_name = $1
		AND   index_name = $2
	`, table, index).Scan(&exists); err != nil {
		__antithesis_instrumentation__.Notify(698187)
		return false, err
	} else {
		__antithesis_instrumentation__.Notify(698188)
	}
	__antithesis_instrumentation__.Notify(698186)
	return exists, nil
}
