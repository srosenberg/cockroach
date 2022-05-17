package streamingccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "net/url"

type StreamAddress string

func (sa StreamAddress) URL() (*url.URL, error) {
	__antithesis_instrumentation__.Notify(24938)
	return url.Parse(string(sa))
}

type PartitionAddress string
