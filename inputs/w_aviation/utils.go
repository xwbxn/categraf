package w_aviation

import (
	"net"
	"os/exec"
	"runtime"
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
