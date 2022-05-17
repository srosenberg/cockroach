package base

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"

	"github.com/cockroachdb/cockroach/pkg/cli/cliflags"
	"github.com/cockroachdb/cockroach/pkg/roachpb"
	"github.com/cockroachdb/cockroach/pkg/util/humanizeutil"
	"github.com/cockroachdb/cockroach/pkg/util/netutil/addr"
	"github.com/cockroachdb/errors"
	"github.com/cockroachdb/errors/oserror"
	"github.com/cockroachdb/pebble"
	"github.com/cockroachdb/redact"
	humanize "github.com/dustin/go-humanize"
	"github.com/spf13/pflag"
)

const MinimumStoreSize = 10 * 64 << 20

func GetAbsoluteStorePath(fieldName string, p string) (string, error) {
	__antithesis_instrumentation__.Notify(1547)
	if p[0] == '~' {
		__antithesis_instrumentation__.Notify(1550)
		return "", fmt.Errorf("%s cannot start with '~': %s", fieldName, p)
	} else {
		__antithesis_instrumentation__.Notify(1551)
	}
	__antithesis_instrumentation__.Notify(1548)

	ret, err := filepath.Abs(p)
	if err != nil {
		__antithesis_instrumentation__.Notify(1552)
		return "", errors.Wrapf(err, "could not find absolute path for %s %s", fieldName, p)
	} else {
		__antithesis_instrumentation__.Notify(1553)
	}
	__antithesis_instrumentation__.Notify(1549)
	return ret, nil
}

type SizeSpec struct {
	InBytes int64
	Percent float64
}

type intInterval struct {
	min *int64
	max *int64
}

type floatInterval struct {
	min *float64
	max *float64
}

func NewSizeSpec(
	field redact.SafeString, value string, bytesRange *intInterval, percentRange *floatInterval,
) (SizeSpec, error) {
	__antithesis_instrumentation__.Notify(1554)
	var size SizeSpec
	if fractionRegex.MatchString(value) {
		__antithesis_instrumentation__.Notify(1556)
		percentFactor := 100.0
		factorValue := value
		if value[len(value)-1] == '%' {
			__antithesis_instrumentation__.Notify(1559)
			percentFactor = 1.0
			factorValue = value[:len(value)-1]
		} else {
			__antithesis_instrumentation__.Notify(1560)
		}
		__antithesis_instrumentation__.Notify(1557)
		var err error
		size.Percent, err = strconv.ParseFloat(factorValue, 64)
		size.Percent *= percentFactor
		if err != nil {
			__antithesis_instrumentation__.Notify(1561)
			return SizeSpec{}, errors.Wrapf(err, "could not parse %s size (%s)", field, value)
		} else {
			__antithesis_instrumentation__.Notify(1562)
		}
		__antithesis_instrumentation__.Notify(1558)
		if percentRange != nil {
			__antithesis_instrumentation__.Notify(1563)
			if (percentRange.min != nil && func() bool {
				__antithesis_instrumentation__.Notify(1564)
				return size.Percent < *percentRange.min == true
			}() == true) || func() bool {
				__antithesis_instrumentation__.Notify(1565)
				return (percentRange.max != nil && func() bool {
					__antithesis_instrumentation__.Notify(1566)
					return size.Percent > *percentRange.max == true
				}() == true) == true
			}() == true {
				__antithesis_instrumentation__.Notify(1567)
				return SizeSpec{}, errors.Newf(
					"%s size (%s) must be between %f%% and %f%%",
					field,
					value,
					*percentRange.min,
					*percentRange.max,
				)
			} else {
				__antithesis_instrumentation__.Notify(1568)
			}
		} else {
			__antithesis_instrumentation__.Notify(1569)
		}
	} else {
		__antithesis_instrumentation__.Notify(1570)
		var err error
		size.InBytes, err = humanizeutil.ParseBytes(value)
		if err != nil {
			__antithesis_instrumentation__.Notify(1572)
			return SizeSpec{}, errors.Wrapf(err, "could not parse %s size (%s)", field, value)
		} else {
			__antithesis_instrumentation__.Notify(1573)
		}
		__antithesis_instrumentation__.Notify(1571)
		if bytesRange != nil {
			__antithesis_instrumentation__.Notify(1574)
			if bytesRange.min != nil && func() bool {
				__antithesis_instrumentation__.Notify(1576)
				return size.InBytes < *bytesRange.min == true
			}() == true {
				__antithesis_instrumentation__.Notify(1577)
				return SizeSpec{}, errors.Newf("%s size (%s) must be larger than %s",
					field, value, humanizeutil.IBytes(*bytesRange.min))
			} else {
				__antithesis_instrumentation__.Notify(1578)
			}
			__antithesis_instrumentation__.Notify(1575)
			if bytesRange.max != nil && func() bool {
				__antithesis_instrumentation__.Notify(1579)
				return size.InBytes > *bytesRange.max == true
			}() == true {
				__antithesis_instrumentation__.Notify(1580)
				return SizeSpec{}, errors.Newf("%s size (%s) must be smaller than %s",
					field, value, humanizeutil.IBytes(*bytesRange.max))
			} else {
				__antithesis_instrumentation__.Notify(1581)
			}
		} else {
			__antithesis_instrumentation__.Notify(1582)
		}
	}
	__antithesis_instrumentation__.Notify(1555)
	return size, nil
}

