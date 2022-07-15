package venus

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"

	"flashcat.cloud/categraf/config"
	"flashcat.cloud/categraf/inputs"
	"flashcat.cloud/categraf/pkg/conv"
	"flashcat.cloud/categraf/types"
	"github.com/gaochao1/sw"
	"github.com/toolkits/pkg/container/list"
	go_snmp "github.com/ulricqin/gosnmp"
)

const inputName = "venus"

type VenusADC struct {
	config.Interval
	counter   uint64
	waitgrp   sync.WaitGroup
	Instances []*Instance `toml:"instances"`
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &VenusADC{}
	})
}

func (r *VenusADC) Prefix() string {
	return inputName
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

func (r *VenusADC) Drop() {}

func (r *VenusADC) Gather(slist *list.SafeList) {
	atomic.AddUint64(&r.counter, 1)

	for i := range r.Instances {
		ins := r.Instances[i]

		r.waitgrp.Add(1)
		go func(slist *list.SafeList, ins *Instance) {
			defer r.waitgrp.Done()

			if ins.IntervalTimes > 0 {
				counter := atomic.LoadUint64(&r.counter)
				if counter%uint64(ins.IntervalTimes) != 0 {
					return
				}
			}

			ins.gatherOnce(slist)
		}(slist, ins)
	}

	r.waitgrp.Wait()
}

type Instance struct {
	Labels        map[string]string `toml:"labels"`
	IntervalTimes int64             `toml:"interval_times"`
	IP            string            `toml:"ip"`
	SnmpTimeoutMs int64             `toml:"snmp_timeout_ms"`
	Community     string            `toml:"community"`
}

func (ins *Instance) Init() error {
	return nil
}

func (ins *Instance) gatherOnce(slist *list.SafeList) {
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

	inputs.PushSamples(slist, fields, tags)
}
