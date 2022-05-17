package tabledesc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"time"

	"github.com/cockroachdb/cockroach/pkg/sql/catalog/catpb"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgcode"
	"github.com/cockroachdb/cockroach/pkg/sql/pgwire/pgerror"
	"github.com/robfig/cron/v3"
)

func ValidateRowLevelTTL(ttl *catpb.RowLevelTTL) error {
	__antithesis_instrumentation__.Notify(270989)
	if ttl == nil {
		__antithesis_instrumentation__.Notify(270998)
		return nil
	} else {
		__antithesis_instrumentation__.Notify(270999)
	}
	__antithesis_instrumentation__.Notify(270990)
	if ttl.DurationExpr == "" {
		__antithesis_instrumentation__.Notify(271000)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			`"ttl_expire_after" must be set`,
		)
	} else {
		__antithesis_instrumentation__.Notify(271001)
	}
	__antithesis_instrumentation__.Notify(270991)
	if ttl.DeleteBatchSize != 0 {
		__antithesis_instrumentation__.Notify(271002)
		if err := ValidateTTLBatchSize("ttl_delete_batch_size", ttl.DeleteBatchSize); err != nil {
			__antithesis_instrumentation__.Notify(271003)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271004)
		}
	} else {
		__antithesis_instrumentation__.Notify(271005)
	}
	__antithesis_instrumentation__.Notify(270992)
	if ttl.SelectBatchSize != 0 {
		__antithesis_instrumentation__.Notify(271006)
		if err := ValidateTTLBatchSize("ttl_select_batch_size", ttl.SelectBatchSize); err != nil {
			__antithesis_instrumentation__.Notify(271007)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271008)
		}
	} else {
		__antithesis_instrumentation__.Notify(271009)
	}
	__antithesis_instrumentation__.Notify(270993)
	if ttl.DeletionCron != "" {
		__antithesis_instrumentation__.Notify(271010)
		if err := ValidateTTLCronExpr("ttl_job_cron", ttl.DeletionCron); err != nil {
			__antithesis_instrumentation__.Notify(271011)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271012)
		}
	} else {
		__antithesis_instrumentation__.Notify(271013)
	}
	__antithesis_instrumentation__.Notify(270994)
	if ttl.RangeConcurrency != 0 {
		__antithesis_instrumentation__.Notify(271014)
		if err := ValidateTTLRangeConcurrency("ttl_range_concurrency", ttl.RangeConcurrency); err != nil {
			__antithesis_instrumentation__.Notify(271015)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271016)
		}
	} else {
		__antithesis_instrumentation__.Notify(271017)
	}
	__antithesis_instrumentation__.Notify(270995)
	if ttl.DeleteRateLimit != 0 {
		__antithesis_instrumentation__.Notify(271018)
		if err := ValidateTTLRateLimit("ttl_delete_rate_limit", ttl.DeleteRateLimit); err != nil {
			__antithesis_instrumentation__.Notify(271019)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271020)
		}
	} else {
		__antithesis_instrumentation__.Notify(271021)
	}
	__antithesis_instrumentation__.Notify(270996)
	if ttl.RowStatsPollInterval != 0 {
		__antithesis_instrumentation__.Notify(271022)
		if err := ValidateTTLRowStatsPollInterval("ttl_row_stats_poll_interval", ttl.RowStatsPollInterval); err != nil {
			__antithesis_instrumentation__.Notify(271023)
			return err
		} else {
			__antithesis_instrumentation__.Notify(271024)
		}
	} else {
		__antithesis_instrumentation__.Notify(271025)
	}
	__antithesis_instrumentation__.Notify(270997)
	return nil
}

func ValidateTTLBatchSize(key string, val int64) error {
	__antithesis_instrumentation__.Notify(271026)
	if val <= 0 {
		__antithesis_instrumentation__.Notify(271028)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			`"%s" must be at least 1`,
			key,
		)
	} else {
		__antithesis_instrumentation__.Notify(271029)
	}
	__antithesis_instrumentation__.Notify(271027)
	return nil
}

func ValidateTTLRangeConcurrency(key string, val int64) error {
	__antithesis_instrumentation__.Notify(271030)
	if val <= 0 {
		__antithesis_instrumentation__.Notify(271032)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			`"%s" must be at least 1`,
			key,
		)
	} else {
		__antithesis_instrumentation__.Notify(271033)
	}
	__antithesis_instrumentation__.Notify(271031)
	return nil
}

func ValidateTTLCronExpr(key string, str string) error {
	__antithesis_instrumentation__.Notify(271034)
	if _, err := cron.ParseStandard(str); err != nil {
		__antithesis_instrumentation__.Notify(271036)
		return pgerror.Wrapf(
			err,
			pgcode.InvalidParameterValue,
			`invalid cron expression for "%s"`,
			key,
		)
	} else {
		__antithesis_instrumentation__.Notify(271037)
	}
	__antithesis_instrumentation__.Notify(271035)
	return nil
}

func ValidateTTLRowStatsPollInterval(key string, val time.Duration) error {
	__antithesis_instrumentation__.Notify(271038)
	if val <= 0 {
		__antithesis_instrumentation__.Notify(271040)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			`"%s" must be at least 1`,
			key,
		)
	} else {
		__antithesis_instrumentation__.Notify(271041)
	}
	__antithesis_instrumentation__.Notify(271039)
	return nil
}

func ValidateTTLRateLimit(key string, val int64) error {
	__antithesis_instrumentation__.Notify(271042)
	if val <= 0 {
		__antithesis_instrumentation__.Notify(271044)
		return pgerror.Newf(
			pgcode.InvalidParameterValue,
			`"%s" must be at least 1`,
			key,
		)
	} else {
		__antithesis_instrumentation__.Notify(271045)
	}
	__antithesis_instrumentation__.Notify(271043)
	return nil
}
