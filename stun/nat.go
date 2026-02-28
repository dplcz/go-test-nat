package stun

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	mRand "math/rand/v2"
	"net"
	"strconv"
	"strings"
	"time"
)

type retVal struct {
	resp         bool
	externalIp   string
	externalPort int
	sourceIp     string
	sourcePort   int
	changedIp    string
	changedPort  int
}

const letters = "0123456789ABCDEF"
const (
	Blocked              = "Blocked"
	OpenInternet         = "Open Internet"
	FullCone             = "Full Cone"
	SymmetricUDPFirewall = "Symmetric UDP Firewall"
	RestrictNAT          = "Restrict NAT"
	RestrictPortNAT      = "Restrict Port NAT"
	SymmetricNAT         = "Symmetric NAT"
	ChangedAddressError  = "Meet an error, when do Test1 on Changed IP and Port"

	MappedAddress    = "0001"
	ResponseAddress  = "0002"
	ChangeRequest    = "0003"
	SourceAddress    = "0004"
	ChangedAddress   = "0005"
	Username         = "0006"
	Password         = "0007"
	MessageIntegrity = "0008"
	ErrorCode        = "0009"
	UnknownAttribute = "000A"
	ReflectedFrom    = "000B"
	XorOnly          = "0021"
	XorMappedAddress = "8020"
	ServerName       = "8022"
	SecondaryAddress = "8050" // Non standard extension

	BindRequestMsg               = "0001"
	BindResponseMsg              = "0101"
	BindErrorResponseMsg         = "0111"
	SharedSecretRequestMsg       = "0002"
	SharedSecretResponseMsg      = "0102"
	SharedSecretErrorResponseMsg = "0112"
)

var dictMsgTypeToVal map[string]string
var TimeOut = errors.New("timeout")

var stunServers = []string{
	"stun.voipbuster.com",
	"stun.ekiga.net",
	"stun.ideasip.com",
	"stun.voiparound.com",
	"stun.voipstunt.com",
	"stun.voxgratia.org",
}

func init() {
	dictMsgTypeToVal = make(map[string]string)
	dictMsgTypeToVal[BindRequestMsg] = "BindRequestMsg"
	dictMsgTypeToVal[BindResponseMsg] = "BindResponseMsg"
	dictMsgTypeToVal[BindErrorResponseMsg] = "BindErrorResponseMsg"
	dictMsgTypeToVal[SharedSecretRequestMsg] = "SharedSecretRequestMsg"
	dictMsgTypeToVal[SharedSecretResponseMsg] = "SharedSecretResponseMsg"
	dictMsgTypeToVal[SharedSecretErrorResponseMsg] = "SharedSecretErrorResponseMsg"
}

func getTranId() (string, error) {
	id := make([]byte, 32)
	if _, err := rand.Read(id); err != nil {
		return "", err
	}
	for i, b := range id {
		id[i] = letters[b%byte(len(letters))]
	}
	return string(id), nil
}

func stunTest(conn *net.UDPConn, host string, port int, sendData string) (retVal, error) {
	var i int
	var (
		msgType     string
		bindRespMsg bool
		tranIdMatch bool
	)
	tempVal := retVal{}
	strLen := "0000"
	// get a random session id
	tranId, err := getTranId()
	if err != nil {
		tempVal.resp = false
		return tempVal, err
	}
	// packet hex data
	strData := fmt.Sprintf("%s%s%s%s", BindRequestMsg, strLen, tranId, sendData)
	bytesData, err := hex.DecodeString(strData)
	if err != nil {
		return tempVal, err
	}
	// send data to stun server
	destAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("%s:%d", host, port))
	_, err = conn.WriteToUDP(bytesData, destAddr)
	if err != nil {
		tempVal.resp = false
		return tempVal, err
	}
	buf := make([]byte, 2048)
	// set socket timeout
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	for i = 0; i < 3; i++ {
		_, _, err = conn.ReadFromUDP(buf)
		if err, ok := err.(net.Error); ok && err.Timeout() {
			tempVal.resp = false
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))
			continue
		}
		// parse response data
		msgType = hex.EncodeToString(buf[:2])
		bindRespMsg = dictMsgTypeToVal[msgType] == "BindResponseMsg"
		tranIdMatch = strings.ToUpper(hex.EncodeToString(buf[4:20])) == tranId
		break
	}
	if i == 3 {
		tempVal.resp = false
		return tempVal, TimeOut
	}
	if bindRespMsg && tranIdMatch {
		tempVal.resp = true
		lenMessage, _ := strconv.ParseInt(hex.EncodeToString(buf[2:4]), 16, 32)
		lenRemain := lenMessage
		base := 20
		for {
			if lenRemain <= 0 {
				break
			}
			attrType := hex.EncodeToString(buf[base : base+2])
			attrLen, _ := strconv.ParseInt(hex.EncodeToString(buf[base+2:base+4]), 16, 32)
			newPort, _ := strconv.ParseInt(hex.EncodeToString(buf[base+6:base+8]), 16, 32)
			var tempIp []string
			temp, _ := strconv.ParseInt(hex.EncodeToString(buf[base+8:base+9]), 16, 32)
			tempIp = append(tempIp, fmt.Sprintf("%d", temp))
			temp, _ = strconv.ParseInt(hex.EncodeToString(buf[base+9:base+10]), 16, 32)
			tempIp = append(tempIp, fmt.Sprintf("%d", temp))
			temp, _ = strconv.ParseInt(hex.EncodeToString(buf[base+10:base+11]), 16, 32)
			tempIp = append(tempIp, fmt.Sprintf("%d", temp))
			temp, _ = strconv.ParseInt(hex.EncodeToString(buf[base+11:base+12]), 16, 32)
			tempIp = append(tempIp, fmt.Sprintf("%d", temp))
			newIp := strings.Join(tempIp, ".")

			switch attrType {
			case MappedAddress:
				tempVal.externalIp = newIp
				tempVal.externalPort = int(newPort)
			case SourceAddress:
				tempVal.sourceIp = newIp
				tempVal.sourcePort = int(newPort)
			case ChangedAddress:
				tempVal.changedIp = newIp
				tempVal.changedPort = int(newPort)
			}
			base = base + 4 + int(attrLen)
			lenRemain = lenRemain - (4 + attrLen)
		}

	}
	return tempVal, nil
}

