package cli

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cockroachdb/cockroach/pkg/cli/clierrorplus"
	"github.com/cockroachdb/cockroach/pkg/cli/clisqlexec"
	"github.com/cockroachdb/cockroach/pkg/security"
	"github.com/cockroachdb/errors"
	"github.com/spf13/cobra"
)

const defaultKeySize = 2048

const defaultCALifetime = 10 * 366 * 24 * time.Hour
const defaultCertLifetime = 5 * 366 * 24 * time.Hour

var createCACertCmd = &cobra.Command{
	Use:   "create-ca --certs-dir=<path to cockroach certs dir> --ca-key=<path-to-ca-key>",
	Short: "create CA certificate and key",
	Long: `
Generate a CA certificate "<certs-dir>/ca.crt" and CA key "<ca-key>".
The certs directory is created if it does not exist.

If the CA key exists and --allow-ca-key-reuse is true, the key is used.
If the CA certificate exists and --overwrite is true, the new CA certificate is prepended to it.
`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runCreateCACert),
}

func runCreateCACert(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(27996)
	return errors.Wrap(
		security.CreateCAPair(
			certCtx.certsDir,
			certCtx.caKey,
			certCtx.keySize,
			certCtx.caCertificateLifetime,
			certCtx.allowCAKeyReuse,
			certCtx.overwriteFiles),
		"failed to generate CA cert and key")
}

var createClientCACertCmd = &cobra.Command{
	Use:   "create-client-ca --certs-dir=<path to cockroach certs dir> --ca-key=<path-to-client-ca-key>",
	Short: "create client CA certificate and key",
	Long: `
Generate a client CA certificate "<certs-dir>/ca-client.crt" and CA key "<client-ca-key>".
The certs directory is created if it does not exist.

If the CA key exists and --allow-ca-key-reuse is true, the key is used.
If the CA certificate exists and --overwrite is true, the new CA certificate is prepended to it.

The client CA is optional and should only be used when separate CAs are desired for server certificates
and client certificates.

If the client CA exists, a client.node.crt client certificate must be created using:
  cockroach cert create-client node

Once the client.node.crt exists, all client certificates will be verified using the client CA.
`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runCreateClientCACert),
}

func runCreateClientCACert(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(27997)
	return errors.Wrap(
		security.CreateClientCAPair(
			certCtx.certsDir,
			certCtx.caKey,
			certCtx.keySize,
			certCtx.caCertificateLifetime,
			certCtx.allowCAKeyReuse,
			certCtx.overwriteFiles),
		"failed to generate client CA cert and key")
}

var createNodeCertCmd = &cobra.Command{
	Use:   "create-node --certs-dir=<path to cockroach certs dir> --ca-key=<path-to-ca-key> <host 1> <host 2> ... <host N>",
	Short: "create node certificate and key",
	Long: `
Generate a node certificate "<certs-dir>/node.crt" and key "<certs-dir>/node.key".

If --overwrite is true, any existing files are overwritten.

At least one host should be passed in (either IP address or dns name).

Requires a CA cert in "<certs-dir>/ca.crt" and matching key in "--ca-key".
If "ca.crt" contains more than one certificate, the first is used.
Creation fails if the CA expiration time is before the desired certificate expiration.
`,
	Args: func(cmd *cobra.Command, args []string) error {
		__antithesis_instrumentation__.Notify(27998)
		if len(args) == 0 {
			__antithesis_instrumentation__.Notify(28000)
			return errors.Errorf("create-node requires at least one host name or address, none was specified")
		} else {
			__antithesis_instrumentation__.Notify(28001)
		}
		__antithesis_instrumentation__.Notify(27999)
		return nil
	},
	RunE: clierrorplus.MaybeDecorateError(runCreateNodeCert),
}

func runCreateNodeCert(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(28002)
	return errors.Wrap(
		security.CreateNodePair(
			certCtx.certsDir,
			certCtx.caKey,
			certCtx.keySize,
			certCtx.certificateLifetime,
			certCtx.overwriteFiles,
			args),
		"failed to generate node certificate and key")
}

