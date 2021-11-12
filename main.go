package main

import (
	"encoding/json"
	"flag"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strings"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/regions"
	dnspod "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/dnspod/v20210323"
)

type Config struct {
	InterfaceName string `json:"interface_name"`
	SecretID      string `json:"secret_id"`
	SecretKey     string `json:"secret_key"`
	Domain        string `json:"domain"`
	SubDomain     string `json:"sub_domain"`
}

var file = flag.String("c", "ddns.json", "")

func main() {
	flag.Parse()
	content, err := ioutil.ReadFile(*file)
	if err != nil {
		log.Fatalln(err)
	}
	cfg := new(Config)
	if err := json.Unmarshal(content, cfg); err != nil {
		log.Fatalln(err)
	}

	ip := GetIPByInterfaceName(cfg.InterfaceName)

	if ok := cacheExist(ip); ok {
		log.Println("缓存存在，不用更新")
		return
	}

	credential := common.NewCredential(cfg.SecretID, cfg.SecretKey)
	client, err := dnspod.NewClient(credential, regions.Guangzhou, profile.NewClientProfile())
	if err != nil {
		log.Fatalln(err)
	}

	var recordId *uint64
	{
		request := dnspod.NewDescribeRecordListRequest()
		request.Domain = &cfg.Domain
		response, err := client.DescribeRecordList(request)

		if err != nil {
			log.Fatalf("An API error has returned: %s", err)
		}
		for _, r := range response.Response.RecordList {
			if *r.Name == cfg.SubDomain {
				recordId = r.RecordId
				break
			}
		}
	}
	{
		request := dnspod.NewModifyDynamicDNSRequest()
		request.Domain = &cfg.Domain
		request.SubDomain = &cfg.SubDomain
		request.RecordId = recordId
		str := "默认"
		request.RecordLine = &str
		request.Value = &ip
		_, err := client.ModifyDynamicDNS(request)
		if err != nil {
			log.Fatalf("An API error has returned: %s", err)
		}
		log.Println("更新成功")
	}
}

func GetIPByInterfaceName(name string) string {
	wlo1, err := net.InterfaceByName(name)
	if err != nil {
		panic(err)
	}
	addrs, err := wlo1.Addrs()
	if err != nil {
		panic(err)
	}
	ipv4 := addrs[0].String()
	p := strings.Index(ipv4, "/")
	if p == -1 {
		panic(ipv4)
	}
	return ipv4[:p]
}

func cacheExist(ip string) bool {
	f, err := os.OpenFile("/tmp/ddns.cache", os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		log.Fatalln(err)
	}
	old, err := io.ReadAll(f)
	if err != nil {
		log.Fatalln(err)
	}
	if ip == string(old) {
		return true
	}
	if err := f.Truncate(0); err != nil {
		log.Fatalln(err)
	}
	if _, err := f.WriteString(ip); err != nil {
		log.Fatalln(err)
	}

	return false
}
