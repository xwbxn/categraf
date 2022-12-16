package agent

import (
	"errors"
	"log"

	// auto registry
	_ "flashcat.cloud/categraf/inputs/arms"
	_ "flashcat.cloud/categraf/inputs/arp_packet"
	_ "flashcat.cloud/categraf/inputs/cola"
	_ "flashcat.cloud/categraf/inputs/conntrack"
	_ "flashcat.cloud/categraf/inputs/cpu"
	_ "flashcat.cloud/categraf/inputs/disk"
	_ "flashcat.cloud/categraf/inputs/diskio"
	_ "flashcat.cloud/categraf/inputs/dns_query"
	_ "flashcat.cloud/categraf/inputs/docker"
	_ "flashcat.cloud/categraf/inputs/elasticsearch"
	_ "flashcat.cloud/categraf/inputs/exec"
	_ "flashcat.cloud/categraf/inputs/greenplum"
	_ "flashcat.cloud/categraf/inputs/http_response"
	_ "flashcat.cloud/categraf/inputs/ipvs"
	_ "flashcat.cloud/categraf/inputs/jenkins"
	_ "flashcat.cloud/categraf/inputs/jolokia_agent"
	_ "flashcat.cloud/categraf/inputs/jolokia_proxy"
	_ "flashcat.cloud/categraf/inputs/kafka"
	_ "flashcat.cloud/categraf/inputs/kernel"
	_ "flashcat.cloud/categraf/inputs/kernel_vmstat"
	_ "flashcat.cloud/categraf/inputs/kubernetes"
	_ "flashcat.cloud/categraf/inputs/linux_sysctl_fs"
	_ "flashcat.cloud/categraf/inputs/logstash"
	_ "flashcat.cloud/categraf/inputs/mem"
	_ "flashcat.cloud/categraf/inputs/mongodb"
	_ "flashcat.cloud/categraf/inputs/mtail"
	_ "flashcat.cloud/categraf/inputs/mysql"
	_ "flashcat.cloud/categraf/inputs/net"
	_ "flashcat.cloud/categraf/inputs/net_response"
	_ "flashcat.cloud/categraf/inputs/netstat"
	_ "flashcat.cloud/categraf/inputs/netstat_filter"
	_ "flashcat.cloud/categraf/inputs/nfsclient"
	_ "flashcat.cloud/categraf/inputs/nginx"
	_ "flashcat.cloud/categraf/inputs/nginx_upstream_check"
	_ "flashcat.cloud/categraf/inputs/ntp"
	_ "flashcat.cloud/categraf/inputs/nvidia_smi"
	_ "flashcat.cloud/categraf/inputs/oracle"
	_ "flashcat.cloud/categraf/inputs/phpfpm"
	_ "flashcat.cloud/categraf/inputs/ping"
	_ "flashcat.cloud/categraf/inputs/postgresql"
	_ "flashcat.cloud/categraf/inputs/processes"
	_ "flashcat.cloud/categraf/inputs/procstat"
	_ "flashcat.cloud/categraf/inputs/prometheus"
	_ "flashcat.cloud/categraf/inputs/rabbitmq"
	_ "flashcat.cloud/categraf/inputs/redis"
	_ "flashcat.cloud/categraf/inputs/redis_sentinel"
	_ "flashcat.cloud/categraf/inputs/rocketmq_offset"
	_ "flashcat.cloud/categraf/inputs/sentinel"
	_ "flashcat.cloud/categraf/inputs/snmp"
	_ "flashcat.cloud/categraf/inputs/sqlserver"
	_ "flashcat.cloud/categraf/inputs/switch_legacy"
	_ "flashcat.cloud/categraf/inputs/system"
	_ "flashcat.cloud/categraf/inputs/tomcat"
	_ "flashcat.cloud/categraf/inputs/venus"
	_ "flashcat.cloud/categraf/inputs/vsphere"
	_ "flashcat.cloud/categraf/inputs/zookeeper"
)

type Agent struct {
	agents []AgentModule
}

// AgentModule is the interface for agent modules
// Use NewXXXAgent() to create a new agent module
// if the agent module is not needed, return nil
type AgentModule interface {
	Start() error
	Stop() error
}

func NewAgent() (*Agent, error) {
	agent := &Agent{
		agents: []AgentModule{
			NewMetricsAgent(),
			NewTracesAgent(),
			NewLogsAgent(),
			NewPrometheusAgent(),
			NewIbexAgent(),
		},
	}
	for _, ag := range agent.agents {
		if ag != nil {
			return agent, nil
		}
	}
	return nil, errors.New("no valid running agents, please check configuration")
}

func (a *Agent) Start() {
	log.Println("I! agent starting")
	for _, agent := range a.agents {
		if agent == nil {
			continue
		}
		if err := agent.Start(); err != nil {
			log.Printf("E! start [%T] err: [%+v]", agent, err)
		} else {
			log.Printf("I! [%T] started", agent)
		}
	}
	log.Println("I! agent started")
}

func (a *Agent) Stop() {
	log.Println("I! agent stopping")
	for _, agent := range a.agents {
		if agent == nil {
			continue
		}
		if err := agent.Stop(); err != nil {
			log.Printf("E! stop [%T] err: [%+v]", agent, err)
		} else {
			log.Printf("I! [%T] stopped", agent)
		}
	}
	log.Println("I! agent stopped")
}

func (a *Agent) Reload() {
	log.Println("I! agent reloading")
	a.Stop()
	a.Start()
	log.Println("I! agent reloaded")
}