func (ss *SizeSpec) String() string {
	__antithesis_instrumentation__.Notify(1583)
	var buffer bytes.Buffer
	if ss.InBytes != 0 {
		__antithesis_instrumentation__.Notify(1586)
		fmt.Fprintf(&buffer, "--size=%s,", humanizeutil.IBytes(ss.InBytes))
	} else {
		__antithesis_instrumentation__.Notify(1587)
	}
	__antithesis_instrumentation__.Notify(1584)
	if ss.Percent != 0 {
		__antithesis_instrumentation__.Notify(1588)
		fmt.Fprintf(&buffer, "--size=%s%%,", humanize.Ftoa(ss.Percent))
	} else {
		__antithesis_instrumentation__.Notify(1589)
	}
	__antithesis_instrumentation__.Notify(1585)
	return buffer.String()
}

func (ss *SizeSpec) Type() string {
	__antithesis_instrumentation__.Notify(1590)
	return "SizeSpec"
}

var _ pflag.Value = &SizeSpec{}

func (ss *SizeSpec) Set(value string) error {
	__antithesis_instrumentation__.Notify(1591)
	spec, err := NewSizeSpec("specified", value, nil, nil)
	if err != nil {
		__antithesis_instrumentation__.Notify(1593)
		return err
	} else {
		__antithesis_instrumentation__.Notify(1594)
	}
	__antithesis_instrumentation__.Notify(1592)
	ss.InBytes = spec.InBytes
	ss.Percent = spec.Percent
	return nil
}

type StoreSpec struct {
	Path        string
	Size        SizeSpec
	BallastSize *SizeSpec
	InMemory    bool
	Attributes  roachpb.Attributes

	StickyInMemoryEngineID string

	UseFileRegistry bool

	RocksDBOptions string

	PebbleOptions string

	EncryptionOptions []byte
}

