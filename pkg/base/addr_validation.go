package base

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/util/log"
	"github.com/cockroachdb/errors"
)

func (cfg *Config) ValidateAddrs(ctx context.Context) error {
	__antithesis_instrumentation__.Notify(1277)

	advHost, advPort, err := validateAdvertiseAddr(ctx,
		cfg.AdvertiseAddr, cfg.Addr, "", cliflags.ListenAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(1284)
		return invalidFlagErr(err, cliflags.AdvertiseAddr)
	} else {
		__antithesis_instrumentation__.Notify(1285)
	}
	__antithesis_instrumentation__.Notify(1278)
	cfg.AdvertiseAddr = net.JoinHostPort(advHost, advPort)

	listenHost, listenPort, err := validateListenAddr(ctx, cfg.Addr, "")
	if err != nil {
		__antithesis_instrumentation__.Notify(1286)
		return invalidFlagErr(err, cliflags.ListenAddr)
	} else {
		__antithesis_instrumentation__.Notify(1287)
	}
	__antithesis_instrumentation__.Notify(1279)
	cfg.Addr = net.JoinHostPort(listenHost, listenPort)

	advSQLHost, advSQLPort, err := validateAdvertiseAddr(ctx,
		cfg.SQLAdvertiseAddr, cfg.SQLAddr, advHost, cliflags.ListenSQLAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(1288)
		return invalidFlagErr(err, cliflags.SQLAdvertiseAddr)
	} else {
		__antithesis_instrumentation__.Notify(1289)
	}
	__antithesis_instrumentation__.Notify(1280)
	cfg.SQLAdvertiseAddr = net.JoinHostPort(advSQLHost, advSQLPort)

	sqlHost, sqlPort, err := validateListenAddr(ctx, cfg.SQLAddr, listenHost)
	if err != nil {
		__antithesis_instrumentation__.Notify(1290)
		return invalidFlagErr(err, cliflags.ListenSQLAddr)
	} else {
		__antithesis_instrumentation__.Notify(1291)
	}
	__antithesis_instrumentation__.Notify(1281)
	cfg.SQLAddr = net.JoinHostPort(sqlHost, sqlPort)

	advHTTPHost, advHTTPPort, err := validateAdvertiseAddr(ctx,
		cfg.HTTPAdvertiseAddr, cfg.HTTPAddr, advHost, cliflags.ListenHTTPAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(1292)
		return errors.Wrap(err, "cannot compute public HTTP address")
	} else {
		__antithesis_instrumentation__.Notify(1293)
	}
	__antithesis_instrumentation__.Notify(1282)
	cfg.HTTPAdvertiseAddr = net.JoinHostPort(advHTTPHost, advHTTPPort)

	httpHost, httpPort, err := validateListenAddr(ctx, cfg.HTTPAddr, listenHost)
	if err != nil {
		__antithesis_instrumentation__.Notify(1294)
		return invalidFlagErr(err, cliflags.ListenHTTPAddr)
	} else {
		__antithesis_instrumentation__.Notify(1295)
	}
	__antithesis_instrumentation__.Notify(1283)
	cfg.HTTPAddr = net.JoinHostPort(httpHost, httpPort)
	return nil
}

