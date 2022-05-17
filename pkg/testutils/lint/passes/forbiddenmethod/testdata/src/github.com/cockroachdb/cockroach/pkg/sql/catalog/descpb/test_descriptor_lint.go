package descpb

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

type Descriptor struct{}

type TableDescriptor struct{}

type DatabaseDescriptor struct{}

type TypeDescriptor struct{}

type SchemaDescriptor struct{}

func (m *Descriptor) GetTable() *TableDescriptor {
	__antithesis_instrumentation__.Notify(644780)
	return nil
}

func (m *Descriptor) GetDatabase() *DatabaseDescriptor {
	__antithesis_instrumentation__.Notify(644781)
	return nil
}

func (m *Descriptor) GetType() *TypeDescriptor {
	__antithesis_instrumentation__.Notify(644782)
	return nil
}

func (m *Descriptor) GetSchema() *SchemaDescriptor {
	__antithesis_instrumentation__.Notify(644783)
	return nil
}
