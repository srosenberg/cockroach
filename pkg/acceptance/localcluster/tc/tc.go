// Package tc contains utility methods for using the Linux tc (traffic control)
// command to mess with the network links between cockroach nodes running on
// the local machine.
//
// Requires passwordless sudo in order to run tc.
//
// Does not work on OS X due to the lack of the tc command (and even an
// alternative wouldn't work for the current use case of this code, which also
// requires being able to bind to multiple localhost addresses).
package tc

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
)

const (
	rootHandle   = 1
	defaultClass = 1
)

type Controller struct {
	interfaces []string
	nextClass  int
}

func NewController(interfaces ...string) *Controller {
	__antithesis_instrumentation__.Notify(1127)
	return &Controller{
		interfaces: interfaces,
		nextClass:  defaultClass + 1,
	}
}

func (c *Controller) Init() error {
	__antithesis_instrumentation__.Notify(1128)
	for _, ifce := range c.interfaces {
		__antithesis_instrumentation__.Notify(1130)
		_, _ = exec.Command("sudo", strings.Split(fmt.Sprintf("tc qdisc del dev %s root", ifce), " ")...).Output()
		out, err := exec.Command("sudo", strings.Split(fmt.Sprintf("tc qdisc add dev %s root handle %d: htb default %d",
			ifce, rootHandle, defaultClass), " ")...).Output()
		if err != nil {
			__antithesis_instrumentation__.Notify(1132)
			return errors.Wrapf(err, "failed to create root tc qdisc for %q: %s", ifce, out)
		} else {
			__antithesis_instrumentation__.Notify(1133)
		}
		__antithesis_instrumentation__.Notify(1131)

		out, err = exec.Command("sudo", strings.Split(fmt.Sprintf("tc class add dev %s parent %d: classid %d:%d htb rate 100mbit",
			ifce, rootHandle, rootHandle, defaultClass), " ")...).Output()
		if err != nil {
			__antithesis_instrumentation__.Notify(1134)
			return errors.Wrapf(err, "failed to create root tc class for %q: %s", ifce, out)
		} else {
			__antithesis_instrumentation__.Notify(1135)
		}
	}
	__antithesis_instrumentation__.Notify(1129)
	return nil
}

func (c *Controller) AddLatency(srcIP, dstIP string, latency time.Duration) error {
	__antithesis_instrumentation__.Notify(1136)
	class := c.nextClass
	handle := class * 10
	c.nextClass++
	for _, ifce := range c.interfaces {
		__antithesis_instrumentation__.Notify(1138)
		out, err := exec.Command("sudo", strings.Split(fmt.Sprintf("tc class add dev %s parent %d: classid %d:%d htb rate 100mbit",
			ifce, rootHandle, rootHandle, class), " ")...).Output()
		if err != nil {
			__antithesis_instrumentation__.Notify(1141)
			return errors.Wrapf(err, "failed to add tc class %d: %s", class, out)
		} else {
			__antithesis_instrumentation__.Notify(1142)
		}
		__antithesis_instrumentation__.Notify(1139)
		out, err = exec.Command("sudo", strings.Split(fmt.Sprintf("tc qdisc add dev %s parent %d:%d handle %d: netem delay %v",
			ifce, rootHandle, class, handle, latency), " ")...).Output()
		if err != nil {
			__antithesis_instrumentation__.Notify(1143)
			return errors.Wrapf(err, "failed to add tc netem delay of %v: %s", latency, out)
		} else {
			__antithesis_instrumentation__.Notify(1144)
		}
		__antithesis_instrumentation__.Notify(1140)
		out, err = exec.Command("sudo", strings.Split(fmt.Sprintf("tc filter add dev %s parent %d: protocol ip u32 match ip src %s/32 match ip dst %s/32 flowid %d:%d",
			ifce, rootHandle, srcIP, dstIP, rootHandle, class), " ")...).Output()
		if err != nil {
			__antithesis_instrumentation__.Notify(1145)
			return errors.Wrapf(err, "failed to add tc filter rule between %s and %s: %s", srcIP, dstIP, out)
		} else {
			__antithesis_instrumentation__.Notify(1146)
		}
	}
	__antithesis_instrumentation__.Notify(1137)
	return nil
}

func (c *Controller) CleanUp() error {
	__antithesis_instrumentation__.Notify(1147)
	var err error
	for _, ifce := range c.interfaces {
		__antithesis_instrumentation__.Notify(1149)
		out, thisErr := exec.Command("sudo", strings.Split(fmt.Sprintf("tc qdisc del dev %s root", ifce), " ")...).Output()
		if err != nil {
			__antithesis_instrumentation__.Notify(1150)
			err = errors.CombineErrors(err, errors.Wrapf(
				thisErr, "failed to remove tc rules for %q -- you may have to remove them manually: %s", ifce, out))
		} else {
			__antithesis_instrumentation__.Notify(1151)
		}
	}
	__antithesis_instrumentation__.Notify(1148)
	return err
}