func UpdateAddrs(ctx context.Context, addr, advAddr *string, ln net.Addr) error {
	__antithesis_instrumentation__.Notify(1296)
	desiredHost, _, err := net.SplitHostPort(*addr)
	if err != nil {
		__antithesis_instrumentation__.Notify(1302)
		return err
	} else {
		__antithesis_instrumentation__.Notify(1303)
	}
	__antithesis_instrumentation__.Notify(1297)

	lnAddr := ln.String()
	lnHost, lnPort, err := net.SplitHostPort(lnAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(1304)
		return err
	} else {
		__antithesis_instrumentation__.Notify(1305)
	}
	__antithesis_instrumentation__.Notify(1298)
	requestedAll := (desiredHost == "" || func() bool {
		__antithesis_instrumentation__.Notify(1306)
		return desiredHost == "0.0.0.0" == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(1307)
		return desiredHost == "::" == true
	}() == true)
	listenedAll := (lnHost == "" || func() bool {
		__antithesis_instrumentation__.Notify(1308)
		return lnHost == "0.0.0.0" == true
	}() == true || func() bool {
		__antithesis_instrumentation__.Notify(1309)
		return lnHost == "::" == true
	}() == true)
	if (requestedAll && func() bool {
		__antithesis_instrumentation__.Notify(1310)
		return !listenedAll == true
	}() == true) || func() bool {
		__antithesis_instrumentation__.Notify(1311)
		return (!requestedAll && func() bool {
			__antithesis_instrumentation__.Notify(1312)
			return desiredHost != lnHost == true
		}() == true) == true
	}() == true {
		__antithesis_instrumentation__.Notify(1313)
		log.Warningf(ctx, "requested to listen on %q, actually listening on %q", desiredHost, lnHost)
	} else {
		__antithesis_instrumentation__.Notify(1314)
	}
	__antithesis_instrumentation__.Notify(1299)
	*addr = net.JoinHostPort(lnHost, lnPort)

	advHost, advPort, err := net.SplitHostPort(*advAddr)
	if err != nil {
		__antithesis_instrumentation__.Notify(1315)
		return err
	} else {
		__antithesis_instrumentation__.Notify(1316)
	}
	__antithesis_instrumentation__.Notify(1300)
	if advPort == "" || func() bool {
		__antithesis_instrumentation__.Notify(1317)
		return advPort == "0" == true
	}() == true {
		__antithesis_instrumentation__.Notify(1318)
		advPort = lnPort
	} else {
		__antithesis_instrumentation__.Notify(1319)
	}
	__antithesis_instrumentation__.Notify(1301)
	*advAddr = net.JoinHostPort(advHost, advPort)
	return nil
}

func validateAdvertiseAddr(
	ctx context.Context, advAddr, listenAddr, defaultHost string, listenFlag cliflags.FlagInfo,
) (string, string, error) {
	__antithesis_instrumentation__.Notify(1320)
	listenHost, listenPort, err := getListenAddr(listenAddr, defaultHost)
	if err != nil {
		__antithesis_instrumentation__.Notify(1326)
		return "", "", invalidFlagErr(err, listenFlag)
	} else {
		__antithesis_instrumentation__.Notify(1327)
	}
	__antithesis_instrumentation__.Notify(1321)

	advHost, advPort := "", ""
	if advAddr != "" {
		__antithesis_instrumentation__.Notify(1328)
		var err error
		advHost, advPort, err = net.SplitHostPort(advAddr)
		if err != nil {
			__antithesis_instrumentation__.Notify(1329)
			return "", "", err
		} else {
			__antithesis_instrumentation__.Notify(1330)
		}
	} else {
		__antithesis_instrumentation__.Notify(1331)
	}
	__antithesis_instrumentation__.Notify(1322)

	if advPort == "" || func() bool {
		__antithesis_instrumentation__.Notify(1332)
		return advPort == "0" == true
	}() == true {
		__antithesis_instrumentation__.Notify(1333)
		advPort = listenPort
	} else {
		__antithesis_instrumentation__.Notify(1334)
	}
	__antithesis_instrumentation__.Notify(1323)

	portNumber, err := net.DefaultResolver.LookupPort(ctx, "tcp", advPort)
	if err != nil {
		__antithesis_instrumentation__.Notify(1335)
		return "", "", err
	} else {
		__antithesis_instrumentation__.Notify(1336)
	}
	__antithesis_instrumentation__.Notify(1324)
	advPort = strconv.Itoa(portNumber)

	if advHost == "" {
		__antithesis_instrumentation__.Notify(1337)
		if listenHost != "" {
			__antithesis_instrumentation__.Notify(1338)

			advHost = listenHost
		} else {
			__antithesis_instrumentation__.Notify(1339)

			var err error
			advHost, err = os.Hostname()
			if err != nil {
				__antithesis_instrumentation__.Notify(1341)
				return "", "", err
			} else {
				__antithesis_instrumentation__.Notify(1342)
			}
			__antithesis_instrumentation__.Notify(1340)

			_, err = net.DefaultResolver.LookupIPAddr(ctx, advHost)
			if err != nil {
				__antithesis_instrumentation__.Notify(1343)
				return "", "", err
			} else {
				__antithesis_instrumentation__.Notify(1344)
			}
		}
	} else {
		__antithesis_instrumentation__.Notify(1345)
	}
	__antithesis_instrumentation__.Notify(1325)
	return advHost, advPort, nil
}

