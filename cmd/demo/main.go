package main

import (
	"fmt"

	"github.com/snowdiceX/exporter"
	"github.com/spf13/viper"
)

// nolint
const (
	MetricAddr         = ":29014"
	MetricPath         = "/metrics"
	KeyPrefix          = "transfer_"
	KeySuggestGasPrice = "suggest_gas_price"
	KeyErrors          = "errors"
)

func init() {
	initMcs := []*exporter.MetricConfig{
		&exporter.MetricConfig{
			Key:  KeyErrors,
			Type: "CounterMetric",
			Help: "Count the errors from start"},
		&exporter.MetricConfig{
			Key:    KeySuggestGasPrice,
			Type:   "GaugeMetric",
			Help:   "Eth suggest gas price",
			Labels: []string{"network"}}}

	viper.Set(exporter.KeyMetricAddr, MetricAddr)
	viper.Set(exporter.KeyMetricPath, MetricPath)
	viper.Set(exporter.KeyMetricPrefix, KeyPrefix)

	viper.Set(exporter.KeyMetricType, initMcs)
}

// StartMetrics starts the prometheus metrics exporter
func StartMetrics() {
	errChannel := make(chan error, 10)
	startLog(errChannel)

	exporter.StartMetrics(errChannel)

	exporter.Set(KeySuggestGasPrice, 0, "ethereum")
	exporter.Set(KeyErrors, 0)
}

func startLog(errChannel <-chan error) {
	go func() {
		for {
			select {
			case e, ok := <-errChannel:
				{
					if ok && e != nil {
						fmt.Println(e)
					}
				}
			}
		}
	}()
}

func main() {
	StartMetrics()
	fmt.Println("exporter started")
	select {}
}
