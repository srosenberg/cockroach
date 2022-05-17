package catpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

func (m *RowLevelTTL) DeletionCronOrDefault() string {
	__antithesis_instrumentation__.Notify(250658)
	if override := m.DeletionCron; override != "" {
		__antithesis_instrumentation__.Notify(250660)
		return override
	} else {
		__antithesis_instrumentation__.Notify(250661)
	}
	__antithesis_instrumentation__.Notify(250659)
	return "@hourly"
}
