package tpcc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	gosql "database/sql"

	"github.com/cockroachdb/errors"
)

type Check struct {
	Name string

	Fn        func(db *gosql.DB, asOfSystemTime string) error
	Expensive bool
}

func AllChecks() []Check {
	__antithesis_instrumentation__.Notify(697536)
	return []Check{
		{"3.3.2.1", check3321, false},
		{"3.3.2.2", check3322, false},
		{"3.3.2.3", check3323, false},
		{"3.3.2.4", check3324, false},
		{"3.3.2.5", check3325, false},
		{"3.3.2.6", check3326, true},
		{"3.3.2.7", check3327, false},
		{"3.3.2.8", check3328, false},
		{"3.3.2.9", check3329, false},
	}
}

func check3321(db *gosql.DB, asOfSystemTime string) error {
	__antithesis_instrumentation__.Notify(697537)

	txn, err := beginAsOfSystemTime(db, asOfSystemTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(697542)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697543)
	}
	__antithesis_instrumentation__.Notify(697538)
	defer func() { __antithesis_instrumentation__.Notify(697544); _ = txn.Rollback() }()
	__antithesis_instrumentation__.Notify(697539)
	row := txn.QueryRow(`
SELECT
    count(*)
FROM
    warehouse
    FULL JOIN (
            SELECT
                d_w_id, sum(d_ytd) AS sum_d_ytd
            FROM
                district
            GROUP BY
                d_w_id
        ) ON w_id = d_w_id
WHERE
    w_ytd != sum_d_ytd
`)
	var i int
	if err := row.Scan(&i); err != nil {
		__antithesis_instrumentation__.Notify(697545)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697546)
	}
	__antithesis_instrumentation__.Notify(697540)

	if i != 0 {
		__antithesis_instrumentation__.Notify(697547)
		return errors.Errorf("%d rows returned, expected zero", i)
	} else {
		__antithesis_instrumentation__.Notify(697548)
	}
	__antithesis_instrumentation__.Notify(697541)

	return nil
}