func (ss StoreSpec) String() string {
	__antithesis_instrumentation__.Notify(1595)

	var buffer bytes.Buffer
	if len(ss.Path) != 0 {
		__antithesis_instrumentation__.Notify(1604)
		fmt.Fprintf(&buffer, "path=%s,", ss.Path)
	} else {
		__antithesis_instrumentation__.Notify(1605)
	}
	__antithesis_instrumentation__.Notify(1596)
	if ss.InMemory {
		__antithesis_instrumentation__.Notify(1606)
		fmt.Fprint(&buffer, "type=mem,")
	} else {
		__antithesis_instrumentation__.Notify(1607)
	}
	__antithesis_instrumentation__.Notify(1597)
	if ss.Size.InBytes > 0 {
		__antithesis_instrumentation__.Notify(1608)
		fmt.Fprintf(&buffer, "size=%s,", humanizeutil.IBytes(ss.Size.InBytes))
	} else {
		__antithesis_instrumentation__.Notify(1609)
	}
	__antithesis_instrumentation__.Notify(1598)
	if ss.Size.Percent > 0 {
		__antithesis_instrumentation__.Notify(1610)
		fmt.Fprintf(&buffer, "size=%s%%,", humanize.Ftoa(ss.Size.Percent))
	} else {
		__antithesis_instrumentation__.Notify(1611)
	}
	__antithesis_instrumentation__.Notify(1599)
	if ss.BallastSize != nil {
		__antithesis_instrumentation__.Notify(1612)
		if ss.BallastSize.InBytes > 0 {
			__antithesis_instrumentation__.Notify(1614)
			fmt.Fprintf(&buffer, "ballast-size=%s,", humanizeutil.IBytes(ss.BallastSize.InBytes))
		} else {
			__antithesis_instrumentation__.Notify(1615)
		}
		__antithesis_instrumentation__.Notify(1613)
		if ss.BallastSize.Percent > 0 {
			__antithesis_instrumentation__.Notify(1616)
			fmt.Fprintf(&buffer, "ballast-size=%s%%,", humanize.Ftoa(ss.BallastSize.Percent))
		} else {
			__antithesis_instrumentation__.Notify(1617)
		}
	} else {
		__antithesis_instrumentation__.Notify(1618)
	}
	__antithesis_instrumentation__.Notify(1600)
	if len(ss.Attributes.Attrs) > 0 {
		__antithesis_instrumentation__.Notify(1619)
		fmt.Fprint(&buffer, "attrs=")
		for i, attr := range ss.Attributes.Attrs {
			__antithesis_instrumentation__.Notify(1621)
			if i != 0 {
				__antithesis_instrumentation__.Notify(1623)
				fmt.Fprint(&buffer, ":")
			} else {
				__antithesis_instrumentation__.Notify(1624)
			}
			__antithesis_instrumentation__.Notify(1622)
			buffer.WriteString(attr)
		}
		__antithesis_instrumentation__.Notify(1620)
		fmt.Fprintf(&buffer, ",")
	} else {
		__antithesis_instrumentation__.Notify(1625)
	}
	__antithesis_instrumentation__.Notify(1601)
	if len(ss.PebbleOptions) > 0 {
		__antithesis_instrumentation__.Notify(1626)
		optsStr := strings.Replace(ss.PebbleOptions, "\n", " ", -1)
		fmt.Fprint(&buffer, "pebble=")
		fmt.Fprint(&buffer, optsStr)
		fmt.Fprint(&buffer, ",")
	} else {
		__antithesis_instrumentation__.Notify(1627)
	}
	__antithesis_instrumentation__.Notify(1602)

	if l := buffer.Len(); l > 0 {
		__antithesis_instrumentation__.Notify(1628)
		buffer.Truncate(l - 1)
	} else {
		__antithesis_instrumentation__.Notify(1629)
	}
	__antithesis_instrumentation__.Notify(1603)
	return buffer.String()
}

func (ss StoreSpec) IsEncrypted() bool {
	__antithesis_instrumentation__.Notify(1630)
	return len(ss.EncryptionOptions) > 0
}

var fractionRegex = regexp.MustCompile(`^([-]?([0-9]+\.[0-9]*|[0-9]*\.[0-9]+|[0-9]+(\.[0-9]*)?%))$`)

