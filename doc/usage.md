# 多机执行ssh/上传/下载工具gossh使用手册

1. 下载 `https://github.com/bingoohuang/gossh/releases` 对应的版本，并且解压缩，重命名为gossh，并且移动到$PATH环境变量设定的目录下
1. 运行 `gossh -h`查看帮助

## 示例

1. 按主机执行远程命令

    ```bash
    $ gossh -H 1.1.1.1:61084,1.1.1.1:61085 -u=root --pass="111111" --cmds="%host date,%host uname -a" -e 1

    ---> 1.1.1.1:61084 <---

    Last login: Sun Apr 26 15:37:30 2020 from 60.247.93.190
    [root@tb10 ~]# date
    Sun Apr 26 15:39:40 CST 2020
    [root@tb10 ~]# uname -a
    Linux tb10 3.10.0-514.26.2.el7.x86_64 #1 SMP Tue Jul 4 15:04:05 UTC 2017 x86_64 x86_64 x86_64 GNU/Linux

    ---> 1.1.1.1:61085 <---

    Last login: Sun Apr 26 15:37:31 2020 from 60.247.93.190
    [root@tb11 ~]# date
    Sun Apr 26 15:39:41 CST 2020
    [root@tb11 ~]# uname -a
    Linux tb11 3.10.0-514.26.2.el7.x86_64 #1 SMP Tue Jul 4 15:04:05 UTC 2017 x86_64 x86_64 x86_64 GNU/Linux
    ```

1. 编辑主机列表(例如hosts.txt)， 和命令列表(例如cmds.txt)，然后指定执行列表文件执行

    ```bash
    $ cat hosts.txt
    123.206.185.162:61084 root/111111
    123.206.185.162:61085 root/111111
    $ cat cmds.txt
    %host date
    %host uname -a
    $ gossh --hostsFile hosts.txt --cmdsFile cmds.txt -e 1

    ---> 123.206.185.162:61084 <---

    Last login: Sun Apr 26 15:45:22 2020 from 60.247.93.190
    [root@tb10 ~]# date
    Sun Apr 26 15:45:48 CST 2020
    [root@tb10 ~]# uname -a
    Linux tb10 3.10.0-514.26.2.el7.x86_64 #1 SMP Tue Jul 4 15:04:05 UTC 2017 x86_64 x86_64 x86_64 GNU/Linux

    ---> 123.206.185.162:61085 <---

    Last login: Sun Apr 26 15:45:23 2020 from 60.247.93.190
    [root@tb11 ~]# date
    Sun Apr 26 15:45:48 CST 2020
    [root@tb11 ~]# uname -a
    Linux tb11 3.10.0-514.26.2.el7.x86_64 #1 SMP Tue Jul 4 15:04:05 UTC 2017 x86_64 x86_64 x86_64 GNU/Linux
    ```

1. 向远程主机上传命令 `%host %ul node.deployer-4.2.18.tar.gz /tmp`
1. 从远程主机下载命令 `%host %dl /tmp/node.deployer-4.2.18.tar.gz .`
1. 执行本机命令 `cd ~/Downloads/` 注意: 没有`%host`开头
1. 设置命令超时 `--cmdTimeout=60s`
