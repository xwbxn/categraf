//go:build linux || freebsd || darwin || openbsd
// +build linux freebsd darwin openbsd

package w_aviation

import (
	"regexp"

	"flashcat.cloud/categraf/inputs/w_aviation/common"
	"flashcat.cloud/categraf/types"
	"github.com/jaypipes/ghw"
)

func GetGateway() string {
	output, err := common.Invoke{}.Command("ip", "route")
	if err != nil {
		return ""
	}
	re, _ := regexp.Compile(`default.*\s(\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3})\s`)
	gw := re.FindStringSubmatch(string(output))
	if len(gw) < 2 {
		return ""
	}
	return gw[1]
}

func (ins *Instance) GetBusInfo(slist *types.SampleList) error {
	pci, err := ghw.PCI()
	if err != nil {
		return err
	}
	for _, device := range pci.Devices {

		fields := map[string]interface{}{
			"bus": 1, // os
		}
		tags := map[string]string{
			"name":    device.Class.Name,
			"vendor":  device.Vendor.Name,
			"product": device.Product.Name,
		}
		slist.PushSamples(inputName, fields, tags)
	}
	return nil
}
