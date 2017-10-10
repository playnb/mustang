package until

import (
	"strconv"
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
