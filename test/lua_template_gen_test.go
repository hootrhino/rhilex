package test

import (
	"os"
	"testing"
	"text/template"
)

var s = `
---@diagnostic disable: undefined-global
-- Success
function Success()
end
-- Failed
function Failed(error)
    print("Error:", error)
end

-- Actions
Actions = {function(args)
    local t = rhilexlib:J2T(data)
    local V0 = rhilexlib:MB(">{{.a}}:16 {{.b}}:16 {{.c}}:16 {{.d}}:16 {{.e}}:16", t['value'], false)
    local a = rhilexlib:T2J(V0['{{.a}}'])
    local b = rhilexlib:T2J(V0['{{.b}}'])
    local c = rhilexlib:T2J(V0['{{.c}}'])
    local d = rhilexlib:T2J(V0['{{.d}}'])
    local e = rhilexlib:T2J(V0['{{.e}}'])
    print('{{.a}} ==> ', {{.a}}, ' ->', rhilexlib:B2I64('>', rhilexlib:BS2B(a)))
    print('{{.b}} ==> ', {{.b}}, ' ->', rhilexlib:B2I64('>', rhilexlib:BS2B(b)))
    print('{{.c}} ==> ', {{.c}}, ' ->', rhilexlib:B2I64('>', rhilexlib:BS2B(c)))
    print('{{.d}} ==> ', {{.d}}, ' ->', rhilexlib:B2I64('>', rhilexlib:BS2B(d)))
    print('{{.e}} ==> ', {{.e}}, ' ->', rhilexlib:B2I64('>', rhilexlib:BS2B(e)))
    return true, args
end}

`

func Test_gen_template(*testing.T) {
	t := template.New("test")
	t = template.Must(t.Parse(s))

	t.Execute(os.Stdout, map[string]string{
		"a": "va",
		"b": "vb",
		"c": "vc",
		"d": "vd",
		"e": "ve",
	})
}
