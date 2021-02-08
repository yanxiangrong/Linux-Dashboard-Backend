package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
	"github.com/shirou/gopsutil/mem"
	"log"
	"net/http"
	"time"
)

type State struct {
	CpuUsage int `json:"cpu_usage_per"`
	MemUsage int `json:"mem-usage_per"`
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
	//todo Host, Disk, IO
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
	cpuLoad, err := load.Avg()
	if err != nil {
		log.Fatalln(err.Error())
	}
	info.Cpu.Cores = int(cpuInfos[0].Cores)
	info.Cpu.ModelName = cpuInfos[0].ModelName
	info.Cpu.Load1 = float32(cpuLoad.Load1)
	info.Cpu.Load5 = float32(cpuLoad.Load5)
	info.Cpu.Load15 = float32(cpuLoad.Load15)
	//todo Mem, Host, Disk, IO
	return info
}

func MemUsage() int {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Fatalln(err.Error())
	}
	fmt.Printf("mem info:%v\n", memInfo)
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
