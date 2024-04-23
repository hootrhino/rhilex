# 固件证书管理器
固件证书管理器，用来防止盗版或者破解。开源版不限制使用，商业版有单独的证书管理器。
## 操作指南

---
**添加证书到配置路径的说明**
1. **下载证书压缩包**
   - 执行：`rhilex.exe active -H "http://127.0.0.1:6677" -U admin -P 123456`
   - 你将收到一个名为`license.zip`的压缩文件下载链接。
2. **解压证书压缩包**
   - 找到下载的`license.zip`文件。
   - 右键点击文件，选择“解压到当前文件夹”或类似的选项（取决于你使用的压缩软件）。
   - 解压后，你将得到一个名为`license.lic`和`license.key`的证书文件。
3. **确定配置路径**
   - 根据你的应用程序或系统，确定证书需要放置的配置路径。这通常是在应用程序的配置文件夹内，或者是在操作系统的某个特定目录下。
   - 如果你不确定配置路径，请查阅应用程序的文档或联系技术支持获取帮助。
4. **将证书移动到配置路径**
   - 打开文件资源管理器，导航到你刚刚解压出的`license.lic`文件所在的位置。
   - 将`license.lic`文件拖放到配置路径的文件夹中，或者使用剪切和粘贴操作将文件移动到配置路径。
5. **确认证书安装**
   - 一旦证书文件被放置在正确的配置路径中，应用程序应该能够自动检测到并开始使用新的证书。
   - 如果你需要重启应用程序或服务来使证书生效，请按照应用程序的说明进行操作。
6. **验证证书**
   - 打开应用程序，检查是否显示证书已安装且有效的提示。
   - 如果应用程序有任何证书验证工具或设置，请使用它们来确保证书已正确安装。

## 测试
下面是一个非常简单的激活服务器方便测试：
```py
from flask import Flask, request, send_file
import zipfile
import os

app = Flask(__name__)

zip_file_path = "license.zip"

with zipfile.ZipFile(zip_file_path, "w") as zipf:
    zipf.write("../config/license.lic", arcname="license.lic")
    zipf.write("../config/license.key", arcname="license.key")


@app.route("/", methods=["POST"])
def handle_post():
    json_data = request.json
    print(json_data)
    return send_file(zip_file_path, as_attachment=True)


if __name__ == "__main__":
    app.run(port=6677)

```