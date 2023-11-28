package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
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

    if *logFile != "" {
        var file, fileErr = os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
        if fileErr != nil {
            log.Fatalf("Failed to create/open log file: %s\n", fileErr)
        }
        log.SetOutput(file)
    }

    if *thisPort == 0 {
        fmt.Print("Port must be specified with -port\n")
        os.Exit(1)
    }

    var portStrs = flag.Args();
    for _, portStr := range portStrs {
        var port uint

        var _, err = fmt.Sscanf(portStr, "%d", &port)
        if err != nil {
            fmt.Printf("Failed to parse port %s: %v\n", portStr, err)
            os.Exit(1)
        }
        if *thisPort != port {
            ports = append(ports, port)
        }
    }
    if len(ports) <= 0 {
        log.Panicf("No nodes to connect to")
    }

    if *isNode {
        node.Spawn(*thisPort, ports, *seconds)
        for {}
    }

    if *seconds != 0 {
        go func() {
            time.Sleep(time.Duration(*seconds) * time.Second)
            sendEnd()
        }()
    }

    var scanner = bufio.NewScanner(os.Stdin)
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
        }

        if cmd == "result" {
            var outcome = sendResult()
            log.Printf("Result: %d\n", outcome)
        }

        if cmd == "end" {
            sendEnd()
            log.Printf("Auction ended\n")
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

        return ack.Ack
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

    for port := range ports {
        var connAttempt, connErr = grpc.Dial(fmt.Sprintf("localhost:%d", port))
        if connErr != nil {
            continue
        }
        conn = connAttempt
        client = pb.NewAuctionClient(conn)
        break
    }
    
    if client == nil {
        log.Panicf("Failed to connect to any node")
    }
}

