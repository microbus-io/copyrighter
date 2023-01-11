/*
Copyright 2023 Microbus Open Source Software and various contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package main

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_FirstComment(t *testing.T) {
	source := `
/*
Right
*/
// Wrong
package x
`
	comment, ok, from, to := firstComment(source, languages[".go"])
	if assert.True(t, ok) {
		assert.Equal(t, "Right", comment)
		assert.Equal(t, from, 1)
		assert.Equal(t, to, 4)
	}

	source = `
package example
/*
Right
*/
// Wrong
var x
`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assert.True(t, ok) {
		assert.Equal(t, "Right", comment)
		assert.Equal(t, from, 2)
		assert.Equal(t, to, 5)
	}

	source = `/*
Right
 Right
*/
`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assert.True(t, ok) {
		assert.Equal(t, "Right\n Right", comment)
		assert.Equal(t, from, 0)
		assert.Equal(t, to, 4)
	}

	source = `/* Wrong */
// Right
package example
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assert.True(t, ok) {
		assert.Equal(t, "Right", comment)
		assert.Equal(t, from, 1)
		assert.Equal(t, to, 2)
	}

	source = `// Right
// Right
package example
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assert.True(t, ok) {
		assert.Equal(t, "Right\nRight", comment)
		assert.Equal(t, from, 0)
		assert.Equal(t, to, 2)
	}

	source = `
// Right
// Right
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assert.True(t, ok) {
		assert.Equal(t, "Right\nRight", comment)
		assert.Equal(t, from, 1)
		assert.Equal(t, to, 3)
	}

	source = `
// Right
//  Right
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assert.True(t, ok) {
		assert.Equal(t, "Right\n Right", comment)
		assert.Equal(t, from, 1)
		assert.Equal(t, to, 3)
	}

	source = `
// Right
/*
Wrong
*/
// Wrong
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assert.True(t, ok) {
		assert.Equal(t, "Right", comment)
		assert.Equal(t, from, 1)
		assert.Equal(t, to, 2)
	}

	source = `
package example
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assert.False(t, ok) {
		assert.Equal(t, "", comment)
		assert.Equal(t, from, 0)
		assert.Equal(t, to, 0)
	}

	source = `/*
package example
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assert.False(t, ok) {
		assert.Equal(t, "", comment)
		assert.Equal(t, from, 0)
		assert.Equal(t, to, 0)
	}

	source = `package example // Wrong
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assert.False(t, ok) {
		assert.Equal(t, "", comment)
		assert.Equal(t, from, 0)
		assert.Equal(t, to, 0)
	}
}

func Test_ProcessDir(t *testing.T) {
	_ = os.WriteFile("test.py", []byte(`print("hello")`), 0666)
	defer os.Remove("test.py")
	_ = os.WriteFile("test.cs", []byte(`namespace HelloWorld{}`), 0666)
	defer os.Remove("test.cs")

	flagExcludeMap = map[string]bool{".go": true}
	err := processDir(".", "Foo\nBar")
	if !assert.NoError(t, err) {
		return
	}

	f, err := os.ReadFile("test.py")
	if assert.NoError(t, err) {
		assert.True(t, bytes.Contains(f, []byte("# Foo\n# Bar\n")))
	}
	f, err = os.ReadFile("test.cs")
	if assert.NoError(t, err) {
		assert.True(t, bytes.Contains(f, []byte("/*\nFoo\nBar\n*/")))
	}
}

func Test_NewComment(t *testing.T) {
	source := `// Package example does something
package example

var x
`
	var sb strings.Builder
	ok, err := process(strings.NewReader(source), &sb, languages[".go"], "Copyright notice")
	if assert.True(t, ok) && assert.NoError(t, err) {
		result := `/*
Copyright notice
*/

// Package example does something
package example

var x
`
		assert.Equal(t, result, sb.String())
	}
}

func Test_ReplaceCommentTopLine(t *testing.T) {
	source := `// Old copyright notice
package example

var x
`
	var sb strings.Builder
	ok, err := process(strings.NewReader(source), &sb, languages[".go"], "Copyright notice")
	if assert.True(t, ok) && assert.NoError(t, err) {
		result := `/*
Copyright notice
*/
package example

var x
`
		assert.Equal(t, result, sb.String())
	}
}

func Test_ReplaceCommentMiddle(t *testing.T) {
	source := `package example
/*
Old copyright notice
*/
var x
`
	var sb strings.Builder
	ok, err := process(strings.NewReader(source), &sb, languages[".go"], "Copyright notice")
	if assert.True(t, ok) && assert.NoError(t, err) {
		result := `package example
/*
Copyright notice
*/
var x
`
		assert.Equal(t, result, sb.String())
	}
}

func Test_Empty(t *testing.T) {
	source := ``
	var sb strings.Builder
	ok, err := process(strings.NewReader(source), &sb, languages[".go"], "Copyright notice")
	if assert.True(t, ok) && assert.NoError(t, err) {
		result := `/*
Copyright notice
*/
`
		assert.Equal(t, result, sb.String())
	}
}

func Test_CarriageReturn(t *testing.T) {
	source := "package example\r\n" +
		"\r\n" +
		"var x\r\n"
	var sb strings.Builder
	ok, err := process(strings.NewReader(source), &sb, languages[".go"], "Copyright\nnotice")
	if assert.True(t, ok) && assert.NoError(t, err) {
		result := "/*\r\n" +
			"Copyright\r\nnotice\r\n" +
			"*/\r\n" +
			"\r\n" +
			"package example\r\n" +
			"\r\n" +
			"var x\r\n"
		assert.Equal(t, result, sb.String())
	}
}
