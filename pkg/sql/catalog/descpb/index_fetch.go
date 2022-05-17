package descpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

const IndexFetchSpecVersionInitial = 1

func (s *IndexFetchSpec) KeyColumns() []IndexFetchSpec_KeyColumn {
	__antithesis_instrumentation__.Notify(251638)
	return s.KeyAndSuffixColumns[:len(s.KeyAndSuffixColumns)-int(s.NumKeySuffixColumns)]
}

func (s *IndexFetchSpec) KeyFullColumns() []IndexFetchSpec_KeyColumn {
	__antithesis_instrumentation__.Notify(251639)
	if s.IsUniqueIndex {
		__antithesis_instrumentation__.Notify(251641)

		return s.KeyColumns()
	} else {
		__antithesis_instrumentation__.Notify(251642)
	}
	__antithesis_instrumentation__.Notify(251640)
	return s.KeyAndSuffixColumns
}

func (s *IndexFetchSpec) KeySuffixColumns() []IndexFetchSpec_KeyColumn {
	__antithesis_instrumentation__.Notify(251643)
	return s.KeyAndSuffixColumns[len(s.KeyAndSuffixColumns)-int(s.NumKeySuffixColumns):]
}
