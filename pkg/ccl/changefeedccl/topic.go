package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"strings"

	"github.com/cockroachdb/cockroach/pkg/jobs/jobspb"
	"github.com/cockroachdb/cockroach/pkg/sql/catalog/descpb"
	"github.com/cockroachdb/errors"
)

type TopicDescriptor interface {
	GetNameComponents() []string

	GetTopicIdentifier() TopicIdentifier

	GetVersion() descpb.DescriptorVersion

	GetTargetSpecification() jobspb.ChangefeedTargetSpecification
}

type TopicIdentifier struct {
	TableID  descpb.ID
	FamilyID descpb.FamilyID
}

type TopicNamer struct {
	join       byte
	prefix     string
	singleName string
	sanitize   func(string) string

	DisplayNames map[jobspb.ChangefeedTargetSpecification]string

	FullNames map[TopicIdentifier]string

	sliceCache []string
}

type TopicNameOption interface {
	set(*TopicNamer)
}

type optJoinByte byte

func (o optJoinByte) set(tn *TopicNamer) {
	__antithesis_instrumentation__.Notify(19016)
	tn.join = byte(o)
}

func WithJoinByte(b byte) TopicNameOption {
	__antithesis_instrumentation__.Notify(19017)
	return optJoinByte(b)
}

type optPrefix string

func (o optPrefix) set(tn *TopicNamer) {
	__antithesis_instrumentation__.Notify(19018)
	tn.prefix = string(o)
}

func WithPrefix(s string) TopicNameOption {
	__antithesis_instrumentation__.Notify(19019)
	return optPrefix(s)
}

type optSingleName string

func (o optSingleName) set(tn *TopicNamer) {
	__antithesis_instrumentation__.Notify(19020)
	tn.singleName = string(o)
}

func WithSingleName(s string) TopicNameOption {
	__antithesis_instrumentation__.Notify(19021)
	return optSingleName(s)
}

type optSanitize func(string) string

func (o optSanitize) set(tn *TopicNamer) {
	__antithesis_instrumentation__.Notify(19022)
	tn.sanitize = o
}

func WithSanitizeFn(fn func(string) string) TopicNameOption {
	__antithesis_instrumentation__.Notify(19023)
	return optSanitize(fn)
}

func MakeTopicNamer(
	specs []jobspb.ChangefeedTargetSpecification, opts ...TopicNameOption,
) (*TopicNamer, error) {
	__antithesis_instrumentation__.Notify(19024)
	tn := &TopicNamer{
		join:         '.',
		DisplayNames: make(map[jobspb.ChangefeedTargetSpecification]string, len(specs)),
		FullNames:    make(map[TopicIdentifier]string),
	}
	for _, opt := range opts {
		__antithesis_instrumentation__.Notify(19027)
		opt.set(tn)
	}
	__antithesis_instrumentation__.Notify(19025)
	for _, s := range specs {
		__antithesis_instrumentation__.Notify(19028)
		name, err := tn.makeDisplayName(s)
		if err != nil {
			__antithesis_instrumentation__.Notify(19030)
			return nil, err
		} else {
			__antithesis_instrumentation__.Notify(19031)
		}
		__antithesis_instrumentation__.Notify(19029)
		tn.DisplayNames[s] = name
	}
	__antithesis_instrumentation__.Notify(19026)

	return tn, nil

}

const familyPlaceholder = "{family}"

func (tn *TopicNamer) Name(td TopicDescriptor) (string, error) {
	__antithesis_instrumentation__.Notify(19032)
	if name, ok := tn.FullNames[td.GetTopicIdentifier()]; ok {
		__antithesis_instrumentation__.Notify(19034)
		return name, nil
	} else {
		__antithesis_instrumentation__.Notify(19035)
	}
	__antithesis_instrumentation__.Notify(19033)
	name, err := tn.makeName(td.GetTargetSpecification(), td)
	tn.FullNames[td.GetTopicIdentifier()] = name
	return name, err
}

