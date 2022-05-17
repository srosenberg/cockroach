package tests

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	gosql "database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/cluster"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/option"
	"github.com/cockroachdb/cockroach/pkg/cmd/roachtest/test"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/roachprod/install"
	"github.com/cockroachdb/cockroach/pkg/util/retry"
	"github.com/stretchr/testify/require"
)

func runEventLog(ctx context.Context, t test.Test, c cluster.Cluster) {
	__antithesis_instrumentation__.Notify(47641)
	type nodeEventInfo struct {
		NodeID roachpb.NodeID
	}

	c.Put(ctx, t.Cockroach(), "./cockroach")
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings())

	db := c.Conn(ctx, t.L(), 1)
	defer db.Close()
	err := WaitFor3XReplication(ctx, t, db)
	require.NoError(t, err)

	err = retry.ForDuration(10*time.Second, func() error {
		__antithesis_instrumentation__.Notify(47645)
		rows, err := db.Query(
			`SELECT "targetID", info FROM system.eventlog WHERE "eventType" = 'node_join'`,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(47650)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47651)
		}
		__antithesis_instrumentation__.Notify(47646)
		seenIds := make(map[int64]struct{})
		for rows.Next() {
			__antithesis_instrumentation__.Notify(47652)
			var targetID int64
			var infoStr gosql.NullString
			if err := rows.Scan(&targetID, &infoStr); err != nil {
				__antithesis_instrumentation__.Notify(47658)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47659)
			}
			__antithesis_instrumentation__.Notify(47653)

			if !infoStr.Valid {
				__antithesis_instrumentation__.Notify(47660)
				t.Fatalf("info not recorded for node join, target node %d", targetID)
			} else {
				__antithesis_instrumentation__.Notify(47661)
			}
			__antithesis_instrumentation__.Notify(47654)
			var info nodeEventInfo
			if err := json.Unmarshal([]byte(infoStr.String), &info); err != nil {
				__antithesis_instrumentation__.Notify(47662)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47663)
			}
			__antithesis_instrumentation__.Notify(47655)
			if a, e := int64(info.NodeID), targetID; a != e {
				__antithesis_instrumentation__.Notify(47664)
				t.Fatalf("Node join with targetID %d had descriptor for wrong node %d", e, a)
			} else {
				__antithesis_instrumentation__.Notify(47665)
			}
			__antithesis_instrumentation__.Notify(47656)

			if _, ok := seenIds[targetID]; ok {
				__antithesis_instrumentation__.Notify(47666)
				t.Fatalf("Node ID %d seen in two different node join messages", targetID)
			} else {
				__antithesis_instrumentation__.Notify(47667)
			}
			__antithesis_instrumentation__.Notify(47657)
			seenIds[targetID] = struct{}{}
		}
		__antithesis_instrumentation__.Notify(47647)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(47668)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47669)
		}
		__antithesis_instrumentation__.Notify(47648)
		if c.Spec().NodeCount != len(seenIds) {
			__antithesis_instrumentation__.Notify(47670)
			return fmt.Errorf("expected %d node join messages, found %d: %v",
				c.Spec().NodeCount, len(seenIds), seenIds)
		} else {
			__antithesis_instrumentation__.Notify(47671)
		}
		__antithesis_instrumentation__.Notify(47649)
		return nil
	})
	__antithesis_instrumentation__.Notify(47642)
	if err != nil {
		__antithesis_instrumentation__.Notify(47672)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(47673)
	}
	__antithesis_instrumentation__.Notify(47643)

	c.Stop(ctx, t.L(), option.DefaultStopOpts(), c.Node(3))
	c.Start(ctx, t.L(), option.DefaultStartOpts(), install.MakeClusterSettings(), c.Node(3))

	err = retry.ForDuration(10*time.Second, func() error {
		__antithesis_instrumentation__.Notify(47674)

		rows, err := db.Query(
			`SELECT "targetID", info FROM system.eventlog WHERE "eventType" = 'node_restart'`,
		)
		if err != nil {
			__antithesis_instrumentation__.Notify(47679)
			return err
		} else {
			__antithesis_instrumentation__.Notify(47680)
		}
		__antithesis_instrumentation__.Notify(47675)

		seenCount := 0
		for rows.Next() {
			__antithesis_instrumentation__.Notify(47681)
			var targetID int64
			var infoStr gosql.NullString
			if err := rows.Scan(&targetID, &infoStr); err != nil {
				__antithesis_instrumentation__.Notify(47686)
				return err
			} else {
				__antithesis_instrumentation__.Notify(47687)
			}
			__antithesis_instrumentation__.Notify(47682)

			if !infoStr.Valid {
				__antithesis_instrumentation__.Notify(47688)
				t.Fatalf("info not recorded for node join, target node %d", targetID)
			} else {
				__antithesis_instrumentation__.Notify(47689)
			}
			__antithesis_instrumentation__.Notify(47683)
			var info nodeEventInfo
			if err := json.Unmarshal([]byte(infoStr.String), &info); err != nil {
				__antithesis_instrumentation__.Notify(47690)
				t.Fatal(err)
			} else {
				__antithesis_instrumentation__.Notify(47691)
			}
			__antithesis_instrumentation__.Notify(47684)
			if a, e := int64(info.NodeID), targetID; a != e {
				__antithesis_instrumentation__.Notify(47692)
				t.Fatalf("node join with targetID %d had descriptor for wrong node %d", e, a)
			} else {
				__antithesis_instrumentation__.Notify(47693)
			}
			__antithesis_instrumentation__.Notify(47685)

			seenCount++
		}
		__antithesis_instrumentation__.Notify(47676)
		if err := rows.Err(); err != nil {
			__antithesis_instrumentation__.Notify(47694)
			t.Fatal(err)
		} else {
			__antithesis_instrumentation__.Notify(47695)
		}
		__antithesis_instrumentation__.Notify(47677)
		if seenCount != 1 {
			__antithesis_instrumentation__.Notify(47696)
			return fmt.Errorf("expected one node restart event, found %d", seenCount)
		} else {
			__antithesis_instrumentation__.Notify(47697)
		}
		__antithesis_instrumentation__.Notify(47678)
		return nil
	})
	__antithesis_instrumentation__.Notify(47644)
	if err != nil {
		__antithesis_instrumentation__.Notify(47698)
		t.Fatal(err)
	} else {
		__antithesis_instrumentation__.Notify(47699)
	}
}
