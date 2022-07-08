package arms

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"flashcat.cloud/categraf/config"
	"flashcat.cloud/categraf/inputs"
	"flashcat.cloud/categraf/types"
	"github.com/toolkits/pkg/container/list"
)

const inputName = "arms"

type Arms struct {
	config.Interval
	counter   uint64
	waitgrp   sync.WaitGroup
	Instances []*Instance `toml:"instances"`
}

func init() {
	inputs.Add(inputName, func() inputs.Input {
		return &Arms{}
	})
}

func (r *Arms) Prefix() string {
	return inputName
}

func (r *Arms) Init() error {
	if len(r.Instances) == 0 {
		return types.ErrInstancesEmpty
	}

	for i := 0; i < len(r.Instances); i++ {
		if err := r.Instances[i].Init(); err != nil {
			return err
		}
	}

	return nil
}

func (r *Arms) Drop() {}

func (r *Arms) Gather(slist *list.SafeList) {
	atomic.AddUint64(&r.counter, 1)

	for i := range r.Instances {
		ins := r.Instances[i]

		r.waitgrp.Add(1)
		go func(slist *list.SafeList, ins *Instance) {
			defer r.waitgrp.Done()

			if ins.IntervalTimes > 0 {
				counter := atomic.LoadUint64(&r.counter)
				if counter%uint64(ins.IntervalTimes) != 0 {
					return
				}
			}
			ins.gatherOnce(slist)
		}(slist, ins)
	}

	r.waitgrp.Wait()
}

type ArmsMetric struct {
	COUNT   float32
	ERROR   float32
	RT      float32
	ERRRATE float32
}

type ArmsData struct {
	Metrics    []ArmsMetric `json:"data"`
	Successful bool
}

type ArmsResponse struct {
	Code    uint
	Data    ArmsData
	Success bool
}

type Instance struct {
	Labels        map[string]string `toml:"labels"`
	IntervalTimes int64             `toml:"interval_times"`

	ApiUrl    string   `toml:"api_url"`
	UserId    string   `toml:"user_id"`
	RegionId  string   `toml:"region_id"`
	Endpoints []string `toml:"endpoints"`
	Offset    string   `toml:"offset"`
}

func (ins *Instance) Init() error {
	return nil
}

func (ins *Instance) gatherOnce(slist *list.SafeList) {
	offset, _ := time.ParseDuration(ins.Offset)
	startTime := time.Now().Add(offset)
	endTime := startTime.Add(time.Minute * 1)

	payload := url.Values{}
	payload.Add("_userId", ins.UserId)
	payload.Add("regionId", ins.RegionId)
	payload.Add("startTime", strconv.Itoa(int(startTime.UnixMilli())))
	payload.Add("endTime", strconv.Itoa(int(endTime.UnixMilli())))
	payload.Add("intervalInSec", "60")
	payload.Add("metric", "APPSTAT.TXN")
	payload.Add("measures", "[COUNT,ERRRATE,RT,ERROR]")
	payload.Add("dimensions", "[rpc]")

	for _, v := range ins.Endpoints {
		filterStr := fmt.Sprintf("[{key=regionId,value=%s},{key=rpc,value=%s}]", ins.RegionId, v)
		payload.Set("filters", filterStr)
		response, err := http.PostForm(ins.ApiUrl, payload)
		if err != nil {
			continue
		}

		rv, err := ioutil.ReadAll(response.Body)
		if err != nil {
			continue
		}
		defer response.Body.Close()

		data := ArmsResponse{}
		err = json.Unmarshal(rv, &data)
		if err != nil {
			continue
		}

		tags := map[string]string{
			"endpoint": v,
		}

		if len(data.Data.Metrics) == 0 {
			continue
		}
		metric := data.Data.Metrics[0]
		fields := map[string]interface{}{
			"count":   metric.COUNT,
			"rt":      metric.RT,
			"error":   metric.ERROR,
			"errrate": metric.ERRRATE,
		}

		types.PushSamples(slist, fields, tags)
	}
}
