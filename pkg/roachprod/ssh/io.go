package ssh

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "io"

type ProgressWriter struct {
	Writer   io.Writer
	Done     int64
	Total    int64
	Progress func(float64)
}

func (p *ProgressWriter) Write(b []byte) (int, error) {
	__antithesis_instrumentation__.Notify(182528)
	n, err := p.Writer.Write(b)
	if err == nil {
		__antithesis_instrumentation__.Notify(182530)
		p.Done += int64(n)
		p.Progress(float64(p.Done) / float64(p.Total))
	} else {
		__antithesis_instrumentation__.Notify(182531)
	}
	__antithesis_instrumentation__.Notify(182529)
	return n, err
}

var InsecureIgnoreHostKey bool
