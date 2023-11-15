package w_aviation

import (
	"os/exec"
	"strings"
)

func output_command(cmd string, args []string) string {
	outputData, err := exec.Command(cmd, args...).Output()
	if err != nil {
		return "无法获取"
	}
	outputs := strings.Split(string(outputData), string([]rune{13, 13, 10}))
	output := outputs[1]
	// fmt.Println(string(outputData))
	return output
}
 