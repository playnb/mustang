package util

import (
	"strconv"
	"strings"
	//"flag"
	"fmt"
	//"io"
	"net"
	//"net/http"
	//"os"
)

func Get_IP(eth int) string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Oops:" + err.Error())
		return "127.0.0.1"
	}
	index := 0
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				if index == eth {
					return ipnet.IP.To4().String()
				}
				//fmt.Println(ipnet.IP.To4().String())
				index++
			}
		}
	}
	return "127.0.0.1"
}

func Show_IPs() {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println("Oops:" + err.Error())
	}
	index := 0
	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				fmt.Printf("Index:%d, IP:%s \n", index, ipnet.IP.To4().String())
				index++
			}
		}
	}
}

func GetRandomAddr(eth int, from int, to int) string {
	ip := Get_IP(eth)
	addr := ""
	port := from
	for {
		port++
		if port > to {
			return ""
		}
		addr = ip + ":" + strconv.Itoa(port)
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
		} else {
			return addr
		}
	}
}

func GetRandomPort(from int, to int) int {
	port := from
	ip := "127.0.0.1"
	for {
		port++
		if port > to {
			return 0
		}
		addr := ip + ":" + strconv.Itoa(port)
		conn, err := net.Dial("tcp", addr)
		if err == nil {
			conn.Close()
		} else {
			return port
		}
	}
}

func InetNtoA(ipnr int64) net.IP {
	var bytes [4]byte
	bytes[0] = byte(ipnr & 0xFF)
	bytes[1] = byte((ipnr >> 8) & 0xFF)
	bytes[2] = byte((ipnr >> 16) & 0xFF)
	bytes[3] = byte((ipnr >> 24) & 0xFF)

	return net.IPv4(bytes[3], bytes[2], bytes[1], bytes[0])
}

func InetAtoN(ipnr net.IP) int64 {
	bits := strings.Split(ipnr.String(), ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}

func InetAStrtoN(ipnr string) int64 {
	bits := strings.Split(ipnr, ".")

	b0, _ := strconv.Atoi(bits[0])
	b1, _ := strconv.Atoi(bits[1])
	b2, _ := strconv.Atoi(bits[2])
	b3, _ := strconv.Atoi(bits[3])

	var sum int64

	sum += int64(b0) << 24
	sum += int64(b1) << 16
	sum += int64(b2) << 8
	sum += int64(b3)

	return sum
}