func check3322(db *gosql.DB, asOfSystemTime string) (retErr error) {
	__antithesis_instrumentation__.Notify(697549)

	txn, err := beginAsOfSystemTime(db, asOfSystemTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(697561)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697562)
	}
	__antithesis_instrumentation__.Notify(697550)
	ts, err := selectTimestamp(txn)
	_ = txn.Rollback()
	if err != nil {
		__antithesis_instrumentation__.Notify(697563)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697564)
	}
	__antithesis_instrumentation__.Notify(697551)
	districtRowsQuery := `
SELECT
    d_next_o_id
FROM
    district AS OF SYSTEM TIME '` + ts + `'
ORDER BY
    d_w_id, d_id`
	districtRows, err := db.Query(districtRowsQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(697565)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697566)
	}
	__antithesis_instrumentation__.Notify(697552)
	defer func() {
		__antithesis_instrumentation__.Notify(697567)
		retErr = errors.CombineErrors(retErr, districtRows.Close())
	}()
	__antithesis_instrumentation__.Notify(697553)
	newOrderQuery := `
SELECT
    max(no_o_id)
FROM
    new_order AS OF SYSTEM TIME '` + ts + `'
GROUP BY
    no_d_id, no_w_id
ORDER BY
    no_w_id, no_d_id;`
	newOrderRows, err := db.Query(newOrderQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(697568)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697569)
	}
	__antithesis_instrumentation__.Notify(697554)
	defer func() {
		__antithesis_instrumentation__.Notify(697570)
		retErr = errors.CombineErrors(retErr, newOrderRows.Close())
	}()
	__antithesis_instrumentation__.Notify(697555)
	orderRowsQuery := `
SELECT
    max(o_id)
FROM
    "order" AS OF SYSTEM TIME '` + ts + `'
GROUP BY
    o_d_id, o_w_id
ORDER BY
    o_w_id, o_d_id`
	orderRows, err := db.Query(orderRowsQuery)
	if err != nil {
		__antithesis_instrumentation__.Notify(697571)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697572)
	}
	__antithesis_instrumentation__.Notify(697556)
	defer func() {
		__antithesis_instrumentation__.Notify(697573)
		retErr = errors.CombineErrors(retErr, orderRows.Close())
	}()
	__antithesis_instrumentation__.Notify(697557)
	var district, newOrder, order float64
	var i int
	for ; districtRows.Next() && func() bool {
		__antithesis_instrumentation__.Notify(697574)
		return newOrderRows.Next() == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(697575)
		return orderRows.Next() == true
	}() == true; i++ {
		__antithesis_instrumentation__.Notify(697576)
		if err := districtRows.Scan(&district); err != nil {
			__antithesis_instrumentation__.Notify(697580)
			return err
		} else {
			__antithesis_instrumentation__.Notify(697581)
		}
		__antithesis_instrumentation__.Notify(697577)
		if err := newOrderRows.Scan(&newOrder); err != nil {
			__antithesis_instrumentation__.Notify(697582)
			return err
		} else {
			__antithesis_instrumentation__.Notify(697583)
		}
		__antithesis_instrumentation__.Notify(697578)
		if err := orderRows.Scan(&order); err != nil {
			__antithesis_instrumentation__.Notify(697584)
			return err
		} else {
			__antithesis_instrumentation__.Notify(697585)
		}
		__antithesis_instrumentation__.Notify(697579)

		if (order != newOrder) || func() bool {
			__antithesis_instrumentation__.Notify(697586)
			return (order != (district - 1)) == true
		}() == true {
			__antithesis_instrumentation__.Notify(697587)
			return errors.Errorf("inequality at idx %d: order: %f, newOrder: %f, district-1: %f",
				i, order, newOrder, district-1)
		} else {
			__antithesis_instrumentation__.Notify(697588)
		}
	}
	__antithesis_instrumentation__.Notify(697558)
	if districtRows.Next() || func() bool {
		__antithesis_instrumentation__.Notify(697589)
		return newOrderRows.Next() == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(697590)
		return orderRows.Next() == true
	}() == true {
		__antithesis_instrumentation__.Notify(697591)
		return errors.New("length mismatch between rows")
	} else {
		__antithesis_instrumentation__.Notify(697592)
	}
	__antithesis_instrumentation__.Notify(697559)
	if i == 0 {
		__antithesis_instrumentation__.Notify(697593)
		return errors.Errorf("zero rows")
	} else {
		__antithesis_instrumentation__.Notify(697594)
	}
	__antithesis_instrumentation__.Notify(697560)
	retErr = errors.CombineErrors(retErr, districtRows.Err())
	retErr = errors.CombineErrors(retErr, newOrderRows.Err())
	return errors.CombineErrors(retErr, orderRows.Err())
}

func check3323(db *gosql.DB, asOfSystemTime string) error {
	__antithesis_instrumentation__.Notify(697595)

	txn, err := beginAsOfSystemTime(db, asOfSystemTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(697600)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697601)
	}
	__antithesis_instrumentation__.Notify(697596)
	defer func() { __antithesis_instrumentation__.Notify(697602); _ = txn.Rollback() }()
	__antithesis_instrumentation__.Notify(697597)
	row := txn.QueryRow(`
SELECT
    count(*)
FROM
    (
        SELECT
            max(no_o_id) - min(no_o_id) - count(*) AS nod
        FROM
            new_order
        GROUP BY
            no_w_id, no_d_id
    )
WHERE
    nod != -1`)

	var i int
	if err := row.Scan(&i); err != nil {
		__antithesis_instrumentation__.Notify(697603)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697604)
	}
	__antithesis_instrumentation__.Notify(697598)

	if i != 0 {
		__antithesis_instrumentation__.Notify(697605)
		return errors.Errorf("%d rows returned, expected zero", i)
	} else {
		__antithesis_instrumentation__.Notify(697606)
	}
	__antithesis_instrumentation__.Notify(697599)

	return nil
}

