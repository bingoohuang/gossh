# gossh

[![Travis CI](https://img.shields.io/travis/bingoohuang/gossh/master.svg?style=flat-square)](https://travis-ci.com/bingoohuang/gossh)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/bingoohuang/gossh/blob/master/LICENSE.md)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/bingoohuang/gossh)
[![Coverage Status](http://codecov.io/github/bingoohuang/gossh/coverage.svg?branch=master)](http://codecov.io/github/bingoohuang/gossh?branch=master)
[![goreport](https://www.goreportcard.com/badge/github.com/bingoohuang/gossh)](https://www.goreportcard.com/report/github.com/bingoohuang/gossh)

execute shell scripts among multiple ssh servers

## Usage demo

```bash
$ gossh --quoteReplace=%q --bangReplace=%b --hosts="192.168.1.1:8022 app/app" --cmds="%host MYSQL_PWD='%babcdefg' mysql -h127.0.0.1 -uroot -e %qshow variables like 'server%'%q"

--- 192.168.1.1:8022 ---
$ MYSQL_PWD='!abcdefg' mysql -h127.0.0.1 -uroot -e "show variables like 'server%'"
+----------------+--------------------------------------+
| Variable_name  | Value                                |
+----------------+--------------------------------------+
| server_id      | 1                                    |
| server_id_bits | 32                                   |
| server_uuid    | 43e9cbe5-b38a-11e9-8570-04d4c439354e |
+----------------+--------------------------------------+
```

```bash
gossh -h="192.168.1.(9 18):8022 app/app id=(9 18)" --cmds="%host-9 MYSQL_PWD='\!abcdefg' mysql -u root -h 127.0.0.1 -vvv -e 'show slave status\G'"
gossh -h="192.168.1.9:8022 app/app id=9, 192.168.1.18:8022 app/app id=18" --cmds="%host-9 %ul ~/go/bin/linux_amd64/mci ./mci,%host-9 ./mci/mci -v"
gossh -h="192.168.1.9:8022 app/app id=9, 192.168.1.18:8022 app/app id=18" --cmds="%host-9 %dl ./mci/mci ."
```

proxy supported:

```bash
gossh --hosts="192.168.1.3:6022 huangjinbing/123 id=1, 192.168.9.1:22 user proxy=1" --cmds="%host-2 %dl 1.log 10.log"
gossh --hosts="192.168.1.3:6022 huangjinbing/123 id=1, 192.168.9.1:22 user proxy=1" --cmds="%host-2 cat 1.log"
```

```bash
gossh --hostsFile ~/hosts.txt --cmdsFile ~/cmds.txt --user root --pass "{PBE}H3y5VaKfj-vxSJ5JUHL0R-CBtZTkR2UR"
```

```text
# hosts.txt
13.26.15.12:(1061-1063)
13.26.15.13:222
13.26.15.14
```

```text
# cmds.txt
%host pwd
%host hostname -I
```

## Substitute ResultVars

1. define result variables like `... => @varName`
1. use the variables like `echo @varName`

```toml
#printConfig = false
#passphrase="xxxx"

hosts = [
"12.26.85.62:1082 root/111",
"12.26.85.62:1083 root/222",
"12.26.85.62:1084 root/333",
]

# 全部命令都默认成远程执行，相当于自动添加了%host标识。
globalRemote = true
cmdTimeout = "15s"

cmds = [
    "date '+%Y%m%d' => @today",
    "sh /tmp/hostdailycheck.sh",
    "%dl /tmp/log/HostDailyCheck-*-@today.txt ./dailychecks@today/",
]
```

```bash
$  uuidgen
6F948925-429E-4D2C-B551-C9C6D12E5062
$  gossh --pbe hello,word -p C9C6D12E5062
+---+-------+-----------------------------+
| # | PLAIN | ENCRYPTED                   |
+---+-------+-----------------------------+
| 1 | hello | {PBE}eiRMlsZPLikVYpZMcHicyg |
| 2 | word  | {PBE}lAHH0UfuqZ0YtV_5VE77uw |
+---+-------+-----------------------------+
$  gossh --pbe hello,word -p C9C6D12E5062
+---+-------+-----------------------------+
| # | PLAIN | ENCRYPTED                   |
+---+-------+-----------------------------+
| 1 | hello | {PBE}6RGab13x5WfzFP0NpA_suA |
| 2 | word  | {PBE}qmPJAysHSmnfQEK-a6JM0A |
+---+-------+-----------------------------+

$  gossh --ebp 6RGab13x5WfzFP0NpA_suA,qmPJAysHSmnfQEK-a6JM0A -p C9C6D12E5062
+---+------------------------+-------+
| # | ENCRYPTED              | PLAIN |
+---+------------------------+-------+
| 1 | 6RGab13x5WfzFP0NpA_suA | hello |
| 2 | qmPJAysHSmnfQEK-a6JM0A | word  |
+---+------------------------+-------+
$  gossh --ebp {PBE}eiRMlsZPLikVYpZMcHicyg,{PBE}lAHH0UfuqZ0YtV_5VE77uw -p C9C6D12E5062
+---+-----------------------------+-------+
| # | ENCRYPTED                   | PLAIN |
+---+-----------------------------+-------+
| 1 | {PBE}eiRMlsZPLikVYpZMcHicyg | hello |
| 2 | {PBE}lAHH0UfuqZ0YtV_5VE77uw | word  |
+---+-----------------------------+-------+
```

## resources

1. [Bootstrap](https://getbootstrap.com/)
1. [Bootstrap 4 Password Show Hide](https://codepen.io/Qanser/pen/dVRGJv)
1. [Tables](https://getbootstrap.com/docs/4.3/content/tables/)
1. [Golang SSH Client: Multiple Commands, Crypto & Goexpect Examples](http://networkbit.ch/golang-ssh-client/)
1. [bramvdbogaerde/go-scp](https://github.com/bramvdbogaerde/go-scp)
1. [golang 批量scp 远程传输文件](https://www.jianshu.com/p/f9d6dfefb63d)
1. [PBEWithMD5AndDES in go](https://github.com/LucasSloan/passwordbasedencryption)
1. [like python-sh, for easy call shell with golang](https://github.com/codeskyblue/go-sh)
1. [A scp client library written in Go. The remote server must have the scp command](https://github.com/hnakamur/go-scp)
1. [How the SCP protocol works](https://chuacw.ath.cx/blogs/chuacw/archive/2019/02/04/how-the-scp-protocol-works.aspx)
1. [Golang SFTP Client: Download File, Upload File Example](http://networkbit.ch/golang-sftp-client/)
1. [package sftp](https://godoc.org/github.com/pkg/sftp)
1. [sftp/example_test.go](https://github.com/pkg/sftp/blob/master/example_test.go)
1. [Golang Client Examples](https://golang.hotexamples.com/examples/github.com.pkg.sftp/Client/-/golang-client-class-examples.html)
1. [go语言使用sftp包上传文件和文件夹到远程服务器](https://blog.csdn.net/fu_qin/article/details/78741854)
1. [Implements support for double star (**) matches in golang's path.Match and filepath.Glob.](https://github.com/bmatcuk/doublestar)
1. [easyssh-proxy provides a simple implementation of some SSH protocol features in Go](https://github.com/appleboy/easyssh-proxy)
1. [List selection type alternative ssh/scp/sftp client. Pure Go.](https://github.com/blacknon/lssh)
1. [A library to handle ssh easily with Golang.It can do multiple proxy, x11 forwarding, etc.](https://github.com/blacknon/go-sshlib), [go-sshlib doc](https://godoc.org/github.com/blacknon/go-sshlib)
1. [An auditing / logging SSH relay for a jump box / bastion host.](https://github.com/iamacarpet/ssh-bastion)
1. [A curated list of SSH resources.](https://github.com/moul/awesome-ssh)
1. [melbahja/goph The native golang ssh client to execute your commands over ssh connection](https://github.com/melbahja/goph)
1. [yahoo/vssh Go library to handle tens of thousands SSH connections and execute the command(s) with higher-level API for building network device / server automation.](https://github.com/yahoo/vssh)
