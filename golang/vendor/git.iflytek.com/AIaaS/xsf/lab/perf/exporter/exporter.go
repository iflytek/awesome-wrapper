package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

var server = flag.String("srv", "server", "server")
var port = flag.Int("port", 1999, "port")
var dbg = flag.Int("dbg", 0, "debug")

func main() {
	flag.Parse()
	pids := getPid(*server)
	getMetrics := func() string {
		w := bytes.Buffer{}
		for _, pid := range pids {
			cpu, mem, err := GetCpuUsageWithExec(pid)
			if nil != err {
				panic(err)
			}
			w.WriteString(fmt.Sprintf("%v_cpu{pid=\"%v\"} %v\n", *server, pid, cpu))
			w.WriteString(fmt.Sprintf("%v_mem{pid=\"%v\"} %v\n", *server, pid, mem))
		}
		return w.String()
	}

	http.HandleFunc("/metrics", func(writer http.ResponseWriter, request *http.Request) {
		_, _ = fmt.Fprint(writer, getMetrics())
	})

	fmt.Println("metrics listen :", *port)
	if err := http.ListenAndServe(fmt.Sprintf(":%v", *port), nil); err != nil {
		log.Fatalf("can't listen ,err:%v\n", err)
	}
}
func getPid(srv string) []int {
	pidData, pidErr := exec.Command("pidof", srv).Output()
	if nil != pidErr {
		panic(pidErr)
	}

	fmt.Printf("server:%v,pids:%v", srv, string(pidData))

	var res []int
	for _, pid := range strings.Split(strings.Replace(string(pidData), "\n", "", -1), " ") {
		r, e := strconv.Atoi(pid)
		if nil != e {
			panic(e)
		}
		res = append(res, r)
	}
	return res
}
func GetCpuUsageWithExec(pid int) (cpu float64, mem float64, err error) {
	getCpuAndMem := func(topData string) (cpu, mem float64, err error) {
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
		if nil != err {
			panic(err)
		}
		return cpu, mem, err
	}
	currentPid := pid
	topRst, topErr := exec.Command("top", "-p "+strconv.Itoa(currentPid), "-b", "-n 1").Output()
	if topErr != nil {
		return 0, 0, topErr
	}
	r := bufio.NewReader(strings.NewReader(string(topRst)))
	pidTrue := false
	for {
		c, _, e := r.ReadLine()
		if e == io.EOF {
			break
		}
		if e != nil && e != io.EOF {
			return 0, 0, e
		}
		if strings.Contains(string(c), "PID") {
			pidTrue = true
			continue
		}
		if pidTrue {
			if *dbg != 0 {
				fmt.Printf("rawData:%v\n", string(c))
			}
			cpu, mem, err = getCpuAndMem(string(c))
			return cpu, mem, err
		}
	}
	return
}
