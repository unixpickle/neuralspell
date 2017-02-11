# neuralspell

Spelling reveals a lot about humanity in general: it's messy, inconsistent, and self-contradictory. Despite all of this, people have various mental heuristics that help them spell or pronounce new words. This is the behavior I want to instill in a neural network.

The goal of this project is to train recurrent neural networks on two tasks:

 * **Spelling:** converting phonetic transcriptions to an English spelling.
 * **Pronunciation:** converting English spellings to phonetic transcriptions.

# Usage

First, you should install and configure [Go](https://golang.org/doc/install). Make sure your GOPATH is setup. Next, download the code as follows:

```
$ go get -d -u github.com/unixpickle/neuralspell/...
```

The repository includes sub-directories for various commands. The [train](#Training) command is the first thing you will want to use. After that, you can use the [eval](#Feeding-input) command to run the network on a new word or phonetic transcription.

## Training

```
$ cd $GOPATH/src/github.com/unixpickle/neuralspell/train
$ go run *.go
```

This will train a new network on the spelling task. Pressing Ctrl+C will gracefully stop training and save the result to a file called `out_net`. Make sure not to press Ctrl+C more than once. To train on the pronunciation task, add `-task pronounce`:

```
$ go run *.go -task pronounce
```

I recommend training for at least an hour per task. See `-help` for more information on training.

## Feeding input

To evaluate the network on new inputs, you can use the `eval` command. Note that you must already have [trained](#Training) the network on the task at hand.

```
$ cd $GOPATH/src/github.com/unixpickle/neuralspell/eval
$ go run *.go -input dɔg
dog
$ go run *.go -input dog -task spell
dɔg
```

If you have not trained the network, the command will probably take a long time to run. This is because, to decode the output of the network, `eval` uses a technique called prefix search decoding. If the network was not trained very much, prefix search decoding has to search a vast array of possible decodings.

# How it works

The technical side of the project is fairly unoriginal. I use an architecture similar to one that might be used for [neural speech recognition](http://www.cs.toronto.edu/~graves/icml_2006.pdf). In this architecture, a bidirectional-RNN "reads" the input and produces a labeling using Connectionist Temporal Classification.

There were other implementation routes I could have taken. I think it would be particularly interesting to have two "encoder" networks that convert spelling/phonetics to a universal vector representation, and then have two "decoder" networks turn said representation into spelling/phonetics.
