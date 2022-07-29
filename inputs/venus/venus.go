package venus

import (
	"fmt"
	"strings"

	"flashcat.cloud/categraf/config"
	"flashcat.cloud/categraf/inputs"
	"flashcat.cloud/categraf/pkg/conv"
	"flashcat.cloud/categraf/types"
	"github.com/gaochao1/sw"
	go_snmp "github.com/ulricqin/gosnmp"
)

const inputName = "venus"

type VenusADC struct {
	config.PluginConfig
	Instances []*Instance `toml:"instances"`
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &VenusADC{}
	})
}

func (r *VenusADC) Init() error {
	if len(r.Instances) == 0 {
		return types.ErrInstancesEmpty
	}

	for i := 0; i < len(r.Instances); i++ {
		if err := r.Instances[i].Init(); err != nil {
			return err
		}
	}

	return nil
}

func (r *VenusADC) GetInstances() []inputs.Instance {
	ret := make([]inputs.Instance, len(r.Instances))
	for i := 0; i < len(r.Instances); i++ {
		ret[i] = r.Instances[i]
	}
	return ret
}

func (r *VenusADC) Drop() {}

func (r *VenusADC) Gather(slist *types.SampleList) {}

type Instance struct {
	config.InstanceConfig
	Labels        map[string]string `toml:"labels"`
	IntervalTimes int64             `toml:"interval_times"`
	IP            string            `toml:"ip"`
	SnmpTimeoutMs int64             `toml:"snmp_timeout_ms"`
	Community     string            `toml:"community"`
}

func (ins *Instance) Init() error {
	return nil
}

func (ins *Instance) Gather(slist *types.SampleList) {
	var err error
	var snmpPDUs []go_snmp.SnmpPDU
	var fields map[string]interface{} = make(map[string]interface{})
	var tags map[string]string

	// cpu usage
	snmpPDUs, err = sw.RunSnmp(ins.IP, ins.Community, "1.3.6.1.4.1.15227.1.3.1.1.1.0", "get", int(ins.SnmpTimeoutMs))
	if len(snmpPDUs) > 0 && err == nil {
		value := fmt.Sprintf("%v", snmpPDUs[0].Value)
		value = strings.ReplaceAll(value, "%", "")
		metric, _ := conv.ToFloat64(value)
		fields["cpu_usage"] = metric
	}

	// mem usage
	snmpPDUs, err = sw.RunSnmp(ins.IP, ins.Community, "1.3.6.1.4.1.15227.1.3.1.1.2.0", "get", int(ins.SnmpTimeoutMs))
	if len(snmpPDUs) > 0 && err == nil {
		value := fmt.Sprintf("%v", snmpPDUs[0].Value)
		value = strings.ReplaceAll(value, "%", "")
		metric, _ := conv.ToFloat64(value)
		fields["mem_usage"] = metric
	}

	// sessions
	snmpPDUs, err = sw.RunSnmp(ins.IP, ins.Community, "1.3.6.1.4.1.15227.1.3.1.1.3.0", "get", int(ins.SnmpTimeoutMs))
	if len(snmpPDUs) > 0 && err == nil {
		metric, _ := conv.ToFloat64(snmpPDUs[0].Value)
		fields["sessions"] = metric
	}

	// forwardRate kbps
	snmpPDUs, err = sw.RunSnmp(ins.IP, ins.Community, "1.3.6.1.4.1.15227.1.3.1.1.5.0", "get", int(ins.SnmpTimeoutMs))
	if len(snmpPDUs) > 0 && err == nil {
		value := fmt.Sprintf("%v", snmpPDUs[0].Value)
		value = strings.ReplaceAll(value, "kbps", "")
		metric, _ := conv.ToFloat64(value)
		fields["forward_rate"] = metric
	}

	slist.PushSamples(inputName, fields, tags)
}