func getStunServer(conn *net.UDPConn) string {
	mRand.Shuffle(len(stunServers), func(i, j int) {
		stunServers[i], stunServers[j] = stunServers[j], stunServers[i]
	})
	for _, stunServer := range stunServers {
		_, err := stunTest(conn, stunServer, 3478, "")
		if err != nil {
			continue
		} else {
			return stunServer
		}
	}
	return ""
}

func getNatType(conn *net.UDPConn, sourceIp string, sourcePort int) (string, retVal, error) {
	var ret retVal
	var err error
	var server string
	server = getStunServer(conn)
	log.Println("Do Test1")
	resp := false

	ret, err = stunTest(conn, server, 3478, "")
	if err != nil {
		return "", ret, err
	}
	resp = ret.resp

	if !resp {
		return Blocked, ret, nil
	}
	log.Println("Result: ", ret)
	exIp := ret.externalIp
	exPort := ret.externalPort
	changedIp := ret.changedIp
	changedPort := ret.changedPort
	if ret.externalIp == sourceIp {
		temp := []string{ChangeRequest, SourceAddress, "00000006"}
		changedReq := strings.Join(temp, "")
		ret, err = stunTest(conn, server, 3478, changedReq)
		if err != nil {
			return "", ret, err
		}
		log.Println("Result: ", ret)
	} else {
		temp := []string{ChangeRequest, SourceAddress, "00000006"}
		changedReq := strings.Join(temp, "")
		log.Println("Do Test2")
		ret, err = stunTest(conn, server, 3478, changedReq)
		if err != nil && !errors.Is(err, TimeOut) {
			return "", ret, err
		}
		if ret.resp {
			return FullCone, ret, nil
		} else {
			log.Println("Do Test1")
			ret, err = stunTest(conn, changedIp, changedPort, "")
			if err != nil {
				return "", ret, err
			}
			log.Println("Result: ", ret)
			if !ret.resp {
				return ChangedAddressError, ret, nil
			} else {
				if exIp == ret.externalIp && exPort == ret.externalPort {
					temp = []string{ChangeRequest, SourceAddress, "00000002"}
					changedPortReq := strings.Join(temp, "")
					log.Println("Do Test3")
					ret, err = stunTest(conn, changedIp, 3478, changedPortReq)
					log.Println("Result: ", ret)
					if ret.resp {
						return RestrictNAT, ret, nil
					} else {
						return RestrictPortNAT, ret, nil
					}
				} else {
					return SymmetricNAT, ret, nil
				}
			}
		}

	}
	return "", ret, nil
}

func GetIpInfo() (string, string, int, error) {
	serverAddr, err := net.ResolveUDPAddr("udp", ":")
	if err != nil {
		log.Fatal(err)
		return "", "", 0, err
	}
	conn, err := net.ListenUDP("udp", serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	natType, ret, err := getNatType(conn, string(serverAddr.IP), serverAddr.Port)
	if err != nil {
		log.Fatal(err)
		return "", "", 0, err
	}
	return natType, ret.externalIp, ret.externalPort, nil
}
