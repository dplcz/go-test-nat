package main

import (
	"go-test-nat/stun"
	"log"
)

func main() {
	natType, externalIp, externalPort, err := stun.GetIpInfo()
	if err != nil {
		log.Fatal(err)
		return
	}
	log.Println("natType:", natType)
	log.Println("externalIp:", externalIp)
	log.Println("externalPort:", externalPort)
}
