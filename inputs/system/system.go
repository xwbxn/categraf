package system

import (
	"log"
	"os"
	"strings"

	"flashcat.cloud/categraf/config"
	"flashcat.cloud/categraf/inputs"
	"flashcat.cloud/categraf/types"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/load"
	"github.com/toolkits/pkg/container/list"
)

const inputName = "system"

type SystemStats struct {
	config.Interval
	CollectUserNumber bool `toml:"collect_user_number"`
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &SystemStats{}
	})
}

func (s *SystemStats) Prefix() string {
	return inputName
}

func (s *SystemStats) Init() error {
	return nil
}

func (s *SystemStats) Drop() {}

func (s *SystemStats) Gather(slist *list.SafeList) {
	loadavg, err := load.Avg()
	if err != nil && !strings.Contains(err.Error(), "not implemented") {
		log.Println("E! failed to gather system load:", err)
		return
	}

	numCPUs, err := cpu.Counts(true)
	if err != nil {
		log.Println("E! failed to gather cpu number:", err)
		return
	}

	fields := map[string]interface{}{
		"load1":        loadavg.Load1,
		"load5":        loadavg.Load5,
		"load15":       loadavg.Load15,
		"n_cpus":       numCPUs,
		"load_norm_1":  loadavg.Load1 / float64(numCPUs),
		"load_norm_5":  loadavg.Load5 / float64(numCPUs),
		"load_norm_15": loadavg.Load15 / float64(numCPUs),
	}

	uptime, err := host.Uptime()
	if err != nil {
		log.Println("E! failed to get host uptime:", err)
	} else {
		fields["uptime"] = uptime
	}

	if s.CollectUserNumber {
		users, err := host.Users()
		if err == nil {
			fields["n_users"] = len(users)
		} else if os.IsNotExist(err) {
			log.Println("W! reading os users:", err)
		} else if os.IsPermission(err) {
			log.Println("W! reading os users:", err)
		}
	}

	types.PushSamples(slist, fields)
}