func check3324(db *gosql.DB, asOfSystemTime string) (retErr error) {
	__antithesis_instrumentation__.Notify(697607)

	txn, err := beginAsOfSystemTime(db, asOfSystemTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(697619)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697620)
	}
	__antithesis_instrumentation__.Notify(697608)

	ts, err := selectTimestamp(txn)
	_ = txn.Rollback()
	if err != nil {
		__antithesis_instrumentation__.Notify(697621)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697622)
	}
	__antithesis_instrumentation__.Notify(697609)
	leftRows, err := db.Query(`
SELECT
    sum(o_ol_cnt)
FROM
    "order" AS OF SYSTEM TIME '` + ts + `'
GROUP BY
    o_w_id, o_d_id
ORDER BY
    o_w_id, o_d_id`)
	if err != nil {
		__antithesis_instrumentation__.Notify(697623)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697624)
	}
	__antithesis_instrumentation__.Notify(697610)
	defer func() {
		__antithesis_instrumentation__.Notify(697625)
		retErr = errors.CombineErrors(retErr, leftRows.Close())
	}()
	__antithesis_instrumentation__.Notify(697611)
	rightRows, err := db.Query(`
SELECT
    count(*)
FROM
    order_line AS OF SYSTEM TIME '` + ts + `'
GROUP BY
    ol_w_id, ol_d_id
ORDER BY
    ol_w_id, ol_d_id`)
	if err != nil {
		__antithesis_instrumentation__.Notify(697626)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697627)
	}
	__antithesis_instrumentation__.Notify(697612)
	defer func() {
		__antithesis_instrumentation__.Notify(697628)
		retErr = errors.CombineErrors(retErr, rightRows.Close())
	}()
	__antithesis_instrumentation__.Notify(697613)
	var i int
	var left, right int64
	for ; leftRows.Next() && func() bool {
		__antithesis_instrumentation__.Notify(697629)
		return rightRows.Next() == true
	}() == true; i++ {
		__antithesis_instrumentation__.Notify(697630)
		if err := leftRows.Scan(&left); err != nil {
			__antithesis_instrumentation__.Notify(697633)
			return err
		} else {
			__antithesis_instrumentation__.Notify(697634)
		}
		__antithesis_instrumentation__.Notify(697631)
		if err := rightRows.Scan(&right); err != nil {
			__antithesis_instrumentation__.Notify(697635)
			return err
		} else {
			__antithesis_instrumentation__.Notify(697636)
		}
		__antithesis_instrumentation__.Notify(697632)
		if left != right {
			__antithesis_instrumentation__.Notify(697637)
			return errors.Errorf("order.sum(o_ol_cnt): %d != order_line.count(*): %d", left, right)
		} else {
			__antithesis_instrumentation__.Notify(697638)
		}
	}
	__antithesis_instrumentation__.Notify(697614)
	if leftRows.Next() || func() bool {
		__antithesis_instrumentation__.Notify(697639)
		return rightRows.Next() == true
	}() == true {
		__antithesis_instrumentation__.Notify(697640)
		return errors.Errorf("at %s: length of order.sum(o_ol_cnt) != order_line.count(*)", ts)
	} else {
		__antithesis_instrumentation__.Notify(697641)
	}
	__antithesis_instrumentation__.Notify(697615)
	if i == 0 {
		__antithesis_instrumentation__.Notify(697642)
		return errors.Errorf("0 rows returned")
	} else {
		__antithesis_instrumentation__.Notify(697643)
	}
	__antithesis_instrumentation__.Notify(697616)
	if err := leftRows.Err(); err != nil {
		__antithesis_instrumentation__.Notify(697644)
		return errors.Wrap(err, "on `order`")
	} else {
		__antithesis_instrumentation__.Notify(697645)
	}
	__antithesis_instrumentation__.Notify(697617)
	if err := rightRows.Err(); err != nil {
		__antithesis_instrumentation__.Notify(697646)
		return errors.Wrap(err, "on `order_line`")
	} else {
		__antithesis_instrumentation__.Notify(697647)
	}
	__antithesis_instrumentation__.Notify(697618)
	return nil
}

