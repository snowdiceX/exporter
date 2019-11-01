package exporter

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

// ExportMetric defined an interface about prometheus exporter metric
type ExportMetric interface {
	GetKey() string
	SetKey(string)
	Init()
	GetValue() float64
	SetValue(float64)
	GetValueType() prometheus.ValueType
	SetValueType(prometheus.ValueType)
	GetLabelValues() []string
	SetLabelValues([]string)
	IsContained([]string) bool
}

// ImmutableGaugeMetric stores an immutable value
type ImmutableGaugeMetric struct {
	key         string
	value       float64
	labelValues []string
}

// GetKey returns metric key
func (m *ImmutableGaugeMetric) GetKey() string {
	return m.key
}

// SetKey sets metric key
func (m *ImmutableGaugeMetric) SetKey(k string) {
	m.key = k
}

// Init metric
func (m *ImmutableGaugeMetric) Init() {

}

// GetValue returns value
func (m *ImmutableGaugeMetric) GetValue() float64 {
	return m.value
}

// SetValue sets value
func (m *ImmutableGaugeMetric) SetValue(v float64) {
	m.value = v
}

// GetValueType returns value type
func (m *ImmutableGaugeMetric) GetValueType() prometheus.ValueType {
	return prometheus.GaugeValue
}

// SetValueType sets nothing
func (m *ImmutableGaugeMetric) SetValueType(prometheus.ValueType) {}

// GetLabelValues returns label values
func (m *ImmutableGaugeMetric) GetLabelValues() []string {
	return m.labelValues
}

// SetLabelValues sets label values
func (m *ImmutableGaugeMetric) SetLabelValues(values []string) {
	m.labelValues = values
}

// IsContained determines whether the labels are contained
func (m *ImmutableGaugeMetric) IsContained(labels []string) bool {
	return true
}

// CounterMetric stores a counter value
type CounterMetric struct {
	key         string
	value       float64
	labelValues []string
	mux         sync.RWMutex
}

// GetKey returns metric key
func (m *CounterMetric) GetKey() string {
	return m.key
}

// SetKey sets metric key
func (m *CounterMetric) SetKey(k string) {
	m.key = k
}

// Init metric
func (m *CounterMetric) Init() {

}

// GetValue returns value
func (m *CounterMetric) GetValue() float64 {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return m.value
}

// SetValue sets value
func (m *CounterMetric) SetValue(v float64) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.value += v
}

// GetValueType returns value type
func (m *CounterMetric) GetValueType() prometheus.ValueType {
	return prometheus.CounterValue
}

// SetValueType sets nothing
func (m *CounterMetric) SetValueType(prometheus.ValueType) {}

// GetLabelValues returns label values
func (m *CounterMetric) GetLabelValues() []string {
	return m.labelValues
}

// SetLabelValues sets label values
func (m *CounterMetric) SetLabelValues(values []string) {
	m.labelValues = values
}

// IsContained determines whether the labels are contained
func (m *CounterMetric) IsContained(labels []string) bool {
	return true
}

// GaugeMetric wraps prometheus export data
type GaugeMetric struct {
	key         string
	value       float64
	labelValues []string
	mux         sync.RWMutex
}

// GetKey returns metric key
func (m *GaugeMetric) GetKey() string {
	return m.key
}

// SetKey sets metric key
func (m *GaugeMetric) SetKey(k string) {
	m.key = k
}

// Init metric
func (m *GaugeMetric) Init() {

}

// GetValue returns value
func (m *GaugeMetric) GetValue() float64 {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return m.value
}

// SetValue sets value
func (m *GaugeMetric) SetValue(v float64) {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.value = v
}

// GetValueType returns value type
func (m *GaugeMetric) GetValueType() prometheus.ValueType {
	return prometheus.GaugeValue
}

// SetValueType sets nothing
func (m *GaugeMetric) SetValueType(prometheus.ValueType) {}

// GetLabelValues returns label values
func (m *GaugeMetric) GetLabelValues() []string {
	return m.labelValues
}

// SetLabelValues sets label values
func (m *GaugeMetric) SetLabelValues(values []string) {
	m.labelValues = values
}

