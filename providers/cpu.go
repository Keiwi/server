package providers

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
)

// CPUUsage innehåller information om CPU användningen
type CPUUsage struct {
	Error   string  `json:"error"`
	Procent float64 `json:"procent"`
}

type AlertProviderCPU struct {
	rw     *sync.RWMutex
	Values []float64
	Avg    float64
	Total  uint
}

func (a AlertProviderCPU) Name() string {
	return "CPU"
}

func (a *AlertProviderCPU) Check(resp string) bool {
	usage := CPUUsage{}
	err := json.Unmarshal([]byte(resp), &usage)
	if err != nil || usage.Error != "" {
		return false
	}

	if uint(a.CountValues()) >= a.GetTotal() {
		a.rw.Lock()
		a.Values = append(a.Values[1:], usage.Procent)
		a.rw.Unlock()
	} else {
		a.rw.Lock()
		a.Values = append(a.Values, usage.Procent)
		a.rw.Unlock()
	}

	if a.avg() > float64(a.GetAvg()) {
		return true
	}
	return false
}

func (a AlertProviderCPU) avg() float64 {
	var i float64 = 0
	if a.Total > uint(a.CountValues()) {
		return i
	}
	for _, v := range a.GetValues() {
		i += v
	}
	return i / float64(a.CountValues())
}

func (a AlertProviderCPU) GetValues() []float64 {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.Values
}

func (a AlertProviderCPU) GetTotal() uint {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.Total
}

func (a AlertProviderCPU) GetAvg() float64 {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return a.Avg
}

func (a AlertProviderCPU) CountValues() int {
	a.rw.RLock()
	defer a.rw.RUnlock()
	return len(a.Values)
}

func (a AlertProviderCPU) Value() string {
	return strconv.FormatFloat(a.avg(), 'g', -1, 64)
}

func (a AlertProviderCPU) Message() string {
	return fmt.Sprintf("CPU Usage: %s", a.Value())
}
