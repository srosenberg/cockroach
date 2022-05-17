package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/rpc"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/cockroach/pkg/server/pgurl"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

type urlParser struct {
	cmd    *cobra.Command
	cliCtx *cliContext

	sslStrict bool
}

func (u urlParser) String() string { __antithesis_instrumentation__.Notify(28155); return "" }

func (u urlParser) Type() string {
	__antithesis_instrumentation__.Notify(28156)
	return "<postgres://...>"
}

func (u urlParser) Set(v string) error {
	__antithesis_instrumentation__.Notify(28157)
	return u.setInternal(v, true)
}

func (u urlParser) setInternal(v string, warn bool) error {
	__antithesis_instrumentation__.Notify(28158)
	parsedURL, err := pgurl.Parse(v)
	if err != nil {
		__antithesis_instrumentation__.Notify(28167)
		return err
	} else {
		__antithesis_instrumentation__.Notify(28168)
	}
	__antithesis_instrumentation__.Notify(28159)

	cliCtx := u.cliCtx

	fl := flagSetForCmd(u.cmd)
	if user := parsedURL.GetUsername(); user != "" {
		__antithesis_instrumentation__.Notify(28169)

		if f := fl.Lookup(cliflags.User.Name); f == nil {
			__antithesis_instrumentation__.Notify(28170)

			if warn {
				__antithesis_instrumentation__.Notify(28171)
				fmt.Fprintf(stderr,
					"warning: --url specifies user/password, but command %q does not accept user/password details - details ignored\n",
					u.cmd.Name())
			} else {
				__antithesis_instrumentation__.Notify(28172)
			}
		} else {
			__antithesis_instrumentation__.Notify(28173)

			cliCtx.sqlConnUser = user

			f.Changed = true
		}
	} else {
		__antithesis_instrumentation__.Notify(28174)
	}
	__antithesis_instrumentation__.Notify(28160)

	net, host, port := parsedURL.GetNetworking()
	if host != "" {
		__antithesis_instrumentation__.Notify(28175)
		cliCtx.clientConnHost = host
		fl.Lookup(cliflags.ClientHost.Name).Changed = true
	} else {
		__antithesis_instrumentation__.Notify(28176)
	}
	__antithesis_instrumentation__.Notify(28161)
	if port != "" {
		__antithesis_instrumentation__.Notify(28177)
		cliCtx.clientConnPort = port
		fl.Lookup(cliflags.ClientPort.Name).Changed = true
	} else {
		__antithesis_instrumentation__.Notify(28178)
	}
	__antithesis_instrumentation__.Notify(28162)

	if db := parsedURL.GetDatabase(); db != "" {
		__antithesis_instrumentation__.Notify(28179)
		if f := fl.Lookup(cliflags.Database.Name); f == nil {
			__antithesis_instrumentation__.Notify(28180)

			if warn {
				__antithesis_instrumentation__.Notify(28181)
				fmt.Fprintf(stderr,
					"warning: --url specifies database %q, but command %q does not accept a database name - database name ignored\n",
					db, u.cmd.Name())
			} else {
				__antithesis_instrumentation__.Notify(28182)
			}
		} else {
			__antithesis_instrumentation__.Notify(28183)
			cliCtx.sqlConnDBName = db
			f.Changed = true
		}
	} else {
		__antithesis_instrumentation__.Notify(28184)
	}
	__antithesis_instrumentation__.Notify(28163)

	flInsecure := fl.Lookup(cliflags.ClientInsecure.Name)
	tlsUsed, tlsMode, caCertPath := parsedURL.GetTLSOptions()

	if tlsUsed && func() bool {
		__antithesis_instrumentation__.Notify(28185)
		return tlsMode == pgurl.TLSUnspecified == true
	}() == true && func() bool {
		__antithesis_instrumentation__.Notify(28186)
		return net == pgurl.ProtoTCP == true
	}() == true {
		__antithesis_instrumentation__.Notify(28187)

		if flInsecure.Changed {
			__antithesis_instrumentation__.Notify(28188)
			var tp pgurl.TransportOption
			if cliCtx.Insecure {
				__antithesis_instrumentation__.Notify(28190)
				tp = pgurl.TransportNone()
				tlsUsed = false
			} else {
				__antithesis_instrumentation__.Notify(28191)
				tlsMode = pgurl.TLSVerifyFull
				tp = pgurl.TransportTLS(tlsMode, caCertPath)
			}
			__antithesis_instrumentation__.Notify(28189)
			parsedURL.WithTransport(tp)
		} else {
			__antithesis_instrumentation__.Notify(28192)

			tlsMode = pgurl.TLSVerifyFull
			parsedURL.WithTransport(pgurl.TransportTLS(tlsMode, caCertPath))
		}
	} else {
		__antithesis_instrumentation__.Notify(28193)
	}
	__antithesis_instrumentation__.Notify(28164)

	if !tlsUsed {
		__antithesis_instrumentation__.Notify(28194)
		if u.sslStrict {
			__antithesis_instrumentation__.Notify(28195)

			if err := flInsecure.Value.Set("true"); err != nil {
				__antithesis_instrumentation__.Notify(28196)
				return errors.Wrapf(err, "setting secure connection based on --url")
			} else {
				__antithesis_instrumentation__.Notify(28197)
			}
		} else {
			__antithesis_instrumentation__.Notify(28198)
		}
	} else {
		__antithesis_instrumentation__.Notify(28199)
		if u.sslStrict {
			__antithesis_instrumentation__.Notify(28202)
			switch tlsMode {
			case pgurl.TLSVerifyFull:
				__antithesis_instrumentation__.Notify(28203)

			default:
				__antithesis_instrumentation__.Notify(28204)
				return fmt.Errorf("command %q only supports sslmode=disable or sslmode=verify-full", u.cmd.Name())
			}
		} else {
			__antithesis_instrumentation__.Notify(28205)
		}
		__antithesis_instrumentation__.Notify(28200)
		if err := flInsecure.Value.Set("false"); err != nil {
			__antithesis_instrumentation__.Notify(28206)
			return errors.Wrapf(err, "setting secure connection based on --url")
		} else {
			__antithesis_instrumentation__.Notify(28207)
		}
		__antithesis_instrumentation__.Notify(28201)

		if u.sslStrict {
			__antithesis_instrumentation__.Notify(28208)

			candidateCertsDir := ""
			foundCertsDir := false
			if fl.Lookup(cliflags.CertsDir.Name).Changed {
				__antithesis_instrumentation__.Notify(28214)

				candidateCertsDir = cliCtx.SSLCertsDir
				candidateCertsDir = os.ExpandEnv(candidateCertsDir)
				candidateCertsDir, err = filepath.Abs(candidateCertsDir)
				if err != nil {
					__antithesis_instrumentation__.Notify(28215)
					return err
				} else {
					__antithesis_instrumentation__.Notify(28216)
				}
			} else {
				__antithesis_instrumentation__.Notify(28217)
			}
			__antithesis_instrumentation__.Notify(28209)

			tryCertsDir := func(optName, opt, expectedFilename string) error {
				__antithesis_instrumentation__.Notify(28218)
				if opt == "" {
					__antithesis_instrumentation__.Notify(28223)

					return nil
				} else {
					__antithesis_instrumentation__.Notify(28224)
				}
				__antithesis_instrumentation__.Notify(28219)

				base := filepath.Base(opt)
				if base != expectedFilename {
					__antithesis_instrumentation__.Notify(28225)
					return fmt.Errorf("invalid file name for %q: expected %q, got %q", optName, expectedFilename, base)
				} else {
					__antithesis_instrumentation__.Notify(28226)
				}
				__antithesis_instrumentation__.Notify(28220)

				dir := filepath.Dir(opt)
				dir, err = filepath.Abs(dir)
				if err != nil {
					__antithesis_instrumentation__.Notify(28227)
					return err
				} else {
					__antithesis_instrumentation__.Notify(28228)
				}
				__antithesis_instrumentation__.Notify(28221)
				if candidateCertsDir != "" {
					__antithesis_instrumentation__.Notify(28229)

					if candidateCertsDir != dir {
						__antithesis_instrumentation__.Notify(28230)
						return fmt.Errorf("non-homogeneous certificate directory: %s=%q, expected %q", optName, opt, candidateCertsDir)
					} else {
						__antithesis_instrumentation__.Notify(28231)
					}
				} else {
					__antithesis_instrumentation__.Notify(28232)

					candidateCertsDir = dir
					foundCertsDir = true
				}
				__antithesis_instrumentation__.Notify(28222)

				return nil
			}
			__antithesis_instrumentation__.Notify(28210)

			userName := security.RootUserName()
			if cliCtx.sqlConnUser != "" {
				__antithesis_instrumentation__.Notify(28233)
				userName, _ = security.MakeSQLUsernameFromUserInput(cliCtx.sqlConnUser, security.UsernameValidation)
			} else {
				__antithesis_instrumentation__.Notify(28234)
			}
			__antithesis_instrumentation__.Notify(28211)
			if err := tryCertsDir("sslrootcert", caCertPath, security.CACertFilename()); err != nil {
				__antithesis_instrumentation__.Notify(28235)
				return err
			} else {
				__antithesis_instrumentation__.Notify(28236)
			}
			__antithesis_instrumentation__.Notify(28212)
			if clientCertEnabled, clientCertPath, clientKeyPath := parsedURL.GetAuthnCert(); clientCertEnabled {
				__antithesis_instrumentation__.Notify(28237)
				if err := tryCertsDir("sslcert", clientCertPath, security.ClientCertFilename(userName)); err != nil {
					__antithesis_instrumentation__.Notify(28239)
					return err
				} else {
					__antithesis_instrumentation__.Notify(28240)
				}
				__antithesis_instrumentation__.Notify(28238)
				if err := tryCertsDir("sslkey", clientKeyPath, security.ClientKeyFilename(userName)); err != nil {
					__antithesis_instrumentation__.Notify(28241)
					return err
				} else {
					__antithesis_instrumentation__.Notify(28242)
				}
			} else {
				__antithesis_instrumentation__.Notify(28243)
			}
			__antithesis_instrumentation__.Notify(28213)

			if foundCertsDir {
				__antithesis_instrumentation__.Notify(28244)
				if err := fl.Set(cliflags.CertsDir.Name, candidateCertsDir); err != nil {
					__antithesis_instrumentation__.Notify(28245)
					return errors.Wrapf(err, "extracting certificate directory")
				} else {
					__antithesis_instrumentation__.Notify(28246)
				}
			} else {
				__antithesis_instrumentation__.Notify(28247)
			}
		} else {
			__antithesis_instrumentation__.Notify(28248)
		}
	}
	__antithesis_instrumentation__.Notify(28165)

	if err := parsedURL.Validate(); err != nil {
		__antithesis_instrumentation__.Notify(28249)

		msg := err.Error()
		if details := errors.FlattenDetails(err); details != "" {
			__antithesis_instrumentation__.Notify(28251)
			msg += "\n" + details
		} else {
			__antithesis_instrumentation__.Notify(28252)
		}
		__antithesis_instrumentation__.Notify(28250)
		return fmt.Errorf("%s", msg)
	} else {
		__antithesis_instrumentation__.Notify(28253)
	}
	__antithesis_instrumentation__.Notify(28166)

	cliCtx.sqlConnURL = parsedURL

	return nil
}

