package sentinel

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"

	"flashcat.cloud/categraf/config"
	"flashcat.cloud/categraf/inputs"
	"flashcat.cloud/categraf/types"
)

const inputName = "sentinel"

type Sentinel struct {
	config.PluginConfig
	Instances []*Instance `toml:"instances"`
}

type LoginResponse struct {
	Code    int                    `json:"code"`
	Msg     string                 `json:"msg"`
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data"`
}

type MetricResponse struct {
	Code    int      `json:"code"`
	Msg     string   `json:"msg"`
	Success bool     `json:"success"`
	Data    PageInfo `json:"data"`
}

type PageInfo struct {
	Metric map[string][]MetricsInfo `json:"metric"`
}

type MetricsInfo struct {
	PassQps  int     `json:"passQps"`
	Rt       float32 `json:"rt"`
	BlockQps int     `json:"blockQps"`
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Sentinel{}
	})
}

func (pt *Sentinel) GetInstances() []inputs.Instance {
	ret := make([]inputs.Instance, len(pt.Instances))
	for i := 0; i < len(pt.Instances); i++ {
		ret[i] = pt.Instances[i]
	}
	return ret
}

var cookieStore map[string][]*http.Cookie = make(map[string][]*http.Cookie)

func login(ins *Instance) {
	url := fmt.Sprintf(ins.Url+"/auth/login?password=%s&username=%s", ins.Password, ins.Username)
	resp, err := http.PostForm(url, nil)
	if err != nil {
		log.Println("E! ", inputName, err.Error())
		return
	}

	rv, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println("E! ", inputName, err.Error())
		return
	}
	defer resp.Body.Close()

	data := LoginResponse{}
	err = json.Unmarshal(rv, &data)
	if err != nil {
		log.Println("E! ", inputName, err.Error())
		return
	}

	if data.Code == 0 {
		cookies := resp.Cookies()
		cookieStore[ins.Url] = cookies
	}
	return
}

func (ins *Instance) Gather(slist *types.SampleList) {
	if cookieStore[ins.Url] == nil {
		login(ins)
	}

	for _, endpoint := range ins.Endpoints {
		url := fmt.Sprintf(ins.Url+"/metric/queryTopResourceMetric.json?app=neusoft-naat&desc=true&pageIndex=1&pageSize=1&searchKey=%s", url.QueryEscape(endpoint))
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			log.Println("E! ", inputName, err.Error())
			return
		}

		for _, cookie := range cookieStore[ins.Url] {
			req.AddCookie(cookie)
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("E! ", inputName, err.Error())
			return
		}

		if resp.StatusCode == 401 {
			delete(cookieStore, ins.Url)
			return
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("E! ", inputName, err.Error())
			return
		}
		defer resp.Body.Close()

		data := MetricResponse{}
		err = json.Unmarshal(body, &data)
		if err != nil {
			log.Println("E! ", inputName, err.Error())
			return
		}
		if data.Code == 0 {
			metrics, err := data.Data.Metric[endpoint]
			if !err {
				log.Println("E!", inputName, "json format err", data)
				return
			}
			last := metrics[len(metrics)-1]
			tags := map[string]string{
				"endpoint": endpoint,
			}
			fields := map[string]interface{}{
				"passQps":  last.PassQps,
				"rt":       last.Rt,
				"blockQps": last.BlockQps,
			}
			slist.PushSamples(inputName, fields, tags)
		}
	}
}

type Instance struct {
	config.InstanceConfig
	Url       string   `toml:"url"`
	Username  string   `toml:"username"`
	Password  string   `toml:"password"`
	Endpoints []string `toml:"endpoints"`
}
