# neuralspell

Spelling reveals a lot about humanity: it's messy, inconsistent, and self-contradictory. Despite all of this, people have mental heuristics for spelling and pronouncing new words. I want to see how well neural networks can learn those same heuristics.

The goal of this project is to train recurrent neural networks on two tasks:

 * **Spelling:** converting phonetic transcriptions to an English spelling.
 * **Pronunciation:** converting English spellings to phonetic transcriptions.

# Results

After training for one epoch, here are some results:

<table>
  <tr>
    <th>Word</th>
    <th>Pronunciation</th>
    <th>Network Spelling</th>
    <th>Network Pronunciation</th>
  </tr>
  <tr>
    <td>invader</td>
    <td>InveIdəɹ</td>
    <td>invader</td>
    <td>Invedəɹ</td>
  </tr>
  <tr>
    <td>twelve</td>
    <td>twɛlv</td>
    <td>twelve</td>
    <td>twɛlv</td>
  </tr>
  <tr>
    <td>evaluate</td>
    <td>IvæljueIt</td>
    <td>evaluate</td>
    <td>ɛvʌlueIt</td>
  </tr>
  <tr>
    <td>guilty</td>
    <td>gIlti</td>
    <td>gilty</td>
    <td>gIlti</td>
  </tr>
</table>

# Usage

First, you should install and configure [Go](https://golang.org/doc/install). Make sure your GOPATH is setup.

If you don't want to setup Go yourself, you can use the [Docker](https://www.docker.com) image for Go. It has everything that you will need:

```
$ docker run -it golang:1.7 /bin/bash
```

## Downloading the code

Next, download the code as follows:

```
$ go get -d -u github.com/unixpickle/neuralspell/...
```

The repository includes sub-directories for various commands. The [train](#training) command is the first thing you will want to use. After that, you can use the [eval](#feeding-input) command to run the network on a new word or phonetic transcription.

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

To evaluate the network on new inputs, you can use the `eval` command. Note that you must already have [trained](#training) the network on the task at hand.

```
$ cd $GOPATH/src/github.com/unixpickle/neuralspell/eval
$ go run *.go -phonetics dɔg
Spelling: dog
$ go run *.go -spelling dog -task spell
Pronunciation: dag
```

If you have not trained the network, the command will probably take a long time to run. This is because, to decode the output of the network, `eval` uses a technique called prefix search decoding. If the network was not trained very much, prefix search decoding has to search a vast array of possible decodings.

# How it works

The technical side of the project is fairly unoriginal. I use an architecture similar to one that might be used for [neural speech recognition](http://www.cs.toronto.edu/~graves/icml_2006.pdf). In this architecture, a bidirectional-RNN "reads" the input and produces a labeling using Connectionist Temporal Classification.

There were other implementation routes I could have taken. I think it would be particularly interesting to have two "encoder" networks that convert spelling/phonetics to a universal vector representation, and then have two "decoder" networks turn said representation into spelling/phonetics.
