package main

import (
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/host"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"log"
	"net/http"
	"time"
)

type State struct {
	CpuUsage int `json:"cpu_usage_per"`
	MemUsage int `json:"mem_usage_per"`
}

type Disk struct {
	Devices    string `json:"devices"`
	MountPoint string `json:"mount_point"`
	FsType     string `json:"fs_type"`
	Total      int    `json:"total"`
	Used       int    `json:"used"`
	Free       int    `json:"free"`
}

type Info struct {
	Cpu struct {
		Cores     int     `json:"cores"`
		ModelName string  `json:"model_name"`
		Load1     float32 `json:"load_1"`
		Load5     float32 `json:"load_5"`
		Load15    float32 `json:"load_15"`
	} `json:"cpu"`
	Memory struct {
		Total     int `json:"total"`
		Available int `json:"available"`
		Used      int `json:"used"`
		Free      int `json:"free"`
	}
	Host struct {
		HostName        string `json:"host_name"`
		Uptime          int    `json:"uptime"`
		Process         int    `json:"process"`
		OS              string `json:"os"`
		KernelVersion   string `json:"kernel_version"`
		KernelArch      string `json:"kernel_arch"`
		Platform        string `json:"platform"`
		PlatformVersion string `json:"platform_version"`
	}
	Disks []Disk
}

const HistoryLength = 100

var information Info
var currentState State
var historyState [HistoryLength]State

func HistoryAppend(info State) {
	for i := 1; i < HistoryLength; i++ {
		historyState[i-1] = historyState[i]
	}
	historyState[HistoryLength-1] = info
}

func main() {
	go updateState()

	router := gin.Default()
	//router.StaticFS("/", http.Dir("dist"))
	v1 := router.Group("v1")
	{
		v1.GET("current", getState)
		v1.GET("history", getHistory)
		v1.GET("moreInfo", getMoreInfo)
	}
	err := router.Run(":9527")
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func CpuUsage() int {
	// CPU使用率
	percent, err := cpu.Percent(time.Millisecond*500, false)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return int(percent[0] + .5)
}

func getInfo() Info {
	var info Info
	cpuInfos, err := cpu.Info()
	if err != nil {
		log.Fatalln(err.Error())
	}
	info.Cpu.Cores = int(cpuInfos[0].Cores)
	info.Cpu.ModelName = cpuInfos[0].ModelName

	cpuLoad, err := load.Avg()
	if err != nil {
		log.Fatalln(err.Error())
	}
	info.Cpu.Load1 = float32(cpuLoad.Load1)
	info.Cpu.Load5 = float32(cpuLoad.Load5)
	info.Cpu.Load15 = float32(cpuLoad.Load15)

	hInfo, err := host.Info()
	if err != nil {
		log.Fatalln(err.Error())
	}
	info.Host.HostName = hInfo.Hostname
	info.Host.KernelArch = hInfo.KernelArch
	info.Host.KernelVersion = hInfo.KernelVersion
	info.Host.OS = hInfo.OS
	info.Host.Platform = hInfo.Platform
	info.Host.PlatformVersion = hInfo.PlatformVersion
	info.Host.Process = int(hInfo.Procs)
	info.Host.Uptime = int(hInfo.Uptime)

	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Fatalln(err.Error())
	}
	info.Memory.Total = int(memInfo.Total)
	info.Memory.Free = int(memInfo.Free)
	info.Memory.Used = int(memInfo.Used)
	info.Memory.Available = int(memInfo.Available)

	partitions, err := disk.Partitions(true)
	if err != nil {
		log.Fatalln(err.Error())
	}
	for _, partition := range partitions {
		usage, _ := disk.Usage(partition.Mountpoint)
		info.Disks = append(info.Disks, Disk{
			Devices:    partition.Device,
			FsType:     partition.Fstype,
			MountPoint: partition.Mountpoint,
			Total:      int(usage.Total),
			Used:       int(usage.Used),
			Free:       int(usage.Free),
		})
		//info.Disk[i].Devices = partition.Device
		//info.Disk[i].FsType = partition.Fstype
		//info.Disk[i].MountPoint = partition.Mountpoint
		//usage, _ := disk.Usage(partition.Mountpoint)
		//info.Disk[i].Total = int(usage.Total)
		//info.Disk[i].Used = int(usage.Used)
		//info.Disk[i].Free = int(usage.Free)
	}
	return info
}

func MemUsage() int {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Fatalln(err.Error())
	}
	return int(memInfo.UsedPercent + .5)
}

func updateState() {
	for {
		currentState.CpuUsage = CpuUsage()
		currentState.MemUsage = MemUsage()
		information = getInfo()
		HistoryAppend(currentState)
		//fmt.Println(currentState)
		time.Sleep(500)
	}
}

func getState(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, currentState)
}

func getHistory(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, historyState)
}

func getMoreInfo(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, information)
}
