package main

import (
	"flag"
	"fmt"
	"github.com/kazhmir/gna"
	"io"
	"sync"
	"time"
)

var host *string
var read *bool
var loop *bool
var interval time.Duration
var keepAlive time.Duration
var latency time.Duration

func main() {
	between := flag.Int("i", 20, "Interval between packets in milliseconds")
	minions := flag.Int("m", 1, "Number of minions")
	read = flag.Bool("r", false, "Print server response.")
	host = flag.String("h", "localhost:8888", "Host address")
	noClose := flag.Int("alive", 0, "Keep connection alive for specified seconds after sending packages")
	loop = flag.Bool("l", false, "Loops infinetely while sending packets.")
	ping := flag.Int("ping", 0, "Client Ping in milliseconds")
	flag.Parse()
	interval = time.Millisecond * time.Duration(*between)
	keepAlive = time.Second * time.Duration(*noClose)
	latency = time.Millisecond * time.Duration(*ping)
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Fuck you")
		return
	}

	var wg sync.WaitGroup
	for i := 0; i < *minions; i++ {
		go func() {
			defer wg.Done()
			(&minion{}).start(args)
		}()
		wg.Add(1)
	}
	wg.Wait()
}

type minion struct {
	cli  *gna.Client
	buff []byte
	i, n int
}

func (m *minion) start(data []string) {
	var err error
	m.cli, err = gna.Dial(*host)
	Error(err)
	defer m.cli.Close()
	var nOfPkts int
	m.buff = make([]byte, 64)
	for {
		for i := range data {
			go func() {
				fmt.Println("sending: ", data[i])
				time.Sleep(latency)
				err := m.cli.Send(data[i])
				Error(err)
				nOfPkts++
			}()
			time.Sleep(interval)
		}
		if *read {
			for nOfPkts > 0 {
				dt, err := m.cli.Recv()
				if err == io.EOF {
					return
				}
				Error(err)
				s, ok := dt.(string)
				if ok {
					fmt.Printf("Recv: %s\n", s)
				}
				nOfPkts--
			}
		}
		nOfPkts = 0
		if *loop {
			time.Sleep(interval)
			continue
		}
		break
	}
	time.Sleep(keepAlive)
}

func Error(err error) {
	if err != nil {
		panic(err)
	}
}
