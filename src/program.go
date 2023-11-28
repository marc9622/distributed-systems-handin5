package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	node "github.com/marc9622/distributed-systems-handin5/src/node"
)

func main() {
    var port = flag.Uint("port", 0, "The port of this process")
    var seconds = flag.Uint("time", 0, "The time, in seconds, before this process tells the other processes to end the auction. (0 means never)")
    flag.Parse()

    if *port == 0 {
        log.Fatal("Port must be specified with -port")
    }

    var ports = []uint{}
    {
        var portStrs = flag.Args();
        for _, portStr := range portStrs {
            var port uint

            var _, err = fmt.Sscanf(portStr, "%d", &port)
            if err != nil {
                log.Fatalf("Failed to parse port %s: %v", portStr, err)
            }
            ports = append(ports, port)
        }
    }

    var n = node.Spawn(*port, ports)
    if *seconds != 0 {
        go func() {
            time.Sleep(time.Duration(*seconds) * time.Second)
            var err = n.End()
            if err != nil {
                log.Panicf("Failed to end auction: %v", err)
            }
        }()
    }

    var reader = bufio.NewReader(os.Stdin)
    for {
        var cmd, readErr = reader.ReadString(' ')
        if readErr != nil {
            log.Panicf("Failed to read input: %s", readErr)
        }

        if cmd == "bid" {
            var amountStr, readErr = reader.ReadString(' ')
            if readErr != nil {
                log.Panicf("Failed to read input: %s", readErr)
            }

            var amount uint
            var _, parseErr = fmt.Sscanf(amountStr, "%d", &amount)
            if parseErr != nil {
                log.Fatalf("Failed to parse bidding amount %s: %v", amountStr, parseErr)
            }

            var success, bidErr = n.Bid(amount)
            if bidErr != nil {
                log.Fatalf("Failed to make bidding: %v", bidErr)
            }
            
            if success {
                log.Printf("Bid %d accepted\n", amount)
            } else {
                log.Printf("Bid %d rejected\n", amount)
            }
        }

        if cmd == "result" {
            var outcome, resultErr = n.Result()
            if resultErr != nil {
                log.Fatalf("Failed to get result: %v", resultErr)
            }

            log.Printf("Result: %s\n", outcome)
        }
    }
}

