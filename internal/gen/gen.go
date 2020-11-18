package main

import (
	"bytes"
	"encoding/csv"
	"flag"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

var output = flag.String("output", "dates.go", "output path")

func main() {
	if err := _main(); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("created %s\n", *output)
}

func _main() error {
	var src io.Reader
	if len(os.Args) == 1 {
		u := "https://www8.cao.go.jp/chosei/shukujitsu/syukujitsu.csv"
		req, err := http.NewRequest("GET", u, nil)
		if err != nil {
			return err
		}
		req.Header.Set("User-Agent", "https://github.com/soh335/shukujitsu")

		fmt.Printf("download %s\n", u)
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		if res.StatusCode != http.StatusOK {
			return fmt.Errorf("status code:%d", res.StatusCode)
		}
		src = res.Body
	} else {
		f, err := os.Open(os.Args[1])
		if err != nil {
			return err
		}
		defer f.Close()
		src = f
	}

	r := csv.NewReader(transform.NewReader(src, japanese.ShiftJIS.NewDecoder()))
	// drop first line
	if _, err := r.Read(); err != nil {
		return err
	}

	var buf bytes.Buffer
	buf.WriteString(`// Code generated by internal/gen/gen.go; DO NOT EDIT.`)
	buf.WriteString("\npackage shukujitsu")
	buf.WriteString("\nvar dates = map[string]string{")
	for {
		records, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		l := fmt.Sprintf("\n\"%s\": \"%s\",", records[0], records[1])
		buf.WriteString(l)
	}
	buf.WriteString("\n}")

	formated, err := format.Source(buf.Bytes())
	if err != nil {
		return err
	}
	return ioutil.WriteFile(*output, formated, 0644)
}
