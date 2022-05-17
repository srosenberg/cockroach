package spec

import __antithesis_instrumentation__ "antithesis.com/instrumentation/wrappers"

import "fmt"

func AWSMachineType(cpus int) string {
	__antithesis_instrumentation__.Notify(44538)
	switch {
	case cpus <= 2:
		__antithesis_instrumentation__.Notify(44539)
		return "c5d.large"
	case cpus <= 4:
		__antithesis_instrumentation__.Notify(44540)
		return "c5d.xlarge"
	case cpus <= 8:
		__antithesis_instrumentation__.Notify(44541)
		return "c5d.2xlarge"
	case cpus <= 16:
		__antithesis_instrumentation__.Notify(44542)
		return "c5d.4xlarge"
	case cpus <= 36:
		__antithesis_instrumentation__.Notify(44543)
		return "c5d.9xlarge"
	case cpus <= 72:
		__antithesis_instrumentation__.Notify(44544)
		return "c5d.18xlarge"
	case cpus <= 96:
		__antithesis_instrumentation__.Notify(44545)

		return "m5d.24xlarge"
	default:
		__antithesis_instrumentation__.Notify(44546)
		panic(fmt.Sprintf("no aws machine type with %d cpus", cpus))
	}
}

func GCEMachineType(cpus int) string {
	__antithesis_instrumentation__.Notify(44547)

	if cpus < 16 {
		__antithesis_instrumentation__.Notify(44549)
		return fmt.Sprintf("n1-standard-%d", cpus)
	} else {
		__antithesis_instrumentation__.Notify(44550)
	}
	__antithesis_instrumentation__.Notify(44548)
	return fmt.Sprintf("n1-highcpu-%d", cpus)
}

func AzureMachineType(cpus int) string {
	__antithesis_instrumentation__.Notify(44551)
	switch {
	case cpus <= 2:
		__antithesis_instrumentation__.Notify(44552)
		return "Standard_D2_v3"
	case cpus <= 4:
		__antithesis_instrumentation__.Notify(44553)
		return "Standard_D4_v3"
	case cpus <= 8:
		__antithesis_instrumentation__.Notify(44554)
		return "Standard_D8_v3"
	case cpus <= 16:
		__antithesis_instrumentation__.Notify(44555)
		return "Standard_D16_v3"
	case cpus <= 36:
		__antithesis_instrumentation__.Notify(44556)
		return "Standard_D32_v3"
	case cpus <= 48:
		__antithesis_instrumentation__.Notify(44557)
		return "Standard_D48_v3"
	case cpus <= 64:
		__antithesis_instrumentation__.Notify(44558)
		return "Standard_D64_v3"
	default:
		__antithesis_instrumentation__.Notify(44559)
		panic(fmt.Sprintf("no azure machine type with %d cpus", cpus))
	}
}
