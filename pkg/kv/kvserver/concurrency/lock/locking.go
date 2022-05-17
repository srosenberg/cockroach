// Package lock provides type definitions for locking-related concepts used by
// concurrency control in the key-value layer.
package lock

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import fmt "fmt"

const MaxDurability = Unreplicated

func init() {
	for v := range Durability_name {
		if d := Durability(v); d > MaxDurability {
			panic(fmt.Sprintf("Durability (%s) with value larger than MaxDurability", d))
		}
	}
}

func (Strength) SafeValue() { __antithesis_instrumentation__.Notify(99393) }

func (Durability) SafeValue() { __antithesis_instrumentation__.Notify(99394) }

func (WaitPolicy) SafeValue() { __antithesis_instrumentation__.Notify(99395) }
