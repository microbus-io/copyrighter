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

// Package main runs a code generator that injects a copyright notice to source files.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CommentMarkers are the comment CommentMarkers for a language.
type CommentMarkers struct {
	single     string
	multiBegin string
	multiEnd   string
}

// languages is a map of the markers used to denote comments in each language.
var languages = map[string]CommentMarkers{
	".c":    {"//", "/*", "*/"},
	".cpp":  {"//", "/*", "*/"},
	".cs":   {"//", "/*", "*/"},
	".css":  {"", "/*", "*/"},
	".go":   {"//", "/*", "*/"},
	".html": {"", "<!--", "-->"},
	".java": {"//", "/*", "*/"},
	".js":   {"//", "/*", "*/"},
	".php":  {"//", "/*", "*/"},
	".ps1":  {"#", "<#", "#>"},
	".py":   {"#", "", ""},
	".sh":   {"#", "", ""},
	".sql":  {"--", "/*", "*/"},
	".ts":   {"//", "/*", "*/"},
	".xml":  {"", "<!--", "-->"},
	".yaml": {"#", "", ""},
}

var (
	flagRecurse    bool
	flagVerbose    bool
	flagExclude    string
	flagExcludeMap map[string]bool
)

// main runs a code generator that injects a copyright notice to source files.
func main() {
	// Parse CLI flags
	flag.BoolVar(&flagRecurse, "r", false, "Recurse sub-directories")
	flag.BoolVar(&flagVerbose, "v", false, "Verbose")
	flag.StringVar(&flagExclude, "x", "", "Comma-separated list of extensions to exclude")
	flag.Parse()

	err := mainErr()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

// mainErr scans the current directory for a copyright.go or doc.go file and applies the first comment
// in that file to all other source files in the directory.
func mainErr() error {
	flagExcludeMap = map[string]bool{}
	for _, x := range strings.Split(flagExclude, ",") {
		flagExcludeMap["."+strings.TrimPrefix(x, ".")] = true
	}
	if flagVerbose {
		fmt.Println("Copyrighter")
	}
	// Load the first comment found in copyright.go or doc.go
	b, err := os.ReadFile("copyright.go")
	if err != nil {
		b, err = os.ReadFile("doc.go")
	}
	if err != nil {
		return fmt.Errorf("unable to read copyright.go or doc.go: %w", err)
	}
	source := string(b)
	notice, ok, _, _ := firstComment(source, languages[".go"])
	if !ok {
		return fmt.Errorf("no comment found in copyright.go or doc.go")
	}
	// Apply the comment to the files in the current directory
	err = processDir(".", notice)
	if err != nil {
		return err
	}
	return nil
}

// processDir applies the copyright notice to the source files in the indicated directory.
func processDir(dirPath string, notice string) error {
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("unable to read files in '%s': %w", dirPath, err)
	}
	for _, de := range dirEntries {
		if de.IsDir() {
			if flagRecurse {
				err = processDir(filepath.Join(dirPath, de.Name()), notice)
				if err != nil {
					return err
				}
			}
			continue
		}
		ext := filepath.Ext(de.Name())
		lang, ok := languages[ext]
		if !ok || flagExcludeMap[ext] {
			continue
		}

		fileName := filepath.Join(dirPath, de.Name())
		source, err := os.ReadFile(fileName)
		if err != nil {
			return err
		}
		var toWrite bytes.Buffer
		ok, err = process(bytes.NewReader(source), &toWrite, lang, notice)
		if ok {
			err = os.WriteFile(fileName, toWrite.Bytes(), 0666)
		}
		if err != nil {
			return fmt.Errorf("failed to process '%s': %w", fileName, err)
		}
		if flagVerbose {
			fmt.Println("  " + fileName)
		}
	}
	return nil
}

// process reads the source code from the reader, inserts the copyright notice if appropriate,
// and writes the results to the writer.
func process(r io.Reader, f io.Writer, lang CommentMarkers, notice string) (ok bool, err error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return false, err
	}
	source := string(b)
	firstComment, ok, fromLine, toLine := firstComment(source, lang)
	if ok && firstComment == notice {
		return false, nil
	}
	if ok && !strings.Contains(strings.ToLower(firstComment), "copyright") {
		ok = false
		fromLine = 0
		toLine = 0
	}

	lineSep := "\n"
	if strings.Contains(source, "\r\n") {
		lineSep = "\r\n"
	}

	// Insert lines before the copyright notice to be replaced
	lines := strings.Split(source, "\n")
	for i := 0; i < fromLine; i++ {
		f.Write([]byte(lines[i]))
		if i < len(lines)-1 {
			f.Write([]byte("\n"))
		}
	}
	// Insert the copyright notice
	var newComment string
	if lineSep != "\n" {
		notice = strings.ReplaceAll(notice, "\n", lineSep)
	}
	if lang.multiBegin != "" {
		newComment = lang.multiBegin + lineSep + notice + lineSep + lang.multiEnd + lineSep
	} else {
		newComment = lang.single + " " +
			strings.Join(strings.Split(notice, "\n"), lineSep+lang.single+" ") + lineSep
	}
	_, err = f.Write([]byte(newComment))
	if err != nil {
		return false, err
	}
	if fromLine == 0 && toLine == 0 && len(lines) > 0 && lines[0] != "" {
		f.Write([]byte(lineSep))
	}
	// Insert lines after the copyright notice to be replaced
	for i := toLine; i < len(lines); i++ {
		f.Write([]byte(lines[i]))
		if i < len(lines)-1 {
			f.Write([]byte("\n"))
		}
	}
	return true, nil
}

// firstComment returns the first multi-line comment it finds.
func firstComment(source string, lang CommentMarkers) (comment string, ok bool, fromLine int, toLine int) {
	lines := strings.Split(source, "\n")
	var inMulti, inSingle bool
	for l := 0; l < len(lines); l++ {
		line := lines[l]
		trimmedLine := strings.TrimSpace(line)
		switch {
		case !inSingle && !inMulti:
			if lang.multiBegin != "" && trimmedLine == lang.multiBegin {
				inMulti = true
				fromLine = l
			} else if lang.single != "" && strings.HasPrefix(trimmedLine, lang.single) {
				inSingle = true
				fromLine = l
			}
		case inSingle:
			if !strings.HasPrefix(trimmedLine, lang.single) {
				toLine = l
				for i := fromLine; i < toLine; i++ {
					lines[i] = strings.TrimPrefix(lines[i], lang.single)
					lines[i] = strings.TrimPrefix(lines[i], " ")
				}
				return strings.Join(lines[fromLine:toLine], "\n"), true, fromLine, toLine
			}
			if l == len(lines)-1 {
				toLine = len(lines)
				for i := fromLine; i < toLine; i++ {
					lines[i] = strings.TrimPrefix(lines[i], lang.single)
					lines[i] = strings.TrimPrefix(lines[i], " ")
				}
				return strings.Join(lines[fromLine:toLine], "\n"), true, fromLine, toLine
			}
		case inMulti:
			if trimmedLine == lang.multiEnd {
				toLine = l + 1
				return strings.Join(lines[fromLine+1:toLine-1], "\n"), true, fromLine, toLine
			}
		}
	}
	return "", false, fromLine, toLine
}
