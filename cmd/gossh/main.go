package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/bingoohuang/gossh"
	"github.com/bingoohuang/gossh/cnf"
	"github.com/bingoohuang/gossh/pbe"
	"github.com/bingoohuang/gossh/scp"

	"github.com/gobars/cmd"
	"github.com/golang/glog"
	expect "github.com/google/goexpect"
	homedir "github.com/mitchellh/go-homedir"
)

const (
	timeout = 10 * time.Second
)

func main() {
	pbe.DeclarePflags()
	defer pbe.DealPflag()

	cnf.DeclarePflags()
	cnf.DeclarePflagsByStruct(gossh.Config{})
	cnf.ParsePflags("GOSSH")

	var config gossh.Config
	cnf.LoadByPflag(&config)

	fmt.Printf("%+v", config)
}

func cmdtest() {
	x := "cd ~/GitHub/docker-compose-mysql-master-master/tool/mci; env GOOS=linux GOARCH=amd64 go install ./..."
	home, _ := homedir.Dir()
	x = strings.ReplaceAll(x, " ~", " "+home)

	p := cmd.NewCmdOptions(cmd.Options{Buffered: true, Streaming: true}, "/bin/bash", "-c", x)
	status := p.Start()

FOR:
	for {
		select {
		case so := <-p.Stdout:
			fmt.Println(so)
		case se := <-p.Stderr:
			_, _ = fmt.Fprintln(os.Stderr, se)
		case exitState := <-status:
			fmt.Println("exit status ", exitState.Exit)
			break FOR
		}
	}
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
	promptRE := regexp.MustCompile(`[#$]`)

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

	cmdx := "pwd"
	fmt.Println(cmdx)

	_ = ge.Send(cmdx + "\n")
	result2, _, _ := ge.Expect(promptRE, timeout)
	fmt.Print(result2)

	fmt.Println("whoami")

	_ = ge.Send(("whoami") + "\n")
	result3, _, _ := ge.Expect(promptRE, timeout)
	fmt.Print(result3)
	fmt.Println("exit")

	_ = ge.Send("exit\n")
}
