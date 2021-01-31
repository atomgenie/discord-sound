package main

import (
	"discord-sound/player"
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
	case "player":
		player.Start()
	default:
		flag.Usage()
		return
	}

}