func validateListenAddr(ctx context.Context, addr, defaultHost string) (string, string, error) {
	__antithesis_instrumentation__.Notify(1346)
	host, port, err := getListenAddr(addr, defaultHost)
	if err != nil {
		__antithesis_instrumentation__.Notify(1348)
		return "", "", err
	} else {
		__antithesis_instrumentation__.Notify(1349)
	}
	__antithesis_instrumentation__.Notify(1347)
	return resolveAddr(ctx, host, port)
}

func getListenAddr(addr, defaultHost string) (string, string, error) {
	__antithesis_instrumentation__.Notify(1350)
	host, port := "", ""
	if addr != "" {
		__antithesis_instrumentation__.Notify(1354)
		var err error
		host, port, err = net.SplitHostPort(addr)
		if err != nil {
			__antithesis_instrumentation__.Notify(1355)
			return "", "", err
		} else {
			__antithesis_instrumentation__.Notify(1356)
		}
	} else {
		__antithesis_instrumentation__.Notify(1357)
	}
	__antithesis_instrumentation__.Notify(1351)
	if host == "" {
		__antithesis_instrumentation__.Notify(1358)
		host = defaultHost
	} else {
		__antithesis_instrumentation__.Notify(1359)
	}
	__antithesis_instrumentation__.Notify(1352)
	if port == "" {
		__antithesis_instrumentation__.Notify(1360)
		port = "0"
	} else {
		__antithesis_instrumentation__.Notify(1361)
	}
	__antithesis_instrumentation__.Notify(1353)
	return host, port, nil
}

func resolveAddr(ctx context.Context, host, port string) (string, string, error) {
	__antithesis_instrumentation__.Notify(1362)
	resolver := net.DefaultResolver

	portNumber, err := resolver.LookupPort(ctx, "tcp", port)
	if err != nil {
		__antithesis_instrumentation__.Notify(1365)
		return "", "", err
	} else {
		__antithesis_instrumentation__.Notify(1366)
	}
	__antithesis_instrumentation__.Notify(1363)
	port = strconv.Itoa(portNumber)

	if host == "" {
		__antithesis_instrumentation__.Notify(1367)

		return host, port, nil
	} else {
		__antithesis_instrumentation__.Notify(1368)
	}
	__antithesis_instrumentation__.Notify(1364)

	addr, err := LookupAddr(ctx, resolver, host)
	return addr, port, err
}

func LookupAddr(ctx context.Context, resolver *net.Resolver, host string) (string, error) {
	__antithesis_instrumentation__.Notify(1369)

	addrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		__antithesis_instrumentation__.Notify(1373)
		return "", err
	} else {
		__antithesis_instrumentation__.Notify(1374)
	}
	__antithesis_instrumentation__.Notify(1370)
	if len(addrs) == 0 {
		__antithesis_instrumentation__.Notify(1375)
		return "", fmt.Errorf("cannot resolve %q to an address", host)
	} else {
		__antithesis_instrumentation__.Notify(1376)
	}
	__antithesis_instrumentation__.Notify(1371)

	for _, addr := range addrs {
		__antithesis_instrumentation__.Notify(1377)
		if ip := addr.IP.To4(); ip != nil {
			__antithesis_instrumentation__.Notify(1378)
			return ip.String(), nil
		} else {
			__antithesis_instrumentation__.Notify(1379)
		}
	}
	__antithesis_instrumentation__.Notify(1372)

	return addrs[0].String(), nil
}

func invalidFlagErr(err error, flag cliflags.FlagInfo) error {
	__antithesis_instrumentation__.Notify(1380)
	return errors.Wrapf(err, "invalid --%s", flag.Name)
}