func check3325(db *gosql.DB, asOfSystemTime string) (retErr error) {
	__antithesis_instrumentation__.Notify(697648)

	txn, err := beginAsOfSystemTime(db, asOfSystemTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(697653)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697654)
	}
	__antithesis_instrumentation__.Notify(697649)
	defer func() { __antithesis_instrumentation__.Notify(697655); _ = txn.Rollback() }()
	__antithesis_instrumentation__.Notify(697650)
	firstQuery := txn.QueryRow(`
(SELECT no_w_id, no_d_id, no_o_id FROM new_order)
EXCEPT ALL
(SELECT o_w_id, o_d_id, o_id FROM "order" WHERE o_carrier_id IS NULL)`)
	if err := firstQuery.Scan(); !errors.Is(err, gosql.ErrNoRows) {
		__antithesis_instrumentation__.Notify(697656)
		return errors.Errorf("left EXCEPT right returned nonzero results.")
	} else {
		__antithesis_instrumentation__.Notify(697657)
	}
	__antithesis_instrumentation__.Notify(697651)
	secondQuery := txn.QueryRow(`
(SELECT o_w_id, o_d_id, o_id FROM "order" WHERE o_carrier_id IS NULL)
EXCEPT ALL
(SELECT no_w_id, no_d_id, no_o_id FROM new_order)`)
	if err := secondQuery.Scan(); !errors.Is(err, gosql.ErrNoRows) {
		__antithesis_instrumentation__.Notify(697658)
		return errors.Errorf("right EXCEPT left returned nonzero results.")
	} else {
		__antithesis_instrumentation__.Notify(697659)
	}
	__antithesis_instrumentation__.Notify(697652)
	return nil
}

func check3326(db *gosql.DB, asOfSystemTime string) (retErr error) {
	__antithesis_instrumentation__.Notify(697660)

	txn, err := beginAsOfSystemTime(db, asOfSystemTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(697665)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697666)
	}
	__antithesis_instrumentation__.Notify(697661)
	defer func() { __antithesis_instrumentation__.Notify(697667); _ = txn.Rollback() }()
	__antithesis_instrumentation__.Notify(697662)

	firstQuery := txn.QueryRow(`
(SELECT o_w_id, o_d_id, o_id, o_ol_cnt FROM "order"
  ORDER BY o_w_id, o_d_id, o_id DESC)
EXCEPT ALL
(SELECT ol_w_id, ol_d_id, ol_o_id, count(*) FROM order_line
  GROUP BY (ol_w_id, ol_d_id, ol_o_id)
  ORDER BY ol_w_id, ol_d_id, ol_o_id DESC)`)
	if err := firstQuery.Scan(); !errors.Is(err, gosql.ErrNoRows) {
		__antithesis_instrumentation__.Notify(697668)
		return errors.Errorf("left EXCEPT right returned nonzero results")
	} else {
		__antithesis_instrumentation__.Notify(697669)
	}
	__antithesis_instrumentation__.Notify(697663)
	secondQuery := txn.QueryRow(`
(SELECT ol_w_id, ol_d_id, ol_o_id, count(*) FROM order_line
  GROUP BY (ol_w_id, ol_d_id, ol_o_id) ORDER BY ol_w_id, ol_d_id, ol_o_id DESC)
EXCEPT ALL
(SELECT o_w_id, o_d_id, o_id, o_ol_cnt FROM "order"
  ORDER BY o_w_id, o_d_id, o_id DESC)`)
	if err := secondQuery.Scan(); !errors.Is(err, gosql.ErrNoRows) {
		__antithesis_instrumentation__.Notify(697670)
		return errors.Errorf("right EXCEPT left returned nonzero results")
	} else {
		__antithesis_instrumentation__.Notify(697671)
	}
	__antithesis_instrumentation__.Notify(697664)
	return nil
}

func check3327(db *gosql.DB, asOfSystemTime string) error {
	__antithesis_instrumentation__.Notify(697672)

	txn, err := beginAsOfSystemTime(db, asOfSystemTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(697677)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697678)
	}
	__antithesis_instrumentation__.Notify(697673)
	defer func() { __antithesis_instrumentation__.Notify(697679); _ = txn.Rollback() }()
	__antithesis_instrumentation__.Notify(697674)
	row := txn.QueryRow(`
SELECT count(*) FROM
  (SELECT o_w_id, o_d_id, o_id FROM "order" WHERE o_carrier_id IS NULL)
FULL OUTER JOIN
  (SELECT ol_w_id, ol_d_id, ol_o_id FROM order_line WHERE ol_delivery_d IS NULL)
ON (ol_w_id = o_w_id AND ol_d_id = o_d_id AND ol_o_id = o_id)
WHERE ol_o_id IS NULL OR o_id IS NULL
`)

	var i int
	if err := row.Scan(&i); err != nil {
		__antithesis_instrumentation__.Notify(697680)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697681)
	}
	__antithesis_instrumentation__.Notify(697675)

	if i != 0 {
		__antithesis_instrumentation__.Notify(697682)
		return errors.Errorf("%d rows returned, expected zero", i)
	} else {
		__antithesis_instrumentation__.Notify(697683)
	}
	__antithesis_instrumentation__.Notify(697676)

	return nil
}

