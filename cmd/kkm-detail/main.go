package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/rjeczalik/kkm"
)

var (
	id    = flag.Int("id", 0, "Student card ID suffixed with two-digit university code.")
	kkmid = flag.Int("kkm", 0, "KKM card ID.")
)

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func main() {
	flag.Parse()
	if *id <= 0 || *kkmid <= 0 {
		die("invalid values for -id and/or -kkm flags")
	}
	detail, err := kkm.Details(*id, *kkmid)
	if err != nil {
		die(err)
	}
	p, err := json.Marshal(detail)
	if err != nil {
		die(err)
	}
	var buf bytes.Buffer
	if err = json.Indent(&buf, p, "", "\t"); err != nil {
		die(err)
	}
	buf.WriteTo(os.Stdout)
	fmt.Println()
}
