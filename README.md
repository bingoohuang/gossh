# gossh
execute shell scripts among multiple ssh servers

## usage

```bash
$  uuidgen
6F948925-429E-4D2C-B551-C9C6D12E5062
$  gossh --pbe hello,word --password C9C6D12E5062
+---+-------+-----------------------------+
| # | PLAIN | ENCRYPTED                   |
+---+-------+-----------------------------+
| 1 | hello | {PBE}eiRMlsZPLikVYpZMcHicyg |
| 2 | word  | {PBE}lAHH0UfuqZ0YtV_5VE77uw |
+---+-------+-----------------------------+
$  gossh --pbe hello,word --password C9C6D12E5062
+---+-------+-----------------------------+
| # | PLAIN | ENCRYPTED                   |
+---+-------+-----------------------------+
| 1 | hello | {PBE}6RGab13x5WfzFP0NpA_suA |
| 2 | word  | {PBE}qmPJAysHSmnfQEK-a6JM0A |
+---+-------+-----------------------------+

$  gossh --ebp 6RGab13x5WfzFP0NpA_suA,qmPJAysHSmnfQEK-a6JM0A --password C9C6D12E5062
+---+------------------------+-------+
| # | ENCRYPTED              | PLAIN |
+---+------------------------+-------+
| 1 | 6RGab13x5WfzFP0NpA_suA | hello |
| 2 | qmPJAysHSmnfQEK-a6JM0A | word  |
+---+------------------------+-------+
$  gossh --ebp {PBE}eiRMlsZPLikVYpZMcHicyg,{PBE}lAHH0UfuqZ0YtV_5VE77uw --password C9C6D12E5062
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



## Scripts

```bash
$ export em="\!";export sq="'";export dq='"';
$ export GOSSH_CMDS="%host-9 MYSQL_PWD='${em}QAZ2wsx' ./mci/mysql -u root -h 127.0.0.1 -vvv -e ${dq}show variables like 'server%'${dq}"
$ gossh -h="192.168.136.9:8022 app/app id=9" -h="192.168.136.18:8022 app/app id=18"
executing MYSQL_PWD='!QAZ2wsx' ./mci/mysql -u root -h 127.0.0.1 -vvv -e "show variables like 'server%'" on hosts [192.168.136.9:8022]
Last login: Mon Nov 18 15:06:22 2019 from 192.168.217.48
ONLY Authorized users only! All accesses logged
MYSQL_PWD='!QAZ2wsx' ./mci/mysql -u root -h 127.0.0.1 -vvv -e "show variables like 'server%'"
exit
[app@BJCA-device ~]$ MYSQL_PWD='!QAZ2wsx' ./mci/mysql -u root -h 127.0.0.1 -vvv -e "show variables like 'server%'"
--------------
show variables like 'server%'
--------------

+----------------+--------------------------------------+
| Variable_name  | Value                                |
+----------------+--------------------------------------+
| server_id      | 1                                    |
| server_id_bits | 32                                   |
| server_uuid    | 43e9cbe5-b38a-11e9-8570-04d4c439354e |
+----------------+--------------------------------------+
3 rows in set (0.00 sec)

Bye
[app@BJCA-device ~]$ exit
登出
```

```bash
gossh -h="192.168.136.9:8022 app/app id=9,192.168.136.18:8022 app/app id=18" --cmds="%host-9 MYSQL_PWD='\!QAZ2wsx' ./mci/mysql -u root -h 127.0.0.1 -vvv -e 'show slave status\G'"
gossh -h="192.168.136.9:8022 app/app id=9,192.168.136.18:8022 app/app id=18" --cmds="%host-9 %ul ~/go/bin/linux_amd64/mci ./mci,%host-9 ./mci/mci -v"
gossh -h="192.168.136.9:8022 app/app id=9,192.168.136.18:8022 app/app id=18" --cmds="%host-9 %dl ./mci/mci ."
```