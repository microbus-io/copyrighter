/*
Copyright 2023 Microbus LLC and various contributors

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
	"reflect"
	"strings"
	"testing"
)

func assertTrue(t *testing.T, b bool) bool {
	if !b {
		t.Fail()
		return false
	}
	return true
}

func assertFalse(t *testing.T, b bool) bool {
	if b {
		t.Fail()
		return false
	}
	return true
}

func assertEqual(t *testing.T, expected any, actual any) bool {
	if !reflect.DeepEqual(expected, actual) {
		t.Fail()
		return false
	}
	return true
}

func assertNoError(t *testing.T, err error) bool {
	if err != nil {
		t.Fail()
		return false
	}
	return true
}

func Test_FirstComment(t *testing.T) {
	source := `
/*
Right
*/
// Wrong
package x
`
	comment, ok, from, to := firstComment(source, languages[".go"])
	if assertTrue(t, ok) {
		assertEqual(t, "Right", comment)
		assertEqual(t, from, 1)
		assertEqual(t, to, 4)
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
	if assertTrue(t, ok) {
		assertEqual(t, "Right", comment)
		assertEqual(t, from, 2)
		assertEqual(t, to, 5)
	}

	source = `/*
Right
 Right
*/`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assertTrue(t, ok) {
		assertEqual(t, "Right\n Right", comment)
		assertEqual(t, from, 0)
		assertEqual(t, to, 4)
	}

	source = `/* Wrong */
// Right
package example
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assertTrue(t, ok) {
		assertEqual(t, "Right", comment)
		assertEqual(t, from, 1)
		assertEqual(t, to, 2)
	}

	source = `// Right
// Right
package example
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assertTrue(t, ok) {
		assertEqual(t, "Right\nRight", comment)
		assertEqual(t, from, 0)
		assertEqual(t, to, 2)
	}

	source = `
// Right
// Right`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assertTrue(t, ok) {
		assertEqual(t, "Right\nRight", comment)
		assertEqual(t, from, 1)
		assertEqual(t, to, 3)
	}

	source = `
// Right
//  Right
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assertTrue(t, ok) {
		assertEqual(t, "Right\n Right", comment)
		assertEqual(t, from, 1)
		assertEqual(t, to, 3)
	}

	source = `
// Right
/*
Wrong
*/
// Wrong
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assertTrue(t, ok) {
		assertEqual(t, "Right", comment)
		assertEqual(t, from, 1)
		assertEqual(t, to, 2)
	}

	source = `
package example
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assertFalse(t, ok) {
		assertEqual(t, "", comment)
		assertEqual(t, from, 0)
		assertEqual(t, to, 0)
	}

	source = `/*
package example
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assertFalse(t, ok) {
		assertEqual(t, "", comment)
		assertEqual(t, from, 0)
		assertEqual(t, to, 0)
	}

	source = `package example // Wrong
	`
	comment, ok, from, to = firstComment(source, languages[".go"])
	if assertFalse(t, ok) {
		assertEqual(t, "", comment)
		assertEqual(t, from, 0)
		assertEqual(t, to, 0)
	}
}

func Test_Main(t *testing.T) {
	_ = os.WriteFile("test.py", []byte(`print("hello")`), 0666)
	defer os.Remove("test.py")
	_ = os.Mkdir("testdir", os.ModePerm)
	defer os.RemoveAll("testdir")
	_ = os.WriteFile("testdir/test.cs", []byte(`namespace HelloWorld{}`), 0666)
	defer os.Remove("testdir/test.cs")

	flagExclude = "go"
	flagRecurse = true
	err := mainErr()
	assertNoError(t, err)

	f, err := os.ReadFile("test.py")
	if assertNoError(t, err) {
		assertTrue(t, bytes.Contains(f, []byte("# Copyright")))
	}
	f, err = os.ReadFile("testdir/test.cs")
	if assertNoError(t, err) {
		assertTrue(t, bytes.Contains(f, []byte("/*\nCopyright")))
	}
}

func Test_NewComment(t *testing.T) {
	source := `// Package example does something
package example

var x
`
	var sb strings.Builder
	ok, err := process(strings.NewReader(source), &sb, languages[".go"], "Copyright notice")
	if assertTrue(t, ok) && assertNoError(t, err) {
		result := `/*
Copyright notice
*/

// Package example does something
package example

var x
`
		assertEqual(t, result, sb.String())
	}
}

func Test_ReplaceCommentTopLine(t *testing.T) {
	source := `// Old copyright notice
package example

var x
`
	var sb strings.Builder
	ok, err := process(strings.NewReader(source), &sb, languages[".go"], "Copyright notice")
	if assertTrue(t, ok) && assertNoError(t, err) {
		result := `/*
Copyright notice
*/
package example

var x
`
		assertEqual(t, result, sb.String())
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
	if assertTrue(t, ok) && assertNoError(t, err) {
		result := `package example
/*
Copyright notice
*/
var x
`
		assertEqual(t, result, sb.String())
	}
}

func Test_Empty(t *testing.T) {
	source := ``
	var sb strings.Builder
	ok, err := process(strings.NewReader(source), &sb, languages[".go"], "Copyright notice")
	if assertTrue(t, ok) && assertNoError(t, err) {
		result := `/*
Copyright notice
*/
`
		assertEqual(t, result, sb.String())
	}
}

func Test_CarriageReturn(t *testing.T) {
	source := "package example\r\n" +
		"\r\n" +
		"var x\r\n"
	var sb strings.Builder
	ok, err := process(strings.NewReader(source), &sb, languages[".go"], "Copyright\nnotice")
	if assertTrue(t, ok) && assertNoError(t, err) {
		result := "/*\r\n" +
			"Copyright\r\nnotice\r\n" +
			"*/\r\n" +
			"\r\n" +
			"package example\r\n" +
			"\r\n" +
			"var x\r\n"
		assertEqual(t, result, sb.String())
	}
}
