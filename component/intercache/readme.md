<!--
 Copyright (C) 2023 wwhai

 This program is free software: you can redistribute it and/or modify
 it under the terms of the GNU Affero General Public License as
 published by the Free Software Foundation, either version 3 of the
 License, or (at your option) any later version.

 This program is distributed in the hope that it will be useful,
 but WITHOUT ANY WARRANTY; without even the implied warranty of
 MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 GNU Affero General Public License for more details.

 You should have received a copy of the GNU Affero General Public License
 along with this program.  If not, see <http://www.gnu.org/licenses/>.
-->

# 内部缓存器

![1713353675527](image/readme/1713353675527.png)
主要用来做内存缓存加速使用，比如留给物模型来展示实时值等。是一个K-V存储器，其本质是Map。只保存点位的ID和数值映射关系，因此需要处理好ID作为MAP的K不能冲突了。