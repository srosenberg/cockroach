package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"golang.org/x/sync/errgroup"
)

const (
	tpccWarehouseSchema = `(
		w_id        integer       not null primary key,
		w_name      varchar(10)   not null,
		w_street_1  varchar(20)   not null,
		w_street_2  varchar(20)   not null,
		w_city      varchar(20)   not null,
		w_state     char(2)       not null,
		w_zip       char(9)       not null,
		w_tax       decimal(4,4)  not null,
		w_ytd       decimal(12,2) not null`
	tpccWarehouseColumnFamiliesSuffix = `
		family      f1 (w_id, w_name, w_street_1, w_street_2, w_city, w_state, w_zip, w_ytd),
		family      f2 (w_tax)`

	tpccDistrictSchemaBase = `(
		d_id         integer       not null,
		d_w_id       integer       not null,
		d_name       varchar(10)   not null,
		d_street_1   varchar(20)   not null,
		d_street_2   varchar(20)   not null,
		d_city       varchar(20)   not null,
		d_state      char(2)       not null,
		d_zip        char(9)       not null,
		d_tax        decimal(4,4)  not null,
		d_ytd        decimal(12,2) not null,
		d_next_o_id  integer       not null,
		primary key  (d_w_id, d_id)`
	tpccDistrictColumnFamiliesSuffix = `
		family       static    (d_w_id, d_id, d_name, d_street_1, d_street_2, d_city, d_state, d_zip),
		family       dynamic_1 (d_ytd),
		family       dynamic_2 (d_next_o_id, d_tax)`

	tpccCustomerSchemaBase = `(
		c_id           integer       not null,
		c_d_id         integer       not null,
		c_w_id         integer       not null,
		c_first        varchar(16)   not null,
		c_middle       char(2)       not null,
		c_last         varchar(16)   not null,
		c_street_1     varchar(20)   not null,
		c_street_2     varchar(20)   not null,
		c_city         varchar(20)   not null,
		c_state        char(2)       not null,
		c_zip          char(9)       not null,
		c_phone        char(16)      not null,
		c_since        timestamp     not null,
		c_credit       char(2)       not null,
		c_credit_lim   decimal(12,2) not null,
		c_discount     decimal(4,4)  not null,
		c_balance      decimal(12,2) not null,
		c_ytd_payment  decimal(12,2) not null,
		c_payment_cnt  integer       not null,
		c_delivery_cnt integer       not null,
		c_data         varchar(500)  not null,
		primary key        (c_w_id, c_d_id, c_id),
		index customer_idx (c_w_id, c_d_id, c_last, c_first)`
	tpccCustomerColumnFamiliesSuffix = `
		family static      (
			c_id, c_d_id, c_w_id, c_first, c_middle, c_last, c_street_1, c_street_2,
			c_city, c_state, c_zip, c_phone, c_since, c_credit, c_credit_lim, c_discount
		),
		family dynamic (c_balance, c_ytd_payment, c_payment_cnt, c_data, c_delivery_cnt)`

	tpccHistorySchemaBase = `(
		rowid    uuid    not null default gen_random_uuid(),
		h_c_id   integer not null,
		h_c_d_id integer not null,
		h_c_w_id integer not null,
		h_d_id   integer not null,
		h_w_id   integer not null,
		h_date   timestamp,
		h_amount decimal(6,2),
		h_data   varchar(24),
		primary key (h_w_id, rowid)`
	deprecatedTpccHistorySchemaFkSuffix = `
		index history_customer_fk_idx (h_c_w_id, h_c_d_id, h_c_id),
		index history_district_fk_idx (h_w_id, h_d_id)`

	tpccOrderSchemaBase = `(
		o_id         integer      not null,
		o_d_id       integer      not null,
		o_w_id       integer      not null,
		o_c_id       integer,
		o_entry_d    timestamp,
		o_carrier_id integer,
		o_ol_cnt     integer,
		o_all_local  integer,
		primary key  (o_w_id, o_d_id, o_id DESC),
		unique index order_idx (o_w_id, o_d_id, o_c_id, o_id DESC) storing (o_entry_d, o_carrier_id)
	`

	tpccNewOrderSchema = `(
		no_o_id  integer   not null,
		no_d_id  integer   not null,
		no_w_id  integer   not null,
		primary key (no_w_id, no_d_id, no_o_id)
	`

	tpccItemSchema = `(
		i_id     integer      not null,
		i_im_id  integer,
		i_name   varchar(24),
		i_price  decimal(5,2),
		i_data   varchar(50),
		primary key (i_id)
	`

	tpccStockSchemaBase = `(
		s_i_id       integer       not null,
		s_w_id       integer       not null,
		s_quantity   integer,
		s_dist_01    char(24),
		s_dist_02    char(24),
		s_dist_03    char(24),
		s_dist_04    char(24),
		s_dist_05    char(24),
		s_dist_06    char(24),
		s_dist_07    char(24),
		s_dist_08    char(24),
		s_dist_09    char(24),
		s_dist_10    char(24),
		s_ytd        integer,
		s_order_cnt  integer,
		s_remote_cnt integer,
		s_data       varchar(50),
		primary key (s_w_id, s_i_id)`
	deprecatedTpccStockSchemaFkSuffix = `
		index stock_item_fk_idx (s_i_id)`

	tpccOrderLineSchemaBase = `(
		ol_o_id         integer   not null,
		ol_d_id         integer   not null,
		ol_w_id         integer   not null,
		ol_number       integer   not null,
		ol_i_id         integer   not null,
		ol_supply_w_id  integer,
		ol_delivery_d   timestamp,
		ol_quantity     integer,
		ol_amount       decimal(6,2),
		ol_dist_info    char(24),
		primary key (ol_w_id, ol_d_id, ol_o_id DESC, ol_number)`
	deprecatedTpccOrderLineSchemaFkSuffix = `
		index order_line_stock_fk_idx (ol_supply_w_id, ol_i_id)`

	localityRegionalByRowSuffix = `
		locality regional by row`
	localityGlobalSuffix = `
		locality global`

	endSchema = "\n\t)"
)

