-- Copyright (C) 2024 wwhai
--
-- This program is free software: you can redistribute it and/or modify
-- it under the terms of the GNU Affero General Public License as
-- published by the Free Software Foundation, either version 3 of the
-- License, or (at your option) any later version.
--
-- This program is distributed in the hope that it will be useful,
-- but WITHOUT ANY WARRANTY; without even the implied warranty of
-- MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
-- GNU Affero General Public License for more details.
--
-- You should have received a copy of the GNU Affero General Public License
-- along with this program.  If not, see <https://www.gnu.org/licenses/>.


function Main(arg)
    for i = 1, 10, 1 do
        local err = rds:Save('SCHEMAAAEM8PHB', {
            info = "运行信息",
            temperature = 25.44,
            oxygen = 20.78,
            ph_value = 7.5
        })
        if err ~= nil then
            Throw(err)
            return 0
        end
    end
    return 0
end
