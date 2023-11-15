package net

import (
	"fmt"
	"net"
	"os/exec"
	"runtime"
	"strings"
)

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
	interfaces, _ := net.Interfaces()
	netaddrs := map[string]map[string]string{}
	for _, inter := range interfaces {
		addrs, _ := inter.Addrs()
		var ipv4_ip, ipv6_ip string
		if len(addrs) < 2 {
			ipv4_ip = addrs[0].String()
			ipv6_ip = "无"
			fmt.Println("- ipv4 IP:", addrs[0].String())
		} else {
			ipv4_ip = addrs[len(addrs)-1].String()
			ipv6_ip = addrs[0].String()
		}

		gateway := getGatewayMacByIface(addrs[len(addrs)-1].String(), inter.Name)
		mac := inter.HardwareAddr.String()

		netaddrs[inter.Name] = map[string]string{
			"ipv4_IP": ipv4_ip,
			"ipv6_IP": ipv6_ip,
			"mac":     mac,
			"gateway": gateway,
		}
	}
	return netaddrs
}