var createClientCertCmd = &cobra.Command{
	Use:   "create-client --certs-dir=<path to cockroach certs dir> --ca-key=<path-to-ca-key> <username>",
	Short: "create client certificate and key",
	Long: `
Generate a client certificate "<certs-dir>/client.<username>.crt" and key
"<certs-dir>/client.<username>.key".

If --overwrite is true, any existing files are overwritten.

Requires a CA cert in "<certs-dir>/ca.crt" and matching key in "--ca-key".
If "ca.crt" contains more than one certificate, the first is used.
Creation fails if the CA expiration time is before the desired certificate expiration.
`,
	Args: cobra.ExactArgs(1),
	RunE: clierrorplus.MaybeDecorateError(runCreateClientCert),
}

func runCreateClientCert(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(28003)
	username, err := security.MakeSQLUsernameFromUserInput(args[0], security.UsernameCreation)
	if err != nil {
		__antithesis_instrumentation__.Notify(28005)
		return errors.Wrap(err, "failed to generate client certificate and key")
	} else {
		__antithesis_instrumentation__.Notify(28006)
	}
	__antithesis_instrumentation__.Notify(28004)

	return errors.Wrap(
		security.CreateClientPair(
			certCtx.certsDir,
			certCtx.caKey,
			certCtx.keySize,
			certCtx.certificateLifetime,
			certCtx.overwriteFiles,
			username,
			certCtx.generatePKCS8Key),
		"failed to generate client certificate and key")
}

var listCertsCmd = &cobra.Command{
	Use:   "list",
	Short: "list certs in --certs-dir",
	Long: `
List certificates and keys found in the certificate directory.
`,
	Args: cobra.NoArgs,
	RunE: clierrorplus.MaybeDecorateError(runListCerts),
}

