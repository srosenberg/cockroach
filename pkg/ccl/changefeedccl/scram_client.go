package changefeedccl

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"crypto/sha256"
	"crypto/sha512"

	"github.com/Shopify/sarama"
	"github.com/xdg-go/scram"
)

var (
	sha256ClientGenerator = func() sarama.SCRAMClient {
		__antithesis_instrumentation__.Notify(18042)
		return &scramClient{HashGeneratorFcn: sha256.New}
	}

	sha512ClientGenerator = func() sarama.SCRAMClient {
		__antithesis_instrumentation__.Notify(18043)
		return &scramClient{HashGeneratorFcn: sha512.New}
	}
)

type scramClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

var _ sarama.SCRAMClient = &scramClient{}

func (c *scramClient) Begin(userName, password, authzID string) error {
	__antithesis_instrumentation__.Notify(18044)
	var err error
	c.Client, err = c.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		__antithesis_instrumentation__.Notify(18046)
		return err
	} else {
		__antithesis_instrumentation__.Notify(18047)
	}
	__antithesis_instrumentation__.Notify(18045)
	c.ClientConversation = c.Client.NewConversation()
	return nil
}

func (c *scramClient) Step(challenge string) (string, error) {
	__antithesis_instrumentation__.Notify(18048)
	return c.ClientConversation.Step(challenge)
}

func (c *scramClient) Done() bool {
	__antithesis_instrumentation__.Notify(18049)
	return c.ClientConversation.Done()
}
