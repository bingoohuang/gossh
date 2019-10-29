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