func NewStoreSpec(value string) (StoreSpec, error) {
	__antithesis_instrumentation__.Notify(1631)
	const pathField = "path"
	if len(value) == 0 {
		__antithesis_instrumentation__.Notify(1635)
		return StoreSpec{}, fmt.Errorf("no value specified")
	} else {
		__antithesis_instrumentation__.Notify(1636)
	}
	__antithesis_instrumentation__.Notify(1632)
	var ss StoreSpec
	used := make(map[string]struct{})
	for _, split := range strings.Split(value, ",") {
		__antithesis_instrumentation__.Notify(1637)
		if len(split) == 0 {
			__antithesis_instrumentation__.Notify(1643)
			continue
		} else {
			__antithesis_instrumentation__.Notify(1644)
		}
		__antithesis_instrumentation__.Notify(1638)
		subSplits := strings.SplitN(split, "=", 2)
		var field string
		var value string
		if len(subSplits) == 1 {
			__antithesis_instrumentation__.Notify(1645)
			field = pathField
			value = subSplits[0]
		} else {
			__antithesis_instrumentation__.Notify(1646)
			field = strings.ToLower(subSplits[0])
			value = subSplits[1]
		}
		__antithesis_instrumentation__.Notify(1639)
		if _, ok := used[field]; ok {
			__antithesis_instrumentation__.Notify(1647)
			return StoreSpec{}, fmt.Errorf("%s field was used twice in store definition", field)
		} else {
			__antithesis_instrumentation__.Notify(1648)
		}
		__antithesis_instrumentation__.Notify(1640)
		used[field] = struct{}{}

		if len(field) == 0 {
			__antithesis_instrumentation__.Notify(1649)
			continue
		} else {
			__antithesis_instrumentation__.Notify(1650)
		}
		__antithesis_instrumentation__.Notify(1641)
		if len(value) == 0 {
			__antithesis_instrumentation__.Notify(1651)
			return StoreSpec{}, fmt.Errorf("no value specified for %s", field)
		} else {
			__antithesis_instrumentation__.Notify(1652)
		}
		__antithesis_instrumentation__.Notify(1642)

		switch field {
		case pathField:
			__antithesis_instrumentation__.Notify(1653)
			var err error
			ss.Path, err = GetAbsoluteStorePath(pathField, value)
			if err != nil {
				__antithesis_instrumentation__.Notify(1666)
				return StoreSpec{}, err
			} else {
				__antithesis_instrumentation__.Notify(1667)
			}
		case "size":
			__antithesis_instrumentation__.Notify(1654)
			var err error
			var minBytesAllowed int64 = MinimumStoreSize
			var minPercent float64 = 1
			var maxPercent float64 = 100
			ss.Size, err = NewSizeSpec(
				"store",
				value,
				&intInterval{min: &minBytesAllowed},
				&floatInterval{min: &minPercent, max: &maxPercent},
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(1668)
				return StoreSpec{}, err
			} else {
				__antithesis_instrumentation__.Notify(1669)
			}
		case "ballast-size":
			__antithesis_instrumentation__.Notify(1655)
			var minBytesAllowed int64
			var minPercent float64 = 0
			var maxPercent float64 = 50
			ballastSize, err := NewSizeSpec(
				"ballast",
				value,
				&intInterval{min: &minBytesAllowed},
				&floatInterval{min: &minPercent, max: &maxPercent},
			)
			if err != nil {
				__antithesis_instrumentation__.Notify(1670)
				return StoreSpec{}, err
			} else {
				__antithesis_instrumentation__.Notify(1671)
			}
			__antithesis_instrumentation__.Notify(1656)
			ss.BallastSize = &ballastSize
		case "attrs":
			__antithesis_instrumentation__.Notify(1657)

			attrMap := make(map[string]struct{})
			for _, attribute := range strings.Split(value, ":") {
				__antithesis_instrumentation__.Notify(1672)
				if _, ok := attrMap[attribute]; ok {
					__antithesis_instrumentation__.Notify(1674)
					return StoreSpec{}, fmt.Errorf("duplicate attribute given for store: %s", attribute)
				} else {
					__antithesis_instrumentation__.Notify(1675)
				}
				__antithesis_instrumentation__.Notify(1673)
				attrMap[attribute] = struct{}{}
			}
			__antithesis_instrumentation__.Notify(1658)
			for attribute := range attrMap {
				__antithesis_instrumentation__.Notify(1676)
				ss.Attributes.Attrs = append(ss.Attributes.Attrs, attribute)
			}
			__antithesis_instrumentation__.Notify(1659)
			sort.Strings(ss.Attributes.Attrs)
		case "type":
			__antithesis_instrumentation__.Notify(1660)
			if value == "mem" {
				__antithesis_instrumentation__.Notify(1677)
				ss.InMemory = true
			} else {
				__antithesis_instrumentation__.Notify(1678)
				return StoreSpec{}, fmt.Errorf("%s is not a valid store type", value)
			}
		case "rocksdb":
			__antithesis_instrumentation__.Notify(1661)
			ss.RocksDBOptions = value
		case "pebble":
			__antithesis_instrumentation__.Notify(1662)

			value = strings.TrimSpace(value)
			var buf bytes.Buffer
			for len(value) > 0 {
				__antithesis_instrumentation__.Notify(1679)
				i := strings.IndexFunc(value, func(r rune) bool {
					__antithesis_instrumentation__.Notify(1681)
					return r == '[' || func() bool {
						__antithesis_instrumentation__.Notify(1682)
						return unicode.IsSpace(r) == true
					}() == true
				})
				__antithesis_instrumentation__.Notify(1680)
				switch {
				case i == -1:
					__antithesis_instrumentation__.Notify(1683)
					buf.WriteString(value)
					value = value[len(value):]
				case value[i] == '[':
					__antithesis_instrumentation__.Notify(1684)

					j := i + strings.IndexRune(value[i:], ']')
					buf.WriteString(value[:j+1])
					value = value[j+1:]
				case unicode.IsSpace(rune(value[i])):
					__antithesis_instrumentation__.Notify(1685)

					buf.WriteString(value[:i])
					buf.WriteRune('\n')
					value = strings.TrimSpace(value[i+1:])
				default:
					__antithesis_instrumentation__.Notify(1686)
				}
			}
			__antithesis_instrumentation__.Notify(1663)

			var opts pebble.Options
			if err := opts.Parse(buf.String(), nil); err != nil {
				__antithesis_instrumentation__.Notify(1687)
				return StoreSpec{}, err
			} else {
				__antithesis_instrumentation__.Notify(1688)
			}
			__antithesis_instrumentation__.Notify(1664)
			ss.PebbleOptions = buf.String()
		default:
			__antithesis_instrumentation__.Notify(1665)
			return StoreSpec{}, fmt.Errorf("%s is not a valid store field", field)
		}
	}
	__antithesis_instrumentation__.Notify(1633)
	if ss.InMemory {
		__antithesis_instrumentation__.Notify(1689)

		if ss.Path != "" {
			__antithesis_instrumentation__.Notify(1692)
			return StoreSpec{}, fmt.Errorf("path specified for in memory store")
		} else {
			__antithesis_instrumentation__.Notify(1693)
		}
		__antithesis_instrumentation__.Notify(1690)
		if ss.Size.Percent == 0 && func() bool {
			__antithesis_instrumentation__.Notify(1694)
			return ss.Size.InBytes == 0 == true
		}() == true {
			__antithesis_instrumentation__.Notify(1695)
			return StoreSpec{}, fmt.Errorf("size must be specified for an in memory store")
		} else {
			__antithesis_instrumentation__.Notify(1696)
		}
		__antithesis_instrumentation__.Notify(1691)
		if ss.BallastSize != nil {
			__antithesis_instrumentation__.Notify(1697)
			return StoreSpec{}, fmt.Errorf("ballast-size specified for in memory store")
		} else {
			__antithesis_instrumentation__.Notify(1698)
		}
	} else {
		__antithesis_instrumentation__.Notify(1699)
		if ss.Path == "" {
			__antithesis_instrumentation__.Notify(1700)
			return StoreSpec{}, fmt.Errorf("no path specified")
		} else {
			__antithesis_instrumentation__.Notify(1701)
		}
	}
	__antithesis_instrumentation__.Notify(1634)
	return ss, nil
}

