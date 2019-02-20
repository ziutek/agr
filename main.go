package main

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"9fans.net/go/acme"
)

func die(info string) {
	io.WriteString(os.Stderr, info+"\n")
	os.Exit(1)
}

func checkErr(err error) {
	if err != nil {
		die(err.Error())
	}
}

func main() {
	winid := os.Getenv("winid")
	if winid == "" {
		die("$winid not defined")
	}
	id, err := strconv.ParseUint(winid, 10, 0)
	checkErr(err)

	win, err := acme.Open(int(id), nil)
	checkErr(err)

	tag, err := win.ReadAll("tag")
	checkErr(err)
	fpath := ""
	if fields := bytes.Fields(tag); len(fields) > 0 {
		fpath = string(fields[0])
		fname := filepath.Base(fpath)
		i := strings.LastIndexByte(fname, '.')
		if i == -1 || fname[i+1:] != "go" {
			die("not a .go file")
		}
	}

	_, _, err = win.ReadAddr()
	checkErr(err)
	checkErr(win.Ctl("mark\nnomark\naddr=dot\n"))
	a, b, err := win.ReadAddr()
	checkErr(err)
	width := b - a

	buf := make([]byte, 1024)
	n, err := win.Read("data", buf)
	m := 0
	for width > 0 && m < n {
		r, o := utf8.DecodeRune(buf[m:n])
		if unicode.IsLetter(r) {
			break
		}
		m += o
		width--
	}
	mode := "definition"
	if width > 0 {
		mode = "referrers"
		if m > 0 {
			mode = "describe"
			n -= m
		}
	}
	for err == nil {
		m, err = win.Read("data", buf)
		n += m
	}
	if err != io.EOF {
		checkErr(err)
	}
	_, err = win.Seek("body", 0, io.SeekStart)
	checkErr(err)
	// SeekEnd not supported so can't use: size, err = win.Seek("body", 0, io.SeekEnd)
	size := 0
	for err == nil {
		m, err = win.Read("body", buf)
		size += m
	}
	if err != io.EOF {
		checkErr(err)
	}
	start := int(size) - n

	//fmt.Printf("a=%d b=%d star=%d width=%d size=%d %s\n", a, b, start, width, size, buf)

	pos := fpath + ":#" + strconv.Itoa(start)
	if width > 0 {
		pos += ",#" + strconv.Itoa(start+width)
	}

	if len(os.Args) > 1 {
		cmd := exec.Command("guru", append(os.Args[1:], pos)...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		checkErr(cmd.Run())
		return
	}

	cmd := exec.Command("guru", mode, pos)
	cmd.Stderr = os.Stderr
	output, err := cmd.Output()
	checkErr(err)

	// TODO: reformat guru output for describe
	os.Stdout.Write(output)
}
