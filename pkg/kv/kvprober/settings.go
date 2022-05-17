package kvprober

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/settings"
	"github.com/cockroachdb/errors"
)

var bypassAdmissionControl = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.prober.bypass_admission_control.enabled",
	"set to bypass admission control queue for kvprober requests; "+
		"note that dedicated clusters should have this set as users own capacity planning "+
		"but serverless clusters should not have this set as SREs own capacity planning",
	true,
)

var readEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.prober.read.enabled",
	"whether the KV read prober is enabled",
	false)

var readInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.prober.read.interval",
	"how often each node sends a read probe to the KV layer on average (jitter is added); "+
		"note that a very slow read can block kvprober from sending additional probes; "+
		"kv.prober.read.timeout controls the max time kvprober can be blocked",
	1*time.Minute, func(duration time.Duration) error {
		__antithesis_instrumentation__.Notify(94036)
		if duration <= 0 {
			__antithesis_instrumentation__.Notify(94038)
			return errors.New("param must be >0")
		} else {
			__antithesis_instrumentation__.Notify(94039)
		}
		__antithesis_instrumentation__.Notify(94037)
		return nil
	})

var readTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.prober.read.timeout",

	"if this much time elapses without success, a KV read probe will be treated as an error; "+
		"note that a very slow read can block kvprober from sending additional probes"+
		"this setting controls the max time kvprober can be blocked",
	2*time.Second, func(duration time.Duration) error {
		__antithesis_instrumentation__.Notify(94040)
		if duration <= 0 {
			__antithesis_instrumentation__.Notify(94042)
			return errors.New("param must be >0")
		} else {
			__antithesis_instrumentation__.Notify(94043)
		}
		__antithesis_instrumentation__.Notify(94041)
		return nil
	})

var writeEnabled = settings.RegisterBoolSetting(
	settings.TenantWritable,
	"kv.prober.write.enabled",
	"whether the KV write prober is enabled",
	false)

var writeInterval = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.prober.write.interval",
	"how often each node sends a write probe to the KV layer on average (jitter is added); "+
		"note that a very slow read can block kvprober from sending additional probes; "+
		"kv.prober.write.timeout controls the max time kvprober can be blocked",
	10*time.Second, func(duration time.Duration) error {
		__antithesis_instrumentation__.Notify(94044)
		if duration <= 0 {
			__antithesis_instrumentation__.Notify(94046)
			return errors.New("param must be >0")
		} else {
			__antithesis_instrumentation__.Notify(94047)
		}
		__antithesis_instrumentation__.Notify(94045)
		return nil
	})

var writeTimeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.prober.write.timeout",

	"if this much time elapses without success, a KV write probe will be treated as an error; "+
		"note that a very slow read can block kvprober from sending additional probes"+
		"this setting controls the max time kvprober can be blocked",
	4*time.Second, func(duration time.Duration) error {
		__antithesis_instrumentation__.Notify(94048)
		if duration <= 0 {
			__antithesis_instrumentation__.Notify(94050)
			return errors.New("param must be >0")
		} else {
			__antithesis_instrumentation__.Notify(94051)
		}
		__antithesis_instrumentation__.Notify(94049)
		return nil
	})

var scanMeta2Timeout = settings.RegisterDurationSetting(
	settings.TenantWritable,
	"kv.prober.planner.scan_meta2.timeout",
	"timeout on scanning meta2 via db.Scan with max rows set to "+
		"kv.prober.planner.num_steps_to_plan_at_once",
	2*time.Second, func(duration time.Duration) error {
		__antithesis_instrumentation__.Notify(94052)
		if duration <= 0 {
			__antithesis_instrumentation__.Notify(94054)
			return errors.New("param must be >0")
		} else {
			__antithesis_instrumentation__.Notify(94055)
		}
		__antithesis_instrumentation__.Notify(94053)
		return nil
	})

var numStepsToPlanAtOnce = settings.RegisterIntSetting(
	settings.TenantWritable,
	"kv.prober.planner.num_steps_to_plan_at_once",
	"the number of Steps to plan at once, where a Step is a decision on "+
		"what range to probe; the order of the Steps is randomized within "+
		"each planning run, so setting this to a small number will lead to "+
		"close-to-lexical probing; already made plans are held in memory, so "+
		"large values are advised against",
	100, func(i int64) error {
		__antithesis_instrumentation__.Notify(94056)
		if i <= 0 {
			__antithesis_instrumentation__.Notify(94058)
			return errors.New("param must be >0")
		} else {
			__antithesis_instrumentation__.Notify(94059)
		}
		__antithesis_instrumentation__.Notify(94057)
		return nil
	})