func (tn *TopicNamer) DisplayNamesSlice() []string {
	__antithesis_instrumentation__.Notify(19036)
	if len(tn.sliceCache) > 0 {
		__antithesis_instrumentation__.Notify(19039)
		return tn.sliceCache
	} else {
		__antithesis_instrumentation__.Notify(19040)
	}
	__antithesis_instrumentation__.Notify(19037)
	for _, n := range tn.DisplayNames {
		__antithesis_instrumentation__.Notify(19041)
		tn.sliceCache = append(tn.sliceCache, n)
		if tn.singleName != "" {
			__antithesis_instrumentation__.Notify(19042)
			return tn.sliceCache
		} else {
			__antithesis_instrumentation__.Notify(19043)
		}
	}
	__antithesis_instrumentation__.Notify(19038)
	return tn.sliceCache
}

func (tn *TopicNamer) Each(fn func(string) error) error {
	__antithesis_instrumentation__.Notify(19044)
	for _, name := range tn.DisplayNames {
		__antithesis_instrumentation__.Notify(19046)
		err := fn(name)
		if tn.singleName != "" || func() bool {
			__antithesis_instrumentation__.Notify(19047)
			return err != nil == true
		}() == true {
			__antithesis_instrumentation__.Notify(19048)
			return err
		} else {
			__antithesis_instrumentation__.Notify(19049)
		}
	}
	__antithesis_instrumentation__.Notify(19045)
	return nil
}

func (tn *TopicNamer) makeName(
	s jobspb.ChangefeedTargetSpecification, td TopicDescriptor,
) (string, error) {
	__antithesis_instrumentation__.Notify(19050)
	switch s.Type {
	case jobspb.ChangefeedTargetSpecification_PRIMARY_FAMILY_ONLY:
		__antithesis_instrumentation__.Notify(19051)
		return tn.nameFromComponents(s.StatementTimeName), nil
	case jobspb.ChangefeedTargetSpecification_COLUMN_FAMILY:
		__antithesis_instrumentation__.Notify(19052)
		return tn.nameFromComponents(s.StatementTimeName, s.FamilyName), nil
	case jobspb.ChangefeedTargetSpecification_EACH_FAMILY:
		__antithesis_instrumentation__.Notify(19053)
		if td == nil {
			__antithesis_instrumentation__.Notify(19056)
			return tn.nameFromComponents(s.StatementTimeName, familyPlaceholder), nil
		} else {
			__antithesis_instrumentation__.Notify(19057)
		}
		__antithesis_instrumentation__.Notify(19054)
		return tn.nameFromComponents(td.GetNameComponents()...), nil
	default:
		__antithesis_instrumentation__.Notify(19055)
		return "", errors.AssertionFailedf("unrecognized type %s", s.Type)
	}
}

func (tn *TopicNamer) makeDisplayName(s jobspb.ChangefeedTargetSpecification) (string, error) {
	__antithesis_instrumentation__.Notify(19058)
	return tn.makeName(s, nil)
}

func (tn *TopicNamer) nameFromComponents(components ...string) string {
	__antithesis_instrumentation__.Notify(19059)

	var b strings.Builder
	b.WriteString(tn.prefix)
	if tn.singleName != "" {
		__antithesis_instrumentation__.Notify(19062)
		b.WriteString(tn.singleName)
	} else {
		__antithesis_instrumentation__.Notify(19063)
		for i, c := range components {
			__antithesis_instrumentation__.Notify(19064)
			if i > 0 {
				__antithesis_instrumentation__.Notify(19066)
				b.WriteByte(tn.join)
			} else {
				__antithesis_instrumentation__.Notify(19067)
			}
			__antithesis_instrumentation__.Notify(19065)
			b.WriteString(c)
		}
	}
	__antithesis_instrumentation__.Notify(19060)
	str := b.String()

	if tn.sanitize != nil {
		__antithesis_instrumentation__.Notify(19068)
		return tn.sanitize(str)
	} else {
		__antithesis_instrumentation__.Notify(19069)
	}
	__antithesis_instrumentation__.Notify(19061)

	return str
}
