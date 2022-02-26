# RU Flood

Inspired by https://vug.pl/takeRussiaDown.html, a DDoS attack on russian state-owned or -controlled media, banks, etc.
This is a version compiled to a separate program so it run without a browser and possibly more efficient.
Also allows scaling the power (see Advanced usage).

## Download

Head over to [Releases](https://github.com/zegkljan/ruflood/releases).

## Build it yourself

You only need to do this if you don't trust the precompiled binaries and/or you want to modify the code.

1. Install [Go](https://go.dev/).
2. Clone/download this repository.
3. Navigate to the folder where you have downloaded it, and run `go build`.

## Usage

Just run the program from the command line, default parameters and targeted websites* will be used.

### Advanced usage

```
ruflood [-c | --max-concurrent-requests c] [-r | --request-timeout r] [-i | --print-interval i] [targets...]
```

`targets` is a whitespace separated list of targeted websites. If no target is specified, a default list is used.

`-c | --max-concurrent-requests c` sets the maximum number of concurrently running web requests. Must be > 0. Default is 1000.

`-r | --request-timeout r` sets the timeout of individual requests, in milliseconds. 0 (zero) will turn timeout off, i.e. the program will wait for a response indefinitely. Must be >= 0. Default is 1000.

`-i | --print-interval i` sets the interval at which overall statistics about the flood is printed out. 0 (zero) turns the printing off completely. Must be >= 0. Default is 1000.

---
\* all websites from https://vug.pl/takeRussiaDown.html plus https://rkn.gov.ru/ - Roskomnadzor - a russian bureau for media oversight, known for basically censoring media that do not fall in line with state-dictated narrative.