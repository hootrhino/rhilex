---@diagnostic disable: undefined-global
-- Success
function Success()
end

-- Failed
function Failed(error)
    print("Error:", error)
end

-- Actions
Actions =
{
    --        ┌───────────────────────────────────────────────┐
    -- data = |00 00 00 01|00 00 00 01|00 00 00 01|00 00 00 01|
    --        └───────────────────────────────────────────────┘
    function(args)
        local json = require("json")
        local tb = rhilexlib:MB("<a:16 b:16 c:16 d1:16", data, false)
        local result = {}
        result['a'] = rhilexlib:B2I64(1, rhilexlib:BS2B(tb["a"]))
        result['b'] = rhilexlib:B2I64(1, rhilexlib:BS2B(tb["b"]))
        result['c'] = rhilexlib:B2I64(1, rhilexlib:BS2B(tb["c"]))
        result['d1'] = rhilexlib:B2I64(1, rhilexlib:BS2B(tb["d1"]))
        print("rhilexlib:MB 2:", json.encode(result))
        data:ToMqtt('OUTEND', json.encode(result))
        return true, args
    end
}

