# go-test-nat

A Golang STUN client for getting NAT type and external IP refered to PyStun(https://github.com/jtriley/pystun)

Follow RFC 3489: http://www.ietf.org/rfc/rfc3489.txt

### Use the Command Line
Clone this repo and run these
```
go build
./go-test-nat
```
Output will be like
```
2024/11/22 00:34:15 natType: Symmetric NAT
2024/11/22 00:34:15 externalIp: 223.110.71.25
2024/11/22 00:34:15 externalPort: 34256
```
### Use the Library
`github.com/dplcz/go-test-nat` can be easily used like
```go
import "github.com/dplcz/go-test-nat"

func main() {
    natType, externalIp, externalPort, err := stun.GetIpInfo()
	...
}
```
### License
MIT License - see [LICENSE](LICENSE) for full text

