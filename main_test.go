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
)

func Test_firstCommentInReader(t *testing.T) {
	testCases := []string{
		"/* Foo */\n// Bar\nvar x", "Bar",
		"/* Foo \nBar\n*/\n// Baz\nvar x", "Baz",
		"/*\nFoo \nBar\n*/\n// Baz\nvar x", "Foo\nBar",
		"\npackage something\n\n/*\nFoo \nBar\n*/\n// Baz\nvar x", "Foo\nBar",
		"/*\n Foo\nBar   \n*/\n", " Foo\nBar",
		"// Foo", "Foo",
		"// Foo\n// Bar", "Foo\nBar",
		"// Foo\n// \tBar", "Foo\n\tBar",
		"// Foo\n// Bar\n/* Baz */", "Foo\nBar",
		"// Foo\n// Bar\nvar x", "Foo\nBar",
		"\npackage something\n// Foo\n//   Bar\nvar x", "Foo\n  Bar",
		"", "",
		"package something\n\nvar x\n", "",
		"# Not in the\n# expected language\n", "",
	}

	for i := 0; i < len(testCases); i += 2 {
		comment, ok, _, _, _ := firstCommentInReader(strings.NewReader(testCases[i]), markers{"//", "/*", "*/"})
		if ok && comment != testCases[i+1] {
			t.Log(testCases[i])
			t.FailNow()
		}
		if !ok && testCases[i+1] != "" {
			t.Log(testCases[i])
			t.FailNow()
		}
	}
}

func Test_firstCommentInFile(t *testing.T) {
	_ = os.WriteFile("test.py", []byte(`
# Foo
#   Bar
print("Hello, World!")
`), 0666)
	defer os.Remove("test.py")
	comment, ok, firstLine, lastLine, err := firstCommentInFile("test.py")
	if !ok || comment != "Foo\n  Bar" || err != nil || firstLine != 2 || lastLine != 3 {
		t.FailNow()
	}
}

func Test_processDir(t *testing.T) {
	_ = os.WriteFile("test.py", []byte(`print("hello")`), 0666)
	defer os.Remove("test.py")
	_ = os.WriteFile("test.cs", []byte(`namespace HelloWorld{}`), 0666)
	defer os.Remove("test.cs")

	flagExcludeMap = map[string]bool{".go": true}
	err := processDir(".", "Foo\nBar")
	if err != nil {
		t.FailNow()
	}

	f, err := os.ReadFile("test.py")
	if err != nil || !bytes.Contains(f, []byte("# Foo\n# Bar\n")) {
		t.FailNow()
	}
	f, err = os.ReadFile("test.cs")
	if err != nil || !bytes.Contains(f, []byte("/*\nFoo\nBar\n*/")) {
		t.FailNow()
	}
}
