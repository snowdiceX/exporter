package exporter

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/spf13/viper"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var collector *cassiniCollector

func initCollector() {
	keyPrefix := viper.GetString(KeyMetricPrefix)
	fmt.Println("key prefix: ", keyPrefix)
	var mcs []*MetricConfig
	viper.UnmarshalKey(KeyMetricType, &mcs)

	collector = &cassiniCollector{
		metricConfigs: mcs,
		descs:         make(map[string]*prometheus.Desc)}
	collector.Init()

	for _, metric := range mcs {
		key := fmt.Sprint(keyPrefix, metric.Key)
		fmt.Println("key: ", key)
		collector.descs[metric.Key] = prometheus.NewDesc(
			key,
			metric.Help,
			metric.Labels, nil)
	}
}

// Collector returns a collector
// which exports metrics about status code of network service response
func Collector(errChannel chan<- error) prometheus.Collector {
	collector.SetErrorChannel(errChannel)
	return collector
}

type cassiniCollector struct {
	metricConfigs []*MetricConfig
	descs         map[string]*prometheus.Desc
	mapper        sync.Map
	errChannel    chan<- error
}

func (c *cassiniCollector) Init() {
	t := time.NewTicker(time.Duration(60) * time.Second)

	go func() {
		for {
			select {
			case <-t.C:
				{
					c.mapper.Range(func(k interface{}, v interface{}) bool {
						m, ok := v.(ExportMetric)
						if ok {
							if m.IsTimeout() {
								m.Cancel()
								c.mapper.Delete(k)
							}
						} else {
							c.errChannel <- fmt.Errorf(
								"Collect type assertion error: %s)", k)
						}
						return false
					})
				}
			}
		}
	}()
}

// Describe returns all descriptions of the collector.
func (c *cassiniCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, desc := range c.descs {
		ch <- desc
	}
}

// Collect returns the current state of all metrics of the collector.
func (c *cassiniCollector) Collect(ch chan<- prometheus.Metric) {
	exports := func(k, v interface{}) bool {
		key, ok := k.(string)
		if !ok {
			c.errChannel <- fmt.Errorf("%s%s%s",
				"Collect error: can not convert key(",
				key, ") into a string")
			return true
		}
		var metric ExportMetric
		metric, ok = v.(ExportMetric)
		if !ok {
			// var metrics []*CassiniMetric
			// metrics, ok = v.([]*CassiniMetric)
			// if !ok {
			// 	c.ch <- fmt.Errorf("%s%s%s",
			// 		"Collect error: can not convert value(", key,
			// 		") into a *cassiniMetric or a []*cassiniMetric")
			// 	return true
			// }
			// for _, metric = range metrics {
			// 	c.export(ch, key, metric)
			// }
		} else {
			c.export(ch, metric)
			fmt.Println("Collect: ", key)
		}
		return true
	}
	c.mapper.Range(exports)
}

// SetErrorChannel set a channel for error
func (c *cassiniCollector) SetErrorChannel(errChannel chan<- error) {
	c.errChannel = errChannel
}

func (c *cassiniCollector) Set(key string, value float64,
	labelValues ...string) {
	mapperKey := buildMapperKey(key, labelValues)
	fmt.Println("metric mapper key: ", mapperKey)
	v, loaded := c.mapper.Load(mapperKey)
	if v == nil || !loaded {
		metric := c.createMetric(key)
		if metric == nil {
			return
		}
		metric.SetValue(value)
		metric.SetLabelValues(labelValues)
		if v, loaded = c.mapper.LoadOrStore(mapperKey, metric); !loaded {
			return
		}
	}
	metric, ok := v.(ExportMetric)
	if !ok {
		c.errChannel <- fmt.Errorf("%s%s%s",
			"Count error: can not convert value(",
			key, ") into a *cassiniMetric")
		return
	}
	metric.SetValue(value)
}

// Set the value to the collector mapper
func Set(key string, value float64,
	labelValues ...string) {
	collector.Set(key, value, labelValues...)
}

func (c *cassiniCollector) export(ch chan<- prometheus.Metric,
	metric ExportMetric) {
	desc, ok := c.descs[metric.GetKey()]
	if !ok {
		c.errChannel <- fmt.Errorf(
			"Collect error: can not find desc(%s)",
			metric.GetKey())
		return
	}
	ch <- prometheus.MustNewConstMetric(
		desc,
		metric.GetValueType(),
		metric.GetValue(),
		metric.GetLabelValues()...)
}

func (c *cassiniCollector) createMetric(key string) (metric ExportMetric) {
	mc := c.getMetricConfig(key)
	if mc != nil {
		if strings.EqualFold(mc.Type, "ImmutableGaugeMetric") {
			metric = &ImmutableGaugeMetric{}
		} else if strings.EqualFold(mc.Type, "CounterMetric") {
			metric = &CounterMetric{}
		} else if strings.EqualFold(mc.Type, "TxMaxGaugeMetric") {
			metric = &TxMaxGaugeMetric{}
		} else if strings.EqualFold(mc.Type, "TickerGaugeMetric") {
			metric = &TickerGaugeMetric{}
		} else if strings.EqualFold(mc.Type, "GaugeMetric") {
			metric = &GaugeMetric{}
		}
	} else {
		metric = &GaugeMetric{}
	}
	metric.SetKey(key)
	metric.Init()
	return
}

func (c *cassiniCollector) getMetricConfig(key string) *MetricConfig {
	for _, mc := range c.metricConfigs {
		if strings.EqualFold(key, mc.Key) {
			return mc
		}
	}
	return nil
}

func buildMapperKey(key string, labels []string) string {
	if len(labels) > 0 {
		return fmt.Sprint(key, "_", strings.Join(labels, "_"))
	}
	return key
}

// StartMetrics prometheus exporter("/metrics") service
func StartMetrics(errChannel chan<- error) {

	initCollector()

	go func() {
		prometheus.MustRegister(Collector(errChannel))

		addr := viper.GetString(KeyMetricAddr)
		path := viper.GetString(KeyMetricPath)

		http.Handle(path, promhttp.Handler())
		errChannel <- http.ListenAndServe(addr, nil)
	}()
}
