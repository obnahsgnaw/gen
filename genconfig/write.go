package genconfig

import (
	"bytes"
	"fmt"
	"golang.org/x/tools/imports"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
)

func render(tmpl string, wr io.Writer, data interface{}) error {
	t, err := template.New(tmpl).Parse(tmpl)
	if err != nil {
		return err
	}
	return t.Execute(wr, data)
}

func output(fileName string, content []byte) error {
	result, err := imports.Process(fileName, content, nil)
	if err != nil {
		lines := strings.Split(string(content), "\n")
		errLine, _ := strconv.Atoi(strings.Split(err.Error(), ":")[1])
		startLine, endLine := errLine-5, errLine+5
		fmt.Println("Format fail:", errLine, err)
		if startLine < 0 {
			startLine = 0
		}
		if endLine > len(lines)-1 {
			endLine = len(lines) - 1
		}
		for i := startLine; i <= endLine; i++ {
			fmt.Println(i, lines[i])
		}
		return fmt.Errorf("cannot format file: %w", err)
	}
	return os.WriteFile(fileName, result, 0640)
}

func write(tmpl string, filename string, data interface{}) (writed string, err error) {
	var buf bytes.Buffer

	if filename, err = filepath.Abs(filename); err != nil {
		return
	}

	dir, _ := filepath.Split(filename)

	if _, err = os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			return
		}
	}

	if err = render(tmpl, &buf, data); err != nil {
		return
	}

	if err = output(filename, buf.Bytes()); err != nil {
		return
	}

	writed = filename
	return
}
