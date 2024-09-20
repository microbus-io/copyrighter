/*
Copyright 2023-2024 Microbus LLC and various contributors

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
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// commentMarkers are the comment markers of a language.
type commentMarkers struct {
	single     string
	multiBegin string
	multiEnd   string
}

// patternMatcher is a file matching pattern.
type patternMatcher struct {
	Op  string
	Exp regexp.Regexp
}

// languages is a map of the markers used to denote comments in each language.
var languages = map[string]commentMarkers{
	".bazel": {"#", "", ""},
	".c":     {"//", "/*", "*/"},
	".cpp":   {"//", "/*", "*/"},
	".cs":    {"//", "/*", "*/"},
	".css":   {"", "/*", "*/"},
	".go":    {"//", "/*", "*/"},
	".html":  {"", "<!--", "-->"},
	".java":  {"//", "/*", "*/"},
	".js":    {"//", "/*", "*/"},
	".php":   {"//", "/*", "*/"},
	".ps1":   {"#", "<#", "#>"},
	".py":    {"#", "", ""},
	".sh":    {"#", "", ""},
	".sql":   {"--", "", ""},
	".tf":    {"#", "/*", "*/"},
	".ts":    {"//", "/*", "*/"},
	".xml":   {"", "<!--", "-->"},
	".yaml":  {"#", "", ""},
	".yml":   {"#", "", ""},
}

var (
	flagVerbose bool
)

// main runs a code generator that injects a copyright notice to source files.
func main() {
	// Parse CLI flags
	flag.BoolVar(&flagVerbose, "v", false, "Verbose")
	flag.Parse()

	err := mainErr()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

// mainErr applies a copyright notice to all subdirectories as indicated by the copyright.go file (if present).
func mainErr() error {
	// Load the first comment found in copyright.go
	b, err := os.ReadFile("copyright.go")
	if err != nil {
		return fmt.Errorf("unable to read copyright.go: %w", err)
	}
	source := string(b)
	notice, ok, _, _ := firstComment(source, languages[".go"])
	if !ok {
		return fmt.Errorf("no comment found in copyright.go")
	}
	notice = strings.ReplaceAll(notice, "YYYY", strconv.Itoa(time.Now().Year()))

	// Parse the file matching patterns
	patterns := []patternMatcher{}
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		op := ""
		if strings.HasPrefix(line, "// + ") {
			op = "+"
		} else if strings.HasPrefix(line, "// - ") {
			op = "-"
		}
		if op == "" {
			continue
		}
		p := ""
		for i, r := range []rune(line[5:]) {
			if i == 0 {
				if r == '/' {
					p += `^`
				} else {
					p += `\/`
				}
			}
			if r == '/' {
				p += `\/`
			} else if r == '*' {
				p += `.*`
			} else {
				p += regexp.QuoteMeta(string(r))
			}
		}
		p += "$"
		patterns = append(patterns, patternMatcher{
			Op:  op,
			Exp: *regexp.MustCompile(p),
		})
	}

	// Apply the comment to the files in all subdirectories
	err = processDir(".", notice, patterns)
	if err != nil {
		return err
	}
	return nil
}

// processDir applies the copyright notice to the source files in the indicated directory.
func processDir(dirPath string, notice string, patterns []patternMatcher) error {
	// Skip subdirectories that contain their own copyright.go file
	if dirPath != "." {
		b, err := os.ReadFile(filepath.Join(dirPath, "copyright.go"))
		if err == nil && bytes.Contains(b, []byte("github.com/microbus-io/copyrighter")) {
			if flagVerbose {
				fmt.Println(dirPath + " (skipped)")
			}
			return nil
		}
	}
	if flagVerbose {
		fmt.Println(dirPath)
	}
	// Iterate over files
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("unable to read files in '%s': %w", dirPath, err)
	}
	subDirs := []fs.DirEntry{}
	for _, de := range dirEntries {
		fileName := filepath.Join(dirPath, de.Name())
		// Determine if to process
		ignore := false
		for _, p := range patterns {
			if p.Exp.MatchString("/" + fileName) {
				if p.Op == "-" {
					ignore = true
				}
				if p.Op == "+" {
					ignore = false
				}
			}
		}
		if fileName == "copyright.go" {
			ignore = true
		}
		if ignore {
			if flagVerbose {
				fmt.Printf("  %-32s (ignored)\n", de.Name())
			}
			continue
		}
		// Collect sub directories
		if de.IsDir() {
			subDirs = append(subDirs, de)
			continue
		}
		// Only process known languages
		ext := filepath.Ext(de.Name())
		lang, ok := languages[ext]
		if !ok {
			if flagVerbose {
				fmt.Printf("  %-32s (disregarded)\n", de.Name())
			}
			continue
		}
		source, err := os.ReadFile(fileName)
		if err != nil {
			return err
		}
		var toWrite bytes.Buffer
		ok, err = process(bytes.NewReader(source), &toWrite, lang, notice)
		if err != nil {
			return fmt.Errorf("failed to process '%s': %w", fileName, err)
		}
		if ok {
			if flagVerbose {
				fmt.Printf("  %-32s (copyrighted)\n", de.Name())
			}
			err = os.WriteFile(fileName, toWrite.Bytes(), 0666)
			if err != nil {
				return fmt.Errorf("failed to write back '%s': %w", fileName, err)
			}
		} else {
			if flagVerbose {
				fmt.Printf("  %-32s (unchanged)\n", de.Name())
			}
		}
	}
	// Recurse into sub directories
	for _, de := range subDirs {
		err = processDir(filepath.Join(dirPath, de.Name()), notice, patterns)
		if err != nil {
			return err
		}
	}
	return nil
}

// process reads the source code from the reader, inserts the copyright notice if appropriate,
// and writes the results to the writer.
func process(r io.Reader, f io.Writer, lang commentMarkers, notice string) (ok bool, err error) {
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
func firstComment(source string, lang commentMarkers) (comment string, ok bool, fromLine int, toLine int) {
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
