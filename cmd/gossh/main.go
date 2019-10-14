package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/golang/glog"
	expect "github.com/google/goexpect"
	"golang.org/x/crypto/ssh"
)

const (
	timeout = 10 * time.Second
)

// http://networkbit.ch/golang-ssh-client/
func main() {
	host := ""
	port := "22"
	user := ""
	pass := ""
	cmd1 := "pwd"
	cmd2 := "whoami"
	promptRE := regexp.MustCompile(`\$`)

	// get host public key
	hostKey := getHostKey(host)

	sshClt, err := ssh.Dial("tcp", host+":"+port, &ssh.ClientConfig{
		User: user,
		Auth: []ssh.AuthMethod{ssh.Password(pass)},
		// allow any host key to be used (non-prod)
		//HostKeyCallback: ssh.InsecureIgnoreHostKey(),

		// verify host public key
		HostKeyCallback: ssh.FixedHostKey(hostKey),
	})
	if err != nil {
		glog.Exitf("ssh.Dial(%q) failed: %v", host, err)
	}
	defer sshClt.Close()

	e, _, err := expect.SpawnSSH(sshClt, timeout)
	if err != nil {
		glog.Exit(err)
	}
	defer e.Close()

	e.Expect(promptRE, timeout)
	e.Send(cmd1 + "\n")
	result1, _, _ := e.Expect(promptRE, timeout)
	e.Send(cmd2 + "\n")
	result2, _, _ := e.Expect(promptRE, timeout)
	e.Send("exit\n")

	fmt.Printf("%s: result:\n %s\n\n", cmd1, result1)
	fmt.Printf("%s: result:\n %s\n\n", cmd2, result2)
}

func getHostKey(host string) ssh.PublicKey {
	// parse OpenSSH known_hosts file, ssh or use ssh-keyscan to pull key
	home := os.Getenv("HOME")
	file, err := os.Open(filepath.Join(home, ".ssh", "known_hosts"))
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var hostKey ssh.PublicKey
	for scanner.Scan() {
		fields := strings.Split(scanner.Text(), " ")
		if len(fields) != 3 || !strings.Contains(fields[0], host) {
			continue
		}

		var err error
		hostKey, _, _, _, err = ssh.ParseAuthorizedKey(scanner.Bytes())
		if err != nil {
			log.Fatalf("error parsing %q: %v", fields[2], err)
		}
		break
	}

	if hostKey == nil {
		log.Fatalf("no hostkey found for %s", host)
	}

	return hostKey
}
