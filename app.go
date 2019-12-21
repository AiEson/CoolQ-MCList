package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"
	"unicode/utf8"

	"github.com/Tnze/CoolQ-Golang-SDK/cqp"
	"github.com/Tnze/CoolQ-Golang-SDK/cqp/util"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/chat"
	"github.com/google/uuid"
	"github.com/miaoscraft/SiS/data"
)

func PathExists(path string) (bool, error) {
	/*
	   判断文件或文件夹是否存在
	   如果返回的错误为nil,说明文件或文件夹存在
	   如果返回的错误类型使用os.IsNotExist()判断为true,说明文件或文件夹不存在
	   如果返回的错误为其它类型,则不确定是否在存在
	*/
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func Ping(args []string, ret func(msg string, Numb int)) bool {
	var (
		resp  []byte
		delay time.Duration
		err   error
	)
	addr, port := getAddr(args)

	//SRV解析
	if _, SRV, err := net.LookupSRV("minecraft", "tcp", addr); len(SRV) != 0 && err == nil {
		addr = SRV[0].Target
		port = int(SRV[0].Port)
	}

	if d := data.Config.Ping.Timeout.Duration; d > 0 {
		//启用Timeout
		resp, delay, err = bot.PingAndListTimeout(addr, port, d)
	} else {
		//禁用Timeout
		resp, delay, err = bot.PingAndList(addr, port)
	}
	if err != nil {
		ret(fmt.Sprintf("嘶...请求失败惹！: %v", err), 0)
		return true
	}

	var s status
	err = json.Unmarshal(resp, &s)
	if err != nil {
		ret(fmt.Sprintf("嘶...解码失败惹！: %v", err), 0)
		return true
	}
	online := s.Players.Online
	// 延迟用手动填进去
	s.Delay = delay

	ret(s.String(), online)
	return true
}

// 从[]string获取服务器地址和端口
// 支持的格式有:
// 	[ "ping" "play.miaoscraft.cn" ]
// 	[ "ping" "play.miaoscraft.cn:25565" ]
// 	[ "ping" "play.miaoscraft.cn" "25565" ]
func getAddr(args []string) (addr string, port int) {
	args = args[1:] //去除第一个元素"ping"
	// 默认值
	addr = data.Config.Ping.DefaultServer
	port = 25565

	// 在第二个参数内寻找端口
	if len(args) >= 2 {
		if p, err := strconv.Atoi(args[1]); err == nil {
			port = p
		}
	}

	// 如果有则加载第一个参数
	if len(args) >= 1 {
		addr = args[0]
	}

	// 在冒号后面寻找端口
	f := strings.Split(addr, ":")
	if len(f) >= 2 {
		if p, err := strconv.Atoi(f[1]); err == nil {
			port = p
		}
	}

	// 冒号前面是地址
	addr = f[0]

	return addr, port
}

type status struct {
	Description chat.Message
	Players     struct {
		Max    int
		Online int
		Sample []struct {
			ID   uuid.UUID
			Name string
		}
	}
	Version struct {
		Name     string
		Protocol int
	}
	//favicon ignored

	Delay time.Duration `json:"-"`
}

var tmp = template.Must(template.
	New("PingRet").
	Funcs(CQCodeUtil).
	Parse(`({{ .Players.Online -}}) {{ range .Players.Sample }}{{ .Name}}，{{ end }}`))

var CQCodeUtil = template.FuncMap{
	"escape": util.Escape,
}

func (s status) String() string {
	var sb strings.Builder
	err := tmp.Execute(&sb, s)
	if err != nil {
		return fmt.Sprintf("似乎在渲染文字模版时出现了棘手的问题: %v", err)
	}
	return sb.String()
}

// go:generate cqcfg -c .
// cqp: 名称: McList
// cqp: 版本: 1.0.0:1
// cqp: 作者: AiEson
// cqp: 简介: 一个超棒的Go语言MCLIST插件，它会MCPing多个服务器~
func main() { /*此处应当留空*/ }

func init() {
	cqp.AppID = "moe.10935336.mclist" // TODO: 修改为这个插件的ID
	cqp.PrivateMsg = onPrivateMsg
	cqp.GroupMsg = onGroupMsg
	cqp.Start = onStart
}

func onStart() int32 {
	confDir := cqp.GetAppDir() + "config.json"
	boo, err := PathExists(confDir)
	if boo == false && err == nil {
		file, err := os.Create(confDir)
		if err != nil {
			os.Exit(1)
		}
		fileString := "{}"
		file.WriteString(fileString)
		file.Close()
	}
	confDir2 := cqp.GetAppDir() + "使用说明.txt"
	boo2, err2 := PathExists(confDir2)
	if boo2 == false && err2 == nil {
		file, err := os.Create(confDir2)
		if err != nil {
			os.Exit(1)
		}
		fileString := `MCList使用说明：
请遵循以下格式进行config.json文件的书写
{
	"des":"欢迎来到XXX服务器",
	"servers":[
		{
			"port":"端口",
			"ip":"IP1",
			"name":"简称"
		},

		{
			"port":"端口",
			"ip":"IP2",
			"name":"简称"
		}
	]
}
		
以此类推，可以加入任意数量服务器
*config.json文件实时读取，修改完毕可直接使用，无需重启插件
*config.json文件中每行字符数量不超过65536个
*没了`
		file.WriteString(fileString)
		file.Close()
	}
	return 0
}

func onPrivateMsg(subType, msgID int32, fromQQ int64, msg string, font int32) int32 {
	cqp.SendPrivateMsg(fromQQ, msg) //复读机
	return 0
}

type Serverslice struct {
	Servers []struct {
		IP   string `json:"ip"`
		Port string `json:"port"`
		Name string `json:"name"`
	} `json:"servers"`
	Des string `json:"des"`
}

func substr(str string, start int, end int) string {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		return ""
	}

	if end < 0 || end > length {
		return ""
	}
	return string(rs[start:end])
}

func onGroupMsg(subType, msgID int32, fromGroup, fromQQ int64, fromAnonymous, msg string, font int32) int32 {
	if msg == "/list" {
		cqp.SendGroupMsg(fromGroup, "正在查询...")
		var out, ing string
		confDir := cqp.GetAppDir() + "config.json"
		str, err := ioutil.ReadFile(confDir)
		var slist Serverslice
		var des string
		allNum := 0
		if err == nil {
			json.Unmarshal(str, &slist)
			des = slist.Des
			for key, val := range slist.Servers {
				fmt.Print(key)
				ip := val.IP
				port := val.Port
				name := val.Name
				info := []string{"ping", ip + ":" + port}
				Ping(info, func(msg string, Numb int) {
					allNum += Numb
					msgi := ""
					if Numb != 0 {
						msgi = substr(msg, 0, utf8.RuneCountInString(msg)-2)
					}
					ing += name + msgi + "\n"
				})
				//println("Key：", key, "\tName：", val.ServerName, "\tIP：", val.ServerIP)
			}
		}
		out = des + "，在线人数：" + strconv.Itoa(allNum) + "\n" + ing
		// var out string
		// Ping([]string{"ping", "10935336.ooo:21002"}, func(msg string) {
		// 	out = msg
		// })
		if out == "，在线人数：0\n" {
			out = "请正确配置config.json文件再使用，详情请看data/app/moe.10935336.mclist/使用说明.txt"
		}
		cqp.SendGroupMsg(fromGroup, out)
	}
	return 0
}
