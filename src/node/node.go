package node

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	pb "github.com/marc9622/distributed-systems-handin5/src/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
    pb.UnimplementedAuctionServer
}

func (n *Node) Bid(_ context.Context, amount *pb.Amount) (*pb.Ack, error) {
    if ended {
        return &pb.Ack{Acknowledged: false}, nil
    }

    mutex.Lock()
    var isHigher = uint(amount.Amount) > highest

    // If the bidding amount is not the highest, then we can be sure the leader doesn't have a higher bid.
    if !isHigher {
        mutex.Unlock()
        return &pb.Ack{Acknowledged: false}, nil
    }

    for {
        findLeader()
        if isLeader {
            highest = uint(amount.Amount)
            mutex.Unlock()
            return &pb.Ack{Acknowledged: true}, nil
        }

        // Try to make bid.
        var ack, bidErr = leader.Bid(ctx, amount)
        if bidErr != nil {
            closeLeader()
            continue
        }
        
        // If the bid was rejected, that means the leaders highest was updated.
        if !ack.Acknowledged {
            var res, resErr = leader.Result(ctx, &pb.Void{})
            if resErr != nil {
                closeLeader()
                continue
            }
            highest = uint(res.Amount);
        }

        mutex.Unlock()
        return ack, nil
    }
}

func (n *Node) Result(_ context.Context, void *pb.Void) (*pb.Outcome, error) {
    mutex.Lock()
    for {
        findLeader()
        if isLeader {
            mutex.Unlock()
            return &pb.Outcome{Amount: uint32(highest)}, nil
        }

        var res, resErr = leader.Result(ctx, void)
        if resErr != nil {
            closeLeader()
            continue
        }

        highest = uint(res.Amount)
        mutex.Unlock()
        return res, nil
    }
}

func (n *Node) End(_ context.Context, void *pb.Void) (*pb.Void, error) {
    mutex.Lock()
    for {
        findLeader()
        if isLeader {
            ended = true
            log.Printf("Auction ended\n")
            mutex.Unlock()
            return &pb.Void{}, nil
        }

        var _, endErr = leader.End(ctx, void)
        if endErr != nil {
            closeLeader()
            continue
        }

        ended = true
        log.Printf("Auction ended\n")
        mutex.Unlock()
        return void, nil
    }
}

func (n *Node) Election(_ context.Context, void *pb.Void) (*pb.Void, error) {
}

func closeLeader() {
    conn.Close()
    conn, leader = nil, nil
}

func findLeader() {
    if leader != nil {
        return
    }

    for port := range ports {
        var connAttempt, connErr = grpc.Dial(fmt.Sprintf("localhost:%d", port))
        if connErr != nil {
            continue
        }

        conn = connAttempt
        leader = pb.NewAuctionClient(conn)
        break
    }

    if leader == nil {
        log.Panic("Not implemented")
    }
}

var port uint
var ports []uint
var ended = false
var highest uint = 0
var isLeader = false
var mutex = &sync.Mutex{}

var opt = grpc.WithTransportCredentials(insecure.NewCredentials())
var ctx = context.Background()
var conn *grpc.ClientConn
var leader pb.AuctionClient

func Spawn(_port uint, _ports []uint, seconds uint) {
    port = _port
    ports = _ports
    var n = Node{}

    // Setting up gRPC server
    go func() {
        var grpcServer = grpc.NewServer()

        pb.RegisterAuctionServer(grpcServer, &n)

        var list, listErr = net.Listen("tcp", fmt.Sprintf("localhost:%d", port));
        if listErr != nil {
            log.Panicf("Failed to listen: %v", listErr)
        }
        defer list.Close()
    }()

    // Setting up gRPC client
    go func() {
        var opt = grpc.WithTransportCredentials(insecure.NewCredentials())
        var _ = opt

    }()
}

