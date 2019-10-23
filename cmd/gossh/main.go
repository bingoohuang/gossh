package main

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/bingoohuang/gossh"
	"github.com/bingoohuang/gossh/scp"

	"github.com/golang/glog"
	expect "github.com/google/goexpect"
	"github.com/mitchellh/go-homedir"
)

const (
	timeout = 10 * time.Second
)

func main() {
	scptest()
	sshtest()
}

func scptest() {
	clientConfig, _ := gossh.PasswordKey("root", "bjca2019")
	client := scp.NewConf().CreateClient()

	if err := client.Connect("192.168.136.22:8022", clientConfig); err != nil {
		fmt.Println("Couldn't establish a connection to the remote server ", err)
		return
	}

	defer client.Close()

	fi, _ := homedir.Expand("~/go/bin/linux_amd64/sysinfo")
	f, _ := os.Open(fi)

	defer f.Close()

	stat, _ := os.Stat(fi)

	mod := fmt.Sprintf("0%o", stat.Mode())
	if err := client.CopyFile(f, "./sysinfo", mod); err != nil {
		fmt.Println("Error while copying file ", err)
	}
}

// http://networkbit.ch/golang-ssh-client/
func sshtest() {
	addr := "192.168.136.22:8022"
	promptRE := regexp.MustCompile(`#|\$`)

	clientConfig, _ := gossh.PasswordKey("root", "bjca2019")

	sshClt, err := gossh.DialTCP(addr, clientConfig)
	if err != nil {
		glog.Exitf("ssh.Dial(%q) failed: %v", addr, err)
	}

	defer sshClt.Close()

	ge, _, err := expect.SpawnSSH(sshClt, timeout)
	if err != nil {
		glog.Exit(err)
	}

	defer ge.Close()

	result1, _, _ := ge.Expect(promptRE, timeout)
	fmt.Print(result1)
	fmt.Println("pwd")

	_ = ge.Send("pwd" + "\n")
	result2, _, _ := ge.Expect(promptRE, timeout)
	fmt.Print(result2)
	fmt.Println("whoami")

	_ = ge.Send(("whoami") + "\n")
	result3, _, _ := ge.Expect(promptRE, timeout)
	fmt.Print(result3)
	fmt.Println("exit")

	_ = ge.Send("exit\n")
}