type schemaOptions struct {
	fkClause       string
	familyClause   string
	columnClause   string
	localityClause string
}

type makeSchemaOption func(o *schemaOptions)

func maybeAddFkSuffix(fks bool, suffix string) makeSchemaOption {
	__antithesis_instrumentation__.Notify(697718)
	return func(o *schemaOptions) {
		__antithesis_instrumentation__.Notify(697719)
		if fks {
			__antithesis_instrumentation__.Notify(697720)
			o.fkClause = suffix
		} else {
			__antithesis_instrumentation__.Notify(697721)
		}
	}
}

func maybeAddColumnFamiliesSuffix(separateColumnFamilies bool, suffix string) makeSchemaOption {
	__antithesis_instrumentation__.Notify(697722)
	return func(o *schemaOptions) {
		__antithesis_instrumentation__.Notify(697723)
		if separateColumnFamilies {
			__antithesis_instrumentation__.Notify(697724)
			o.familyClause = suffix
		} else {
			__antithesis_instrumentation__.Notify(697725)
		}
	}
}

func maybeAddLocalityRegionalByRow(
	multiRegionCfg multiRegionConfig, partColName string,
) makeSchemaOption {
	__antithesis_instrumentation__.Notify(697726)
	return func(o *schemaOptions) {
		__antithesis_instrumentation__.Notify(697727)
		if len(multiRegionCfg.regions) > 0 {
			__antithesis_instrumentation__.Notify(697728)

			var b strings.Builder
			fmt.Fprintf(&b, `
               crdb_region crdb_internal_region NOT VISIBLE NOT NULL AS (
                       CASE %s %% %d`, partColName, len(multiRegionCfg.regions))
			for i, region := range multiRegionCfg.regions {
				__antithesis_instrumentation__.Notify(697730)
				fmt.Fprintf(&b, `
                       WHEN %d THEN '%s'`, i, region)
			}
			__antithesis_instrumentation__.Notify(697729)
			b.WriteString(`
                       END
               ) STORED`)
			o.columnClause = b.String()
			o.localityClause = localityRegionalByRowSuffix
		} else {
			__antithesis_instrumentation__.Notify(697731)
		}
	}
}

func makeSchema(base string, opts ...makeSchemaOption) string {
	__antithesis_instrumentation__.Notify(697732)
	var o schemaOptions
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(697738)
		opt(&o)
	}
	__antithesis_instrumentation__.Notify(697733)
	ret := base
	if o.fkClause != "" {
		__antithesis_instrumentation__.Notify(697739)
		ret += "," + o.fkClause
	} else {
		__antithesis_instrumentation__.Notify(697740)
	}
	__antithesis_instrumentation__.Notify(697734)
	if o.familyClause != "" {
		__antithesis_instrumentation__.Notify(697741)
		ret += "," + o.familyClause
	} else {
		__antithesis_instrumentation__.Notify(697742)
	}
	__antithesis_instrumentation__.Notify(697735)
	if o.columnClause != "" {
		__antithesis_instrumentation__.Notify(697743)
		ret += "," + o.columnClause
	} else {
		__antithesis_instrumentation__.Notify(697744)
	}
	__antithesis_instrumentation__.Notify(697736)
	ret += endSchema
	if o.localityClause != "" {
		__antithesis_instrumentation__.Notify(697745)
		ret += o.localityClause
	} else {
		__antithesis_instrumentation__.Notify(697746)
	}
	__antithesis_instrumentation__.Notify(697737)
	return ret
}

func scatterRanges(db *gosql.DB) error {
	__antithesis_instrumentation__.Notify(697747)
	tables := []string{
		`customer`,
		`district`,
		`history`,
		`item`,
		`new_order`,
		`"order"`,
		`order_line`,
		`stock`,
		`warehouse`,
	}

	var g errgroup.Group
	for _, table := range tables {
		__antithesis_instrumentation__.Notify(697749)
		sql := fmt.Sprintf(`ALTER TABLE %s SCATTER`, table)
		g.Go(func() error {
			__antithesis_instrumentation__.Notify(697750)
			if _, err := db.Exec(sql); err != nil {
				__antithesis_instrumentation__.Notify(697752)
				return errors.Wrapf(err, "Couldn't exec %q", sql)
			} else {
				__antithesis_instrumentation__.Notify(697753)
			}
			__antithesis_instrumentation__.Notify(697751)
			return nil
		})
	}
	__antithesis_instrumentation__.Notify(697748)
	return g.Wait()
}
