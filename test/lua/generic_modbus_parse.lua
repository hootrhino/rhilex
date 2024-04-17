-- {
--     "d1":{
--         "tag":"d1",
--         "function":3,
--         "slaverId":1,
--         "address":0,
--         "quantity":2,
--         "value":"AiYBDA=="
--     },
--     "d2":{
--         "tag":"d2",
--         "function":3,
--         "slaverId":2,
--         "address":0,
--         "quantity":2,
--         "value":"AicBCQ=="
--     }
-- }
---@diagnostic disable: undefined-global
-- Success
function Success()
    -- rhilexlib:log("success")
end
-- Failed
function Failed(error)
    rhilexlib:log("Error:", error)
end

-- Actions
Actions = {function(args)
    local jt = rhilexlib:J2T(data)
    for k, v in pairs(jt) do
        local ht = rhilexlib:MB('>hv:16 tv:16', v['value'], false)
        print(k, "Raw value:", ht['hv'], ht['tv'])
        local humi = rhilexlib:B2I64('>', rhilexlib:BS2B(ht['hv']))
        local temp = rhilexlib:B2I64('>', rhilexlib:BS2B(ht['tv']))
        local ts = rhilexlib:TsUnixNano()
        local jsont = {
            method = 'report',
            clientToken = ts,
            timestamp = ts,
            params = {
                temp = temp,
                humi = humi
            }
        }
        print(k, "Parsed value:", rhilexlib:T2J(jsont))
    end
    return true, args
end}
