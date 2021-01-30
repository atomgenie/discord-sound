package main

import (
	"discord-sound/youtubedl"
	"flag"
)

func main() {

	flag.Parse()

	args := flag.Args()

	if len(args) == 0 {
		flag.Usage()
		return
	}

	switch args[0] {
	case "youtubedl":
		youtubedl.Start()
	default:
		flag.Usage()
		return
	}

}
