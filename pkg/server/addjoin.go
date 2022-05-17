package server

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/cockroachdb/cockroach/pkg/kv"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/serverpb"
	"github.com/cockroachdb/cockroach/pkg/sql"
	"github.com/cockroachdb/cockroach/pkg/sql/sem/tree"
	"github.com/cockroachdb/cockroach/pkg/util/uuid"
	"github.com/cockroachdb/errors"
)

var ErrInvalidAddJoinToken = errors.New("invalid add/join token received")

var ErrAddJoinTokenConsumed = errors.New("add/join token consumed but then another error occurred")

func (s *adminServer) RequestCA(
	ctx context.Context, req *serverpb.CARequest,
) (*serverpb.CAResponse, error) {
	__antithesis_instrumentation__.Notify(187491)
	settings := s.server.ClusterSettings()
	if settings == nil {
		__antithesis_instrumentation__.Notify(187495)
		return nil, errors.AssertionFailedf("could not look up cluster settings")
	} else {
		__antithesis_instrumentation__.Notify(187496)
	}
	__antithesis_instrumentation__.Notify(187492)
	if !sql.FeatureTLSAutoJoinEnabled.Get(&settings.SV) {
		__antithesis_instrumentation__.Notify(187497)
		return nil, errors.New("feature disabled by administrator")
	} else {
		__antithesis_instrumentation__.Notify(187498)
	}
	__antithesis_instrumentation__.Notify(187493)

	cm, err := s.server.rpcContext.GetCertificateManager()
	if err != nil {
		__antithesis_instrumentation__.Notify(187499)
		return nil, errors.Wrap(err, "failed to get certificate manager")
	} else {
		__antithesis_instrumentation__.Notify(187500)
	}
	__antithesis_instrumentation__.Notify(187494)
	caCert := cm.CACert().FileContents

	res := &serverpb.CAResponse{
		CaCert: caCert,
	}
	return res, nil
}

func (s *adminServer) consumeJoinToken(ctx context.Context, clientToken security.JoinToken) error {
	__antithesis_instrumentation__.Notify(187501)
	return s.server.db.Txn(ctx, func(ctx context.Context, txn *kv.Txn) error {
		__antithesis_instrumentation__.Notify(187502)
		row, err := s.ie.QueryRow(
			ctx, "select-consume-join-token", txn,
			"SELECT id, secret FROM system.join_tokens WHERE id = $1 AND now() < expiration",
			clientToken.TokenID.String())
		if err != nil {
			__antithesis_instrumentation__.Notify(187506)
			return err
		} else {
			__antithesis_instrumentation__.Notify(187507)
			if len(row) != 2 {
				__antithesis_instrumentation__.Notify(187508)
				return ErrInvalidAddJoinToken
			} else {
				__antithesis_instrumentation__.Notify(187509)
			}
		}
		__antithesis_instrumentation__.Notify(187503)

		secret := *row[1].(*tree.DBytes)
		if !bytes.Equal([]byte(secret), clientToken.SharedSecret) {
			__antithesis_instrumentation__.Notify(187510)
			return errors.New("invalid shared secret")
		} else {
			__antithesis_instrumentation__.Notify(187511)
		}
		__antithesis_instrumentation__.Notify(187504)

		i, err := s.ie.Exec(ctx, "delete-consume-join-token", txn,
			"DELETE FROM system.join_tokens WHERE id = $1",
			clientToken.TokenID.String())
		if err != nil {
			__antithesis_instrumentation__.Notify(187512)
			return err
		} else {
			__antithesis_instrumentation__.Notify(187513)
			if i == 0 {
				__antithesis_instrumentation__.Notify(187514)
				return errors.New("error when consuming join token: no token found")
			} else {
				__antithesis_instrumentation__.Notify(187515)
			}
		}
		__antithesis_instrumentation__.Notify(187505)

		return nil
	})
}

func (s *adminServer) RequestCertBundle(
	ctx context.Context, req *serverpb.CertBundleRequest,
) (*serverpb.CertBundleResponse, error) {
	__antithesis_instrumentation__.Notify(187516)
	settings := s.server.ClusterSettings()
	if settings == nil {
		__antithesis_instrumentation__.Notify(187523)
		return nil, errors.AssertionFailedf("could not look up cluster settings")
	} else {
		__antithesis_instrumentation__.Notify(187524)
	}
	__antithesis_instrumentation__.Notify(187517)
	if !sql.FeatureTLSAutoJoinEnabled.Get(&settings.SV) {
		__antithesis_instrumentation__.Notify(187525)
		return nil, errors.New("feature disabled by administrator")
	} else {
		__antithesis_instrumentation__.Notify(187526)
	}
	__antithesis_instrumentation__.Notify(187518)

	var err error
	var clientToken security.JoinToken
	clientToken.SharedSecret = req.SharedSecret
	clientToken.TokenID, err = uuid.FromString(req.TokenID)
	if err != nil {
		__antithesis_instrumentation__.Notify(187527)
		return nil, ErrInvalidAddJoinToken
	} else {
		__antithesis_instrumentation__.Notify(187528)
	}
	__antithesis_instrumentation__.Notify(187519)

	if err := s.consumeJoinToken(ctx, clientToken); err != nil {
		__antithesis_instrumentation__.Notify(187529)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(187530)
	}
	__antithesis_instrumentation__.Notify(187520)

	certBundle, err := collectLocalCABundle(s.server.cfg.SSLCertsDir)
	if err != nil {
		__antithesis_instrumentation__.Notify(187531)

		return nil, ErrAddJoinTokenConsumed
	} else {
		__antithesis_instrumentation__.Notify(187532)
	}
	__antithesis_instrumentation__.Notify(187521)

	bundleBytes, err := json.Marshal(certBundle)
	if err != nil {
		__antithesis_instrumentation__.Notify(187533)

		return nil, ErrAddJoinTokenConsumed
	} else {
		__antithesis_instrumentation__.Notify(187534)
	}
	__antithesis_instrumentation__.Notify(187522)

	res := &serverpb.CertBundleResponse{
		Bundle: bundleBytes,
	}
	return res, nil
}
