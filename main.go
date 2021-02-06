package main

import (
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
	"log"
	"net/http"
	"time"
)

type Info struct {
	CpuUsagePer int `json:"cpu_usage_per"`
	MemUsagePer int `json:"mem_usage_per"`
}

const HistoryLength = 100

var currentInfo Info
var historyInfo [HistoryLength] Info

func HistoryAppend(info Info) {
	for i := 1; i < HistoryLength; i++ {
		historyInfo[i - 1] = historyInfo[i]
	}
	historyInfo[HistoryLength - 1] = info
}

func main() {
	go updateDate()

	router := gin.Default()
	//router.StaticFS("/", http.Dir("dist"))
	v1 := router.Group("v1")
	{
		v1.GET("current", getInfo)
		v1.GET("history", getHistory)
	}
	err := router.Run(":9527")
	if err != nil {
		log.Fatalln(err.Error())
	}
}

func getCpuPer() int {
	// CPU使用率
	percent, err := cpu.Percent(time.Second, false)
	if err != nil {
		log.Fatalln(err.Error())
	}
	return int(percent[0])
}

func getMemPer() int {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Fatalln(err.Error())
	}
	return int(memInfo.UsedPercent)
}

func updateDate() {
	for {
		currentInfo.CpuUsagePer = getCpuPer()
		currentInfo.MemUsagePer = getMemPer()
		HistoryAppend(currentInfo)
		//fmt.Println(currentInfo)
		time.Sleep(500)
	}
}

func getInfo(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, currentInfo)
}

func getHistory(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, historyInfo)
}