package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/unixpickle/anynet/anyctc"
	"github.com/unixpickle/anynet/anysgd"
	"github.com/unixpickle/anyvec"
	"github.com/unixpickle/essentials"
	"github.com/unixpickle/neuralspell"
	"github.com/unixpickle/rip"
	"github.com/unixpickle/serializer"
)

func main() {
	var netFile string
	var task string
	var dataFile string
	var stepSize float64
	var validation float64
	var batchSize int

	flag.StringVar(&netFile, "out", "out_net", "network file path")
	flag.StringVar(&task, "task", "spell", "task ('spell' or 'pronounce')")
	flag.StringVar(&dataFile, "data", "../dict/cmudict-IPA.txt", "dictionary path")
	flag.Float64Var(&stepSize, "step", 0.001, "SGD step size")
	flag.Float64Var(&validation, "validation", 0.1, "validation fraction")
	flag.IntVar(&batchSize, "batch", 128, "SGD batch size")

	flag.Parse()
	if task != "spell" && task != "pronounce" {
		essentials.Die("unknown task:", task)
	}

	log.Println("Loading dictionary...")
	dictionary, err := neuralspell.ReadDictionary(dataFile)
	if err != nil {
		essentials.Die(err)
	}
	dictionary.InputPhones = (task == "spell")

	rand.Seed(time.Now().UnixNano())
	validationSet, trainingSet := anysgd.HashSplit(dictionary, validation)
	log.Println("Loaded", trainingSet.Len(), "training and", validationSet.Len(),
		"validation samples.")

	var net *neuralspell.Network
	if err = serializer.LoadAny(netFile, &net); err != nil {
		log.Println("Creating new network...")
		net = neuralspell.NewNetwork()
	} else {
		log.Println("Loaded network.")
	}

	log.Println("Training...")
	bidir := net.Speller
	if !dictionary.InputPhones {
		bidir = net.Pronouncer
	}
	trainer := &anyctc.Trainer{
		Func:    bidir.Apply,
		Params:  bidir.Parameters(),
		Average: true,
	}
	var iter int
	sgd := &anysgd.SGD{
		Fetcher:     trainer,
		Gradienter:  trainer,
		Transformer: &anysgd.Adam{},
		Samples:     trainingSet,
		Rater:       anysgd.ConstRater(stepSize),
		BatchSize:   batchSize,
		StatusFunc: func(b anysgd.Batch) {
			if validationSet.Len() > 0 {
				log.Printf("iter %d: cost=%v validation=%v", iter, trainer.LastCost,
					crossValidate(trainer, validationSet, batchSize))
			} else {
				log.Printf("iter %d: cost=%v", iter, trainer.LastCost)
			}
			iter++
		},
	}
	err = sgd.Run(rip.NewRIP().Chan())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	log.Println("Saving network...")
	if err := serializer.SaveAny(netFile, net); err != nil {
		essentials.Die(err)
	}
}

func crossValidate(t *anyctc.Trainer, s anysgd.SampleList, batch int) anyvec.Numeric {
	anysgd.Shuffle(s)
	bs := batch
	if bs > s.Len() {
		bs = s.Len()
	}
	samples := s.Slice(0, bs)
	b, _ := t.Fetch(samples)
	return anyvec.Sum(t.TotalCost(b.(*anyctc.Batch)).Output())
}
