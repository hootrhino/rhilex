-- data = [
--     {"tag":"add1", "id": "001", "value": 0x0001},
--     {"tag":"add2", "id": "002", "value": 0x0002},
-- ]

function ParseData(args)
    -- data: {"in":"AA0011...","out":"AABBCDD..."}
    local DataT, err = rhilexlib:J2T(args)
    if err ~= nil then
        return true, args
    end
    -- Do your business
    rhilexlib:log(DataT['in'])
    rhilexlib:log(DataT['out'])
end
