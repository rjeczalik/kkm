package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/rjeczalik/kkm"
)

var (
	card = flag.String("card", "UJ", "Student card owner - two letter university name. By default UJ.")
	id   = flag.Int("id", 0, "Student card ID. Required")
)

var avail []string

func die(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}

func init() {
	avail = make([]string, 0, len(kkm.CityCardType))
	for k := range kkm.CityCardType {
		avail = append(avail, strings.ToUpper(k))
	}
	sort.Strings(avail)
}

func main() {
	flag.CommandLine.Usage = func() {
		flag.PrintDefaults()
		fmt.Printf("\n Available university acronyms: %v\n", strings.Join(avail, ", "))
		os.Exit(0)
	}
	if len(os.Args) == 1 {
		flag.CommandLine.Usage()
	}
	flag.Parse()
	if *id <= 0 {
		die("invalid student card ID")
	}
	tickets, err := kkm.History(*card, *id)
	if err != nil {
		die(err)
	}
	p, err := json.Marshal(tickets)
	if err != nil {
		die(err)
	}
	var buf bytes.Buffer
	if err = json.Indent(&buf, p, "", "\t"); err != nil {
		die(err)
	}
	buf.WriteTo(os.Stdout)
}
