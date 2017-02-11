package main

import (
	"flag"
	"fmt"

	"github.com/unixpickle/essentials"
	"github.com/unixpickle/neuralspell"
	"github.com/unixpickle/serializer"
)

func main() {
	var netPath string
	var input string
	var task string

	flag.StringVar(&netPath, "net", "../train/out_net", "network file path")
	flag.StringVar(&input, "input", "", "input spelling or phonetics")
	flag.StringVar(&task, "task", "spell", "task ('spell' or 'pronounce')")

	flag.Parse()

	if input == "" {
		essentials.Die("missing -input flag (see -help for more)")
	} else if task != "spell" && task != "pronounce" {
		essentials.Die("unknown task:", task)
	}

	var net *neuralspell.Network
	if err := serializer.LoadAny(netPath, &net); err != nil {
		essentials.Die(err)
	}

	method := net.Spell
	if task != "spell" {
		method = net.Pronounce
	}
	out, err := method(input)
	if err != nil {
		essentials.Die(err)
	} else {
		fmt.Println(out)
	}
}
