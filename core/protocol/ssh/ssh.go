package ssh

import (
	"github.com/gliderlabs/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"fmt"
	"io"
	"strings"
	"HFish/utils/is"
	"HFish/core/rpc/client"
	"HFish/core/report"
	"HFish/utils/log"
	"HFish/utils/conf"
	"HFish/utils/json"
	"github.com/bitly/go-simplejson"
	"HFish/utils/file"
)

func getJson() *simplejson.Json {
	res, err := json.Get("ssh")

	if err != nil {
		log.Pr("HFish", "127.0.0.1", "解析 SSH JSON 文件失败", err)
	}
	return res
}

func Start(addr string) {
	ssh.ListenAndServe(
		addr,
		func(s ssh.Session) {
			res := getJson()

			term := terminal.NewTerminal(s, res.Get("hostname").MustString())
			line := ""
			for {
				line, _ = term.ReadLine()
				if line == "exit" {
					break
				}

				fileName := res.Get("command").Get(line).MustString()

				if (fileName == "") {
					fileName = res.Get("command").Get("default").MustString()
				}

				output := file.ReadLibsText("ssh", fileName)

				fmt.Println(line)

				io.WriteString(s, output+"\n")
			}
		},
		ssh.PasswordAuth(func(s ssh.Context, password string) bool {
			info := s.User() + "&&" + password

			arr := strings.Split(s.RemoteAddr().String(), ":")

			log.Pr("SSH", arr[0], "已经连接")

			// 判断是否为 RPC 客户端
			if is.Rpc() {
				go client.ReportResult("SSH", "", arr[0], info, "0")
			} else {
				go report.ReportSSH(arr[0], "本机", info)
			}

			sshStatus := conf.Get("ssh", "status")

			if (sshStatus == "2") {
				// 高交互模式
				res := getJson()
				accountx := res.Get("account")
				passwordx := res.Get("password")

				if (accountx.MustString() == s.User() && passwordx.MustString() == password) {
					return true
				}
			}

			// 低交互模式，返回账号密码不正确
			return false
		}),
	)
}
