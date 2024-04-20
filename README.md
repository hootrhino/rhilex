# RHILEX
RHILEX 商业版本仓库。

## 构建
安装go1.18以后的版本。

```sh
go get
go build
```

## 跨平台
```
make arm32
make arm64
```

## 发布
```sh
bash ./release_pkg.sh
```