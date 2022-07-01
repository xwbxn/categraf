package inputs

import (
	"flashcat.cloud/categraf/config"
	"flashcat.cloud/categraf/pkg/conv"
	"flashcat.cloud/categraf/types"
	"github.com/toolkits/pkg/container/list"
)

type Input interface {
	Init() error
	Drop()
	Prefix() string
	GetInterval() config.Duration
	Gather(slist *list.SafeList)
}

type Creator func() Input

var InputCreators = map[string]Creator{}

func Add(name string, creator Creator) {
	InputCreators[name] = creator
}

func NewSample(metric string, value interface{}, labels ...map[string]string) *types.Sample {
	floatValue, err := conv.ToFloat64(value)
	if err != nil {
		return nil
	}

	s := &types.Sample{
		Metric: metric,
		Value:  floatValue,
		Labels: make(map[string]string),
	}

	for i := 0; i < len(labels); i++ {
		for k, v := range labels[i] {
			if v == "-" {
				continue
			}
			s.Labels[k] = v
		}
	}

	return s
}

func NewSamples(fields map[string]interface{}, labels ...map[string]string) []*types.Sample {
	count := len(fields)
	samples := make([]*types.Sample, 0, count)

	for metric, value := range fields {
		floatValue, err := conv.ToFloat64(value)
		if err != nil {
			continue
		}
		samples = append(samples, NewSample(metric, floatValue, labels...))
	}

	return samples
}

func PushSamples(slist *list.SafeList, fields map[string]interface{}, labels ...map[string]string) {
	for metric, value := range fields {
		slist.PushFront(NewSample(metric, value, labels...))
	}
}
