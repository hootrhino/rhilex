# Openwrt daemon script
该脚本是RHILEX的`Openwrt系统`操作脚本，处理RHILEX的安装、启动、停止、卸载等。
## 基础使用
### 下载
将安装包解压:
```sh
unzip rhilex-arm32linux-$VERSION.zip -d rhilex
```
### 安装
```sh
./rhilex-openwrt.sh install
```

### 卸载
```sh
./rhilex-openwrt.sh uninstall
```

### 使用
下面的脚本一定要在root权限下执行,或者使用sudo。
```bash
# 启动
./rhilex-openwrt.sh start
# 停止
./rhilex-openwrt.sh stop
# 重启
./rhilex-openwrt.sh restart
# 状态
./rhilex-openwrt.sh status
```

## 守护进程
```sh
# 打开crontab
sudo crontab -e
# 输入
@reboot (export ARCHSUPPORT=RHINOPI && /etc/init.d/rhilex.service start > /var/log/rhilex.log 2>&1)
```