type StoreSpecList struct {
	Specs   []StoreSpec
	updated bool
}

var _ pflag.Value = &StoreSpecList{}

func (ssl StoreSpecList) String() string {
	__antithesis_instrumentation__.Notify(1702)
	var buffer bytes.Buffer
	for _, ss := range ssl.Specs {
		__antithesis_instrumentation__.Notify(1705)
		fmt.Fprintf(&buffer, "--%s=%s ", cliflags.Store.Name, ss)
	}
	__antithesis_instrumentation__.Notify(1703)

	if l := buffer.Len(); l > 0 {
		__antithesis_instrumentation__.Notify(1706)
		buffer.Truncate(l - 1)
	} else {
		__antithesis_instrumentation__.Notify(1707)
	}
	__antithesis_instrumentation__.Notify(1704)
	return buffer.String()
}

const AuxiliaryDir = "auxiliary"

func EmergencyBallastFile(pathJoin func(...string) string, dataDir string) string {
	__antithesis_instrumentation__.Notify(1708)
	return pathJoin(dataDir, AuxiliaryDir, "EMERGENCY_BALLAST")
}

func PreventedStartupFile(dir string) string {
	__antithesis_instrumentation__.Notify(1709)
	return filepath.Join(dir, "_CRITICAL_ALERT.txt")
}

