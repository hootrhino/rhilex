// Copyright (C) 2024 wwhai
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as
// published by the Free Software Foundation, either version 3 of the
// License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package test

import (
	"testing"

	"github.com/alecthomas/repr"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

// A custom lexer for INI files. This illustrates a relatively complex Regexp lexer, as well
// as use of the Unquote filter, which unquotes string tokens.
var (
	iniLexer = lexer.MustSimple([]lexer.SimpleRule{
		{`Ident`, `[a-zA-Z][a-zA-Z_\d]*`},
		{`String`, `"(?:\\.|[^"])*"`},
		{`Float`, `\d+(?:\.\d+)?`},
		{`Punct`, `[][=]`},
		{"comment", `[#;][^\n]*`},
		{"whitespace", `\s+`},
	})
)

type INI struct {
	Properties []*_Property `@@*`
	Sections   []*Section   `@@*`
}

type Section struct {
	Identifier string       `"[" @Ident "]"`
	Properties []*_Property `@@*`
}

type _Property struct {
	Key   string `@Ident "="`
	Value Value  `@@`
}

type Value interface{ value() }

type _String struct {
	String string `@String`
}

func (_String) value() {}

type Number struct {
	Number float64 `@Float`
}

func (Number) value() {}

// go test -timeout 30s -run ^Test_parse_sql github.com/hootrhino/rhilex/test -v -count=1
func Test_parse_sql(t *testing.T) {
	s := `
a = "a"
b = 123

# A comment
[numbers]
a = 10.3
b = 20

; Another comment
[strings]
a = "\"quoted\""
b = "b"
`
	parser := participle.MustBuild[INI](
		participle.Lexer(iniLexer),
		participle.Unquote("String"),
		participle.Union[Value](_String{}, Number{}),
	)
	ini, err := parser.ParseString("", s)
	repr.Println(ini, repr.Indent("  "), repr.OmitEmpty(true))
	if err != nil {
		panic(err)
	}
}
