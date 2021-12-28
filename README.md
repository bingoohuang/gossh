# gossh

[![Travis CI](https://img.shields.io/travis/bingoohuang/gossh/master.svg?style=flat-square)](https://travis-ci.com/bingoohuang/gossh)
[![Software License](https://img.shields.io/badge/License-MIT-orange.svg?style=flat-square)](https://github.com/bingoohuang/gossh/blob/master/LICENSE.md)
[![GoDoc](https://img.shields.io/badge/godoc-reference-blue.svg?style=flat-square)](https://godoc.org/github.com/bingoohuang/gossh)
[![Coverage Status](http://codecov.io/github/bingoohuang/gossh/coverage.svg?branch=master)](http://codecov.io/github/bingoohuang/gossh?branch=master)
[![goreport](https://www.goreportcard.com/badge/github.com/bingoohuang/gossh)](https://www.goreportcard.com/report/github.com/bingoohuang/gossh)

execute shell scripts among multiple ssh servers

## Features

1. Support global variables defined in the host definition, example `@ARCH` in hosts, at 2021-11-30
    ```toml
    # gossh -c xx.toml --group 4
    
    #printConfig = true
    #passphrase="xxxx"
    
    hosts = [
    "admin:{PBE}xxx@192.168.1.2 group=4 @ARCH=amd64", # ARM 测试编译机
    ]
    
    # 全部命令都默认成远程执行，相当于自动添加了%host标识。
    globalRemote = true
    cmdTimeout = "300s"
    # confirm = true
    # exec mode(0: cmd by cmd, 1 host by host).
    execMode = 0
    
    cmds = [
    "%local rm -fr git.commit && make git.commit",
    "%local find ./ -name \".DS_Store\" -exec rm -rf '{}' ';'",
    "%local rm -fr vendor && go mod download && go mod vendor",
    "%local cd .. && rm -fr xyz.tar.gz && tar --exclude .git --exclude .idea -czf xyz.tar.gz xyz",
    "%ul xyz.tar.gz xyzsrc/",
    "cd xyzsrc",
    "[[ -d xyz ]] && rm -fr xyz",
    "tar zxf xyz.tar.gz --warning=no-unknown-keyword --warning=no-timestamp --exclude .git && cd xyz",
    "make install",
    "%local cd xyz && pwd",
    "%local date '+%Y%m%d%H%M%S' => @Now",
    "%dl xyzsrc/xyz/build/xyz build/xyz_@ARCH_@Now",
    "%local pwd && ls -lhd build/xyz_@ARCH_@Now",
    "%local upx build/xyz_@ARCH_@Now && ls -lh build/xyz_@ARCH_@Now",
    "%local rm -fr vendor/",
    ]
    ```
3.

## Usage demo

### repl mode

```bash
gossh -H "user:pass@aa.co" -H "user:pass@bb.co" --repl
log file /Users/bingoobjca/.gossh/logs/.-20211228140231.log created
>>> date

---> @aa.co:22 <---
Last login: Tue Dec 28 14:29:43 2021 from 60.247.93.190
[user@VM-24-15-centos ~]$ date
Tue Dec 28 14:31:52 CST 2021

---> bb.co:22 <---
Last login: Tue Dec 28 14:26:23 2021 from 192.168.225.11
[user@CS1 ~]# date
2021年 12月 28日 星期二 14:28:28 CST

>>> hostname

---> @aa.co:22 <---
[user@VM-24-15-centos ~]$ hostname
VM-24-15-centos

---> bb.co:22 <---
[user@CS1 ~]# hostname
CS1
>>> %local date

---> localhost <---
$ date
2021年12月28日 星期二 14时32分25秒 CST
>>> exit
log file /Users/bingoobjca/.gossh/logs/.-20211228140231.log recorded
```

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

## Config examples

### demo configuration of proxy

```toml
hosts = [
    "12.26.85.0:22 user/pass id=0",
    "12.26.85.1:22 root/na id=1 proxy=0", # proxy by id=0
    "12.26.85.2:22 root/na id=2 proxy=0",
    "12.26.85.3:22 root/na id=3 proxy=0",
]

cmds = [
    # execute on hosts whose id between 1 and 3
    "%host-(1-3) hostname -I",
]
```

### demo configuration of host group example

```toml
# group.toml
hosts = [
    # if no group specified, a group names default will be set.
    "12.26.85.0:22 user/pass group=proxy",
    "12.26.85.1:22 root/na proxy=0 group=g1/g3", # proxy by id=0
    "12.26.85.2:22 root/na proxy=0 group=g2/g3",
    "12.26.85.3:22 root/na proxy=0 group=g1/g2",
]


globalRemote = true

cmds = [
    "hostname -I",
]
```

cli commands:

1. `gossh -c group.toml --group=g1`
1. `gossh -c group.toml --group=g2`
1. `gossh -c group.toml --group=g3`

## demo configuration of tags

```toml
# tags.toml
hosts = [
    "12.26.85.0:22 user/pass",
    "12.26.85.1:22 user/pass",
    "12.26.85.2:22 user/pass",
    "12.26.85.3:22 user/pass",
]

globalRemote = true

# gossh -c tags.toml
cmds = [
    "echo bingoohuang",
]

# gossh -c tags.toml --tag=hostname
hostname-cmds = [
    "hostname -I",
]

# gossh -c tags.toml --tag=date
date-cmds = [
    "date",
]
```

## Substitute ResultVars

1. define result variables like `... => @varName`
1. use the variables like `echo @varName`

Notice:

1. `@varName` which is not capitalized will be limited to the host related.
1. `@VarName` which is capitalized will be global to all hosts.

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
    "date '+%Y%m%d' => @Today",
    "sh /tmp/hostdailycheck.sh",
    "%dl /tmp/log/HostDailyCheck-*-@Today.txt ./dailychecks@today/",
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

1. [A statically-linked ssh server with a reverse connection feature for simple yet powerful remote access. Most useful during HackTheBox challenges, CTFs or similar.](https://github.com/Fahrj/reverse-ssh)
    - CTF（Capture The Flag）中文一般译作夺旗赛，在网络安全领域中指的是网络安全技术人员之间进行技术竞技的一种比赛形式。 https://firmianay.gitbook.io/ctf-all-in-one
    -
    hackthebox是一个非常不错的在线实验平台，能帮助你提升渗透测试技能和黑盒测试技能，平台上有很多靶机，从易到难，各个级别的靶机都有。https://cloud.tencent.com/developer/article/1596548
1. [Rospo is a tool meant to create reliable ssh tunnels. It embeds an ssh server too if you want to reverse proxy a secured shell](https://github.com/ferama/rospo)
1. [Stack Up is a simple deployment tool that performs given set of commands on multiple hosts in parallel. It reads Supfile, a YAML, which defines networks (groups of hosts), commands and targets.](https://github.com/pressly/sup)
2. [Bootstrap](https://getbootstrap.com/)
3. [Bootstrap 4 Password Show Hide](https://codepen.io/Qanser/pen/dVRGJv)
4. [Tables](https://getbootstrap.com/docs/4.3/content/tables/)
5. [Golang SSH Client: Multiple Commands, Crypto & Goexpect Examples](http://networkbit.ch/golang-ssh-client/)
6. [bramvdbogaerde/go-scp](https://github.com/bramvdbogaerde/go-scp)
7. [golang 批量scp 远程传输文件](https://www.jianshu.com/p/f9d6dfefb63d)
8. [PBEWithMD5AndDES in go](https://github.com/LucasSloan/passwordbasedencryption)
9. [like python-sh, for easy call shell with golang](https://github.com/codeskyblue/go-sh)
10. [A scp client library written in Go. The remote server must have the scp command](https://github.com/hnakamur/go-scp)
11. [How the SCP protocol works](https://chuacw.ath.cx/blogs/chuacw/archive/2019/02/04/how-the-scp-protocol-works.aspx)
12. [Golang SFTP Client: Download File, Upload File Example](http://networkbit.ch/golang-sftp-client/)
13. [package sftp](https://godoc.org/github.com/pkg/sftp)
14. [sftp/example_test.go](https://github.com/pkg/sftp/blob/master/example_test.go)
15. [Golang Client Examples](https://golang.hotexamples.com/examples/github.com.pkg.sftp/Client/-/golang-client-class-examples.html)
16. [go语言使用sftp包上传文件和文件夹到远程服务器](https://blog.csdn.net/fu_qin/article/details/78741854)
17. [Implements support for double star (**) matches in golang's path.Match and filepath.Glob.](https://github.com/bmatcuk/doublestar)
18. [easyssh-proxy provides a simple implementation of some SSH protocol features in Go](https://github.com/appleboy/easyssh-proxy)
19. [List selection type alternative ssh/scp/sftp client. Pure Go.](https://github.com/blacknon/lssh)
20. [A library to handle ssh easily with Golang.It can do multiple proxy, x11 forwarding, etc.](https://github.com/blacknon/go-sshlib)
    , [go-sshlib doc](https://godoc.org/github.com/blacknon/go-sshlib)
21. [An auditing / logging SSH relay for a jump box / bastion host.](https://github.com/iamacarpet/ssh-bastion)
22. [A curated list of SSH resources.](https://github.com/moul/awesome-ssh)
23. [melbahja/goph The native golang ssh client to execute your commands over ssh connection](https://github.com/melbahja/goph)
24. [yahoo/vssh Go library to handle tens of thousands SSH connections and execute the command(s) with higher-level API for building network device / server automation.](https://github.com/yahoo/vssh)
