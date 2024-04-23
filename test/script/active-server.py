# Copyright (C) 2024 wwhai
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU Affero General Public License as
# published by the Free Software Foundation, either version 3 of the
# License, or (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU Affero General Public License for more details.
#
# You should have received a copy of the GNU Affero General Public License
# along with this program.  If not, see <https://www.gnu.org/licenses/>.
from flask import Flask, request, send_file
import zipfile
import os

app = Flask(__name__)

# 定义压缩包的路径和文件名
zip_file_path = "license.zip"

# 创建一个zip文件，包含../config/license.lic
with zipfile.ZipFile(zip_file_path, "w") as zipf:
    # 注意：这里的路径是相对路径，需要根据你的文件结构进行调整
    zipf.write("../config/license.lic", arcname="license.lic")
    zipf.write("../config/license.key", arcname="license.key")


@app.route("/", methods=["POST"])
def handle_post():
    # 打印请求体中的JSON
    json_data = request.json
    print(json_data)

    # 返回包含license.lic的zip文件
    return send_file(zip_file_path, as_attachment=True)


if __name__ == "__main__":
    app.run(port=6677)
