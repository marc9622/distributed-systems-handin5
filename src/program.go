package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	node "github.com/marc9622/distributed-systems-handin5/src/node"
	pb "github.com/marc9622/distributed-systems-handin5/src/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
    var thisPort = flag.Uint("port", 0, "The port of this process.")
    var seconds = flag.Uint("time", 0, "The time, in seconds, before this process tells the other processes to end the auction. (0 means never)")
    var logFile = flag.String("log", "", "The file to print logs to. (Otherwise logs are printed to stdout)")
    var isNode = flag.Bool("node", false, "Whether makes this a node instead of client.")
    flag.Parse()

    if *isNode && *thisPort == 0 {
        fmt.Print("Port must be specified with -port\n")
        os.Exit(1)
    }

    // Setup logs
    if *logFile != "" {
        var file, fileErr = os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if fileErr != nil {
            log.Fatalf("Failed to create/open log file: %s\n", fileErr)
        }
        log.SetOutput(file)
    }

    /* Setup slice of ports */ {
        var portStrs = flag.Args();
        for _, portStr := range portStrs {
            var port uint

            var _, err = fmt.Sscanf(portStr, "%d", &port)
            if err != nil {
                log.Printf("Failed to parse port %s: %v\n", portStr, err)
                os.Exit(1)
            }
            if *isNode && *thisPort == port {
                continue
            }
            ports = append(ports, port)
            log.Printf("Added port %d\n", port)
        }
        if len(ports) <= 0 {
            log.Panicf("No nodes to connect to")
        }
    }

    if *isNode {
        ports = append(ports, *thisPort)
        sort.Slice(ports, func(i, j int) bool { return ports[i] < ports[j] })
        for i, port := range ports {
            if port == *thisPort {
                node.Spawn(uint(i), ports, *seconds)
            }
        }
        for {}
    } else {
        if *seconds != 0 {
            go func() {
                time.Sleep(time.Duration(*seconds) * time.Second)
                sendEnd()
            }()
        }

        var scanner = bufio.NewScanner(os.Stdin)
        scanner.Split(bufio.ScanWords)
        for {
            log.Printf("Waiting for input");
            scanner.Scan()
            var readErr = scanner.Err()
            if readErr != nil {
                log.Panicf("Failed to read input: %s", readErr)
            }

            var cmd = scanner.Text()
            if cmd == "bid" {
                scanner.Scan()
                var readErr = scanner.Err()
                if readErr != nil {
                    log.Panicf("Failed to read input: %s", readErr)
                }

                var amountStr = scanner.Text()
                var amount uint
                var _, parseErr = fmt.Sscanf(amountStr, "%d", &amount)
                if parseErr != nil {
                    log.Panicf("Failed to parse bidding amount %s: %v", amountStr, parseErr)
                }

                var success = sendBid(amount)
                if success {
                    log.Printf("Bid %d accepted\n", amount)
                } else {
                    log.Printf("Bid %d rejected\n", amount)
                }
            } else if cmd == "result" {
                var outcome = sendResult()
                log.Printf("Result: %d\n", outcome)
            } else if cmd == "end" {
                sendEnd()
                log.Printf("Auction ended\n")
            } else {
                log.Printf("Unknown command: %s\n", cmd)
            }
        }
    }
}

var ports = []uint{}
var opt = grpc.WithTransportCredentials(insecure.NewCredentials())
var ctx = context.Background()
var conn *grpc.ClientConn
var client pb.AuctionClient

func sendBid(amount uint) bool {
    for {
        findConnection()
        var ack, err = client.Bid(ctx, &pb.Amount{Amount: uint32(amount)})
        if err != nil {
            closeConnection()
            continue
        }

        return ack.Accepted
    }
}

func sendResult() uint {
    for {
        findConnection()
        var outcome, err = client.Result(ctx, &pb.Void{})
        if err != nil {
            closeConnection()
            continue
        }

        log.Printf("Outcome: %v\n", outcome)
        return uint(outcome.Amount)
    }
}

func sendEnd() {
    for {
        findConnection()
        var _, err = client.End(ctx, &pb.Void{})
        if err != nil {
            closeConnection()
            continue
        }

        return
    }
}

func closeConnection() {
    conn.Close()
    conn, client = nil, nil
}

func findConnection() {
    if client != nil {
        return
    }

    for _, port := range ports {
        var address = fmt.Sprintf("localhost:%d", port)
        log.Printf("Connecting to %s\n", address)
        var connAttempt, connErr = grpc.Dial(address, opt)
        if connErr != nil {
            log.Printf("Failed to connect to %s: %v\n", address, connErr)
        }
        conn = connAttempt
        client = pb.NewAuctionClient(conn)
        break
    }
    
    if client == nil {
        log.Panicf("Failed to connect to any node")
    }
}