func (ssl StoreSpecList) PriorCriticalAlertError() (err error) {
	__antithesis_instrumentation__.Notify(1710)
	addError := func(newErr error) {
		__antithesis_instrumentation__.Notify(1713)
		if err == nil {
			__antithesis_instrumentation__.Notify(1715)
			err = errors.New("startup forbidden by prior critical alert")
		} else {
			__antithesis_instrumentation__.Notify(1716)
		}
		__antithesis_instrumentation__.Notify(1714)

		err = errors.WithDetailf(err, "%v", newErr)
	}
	__antithesis_instrumentation__.Notify(1711)
	for _, ss := range ssl.Specs {
		__antithesis_instrumentation__.Notify(1717)
		path := ss.PreventedStartupFile()
		if path == "" {
			__antithesis_instrumentation__.Notify(1720)
			continue
		} else {
			__antithesis_instrumentation__.Notify(1721)
		}
		__antithesis_instrumentation__.Notify(1718)
		b, err := ioutil.ReadFile(path)
		if err != nil {
			__antithesis_instrumentation__.Notify(1722)
			if !oserror.IsNotExist(err) {
				__antithesis_instrumentation__.Notify(1724)
				addError(errors.Wrapf(err, "%s", path))
			} else {
				__antithesis_instrumentation__.Notify(1725)
			}
			__antithesis_instrumentation__.Notify(1723)
			continue
		} else {
			__antithesis_instrumentation__.Notify(1726)
		}
		__antithesis_instrumentation__.Notify(1719)
		addError(errors.Newf("From %s:\n\n%s\n", path, b))
	}
	__antithesis_instrumentation__.Notify(1712)
	return err
}

func (ss StoreSpec) PreventedStartupFile() string {
	__antithesis_instrumentation__.Notify(1727)
	if ss.InMemory {
		__antithesis_instrumentation__.Notify(1729)
		return ""
	} else {
		__antithesis_instrumentation__.Notify(1730)
	}
	__antithesis_instrumentation__.Notify(1728)
	return PreventedStartupFile(filepath.Join(ss.Path, AuxiliaryDir))
}

func (ssl *StoreSpecList) Type() string {
	__antithesis_instrumentation__.Notify(1731)
	return "StoreSpec"
}

func (ssl *StoreSpecList) Set(value string) error {
	__antithesis_instrumentation__.Notify(1732)
	spec, err := NewStoreSpec(value)
	if err != nil {
		__antithesis_instrumentation__.Notify(1735)
		return err
	} else {
		__antithesis_instrumentation__.Notify(1736)
	}
	__antithesis_instrumentation__.Notify(1733)
	if !ssl.updated {
		__antithesis_instrumentation__.Notify(1737)
		ssl.Specs = []StoreSpec{spec}
		ssl.updated = true
	} else {
		__antithesis_instrumentation__.Notify(1738)
		ssl.Specs = append(ssl.Specs, spec)
	}
	__antithesis_instrumentation__.Notify(1734)
	return nil
}

type JoinListType []string

func (jls JoinListType) String() string {
	__antithesis_instrumentation__.Notify(1739)
	var buffer bytes.Buffer
	for _, jl := range jls {
		__antithesis_instrumentation__.Notify(1742)
		fmt.Fprintf(&buffer, "--join=%s ", jl)
	}
	__antithesis_instrumentation__.Notify(1740)

	if l := buffer.Len(); l > 0 {
		__antithesis_instrumentation__.Notify(1743)
		buffer.Truncate(l - 1)
	} else {
		__antithesis_instrumentation__.Notify(1744)
	}
	__antithesis_instrumentation__.Notify(1741)
	return buffer.String()
}

func (jls *JoinListType) Type() string {
	__antithesis_instrumentation__.Notify(1745)
	return "string"
}

func (jls *JoinListType) Set(value string) error {
	__antithesis_instrumentation__.Notify(1746)
	if strings.TrimSpace(value) == "" {
		__antithesis_instrumentation__.Notify(1749)

		return errors.New("no address specified in --join")
	} else {
		__antithesis_instrumentation__.Notify(1750)
	}
	__antithesis_instrumentation__.Notify(1747)
	for _, v := range strings.Split(value, ",") {
		__antithesis_instrumentation__.Notify(1751)
		v = strings.TrimSpace(v)
		if v == "" {
			__antithesis_instrumentation__.Notify(1755)

			continue
		} else {
			__antithesis_instrumentation__.Notify(1756)
		}
		__antithesis_instrumentation__.Notify(1752)

		addr, port, err := addr.SplitHostPort(v, "")
		if err != nil {
			__antithesis_instrumentation__.Notify(1757)
			return err
		} else {
			__antithesis_instrumentation__.Notify(1758)
		}
		__antithesis_instrumentation__.Notify(1753)

		if len(port) == 0 {
			__antithesis_instrumentation__.Notify(1759)
			port = DefaultPort
		} else {
			__antithesis_instrumentation__.Notify(1760)
		}
		__antithesis_instrumentation__.Notify(1754)

		*jls = append(*jls, net.JoinHostPort(addr, port))
	}
	__antithesis_instrumentation__.Notify(1748)
	return nil
}
