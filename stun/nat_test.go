package stun

import (
	"fmt"
	"net"
	"testing"
)

func TestGetTranId(t *testing.T) {
	t.Log(getTranId())
}

func TestStunTest(t *testing.T) {

	serverAddr, err := net.ResolveUDPAddr("udp", ":9999")
	if err != nil {
		t.Error(err)
	}
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Error(err)
	}
	defer conn.Close()
	//for _, server := range stunServers {
	//	_, err := stunTest(conn, server, 3478)
	//	if err != nil {
	//		fmt.Println(err)
	//	}
	//}
}

func TestGetNatType(t *testing.T) {
	serverAddr, err := net.ResolveUDPAddr("udp", ":9999")
	if err != nil {
		t.Error(err)
	}
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		t.Error(err)
	}
	getNatType(conn, "0.0.0.0", 9999)
}

func TestUdp(t *testing.T) {
	// 本地地址和端口
	localAddr := "localhost:9999"

	// 解析本地UDP地址
	udpAddr, err := net.ResolveUDPAddr("udp", localAddr)
	if err != nil {
		fmt.Printf("Failed to resolve UDP address: %v\n", err)
		return
	}

	// 监听本地端口
	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Printf("Failed to listen UDP: %v\n", err)
		return
	}
	defer conn.Close()

	fmt.Printf("UDP server listening on %s\n", localAddr)

	// 目标服务的地址和端口
	remoteAddr := "77.72.169.210:3478"
	remoteUDPAddr, err := net.ResolveUDPAddr("udp", remoteAddr)
	if err != nil {
		fmt.Printf("Failed to resolve remote UDP address: %v\n", err)
		return
	}

	// 要发送的数据
	message := []byte("Hello UDP Server")

	// 直接向目标地址发送数据
	_, err = conn.WriteToUDP(message, remoteUDPAddr)
	if err != nil {
		fmt.Printf("Failed to write to UDP: %v\n", err)
		return
	}

	fmt.Println("Data sent successfully")
}

func TestAll(t *testing.T) {
	t.Log(GetIpInfo())
}
