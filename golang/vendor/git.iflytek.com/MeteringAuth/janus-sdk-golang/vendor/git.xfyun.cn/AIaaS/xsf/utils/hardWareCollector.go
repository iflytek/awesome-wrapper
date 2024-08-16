package utils

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/cpu"
	"io"
	"io/ioutil"
	"os/exec"
	"strconv"
	"strings"
)

var (
	InterErr = errors.New("inter error")
)

func getCpuAndMem(topData string) (cpu, mem float64, err error) {
	var data []string
	for _, v := range strings.Split(topData, " ") {
		if strings.Replace(v, " ", "", -1) != "" {
			data = append(data, v)
		}
	}
	mem, err = strconv.ParseFloat(data[len(data)-3], 10)
	if err != nil {
		return cpu, mem, err
	}
	cpu, err = strconv.ParseFloat(data[len(data)-4], 10)
	return cpu, mem, err
}

type cpuCollector struct {
	info []cpu.InfoStat
}

func newCpuCollector() (*cpuCollector, error) {
	info, infoErr := cpu.Info()
	if infoErr != nil {
		return nil, infoErr
	}
	return &cpuCollector{info: info}, nil
}

func (c *cpuCollector) getModel() ([]string, error) {
	var modes []string
	if len(c.info) == 0 {
		return nil, InterErr
	}
	for _, v := range c.info {
		modes = append(modes, v.ModelName)
	}
	return modes, nil
}

func (c *cpuCollector) getCpus() (int, error) {
	return cpu.Counts(true)
}

func (c *cpuCollector) getCpuMHz() ([]float64, error) {
	if len(c.info) == 0 {
		return nil, InterErr
	}
	var rst []float64
	for _, v := range c.info {
		rst = append(rst, v.Mhz)
	}
	return rst, nil
}

func (c *cpuCollector) inDocker() (bool, error) {
	readLines := func(data []byte) ([]string, error) {
		var rst []string
		r := bufio.NewReader(bytes.NewReader(data))
		for {
			c, _, e := r.ReadLine()
			if e == io.EOF {
				return rst, nil
			}
			if e != nil && e != io.EOF {
				return nil, e
			}
			rst = append(rst, string(c))
		}
	}

	markedDocker := func(data []string) (bool, error) {
		for _, dataItem := range data {
			if strings.Contains(dataItem, "name") {
				continue
			}
			if !strings.Contains(dataItem, "docker") {
				return false, errors.New(dataItem)
			}
		}
		return true, nil
	}

	r, e := ioutil.ReadFile("/proc/1/cgroup")
	if e != nil {
		return false, e
	}

	lines, linesErr := readLines(r)
	if linesErr != nil {
		return false, e
	}

	return markedDocker(lines)

}

func (c *cpuCollector) getCpuLimit() (float64, error) {
	const cfsQuotaUsPath = `/sys/fs/cgroup/cpu/cpu.cfs_quota_us`
	const cfsPeriodUsPath = `/sys/fs/cgroup/cpu/cpu.cfs_period_us`

	readCpuLimit := func(inCfsQuotaUs, inCfsPeriodUs string) (outCfsQuotaUs, outCfsPeriodUs int, err error) {
		cfsQuotaUs, cfsQuotaUsErr := ioutil.ReadFile(inCfsQuotaUs)
		if cfsQuotaUsErr != nil {
			return 0, 0, cfsQuotaUsErr
		}
		outCfsQuotaUs, err = strconv.Atoi(strings.Replace(string(cfsQuotaUs), "\n", "", -1))
		if err != nil {
			return outCfsQuotaUs, 0, err
		}

		cfsPeriodUs, cfsPeriodUsErr := ioutil.ReadFile(inCfsPeriodUs)
		if cfsPeriodUsErr != nil {
			return 0, 0, cfsPeriodUsErr
		}
		outCfsPeriodUs, err = strconv.Atoi(strings.Replace(string(cfsPeriodUs), "\n", "", -1))
		return outCfsQuotaUs, outCfsPeriodUs, err
	}

	outCfsQuotaUs, outCfsPeriodUs, err := readCpuLimit(cfsQuotaUsPath, cfsPeriodUsPath)
	if err != nil {
		return 0, err
	}
	return float64(outCfsQuotaUs) / float64(outCfsPeriodUs), nil
}

func (c *cpuCollector) getCpuUsage() (cpu float64, err error) {
	topRst, topErr := exec.Command("top", "-p 1", "-b", "-n 1").Output()
	if topErr != nil {
		return 0, topErr
	}
	r := bufio.NewReader(strings.NewReader(string(topRst)))
	pidTrue := false
	for {
		c, _, e := r.ReadLine()
		if e == io.EOF {
			break
		}
		if e != nil && e != io.EOF {
			return 0, e
		}
		if strings.Contains(string(c), "PID") {
			pidTrue = true
			continue
		}
		if pidTrue {
			cpu, _, err = getCpuAndMem(string(c))
			return cpu, err
		}
	}
	return
}

type memCollector struct {
}

func newMemCollector() (*memCollector, error) {
	return &memCollector{}, nil
}
func (m *memCollector) getMemUsage() (mem float64, err error) {
	topRst, topErr := exec.Command("top", "-p 1", "-b", "-n 1").Output()
	if topErr != nil {
		return 0, topErr
	}
	r := bufio.NewReader(strings.NewReader(string(topRst)))
	pidTrue := false
	for {
		c, _, e := r.ReadLine()
		if e == io.EOF {
			break
		}
		if e != nil && e != io.EOF {
			return 0, e
		}
		if strings.Contains(string(c), "PID") {
			pidTrue = true
			continue
		}
		if pidTrue {
			_, mem, err = getCpuAndMem(string(c))
			return mem, err
		}
	}
	return
}

type HardWareCollector struct {
	*cpuCollector
	*memCollector
}

func newHardWareCollector() (*HardWareCollector, error) {
	c := &HardWareCollector{}
	err := c.init()
	return c, err
}
func (h *HardWareCollector) init() (err error) {
	h.cpuCollector, err = newCpuCollector()
	if err != nil {
		return err
	}

	h.memCollector, err = newMemCollector()

	return err
}
func HardWareCollectorExample() {
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