func check3328(db *gosql.DB, asOfSystemTime string) error {
	__antithesis_instrumentation__.Notify(697684)

	txn, err := beginAsOfSystemTime(db, asOfSystemTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(697689)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697690)
	}
	__antithesis_instrumentation__.Notify(697685)
	defer func() { __antithesis_instrumentation__.Notify(697691); _ = txn.Rollback() }()
	__antithesis_instrumentation__.Notify(697686)
	row := txn.QueryRow(`
SELECT count(*) FROM
  (SELECT w_id, w_ytd, sum FROM warehouse
  JOIN
  (SELECT h_w_id, sum(h_amount) FROM history GROUP BY h_w_id)
  ON w_id = h_w_id
  WHERE w_ytd != sum
  )
`)

	var i int
	if err := row.Scan(&i); err != nil {
		__antithesis_instrumentation__.Notify(697692)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697693)
	}
	__antithesis_instrumentation__.Notify(697687)

	if i != 0 {
		__antithesis_instrumentation__.Notify(697694)
		return errors.Errorf("%d rows returned, expected zero", i)
	} else {
		__antithesis_instrumentation__.Notify(697695)
	}
	__antithesis_instrumentation__.Notify(697688)

	return nil
}

func check3329(db *gosql.DB, asOfSystemTime string) error {
	__antithesis_instrumentation__.Notify(697696)

	txn, err := beginAsOfSystemTime(db, asOfSystemTime)
	if err != nil {
		__antithesis_instrumentation__.Notify(697701)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697702)
	}
	__antithesis_instrumentation__.Notify(697697)
	defer func() { __antithesis_instrumentation__.Notify(697703); _ = txn.Rollback() }()
	__antithesis_instrumentation__.Notify(697698)
	row := txn.QueryRow(`
SELECT count(*) FROM
  (SELECT d_id, d_ytd, sum FROM district
  JOIN
  (SELECT h_w_id, h_d_id, sum(h_amount) FROM history GROUP BY (h_w_id, h_d_id))
  ON d_id = h_d_id AND d_w_id = h_w_id
  WHERE d_ytd != sum
  )
`)

	var i int
	if err := row.Scan(&i); err != nil {
		__antithesis_instrumentation__.Notify(697704)
		return err
	} else {
		__antithesis_instrumentation__.Notify(697705)
	}
	__antithesis_instrumentation__.Notify(697699)

	if i != 0 {
		__antithesis_instrumentation__.Notify(697706)
		return errors.Errorf("%d rows returned, expected zero", i)
	} else {
		__antithesis_instrumentation__.Notify(697707)
	}
	__antithesis_instrumentation__.Notify(697700)

	return nil
}

func beginAsOfSystemTime(db *gosql.DB, asOfSystemTime string) (txn *gosql.Tx, err error) {
	__antithesis_instrumentation__.Notify(697708)
	txn, err = db.Begin()
	if err != nil {
		__antithesis_instrumentation__.Notify(697711)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(697712)
	}
	__antithesis_instrumentation__.Notify(697709)
	if asOfSystemTime != "" {
		__antithesis_instrumentation__.Notify(697713)
		_, err = txn.Exec("SET TRANSACTION AS OF SYSTEM TIME " + asOfSystemTime)
		if err != nil {
			__antithesis_instrumentation__.Notify(697714)
			_ = txn.Rollback()
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(697715)
		}
	} else {
		__antithesis_instrumentation__.Notify(697716)
	}
	__antithesis_instrumentation__.Notify(697710)
	return txn, nil
}

func selectTimestamp(txn *gosql.Tx) (ts string, err error) {
	__antithesis_instrumentation__.Notify(697717)
	err = txn.QueryRow("SELECT cluster_logical_timestamp()::string").Scan(&ts)
	return ts, err
}
