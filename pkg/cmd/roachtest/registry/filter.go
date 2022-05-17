package registry

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"regexp"
	"strings"
)

type TestFilter struct {
	Name *regexp.Regexp
	Tag  *regexp.Regexp

	RawTag []string
}

func NewTestFilter(filter []string) *TestFilter {
	__antithesis_instrumentation__.Notify(44366)
	var name []string
	var tag []string
	var rawTag []string
	for _, v := range filter {
		__antithesis_instrumentation__.Notify(44370)
		if strings.HasPrefix(v, "tag:") {
			__antithesis_instrumentation__.Notify(44371)
			tag = append(tag, strings.TrimPrefix(v, "tag:"))
			rawTag = append(rawTag, v)
		} else {
			__antithesis_instrumentation__.Notify(44372)
			name = append(name, v)
		}
	}
	__antithesis_instrumentation__.Notify(44367)

	if len(tag) == 0 {
		__antithesis_instrumentation__.Notify(44373)
		tag = []string{DefaultTag}
		rawTag = []string{"tag:" + DefaultTag}
	} else {
		__antithesis_instrumentation__.Notify(44374)
	}
	__antithesis_instrumentation__.Notify(44368)

	makeRE := func(strs []string) *regexp.Regexp {
		__antithesis_instrumentation__.Notify(44375)
		switch len(strs) {
		case 0:
			__antithesis_instrumentation__.Notify(44376)
			return regexp.MustCompile(`.`)
		case 1:
			__antithesis_instrumentation__.Notify(44377)
			return regexp.MustCompile(strs[0])
		default:
			__antithesis_instrumentation__.Notify(44378)
			for i := range strs {
				__antithesis_instrumentation__.Notify(44380)
				strs[i] = "(" + strs[i] + ")"
			}
			__antithesis_instrumentation__.Notify(44379)
			return regexp.MustCompile(strings.Join(strs, "|"))
		}
	}
	__antithesis_instrumentation__.Notify(44369)

	return &TestFilter{
		Name:   makeRE(name),
		Tag:    makeRE(tag),
		RawTag: rawTag,
	}
}
