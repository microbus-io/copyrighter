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

// Package main runs a code generator that injects a copyright notice to source files.
package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// markers are the comment markers for a language.
type markers struct {
	single     string
	multiBegin string
	multiEnd   string
}

// languages is a map of the markers used to denote comments in each language.
var languages = map[string]markers{
	".go":   {"//", "/*", "*/"},
	".js":   {"//", "/*", "*/"},
	".ts":   {"//", "/*", "*/"},
	".cs":   {"//", "/*", "*/"},
	".java": {"//", "/*", "*/"},
	".c":    {"//", "/*", "*/"},
	".cpp":  {"//", "/*", "*/"},
	".php":  {"//", "/*", "*/"},
	".py":   {"#", "", ""},
	".css":  {"", "/*", "*/"},
	".xml":  {"", "<!--", "-->"},
	".html": {"", "<!--", "-->"},
	".yaml": {"#", "", ""},
	".ps1":  {"#", "<#", "#>"},
	".sh":   {"#", "", ""},
}

var (
	flagRecurse    bool
	flagVerbose    bool
	flagExclude    string
	flagExcludeMap map[string]bool
)

// main runs a code generator that injects a copyright notice to source files.
func main() {
	err := mainErr()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(1)
	}
}

// mainErr scans the current directory for a copyright.go file and applies the first comment
// in that file to all other source files in the directory.
func mainErr() error {
	// Parse CLI flags
	flag.BoolVar(&flagRecurse, "r", false, "Recurse sub-directories")
	flag.BoolVar(&flagVerbose, "v", false, "Verbose")
	flag.StringVar(&flagExclude, "x", "", "Comma-separated list of extensions to exclude")
	flag.Parse()
	flagExcludeMap = map[string]bool{}
	for _, x := range strings.Split(flagExclude, ",") {
		flagExcludeMap["."+strings.TrimPrefix(x, ".")] = true
	}
	if flagVerbose {
		fmt.Println("Copyrighter")
	}
	// Load the first found comment in copyright.go
	cwd, _ := os.Getwd()
	noticeToApply, ok, err := firstCommentInFile("copyright.go")
	if err != nil {
		return fmt.Errorf("unable to read '%s': %w", filepath.Join(cwd, "copyright.go"), err)
	}
	if !ok {
		return errors.New("no comment found in copyright.go")
	}
	// Apply the comment to the files in the current directory
	err = processDir(".", noticeToApply)
	if err != nil {
		return err
	}
	return nil
}

// processDir applies the copyright notice to the source files in the indicated directory.
func processDir(dirPath string, noticeToApply string) error {
	cwd, _ := os.Getwd()
	dirEntries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("unable to read files in '%s': %w", filepath.Join(cwd, dirPath), err)
	}
	for _, de := range dirEntries {
		if de.IsDir() {
			if flagRecurse {
				err = processDir(filepath.Join(dirPath, de.Name()), noticeToApply)
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
		firstComment, ok, err := firstCommentInFile(filepath.Join(dirPath, de.Name()))
		if err != nil {
			return err
		}
		if ok && firstComment == noticeToApply {
			continue
		}
		source, err := os.ReadFile(filepath.Join(dirPath, de.Name()))
		if err != nil {
			return fmt.Errorf("unable to read '%s': %w", filepath.Join(cwd, dirPath, de.Name()), err)
		}
		f, err := os.Create(filepath.Join(dirPath, de.Name()))
		if err != nil {
			return fmt.Errorf("unable to create '%s': %w", filepath.Join(cwd, dirPath, de.Name()), err)
		}
		err = func() error {
			defer f.Close()
			if lang.multiBegin != "" {
				_, err := f.WriteString(lang.multiBegin + "\n" +
					noticeToApply + "\n" +
					lang.multiEnd + "\n")
				if err != nil {
					return err
				}
			} else {
				_, err := f.WriteString(lang.single + " " +
					strings.Join(strings.Split(noticeToApply, "\n"), "\n"+lang.single+" ") + "\n",
				)
				if err != nil {
					return err
				}
			}
			if len(source) > 0 && source[0] != '\n' {
				_, err := f.WriteString("\n")
				if err != nil {
					return err
				}
			}
			_, err = f.Write(source)
			if err != nil {
				return err
			}
			return nil
		}()
		if err != nil {
			return fmt.Errorf("failed to overwrite '%s': %w", filepath.Join(cwd, dirPath, de.Name()), err)
		}
		if flagVerbose {
			fmt.Println("  " + filepath.Join(cwd, dirPath, de.Name()))
		}
	}
	return nil
}

// firstCommentInFile returns the first comment it finds in the first 1024 lines in a file.
func firstCommentInFile(filename string) (comment string, ok bool, err error) {
	ext := filepath.Ext(filename)
	lang, ok := languages[ext]
	if !ok {
		return "", false, nil
	}
	file, err := os.Open(filename)
	if err != nil {
		cwd, _ := os.Getwd()
		return "", false, fmt.Errorf("failed to open '%s': %w", filepath.Join(cwd, filename), err)
	}
	defer file.Close()
	return firstCommentInReader(file, lang)
}

// firstCommentInReader returns the first comment it finds in the first 1024 lines in a reader.
func firstCommentInReader(r io.Reader, lang markers) (comment string, ok bool, err error) {
	var aggregated strings.Builder
	var inMulti, inSingle, exit bool
	var lineNum int
	scanner := bufio.NewScanner(r)
	for !exit && lineNum < 1024 && scanner.Scan() {
		lineNum++
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)
		switch {
		case !inSingle && !inMulti:
			if lang.multiBegin != "" && strings.HasPrefix(trimmedLine, lang.multiBegin) {
				inMulti = true
				p := strings.LastIndex(trimmedLine, lang.multiEnd)
				if p > 0 {
					aggregated.WriteString(strings.TrimSpace(trimmedLine[len(lang.multiBegin):p]))
					exit = true
				} else {
					aggregated.WriteString(strings.TrimSpace(trimmedLine[len(lang.multiBegin):]))
				}
			} else if lang.single != "" && strings.HasPrefix(trimmedLine, lang.single) {
				inSingle = true
				aggregated.WriteString(strings.TrimPrefix(trimmedLine[len(lang.single):], " "))
			}
		case inSingle:
			if strings.HasPrefix(trimmedLine, lang.single) {
				if aggregated.Len() > 0 {
					aggregated.WriteString("\n")
				}
				aggregated.WriteString(strings.TrimPrefix(trimmedLine[len(lang.single):], " "))
			} else {
				exit = true
			}
		case inMulti:
			if strings.HasSuffix(trimmedLine, lang.multiEnd) {
				p := strings.LastIndex(line, lang.multiEnd)
				toAdd := strings.TrimRight(line[:p], " ")
				if toAdd != "" {
					aggregated.WriteString("\n")
					aggregated.WriteString(toAdd)
				}
				exit = true
			} else {
				if aggregated.Len() > 0 {
					aggregated.WriteString("\n")
				}
				aggregated.WriteString(strings.TrimRight(line, " "))
			}
		}
	}
	return aggregated.String(), inMulti || inSingle, nil
}