func runListCerts(cmd *cobra.Command, args []string) error {
	__antithesis_instrumentation__.Notify(28007)
	cm, err := security.NewCertificateManager(certCtx.certsDir, security.CommandTLSSettings{})
	if err != nil {
		__antithesis_instrumentation__.Notify(28016)
		return errors.Wrap(err, "cannot load certificates")
	} else {
		__antithesis_instrumentation__.Notify(28017)
	}
	__antithesis_instrumentation__.Notify(28008)

	fmt.Fprintf(os.Stdout, "Certificate directory: %s\n", certCtx.certsDir)

	certTableHeaders := []string{"Usage", "Certificate File", "Key File", "Expires", "Notes", "Error"}
	alignment := "llllll"
	var rows [][]string

	addRow := func(ci *security.CertInfo, notes string) {
		__antithesis_instrumentation__.Notify(28018)
		var errString string
		if ci.Error != nil {
			__antithesis_instrumentation__.Notify(28020)
			errString = ci.Error.Error()
		} else {
			__antithesis_instrumentation__.Notify(28021)
		}
		__antithesis_instrumentation__.Notify(28019)
		rows = append(rows, []string{
			ci.FileUsage.String(),
			ci.Filename,
			ci.KeyFilename,
			ci.ExpirationTime.Format("2006/01/02"),
			notes,
			errString,
		})
	}
	__antithesis_instrumentation__.Notify(28009)

	if cert := cm.CACert(); cert != nil {
		__antithesis_instrumentation__.Notify(28022)
		var notes string
		if cert.Error == nil && func() bool {
			__antithesis_instrumentation__.Notify(28024)
			return len(cert.ParsedCertificates) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(28025)
			notes = fmt.Sprintf("num certs: %d", len(cert.ParsedCertificates))
		} else {
			__antithesis_instrumentation__.Notify(28026)
		}
		__antithesis_instrumentation__.Notify(28023)
		addRow(cert, notes)
	} else {
		__antithesis_instrumentation__.Notify(28027)
	}
	__antithesis_instrumentation__.Notify(28010)

	if cert := cm.ClientCACert(); cert != nil {
		__antithesis_instrumentation__.Notify(28028)
		var notes string
		if cert.Error == nil && func() bool {
			__antithesis_instrumentation__.Notify(28030)
			return len(cert.ParsedCertificates) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(28031)
			notes = fmt.Sprintf("num certs: %d", len(cert.ParsedCertificates))
		} else {
			__antithesis_instrumentation__.Notify(28032)
		}
		__antithesis_instrumentation__.Notify(28029)
		addRow(cert, notes)
	} else {
		__antithesis_instrumentation__.Notify(28033)
	}
	__antithesis_instrumentation__.Notify(28011)

	if cert := cm.UICACert(); cert != nil {
		__antithesis_instrumentation__.Notify(28034)
		var notes string
		if cert.Error == nil && func() bool {
			__antithesis_instrumentation__.Notify(28036)
			return len(cert.ParsedCertificates) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(28037)
			notes = fmt.Sprintf("num certs: %d", len(cert.ParsedCertificates))
		} else {
			__antithesis_instrumentation__.Notify(28038)
		}
		__antithesis_instrumentation__.Notify(28035)
		addRow(cert, notes)
	} else {
		__antithesis_instrumentation__.Notify(28039)
	}
	__antithesis_instrumentation__.Notify(28012)

	if cert := cm.NodeCert(); cert != nil {
		__antithesis_instrumentation__.Notify(28040)
		var addresses []string
		if cert.Error == nil && func() bool {
			__antithesis_instrumentation__.Notify(28042)
			return len(cert.ParsedCertificates) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(28043)
			addresses = cert.ParsedCertificates[0].DNSNames
			for _, ip := range cert.ParsedCertificates[0].IPAddresses {
				__antithesis_instrumentation__.Notify(28044)
				addresses = append(addresses, ip.String())
			}
		} else {
			__antithesis_instrumentation__.Notify(28045)
			addresses = append(addresses, "<unknown>")
		}
		__antithesis_instrumentation__.Notify(28041)

		addRow(cert, fmt.Sprintf("addresses: %s", strings.Join(addresses, ",")))
	} else {
		__antithesis_instrumentation__.Notify(28046)
	}
	__antithesis_instrumentation__.Notify(28013)

	if cert := cm.UICert(); cert != nil {
		__antithesis_instrumentation__.Notify(28047)
		var addresses []string
		if cert.Error == nil && func() bool {
			__antithesis_instrumentation__.Notify(28049)
			return len(cert.ParsedCertificates) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(28050)
			addresses = cert.ParsedCertificates[0].DNSNames
			for _, ip := range cert.ParsedCertificates[0].IPAddresses {
				__antithesis_instrumentation__.Notify(28051)
				addresses = append(addresses, ip.String())
			}
		} else {
			__antithesis_instrumentation__.Notify(28052)
			addresses = append(addresses, "<unknown>")
		}
		__antithesis_instrumentation__.Notify(28048)

		addRow(cert, fmt.Sprintf("addresses: %s", strings.Join(addresses, ",")))
	} else {
		__antithesis_instrumentation__.Notify(28053)
	}
	__antithesis_instrumentation__.Notify(28014)

	for _, cert := range cm.ClientCerts() {
		__antithesis_instrumentation__.Notify(28054)
		var user string
		if cert.Error == nil && func() bool {
			__antithesis_instrumentation__.Notify(28056)
			return len(cert.ParsedCertificates) > 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(28057)
			user = cert.ParsedCertificates[0].Subject.CommonName
		} else {
			__antithesis_instrumentation__.Notify(28058)
			user = "<unknown>"
		}
		__antithesis_instrumentation__.Notify(28055)

		addRow(cert, fmt.Sprintf("user: %s", user))
	}
	__antithesis_instrumentation__.Notify(28015)

	return sqlExecCtx.PrintQueryOutput(os.Stdout, stderr, certTableHeaders, clisqlexec.NewRowSliceIter(rows, alignment))
}

var certCmds = []*cobra.Command{
	createCACertCmd,
	createClientCACertCmd,
	mtCreateTenantCACertCmd,
	createNodeCertCmd,
	createClientCertCmd,
	mtCreateTenantCertCmd,
	mtCreateTenantSigningCertCmd,
	listCertsCmd,
}

var certCmd = func() *cobra.Command {
	__antithesis_instrumentation__.Notify(28059)
	cmd := &cobra.Command{
		Use:   "cert",
		Short: "create ca, node, and client certs",
		RunE:  UsageAndErr,
	}
	cmd.AddCommand(certCmds...)
	return cmd
}()
