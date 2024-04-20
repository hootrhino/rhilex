## 标准库规范

- 最多传5个参数，需要资源的时候第一个参数永远是资源UUID
- 函数必须两个返回值：data，error

## 关于LUA
在 Lua 中，使用 `:` 语法来调用一个方法时，第一个参数是一个特殊的隐藏参数，通常被称为 `self` 或 `this`。这个参数是对对象本身的引用，也就是方法的接收者。在 Lua 中，这个参数不是指针，而是一个普通的 Lua 值，它可以是任何类型的对象，比如表（table）。
例如，假设你有一个表 `M`，它有一个方法 `fun`，你可以这样调用它：
```lua
M:fun()
```
在这个调用中，`M` 会作为 `self` 参数自动传递给 `fun` 方法。在 `fun` 方法的实现中，你可以通过 `self` 来访问 `M` 的成员变量和其他方法。
下面是一个简单的例子：
```lua
-- 定义一个表 M，它有一个成员变量和一个方法
M = {}
M.value = 10
function M:fun()
    print("self.value is:", self.value) -- self 指向 M
end
-- 调用 M 的 fun 方法
M:fun() -- 输出: self.value is: 10
```
在这个例子中，`M:fun()` 调用将 `M` 作为 `self` 参数传递给 `fun` 方法，所以 `self.value` 实际上就是 `M.value`。
需要注意的是，虽然在 Lua 中使用 `self` 来访问方法的接收者是一种约定，但实际上你可以使用任何名称来代表这个参数。例如，如果你不想使用 `self`，你可以这样定义方法：
```lua
function M:fun(receiver)
    print("receiver.value is:", receiver.value) -- receiver 指向 M
end
-- 调用 M 的 fun 方法
M:fun(M) -- 输出: receiver.value is: 10
```
在这个例子中，我们使用了 `receiver` 而不是 `self`，但是通常来说，遵循使用 `self` 的约定会让代码更易于阅读和理解。

## Lua函数传参
下面是一个简单的例子，展示了如何从 C 语言向 Lua 传递参数，并在 Lua 中获取这些参数并打印它们。
首先，我们创建一个 Lua 脚本 `print_params.lua`，它定义了一个函数 `print_params`，该函数接受可变数量的参数并打印它们：
```lua
-- print_params.lua
function print_params(...)
    local args = {...}
    for i, v in ipairs(args) do
        print("Parameter", i, ":", v)
    end
end
```
然后，我们编写一个 C 程序，它将调用 Lua 脚本中的 `print_params` 函数，并传递一些参数：
```c
#include <stdio.h>
#include <lua.h>
#include <lauxlib.h>
#include <lualib.h>
int main(void) {
    lua_State *L = luaL_newstate(); // 创建一个新的 Lua 状态
    luaL_openlibs(L); // 打开标准库
    // 加载 Lua 脚本
    if (luaL_loadfile(L, "print_params.lua") || lua_pcall(L, 0, 0, 0)) {
        printf("Error: %s\n", lua_tostring(L, -1));
        lua_close(L);
        return 1;
    }
    // 获取 print_params 函数
    lua_getglobal(L, "print_params");
    // 将参数压入栈中
    lua_pushstring(L, "Hello"); // 第一个参数
    lua_pushnumber(L, 123); // 第二个参数
    lua_pushboolean(L, 1); // 第三个参数（true）
    // 调用 Lua 函数
    if (lua_pcall(L, 3, 0, 0) != LUA_OK) { // 3 个参数，没有返回值
        printf("Error: %s\n", lua_tostring(L, -1));
        lua_close(L);
        return 1;
    }
    lua_close(L); // 关闭 Lua 状态
    return 0;
}
```
在这个 C 程序中，我们首先创建了一个 Lua 状态，并加载了 `print_params.lua` 脚本。然后，我们使用 `lua_getglobal` 获取了 Lua 中的 `print_params` 函数，并使用 `lua_push*` 函数系列将三个参数压入栈中（一个字符串、一个数字和一个布尔值）。最后，我们使用 `lua_pcall` 来调用 Lua 函数，并传递了三个参数。
当你运行这个 C 程序时，它将执行 Lua 脚本中的 `print_params` 函数，并打印出传递给它的参数。
