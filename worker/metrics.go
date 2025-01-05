package worker

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
)

type Metrics struct {
	Load   load.AvgStat
	CPU    cpu.TimesStat
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

func GetDiskMetrics() disk.UsageStat {
	res, _ := disk.Usage("/")
	return *res
}

func GetMemoryMetrics() mem.VirtualMemoryStat {
	res, _ := mem.VirtualMemory()
	return *res
}

func GetFullMetrics() Metrics {
	return Metrics{
		Load:   GetLoadMetrics(),
		Disk:   GetDiskMetrics(),
		Memory: GetMemoryMetrics(),
		CPU:    GetCPUMetrics(),
	}
}
