package main

import (
	"encoding/hex"
	"flag"
	"fmt"
	"log"

	"github.com/micromdm/nanodep/albc"
)

// overridden by -ldflags -X
var version = "unknown"

func main() {
	var (
		flRaw     = flag.String("raw", "", "hex-encoded raw bypass code")
		flCode    = flag.String("code", "", "dash-separated \"human readable\" bypass code")
		flVersion = flag.Bool("version", false, "print version")
	)
	flag.Parse()

	if *flVersion {
		fmt.Println(version)
		return
	}

	var (
		err error
		bc  albc.BypassCode
	)

	if *flRaw != "" && *flCode != "" {
		log.Fatal("cannot specify both raw and code")
	} else if *flRaw == "" && *flCode == "" {
		bc, err = albc.New()
		if err != nil {
			log.Fatal(err)
		}
	} else if *flRaw != "" {
		b, err := hex.DecodeString(*flRaw)
		if err != nil {
			log.Fatal(err)
		}
		bc, err = albc.NewFromBytes(b)
		if err != nil {
			log.Fatal(err)
		}
	} else if *flCode != "" {
		bc, err = albc.NewFromCode(*flCode)
		if err != nil {
			log.Fatal(err)
		}
	}

	raw := hex.EncodeToString(bc[:])

	code, err := bc.Code()
	if err != nil {
		log.Fatal(err)
	}

	hash, err := bc.Hash()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("%s  raw\n%s  code\n%s  hash\n", raw, code, hash)
}
