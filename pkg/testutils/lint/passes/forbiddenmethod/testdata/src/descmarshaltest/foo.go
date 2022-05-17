package descmarshaltest

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"

func F() {
	__antithesis_instrumentation__.Notify(644762)
	var d descpb.Descriptor
	d.GetDatabase()

	d.GetDatabase()

	d.GetDatabase()

	d.GetTable()

	d.GetTable()

	d.GetTable()

	d.GetType()

	d.GetType()

	d.GetType()

	d.GetSchema()

	d.GetSchema()

	d.GetSchema()

	if t := d.GetTable(); t != nil {
		__antithesis_instrumentation__.Notify(644768)
		panic("foo")
	} else {
		__antithesis_instrumentation__.Notify(644769)
	}
	__antithesis_instrumentation__.Notify(644763)

	if t := d.
		GetTable(); t != nil {
		__antithesis_instrumentation__.Notify(644770)
		panic("foo")
	} else {
		__antithesis_instrumentation__.Notify(644771)
	}
	__antithesis_instrumentation__.Notify(644764)

	if t :=

		d.GetTable(); t != nil {
		__antithesis_instrumentation__.Notify(644772)
		panic("foo")
	} else {
		__antithesis_instrumentation__.Notify(644773)
	}
	__antithesis_instrumentation__.Notify(644765)

	if t := d.GetTable(); t !=

		nil {
		__antithesis_instrumentation__.Notify(644774)
		panic("foo")
	} else {
		__antithesis_instrumentation__.Notify(644775)
	}
	__antithesis_instrumentation__.Notify(644766)

	if t := d.GetTable(); t != nil {
		__antithesis_instrumentation__.Notify(644776)
		panic("foo")
	} else {
		__antithesis_instrumentation__.Notify(644777)
	}
	__antithesis_instrumentation__.Notify(644767)

	if t := d.GetTable(); t != nil {
		__antithesis_instrumentation__.Notify(644778)
		panic("foo")
	} else {
		__antithesis_instrumentation__.Notify(644779)
	}
}
