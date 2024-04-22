# 固件证书管理器
固件证书管理器，用来防止盗版或者破解。开源版不限制使用，商业版有单独的证书管理器。
## 接口

例如要激活：`00002&admin&123456&FF:FF:FF:FF:FF&0&0`:

```sh
curl -X 'GET' \
  'http://106.15.225.172:8000/api/v1/device-active?param=00002%26admin%26123456%26FF%3AFF%3AFF%3AFF%3AFF%260%260' \
  -H 'accept: application/json'
```

会返回一个压缩包文件，里面包含了证书。