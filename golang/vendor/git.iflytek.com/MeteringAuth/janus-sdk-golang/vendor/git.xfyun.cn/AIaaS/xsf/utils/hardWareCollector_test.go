package utils

import (
	"fmt"
	"testing"
)

func Test_HardWareCollector(t *testing.T) {

	checkErr := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	inst, err := newHardWareCollector()
	checkErr(err)

	{
		models, modelsErr := inst.getModel()
		fmt.Printf("models:%v,modelsErr:%v\n", models, modelsErr)
	}
	{
		cpus, cpusErr := inst.getCpus()
		fmt.Printf("cpus:%v,cpusErr:%v\n", cpus, cpusErr)
	}
	{
		cpuMHz, cpuMHzErr := inst.getCpuMHz()
		fmt.Printf("cpuMHz:%v,cpuMHzErr:%v\n", cpuMHz, cpuMHzErr)
	}
	{
		inDocker, inDockerErr := inst.inDocker()
		fmt.Printf("inDocker:%v,inDockerErr:%v\n", inDocker, inDockerErr)
	}
	{
		cpuLimit, cpuLimitErr := inst.getCpuLimit()
		fmt.Printf("cpuLimit:%v,cpuLimitErr:%v\n", cpuLimit, cpuLimitErr)
	}
	{
		cpuUsage, cpuUsageErr := inst.getCpuUsage()
		fmt.Printf("cpuUsage:%v,cpuUsageErr:%v\n", cpuUsage, cpuUsageErr)
	}
	{
		memUsage, memUsageErr := inst.getMemUsage()
		fmt.Printf("memUsage:%v,memUsageErr:%v\n", memUsage, memUsageErr)
	}

}