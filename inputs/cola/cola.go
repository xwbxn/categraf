package cola

import (
	"flashcat.cloud/categraf/config"
	"flashcat.cloud/categraf/inputs"
	"flashcat.cloud/categraf/types"
)

const inputName = "cola"

type Cola struct {
	config.PluginConfig
	Instances []*Instance `toml:"instances"`
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Cola{}
	})
}

func (c *Cola) Clone() inputs.Input {
	return &Cola{}
}

func (c *Cola) Name() string {
	return inputName
}

func (pt *Cola) GetInstances() []inputs.Instance {
	ret := make([]inputs.Instance, len(pt.Instances))
	for i := 0; i < len(pt.Instances); i++ {
		ret[i] = pt.Instances[i]
	}
	return ret
}

type Instance struct {
	config.InstanceConfig
	BaseUrl string   `toml:"base_url"`
	NetLink int      `toml:"net_link"`
	User    string   `toml:"username"`
	Pwd     string   `toml:"password"`
	Table   string   `toml:"table"`
	Fields  []string `toml:"fields"`
	Filter  string   `toml:"filter"`
}

func (i *Instance) Gather(slist *types.SampleList) {
	client := Client{}
	client.BaseUrl = i.BaseUrl
	client.Username = i.User
	client.Password = i.Pwd
	client.Login()
	metrics := client.GetStatistic(i.NetLink, i.Table, i.Fields, i.Filter)
	client.Logout()

	for _, record := range metrics {
		slist.PushSamples(inputName, record, nil)
	}
}
