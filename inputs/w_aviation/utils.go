package w_aviation

import (
	"net"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

func output_command(cmd string, args []string) string {
	outputData, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "无法获取"
	}
	outputs := strings.Split(string(outputData), string([]rune{13, 13, 10}))
	output := outputs[1]
	return output
}

func getGatewayMacByIface(ipv4_s string, name string) string {

	ip_addrs := strings.Split(ipv4_s, "/")
	ip_addr := ip_addrs[0]
	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "route"
		args = []string{"print"}
	} else {
		cmd = "ip"
		args = []string{"route", "show"}
	}

	outputData, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "无法获取"
	}
	output := string(outputData)

	for _, line := range strings.Split(output, "\n") {
		fields := strings.Fields(line)
		if len(fields) < 5 {
			continue
		}
		if (runtime.GOOS == "windows" && fields[3] == ip_addr && len(fields) == 5) || (runtime.GOOS != "windows" && fields[4] == name) {
			if len(fields[2]) > 8 {
				return fields[2]
			}
		}
	}

	return "无法获取"
}

func GetNetInfo() map[string]map[string]string {
	netaddrs := map[string]map[string]string{}

	interfaces, _ := net.Interfaces()
	if interfaces == nil {
		return nil
	}
	for _, inter := range interfaces {
		addrs, _ := inter.Addrs()

		var ipv4_ip, ipv6_ip, gateway string
		if len(addrs) == 0 {
			ipv4_ip = "无"
			ipv6_ip = "无"
			gateway = "无"
		} else if len(addrs) < 2 {
			ipv4_ip = addrs[0].String()
			ipv6_ip = "无"
			gateway = getGatewayMacByIface(addrs[len(addrs)-1].String(), inter.Name)
		} else {
			if runtime.GOOS == "windows" {
				ipv4_ip = addrs[len(addrs)-1].String()
				ipv6_ip = addrs[0].String()
				gateway = getGatewayMacByIface(addrs[len(addrs)-1].String(), inter.Name)
			} else {
				ipv6_ip = addrs[len(addrs)-1].String()
				ipv4_ip = addrs[0].String()
				gateway = getGatewayMacByIface(addrs[0].String(), inter.Name)
			}
		}

		mac := inter.HardwareAddr.String()
		if mac == "" {
			mac = "00:00:00:00:00:00"
		}

		netaddrs[inter.Name] = map[string]string{
			"name":    inter.Name,
			"ipv4_IP": ipv4_ip,
			"ipv6_IP": ipv6_ip,
			"mac":     mac,
			"gateway": gateway,
		}
	}
	return netaddrs
}

//	ID  | SDRType | Type            |SNum| Name             |Status| Reading
//
// 0001 | Full    | Voltage         | 01 | BMC +0.75V       | OK   | 0.74 V
// 0002 | Full    | Voltage         | 02 | BMC +1V          | OK   | 1.01 V
// 0003 | Full    | Voltage         | 03 | BMC +1.5V        | OK   | 1.52 V
// 0004 | Full    | Voltage         | 04 | BMC +1.8V        | OK   | 1.82 V
// extractFieldsFromRegex consumes a regex with named capture groups and returns a kvp map of strings with the results
var re = regexp.MustCompile(`^\s*(?P<id>[^I|^D]*?)\s*\|\s*(?P<sdrtype>.*?)\s*\|\s*(?P<sdr>.*?)\s*\|\s*(?P<snum>.*?)\s*\|\s*(?P<name>.*?)\s*\|\s*(?P<status>.*?)\s*\|\s*(?P<reading>.*?)\s*$`)

func ExtractFieldsFromRegex(input string) map[string]string {
	submatches := re.FindStringSubmatch(input)
	results := make(map[string]string)
	subexpNames := re.SubexpNames()
	if len(subexpNames) > len(submatches) {
		return results
	}
	for i, name := range subexpNames {
		if name != input && name != "" && input != "" {
			results[name] = strings.Trim(submatches[i], "")
		}
	}
	return results
}

func transformReading(val string) float64 {
	arr := strings.Split(val, " ")
	v, err := strconv.ParseFloat(arr[0], 32)
	if err != nil {
		return float64(0)
	}
	return v
}
