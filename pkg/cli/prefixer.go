package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"path/filepath"
	"strings"
)

type filePrefixerOption func(*filePrefixerOptions)

var defaultFilePrefixerOptions = filePrefixerOptions{
	Delimiters: []string{"/", ".", "-"},
	Template:   "${host}> ",
}

type filePrefixerOptions struct {
	Delimiters []string
	Template   string
}

func withTemplate(template string) filePrefixerOption {
	__antithesis_instrumentation__.Notify(33804)
	return func(o *filePrefixerOptions) {
		__antithesis_instrumentation__.Notify(33805)
		o.Template = template
	}
}

type filePrefixer struct {
	template   string
	delimiters []string
}

func newFilePrefixer(opts ...filePrefixerOption) filePrefixer {
	__antithesis_instrumentation__.Notify(33806)
	options := defaultFilePrefixerOptions
	for _, o := range opts {
		__antithesis_instrumentation__.Notify(33808)
		o(&options)
	}
	__antithesis_instrumentation__.Notify(33807)
	return filePrefixer{
		template:   options.Template,
		delimiters: options.Delimiters,
	}
}

func (f filePrefixer) PopulatePrefixes(logFiles []fileInfo) {
	__antithesis_instrumentation__.Notify(33809)

	tPaths := make([][]string, len(logFiles))
	common := map[string]int{}

	for i, lf := range logFiles {
		__antithesis_instrumentation__.Notify(33811)
		seen := map[string]struct{}{}
		tokens := strings.Split(filepath.Dir(lf.path), string(filepath.Separator))
		for _, t := range tokens {
			__antithesis_instrumentation__.Notify(33813)
			if _, ok := seen[t]; !ok {
				__antithesis_instrumentation__.Notify(33814)
				seen[t] = struct{}{}
				common[t]++
			} else {
				__antithesis_instrumentation__.Notify(33815)
			}
		}
		__antithesis_instrumentation__.Notify(33812)
		tPaths[i] = tokens
	}
	__antithesis_instrumentation__.Notify(33810)

	for i, tokens := range tPaths {
		__antithesis_instrumentation__.Notify(33816)
		var filteredTokens []string

		for _, t := range tokens {
			__antithesis_instrumentation__.Notify(33818)
			count := common[t]

			if count < len(logFiles) && func() bool {
				__antithesis_instrumentation__.Notify(33819)
				return !f.hasDelimiters(t) == true
			}() == true {
				__antithesis_instrumentation__.Notify(33820)
				filteredTokens = append(filteredTokens, t)
			} else {
				__antithesis_instrumentation__.Notify(33821)
			}
		}
		__antithesis_instrumentation__.Notify(33817)
		filteredTokens = append(filteredTokens, filepath.Base(logFiles[i].path))

		shortenedFP := filepath.Join(filteredTokens...)
		matches := logFiles[i].pattern.FindStringSubmatchIndex(shortenedFP)
		logFiles[i].prefix = logFiles[i].pattern.ExpandString(nil, f.template, shortenedFP, matches)
	}
}

func (f filePrefixer) hasDelimiters(token string) bool {
	__antithesis_instrumentation__.Notify(33822)
	for _, d := range f.delimiters {
		__antithesis_instrumentation__.Notify(33824)
		if strings.Contains(token, d) {
			__antithesis_instrumentation__.Notify(33825)
			return true
		} else {
			__antithesis_instrumentation__.Notify(33826)
		}
	}
	__antithesis_instrumentation__.Notify(33823)
	return false
}
