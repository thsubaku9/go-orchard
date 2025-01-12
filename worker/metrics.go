package worker

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type CPUMetric struct {
	TimeStat       cpu.TimesStat
	PercentageUsed float64
}
type Metrics struct {
	Load   load.AvgStat
	CPU    CPUMetric
	Disk   disk.UsageStat
	Memory mem.VirtualMemoryStat
}

func GetLoadMetrics() load.AvgStat {
	res, _ := load.Avg()
	return *res

}

func GetCPUMetrics() cpu.TimesStat {
	res, _ := cpu.Times(false)
	return res[0]
}

func getCPUUsage(stats cpu.TimesStat) float64 {
	idle := stats.Idle + stats.Iowait
	nonIdle := stats.User + stats.Nice + stats.System + stats.Irq + stats.Softirq + stats.Steal

	total := idle + nonIdle
	if total == 0 {
		return 0.00
	}
	return (float64(total) - float64(idle)) / float64(total)
}

func GetDiskMetrics() disk.UsageStat {
	res, _ := disk.Usage("/")
	return *res
}

func GetMemoryMetrics() mem.VirtualMemoryStat {
	res, _ := mem.VirtualMemory()
	return *res
}

func GetFullMetrics() Metrics {

	cpuTimeStates := GetCPUMetrics()
	return Metrics{
		Load:   GetLoadMetrics(),
		Disk:   GetDiskMetrics(),
		Memory: GetMemoryMetrics(),
		CPU:    CPUMetric{TimeStat: cpuTimeStates, PercentageUsed: getCPUUsage(cpuTimeStates)},
	}
}

func DeliverPeriodicStats(d time.Duration, bufferSize int) <-chan Metrics {

	ticker := time.NewTicker(d)
	metricChannel := make(chan Metrics, bufferSize)

	go func() {
		for range ticker.C {
			metricChannel <- GetFullMetrics()
		}
	}()

	return metricChannel
}
