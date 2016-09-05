package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/unixpickle/neuralspell"
	"github.com/unixpickle/serializer"
	"github.com/unixpickle/sgd"
	"github.com/unixpickle/speechrecog/ctc"
	"github.com/unixpickle/weakai/neuralnet"
	"github.com/unixpickle/weakai/rnn"
)

const (
	SpellType     = "spell"
	PronounceType = "pronounce"

	TrainingDataFrac  = 0.8
	CostBatchSize     = 100
	TrainingBatchSize = 100
)

func main() {
	if len(os.Args) != 4 {
		dieUsage()
	}
	netType := os.Args[1]
	if netType != SpellType && netType != PronounceType {
		dieUsage()
	}

	log.Println("Loading dictionary...")
	dictionary, err := neuralspell.ReadDictionary(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to read dictionary:", err)
		os.Exit(1)
	}
	dictionary.InputPhones = (netType == SpellType)

	log.Println("Loading/creating network...")
	var network rnn.SeqFunc
	netData, err := ioutil.ReadFile(os.Args[3])
	if err == nil {
		netObj, err := serializer.DeserializeWithType(netData)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Failed to deserialize network:", err)
			os.Exit(1)
		}
		var ok bool
		network, ok = netObj.(rnn.SeqFunc)
		if !ok {
			fmt.Fprintf(os.Stderr, "Type is not an rnn.SeqFunc: %T\n", network)
			os.Exit(1)
		}
	} else {
		if netType == SpellType {
			network = neuralspell.NewSpeller()
		} else {
			network = neuralspell.NewPronouncer()
		}
	}

	log.Println("Training...")
	trainNetwork(netType == SpellType, network, dictionary)

	log.Println("Saving network...")
	serialized, err := serializer.SerializeWithType(network.(serializer.Serializer))
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to serialize:", err)
		os.Exit(1)
	}
	if err := ioutil.WriteFile(os.Args[3], serialized, 0755); err != nil {
		fmt.Fprintln(os.Stderr, "Failed to save network:", err)
		os.Exit(1)
	}
}

func trainNetwork(inPhones bool, net rnn.SeqFunc, samples sgd.SampleSet) {
	gradienter := &sgd.Adam{
		Gradienter: &ctc.RGradienter{
			Learner:        net.(sgd.Learner),
			SeqFunc:        net,
			MaxConcurrency: 2,
			MaxSubBatch:    TrainingBatchSize/2 + (TrainingBatchSize % 2),
		},
	}
	toggleDropout(net, true)

	sgd.ShuffleSampleSet(samples)
	training := samples.Copy().Subset(0, int(float64(samples.Len())*TrainingDataFrac))
	validation := samples.Copy().Subset(training.Len(), samples.Len())

	var epoch int
	sgd.SGDInteractive(gradienter, training, 1e-3, TrainingBatchSize, func() bool {
		toggleDropout(net, false)
		cost := ctc.TotalCost(net, training, CostBatchSize, 2)
		crossCost := ctc.TotalCost(net, validation, CostBatchSize, 2)
		toggleDropout(net, true)
		log.Printf("Epoch %d: cost=%e cross=%e", epoch, cost, crossCost)
		epoch++
		return true
	})
	toggleDropout(net, false)
}

func toggleDropout(net rnn.SeqFunc, dropout bool) {
	bd := net.(*rnn.Bidirectional)
	output := bd.Output.(*rnn.NetworkSeqFunc).Network[0].(*neuralnet.DropoutLayer)
	output.Training = dropout
}

func dieUsage() {
	fmt.Fprintln(os.Stderr, "Usage:", os.Args[0], "<spell | pronounce> dictionary.txt out_net")
	os.Exit(1)
}
