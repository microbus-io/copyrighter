/*
Copyright 2022 Microbus Open Source Software and various contributors

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
	"strings"
	"testing"
)

func Test_firstCommentInReader(t *testing.T) {
	testCases := []string{
		"/* Foo */\n// Bar\nvar x", "Foo",
		"/* Foo \nBar\n*/\n// Baz\nvar x", "Foo\nBar",
		"/*\nFoo \nBar\n*/\n// Baz\nvar x", "Foo\nBar",
		"\npackage something\n\n/*\nFoo \nBar\n*/\n// Baz\nvar x", "Foo\nBar",
		"// Foo", "Foo",
		"// Foo\n// Bar", "Foo\nBar",
		"// Foo\n// \tBar", "Foo\n\tBar",
		"// Foo\n// Bar\n/* Baz */", "Foo\nBar",
		"// Foo\n// Bar\nvar x", "Foo\nBar",
	}

	for i := 0; i < len(testCases); i += 2 {
		comment, ok, _ := firstCommentInReader(strings.NewReader(testCases[i]), markers{"//", "/*", "*/"})
		if !ok || comment != testCases[i+1] {
			t.FailNow()
		}
	}
}
