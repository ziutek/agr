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

	if b == 0 {
		return
	}
	if len(os.Args) > 1 {
		cmd := exec.Command("guru", os.Args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		checkErr(cmd.Run())
		return
	}

	mode := "definition"
	pos := fpath + ":#" + strconv.Itoa(a)
	if b != a {
		_, err = win.Seek("body", int64(a), 0)
		checkErr(err)
		buf := make([]byte, b-a)
		_, err = win.Read("body", buf)
		checkErr(err)
		i := 0
		for i < len(buf) {
			r, n := utf8.DecodeRune(buf[i:])
			if unicode.IsLetter(r) {
				break
			}
			i += n
		}
		if i > 0 {
			mode = "referrers"
			a += i
		}
		pos += ",#" + strconv.Itoa(b)
	}
	cmd := exec.Command("guru", mode, pos)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	checkErr(cmd.Run())
}