func (cliCtx *cliContext) makeClientConnURL() (*pgurl.URL, error) {
	__antithesis_instrumentation__.Notify(28254)
	var purl *pgurl.URL
	if cliCtx.sqlConnURL != nil {
		__antithesis_instrumentation__.Notify(28261)

		purl = cliCtx.sqlConnURL
	} else {
		__antithesis_instrumentation__.Notify(28262)

		purl = pgurl.New()
	}
	__antithesis_instrumentation__.Notify(28255)

	purl.WithDatabase(cliCtx.sqlConnDBName)
	if _, host, port := purl.GetNetworking(); host != cliCtx.clientConnHost || func() bool {
		__antithesis_instrumentation__.Notify(28263)
		return port != cliCtx.clientConnPort == true
	}() == true {
		__antithesis_instrumentation__.Notify(28264)
		purl.WithNet(pgurl.NetTCP(cliCtx.clientConnHost, cliCtx.clientConnPort))
	} else {
		__antithesis_instrumentation__.Notify(28265)
	}
	__antithesis_instrumentation__.Notify(28256)

	userName, err := security.MakeSQLUsernameFromUserInput(cliCtx.sqlConnUser, security.UsernameValidation)
	if err != nil {
		__antithesis_instrumentation__.Notify(28266)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28267)
	}
	__antithesis_instrumentation__.Notify(28257)
	if userName.Undefined() {
		__antithesis_instrumentation__.Notify(28268)
		userName = security.RootUserName()
	} else {
		__antithesis_instrumentation__.Notify(28269)
	}
	__antithesis_instrumentation__.Notify(28258)

	sCtx := rpc.MakeSecurityContext(cliCtx.Config, security.CommandTLSSettings{}, roachpb.SystemTenantID)
	if err := sCtx.LoadSecurityOptions(purl, userName); err != nil {
		__antithesis_instrumentation__.Notify(28270)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28271)
	}
	__antithesis_instrumentation__.Notify(28259)

	if err := purl.Validate(); err != nil {
		__antithesis_instrumentation__.Notify(28272)
		return nil, err
	} else {
		__antithesis_instrumentation__.Notify(28273)
	}
	__antithesis_instrumentation__.Notify(28260)

	return purl, nil
}
