//go:build windows
// +build windows

package w_aviation

import (
	"regexp"

	//探针自带的
	"flashcat.cloud/categraf/inputs/w_aviation/common"
	"flashcat.cloud/categraf/types" //用于打包发送数据

	//wmi,获取windows信息
	"github.com/StackExchange/wmi"
)

func GetGateway() string {
	output, err := common.Invoke{}.Command("route", "print")
	if err != nil {
		return ""
	}
	re, _ := regexp.Compile(`\s+0\.0\.0\.0\s+0\.0\.0\.0\s+(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s`)
	gw := re.FindStringSubmatch(string(output))
	if len(gw) < 2 {
		return ""
	}
	return gw[1]
}

func (ins *Instance) GetBusInfo(slist *types.SampleList) error {
	type BusDevice struct {
		Caption      string
		Manufacturer string
		Name         string
	}
	var devices []BusDevice
	err := wmi.Query("select * from Win32_PnPEntity where DeviceID like '%PCI%'", &devices)
	if err != nil {
		return err
	}
	for _, device := range devices {

		fields := map[string]interface{}{
			"bus": 1, // os
		}
		tags := map[string]string{
			"name":    device.Caption,
			"vendor":  device.Manufacturer,
			"product": device.Name,
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}