// IsContained determines whether the labels are contained
func (m *GaugeMetric) IsContained(labels []string) bool {
	if len(m.labelValues) != len(labels) {
		return false
	}
	for i, v := range labels {
		if !strings.EqualFold(v, m.labelValues[i]) {
			return false
		}
	}
	return true
}

// TickerGaugeMetric wraps prometheus export data with ticker job
type TickerGaugeMetric struct {
	key         string
	calculating float64
	value       float64
	labelValues []string
	mux         sync.RWMutex
}

// GetKey returns metric key
func (m *TickerGaugeMetric) GetKey() string {
	return m.key
}

// SetKey sets metric key
func (m *TickerGaugeMetric) SetKey(k string) {
	m.key = k
}

// Init starts ticker goroutine
func (m *TickerGaugeMetric) Init() {
	t := time.NewTicker(time.Duration(1) * time.Second)

	go func() {
		for {
			select {
			case <-t.C:
				{
					m.SetValue(0)
				}
			}
		}
	}()
}

// GetValue returns value
func (m *TickerGaugeMetric) GetValue() float64 {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return m.value
}

// SetValue sets value
func (m *TickerGaugeMetric) SetValue(v float64) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if v == 0 {
		m.value = m.calculating
		m.calculating = 0
		return
	}
	m.calculating += v
}

// GetValueType returns value type
func (m *TickerGaugeMetric) GetValueType() prometheus.ValueType {
	return prometheus.GaugeValue
}

// SetValueType sets nothing
func (m *TickerGaugeMetric) SetValueType(prometheus.ValueType) {}

// GetLabelValues returns label values
func (m *TickerGaugeMetric) GetLabelValues() []string {
	return m.labelValues
}

// SetLabelValues sets label values
func (m *TickerGaugeMetric) SetLabelValues(values []string) {
	m.labelValues = values
}

// IsContained determines whether the labels are contained
func (m *TickerGaugeMetric) IsContained(labels []string) bool {
	return true
}

// TxMaxGaugeMetric wraps prometheus export data with ticker job
type TxMaxGaugeMetric struct {
	key         string
	max         float64
	value       float64
	labelValues []string
	mux         sync.RWMutex
}

// Init starts ticker goroutine
func (m *TxMaxGaugeMetric) Init() {
	t := time.NewTicker(time.Duration(60) * time.Second)

	go func() {
		for {
			select {
			case <-t.C:
				{
					m.ResetValue()
				}
			}
		}
	}()
}

// GetKey returns metric key
func (m *TxMaxGaugeMetric) GetKey() string {
	return m.key
}

// SetKey sets metric key
func (m *TxMaxGaugeMetric) SetKey(k string) {
	m.key = k
}

// GetValue returns value
func (m *TxMaxGaugeMetric) GetValue() float64 {
	m.mux.RLock()
	defer m.mux.RUnlock()

	return m.value
}

// ResetValue sets value to 0
func (m *TxMaxGaugeMetric) ResetValue() {
	m.mux.Lock()
	defer m.mux.Unlock()

	m.value = m.max
	m.max = 0
	fmt.Println("ticker reset: ", m.key, "; ",
		m.value, "; ", m.max)
}

// SetValue sets value
func (m *TxMaxGaugeMetric) SetValue(v float64) {
	m.mux.Lock()
	defer m.mux.Unlock()

	if m.max < v {
		m.max = v
	}
	fmt.Println("ticker max: ", m.key, "; ",
		m.value, "; ", m.max, "; ", v)
}

// GetValueType returns value type
func (m *TxMaxGaugeMetric) GetValueType() prometheus.ValueType {
	return prometheus.GaugeValue
}

// SetValueType sets nothing
func (m *TxMaxGaugeMetric) SetValueType(prometheus.ValueType) {}

// GetLabelValues returns label values
func (m *TxMaxGaugeMetric) GetLabelValues() []string {
	return m.labelValues
}

// SetLabelValues sets label values
func (m *TxMaxGaugeMetric) SetLabelValues(values []string) {
	m.labelValues = values
}

// IsContained determines whether the labels are contained
func (m *TxMaxGaugeMetric) IsContained(labels []string) bool {
	if len(m.labelValues) != len(labels) {
		return false
	}
	for i, v := range labels {
		if !strings.EqualFold(v, m.labelValues[i]) {
			return false
		}
	}
	return true
}
