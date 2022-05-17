package privilege

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "strconv"

func _() {
	__antithesis_instrumentation__.Notify(563249)

	var x [1]struct{}
	_ = x[ALL-1]
	_ = x[CREATE-2]
	_ = x[DROP-3]
	_ = x[GRANT-4]
	_ = x[SELECT-5]
	_ = x[INSERT-6]
	_ = x[DELETE-7]
	_ = x[UPDATE-8]
	_ = x[USAGE-9]
	_ = x[ZONECONFIG-10]
	_ = x[CONNECT-11]
	_ = x[RULE-12]
}

const _Kind_name = "ALLCREATEDROPGRANTSELECTINSERTDELETEUPDATEUSAGEZONECONFIGCONNECTRULE"

var _Kind_index = [...]uint8{0, 3, 9, 13, 18, 24, 30, 36, 42, 47, 57, 64, 68}

func (i Kind) String() string {
	__antithesis_instrumentation__.Notify(563250)
	i -= 1
	if i >= Kind(len(_Kind_index)-1) {
		__antithesis_instrumentation__.Notify(563252)
		return "Kind(" + strconv.FormatInt(int64(i+1), 10) + ")"
	} else {
		__antithesis_instrumentation__.Notify(563253)
	}
	__antithesis_instrumentation__.Notify(563251)
	return _Kind_name[_Kind_index[i]:_Kind_index[i+1]]
}
