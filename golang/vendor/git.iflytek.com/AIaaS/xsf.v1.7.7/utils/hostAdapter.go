package utils

import (
	"fmt"
	"net"
	"os"
	"runtime"
)

// HostAdapter 如果 host 为 32 位 ipv4 地址，直接返回
// 如果 host 为域名，则解析 host 对应的 ipv4
// 如果 host 为空，则解析网卡 netCard
// 如果两者均为空，则根据本机 hostname
// 需要判断其是否为 ipv4 后再返回
func HostAdapter(host string, netCard string) (string, error) {
	if IsIpv4(host) {
		return host, nil
	}

	return Host2Ip(host, netCard)
}

func Host2Ip(host string, netCard string) (string, error) {
	/*
		1、如果 host 存在，则取 host 对应的 ipv4
	*/
	if host != "" {
		addrs, err := net.LookupHost(host)
		if err != nil {
			return "", err
		}

		for _, addr := range addrs {
			if IsIpv4(addr) {
				return addr, nil
			}
		}

		return "", fmt.Errorf("can't convert host -> %v to ipv4", host)
	}

	/*
		2、如果 host不存在，netCard 存在，则取 netCard 对应的 ipv4
	*/
	if netCard != "" {
		ipMap, err := GetAddrs()
		if err != nil {
			return "", err
		}

		if ip, ok := ipMap[netCard]; ok && IsIpv4(ip) {
			return ip, nil
		}

		return "", fmt.Errorf("can't convert net card -> %v to ipv4", netCard)
	}

	/*
		3、如果是 windows，走 dns 取正在用的 ipv4
	*/
	if runtime.GOOS == "windows" {
		ip, err := getWinIp()

		if err != nil {
			return "", err
		}

		if IsIpv4(ip) {
			return ip, nil
		}

		return "", fmt.Errorf("can't get ipv4 from dns in windows os")
	}

	/*
		4、如果 host 和 netCard 都没传，则
			a、取本机 hostname
			b、调用 LookupHost 查找 hostname 对应的 ipv4
	*/
	hostname, err := os.Hostname()
	if err != nil {
		return "", err
	}

	addrs, err := net.LookupHost(hostname)

	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		if IsIpv4(addr) {
			return addr, nil
		}
	}

	return "", fmt.Errorf("can't convert hostname -> %v to ipv4", hostname)
}

// IsIpv4 检查该 ip 是否为合法的 ipv4 地址
func IsIpv4(ip string) bool {
	trial := net.ParseIP(ip)
	if trial.To4() == nil {
		return false
	}
	return true
}

func getWinIp() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "", err
	}
	if conn == nil {
		return "", fmt.Errorf("nil conn")
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP.String(), nil
}
