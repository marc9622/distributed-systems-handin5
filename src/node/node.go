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
    //mutex.Lock()
    log.Printf(">>> Bid: %d\n", amount.Amount)

    if ended {
        return &pb.Ack{Accepted: false}, nil
    }

    var isHigher = uint(amount.Amount) > highest

    // If the bidding amount is not the highest, then we can be sure the leader doesn't have a higher bid.
    if !isHigher {
        //mutex.Unlock()
        log.Printf("<<< Bid: %d\n", amount.Amount)
        return &pb.Ack{Accepted: false}, nil
    }

    for {
        findLeader()
        if isLeader {
            highest = uint(amount.Amount)
            //mutex.Unlock()
            log.Printf("<<< Bid: %d\n", amount.Amount)
            return &pb.Ack{Accepted: true}, nil
        }

        // Try to make bid.
        var ack, bidErr = leader.Bid(ctx, amount)
        if bidErr != nil {
            closeLeader()
            continue
        }
        
        // If the bid was rejected, that means the leaders highest was updated.
        if !ack.Accepted {
            var res, resErr = leader.Result(ctx, &pb.Void{})
            if resErr != nil {
                closeLeader()
                continue
            }
            highest = uint(res.Amount);
        }

        //mutex.Unlock()
        log.Printf("<<< Bid: %d\n", amount.Amount)
        return ack, nil
    }
}

func (n *Node) Result(_ context.Context, void *pb.Void) (*pb.Outcome, error) {
    //mutex.Lock()
    log.Printf(">>> Result\n")

    for {
        findLeader()
        if isLeader {
            //mutex.Unlock()
            log.Printf("<<< Result\n")
            return &pb.Outcome{Amount: uint32(highest)}, nil
        }

        var res, resErr = leader.Result(ctx, void)
        if resErr != nil {
            closeLeader()
            continue
        }

        highest = uint(res.Amount)
        //mutex.Unlock()
        log.Printf("<<< Result\n")
        return res, nil
    }
}

func (n *Node) End(_ context.Context, void *pb.Void) (*pb.Void, error) {
    //mutex.Lock()
    log.Printf(">>> End\n")

    for {
        findLeader()
        if isLeader {
            ended = true
            log.Printf("<<< End\n")
            //mutex.Unlock()
            return &pb.Void{}, nil
        }

        var _, endErr = leader.End(ctx, void)
        if endErr != nil {
            closeLeader()
            continue
        }

        ended = true
        log.Printf("<<< End\n")
        //mutex.Unlock()
        return void, nil
    }
}

func (n *Node) Election(_ context.Context, void *pb.Void) (*pb.Void, error) {
    //mutex.Lock()
    log.Printf(">>> Election\n")
    findLeader()
    log.Printf("<<< Election\n")
    //mutex.Unlock()
    return &pb.Void{}, nil
}

func (n *Node) Leader(_ context.Context, id *pb.Id) (*pb.Void, error) {
    //mutex.Lock()
    log.Printf(">>> Leader: %d\n", id.Id)

    isLeader = false
    closeLeader()

    var address = fmt.Sprintf("localhost:%d", ports[id.Id])
    var connAttempt, connErr = grpc.Dial(address, opt)
    if connErr != nil {
        //mutex.Unlock()
        log.Printf("<<< Leader: %d\n", id.Id)
        return &pb.Void{}, nil
    }
    conn = connAttempt

    leader = pb.NewAuctionClient(conn)

    //mutex.Unlock()
    log.Printf("<<< Leader: %d\n", id.Id)
    return &pb.Void{}, nil
}

func becomeLeader() {
    isLeader = true;
    log.Printf("I am the leader\n")

    for i := 0; i < int(id); i++ {
        var address = fmt.Sprintf("localhost:%d", ports[i])
        var conn, connErr = grpc.Dial(address, opt)
        if connErr != nil {
            continue
        }

        pb.NewAuctionClient(conn).Leader(ctx, &pb.Id{Id: uint32(id)}) // Ignore errors: we continue either way.
        conn.Close()
    }
}

func closeLeader() {
    if leader != nil {
        conn, leader = nil, nil
    }
}

func findLeader() {
    log.Printf(">>> Find leader\n")
    if isLeader || leader != nil {
        log.Printf("<<< Find leader\n")
        return
    }

    var count = uint(len(ports))

    // If this has the highest port...
    if ports[count-1] == ports[id] {
        becomeLeader()
        log.Printf("<<< Find leader\n")
        return
    }

    // Otherwise call for election to all higher ports
    var responded = false
    for i := count-1; i > id; i-- {
        var address = fmt.Sprintf("localhost:%d", ports[i])
        var conn, connErr = grpc.Dial(address, opt)
        if connErr != nil {
            continue
        }

        var _, elecErr = pb.NewAuctionClient(conn).Election(ctx, &pb.Void{})
        if elecErr != nil {
            continue
        }

        responded = true
    }

    // If no one responded, then this is the leader.
    if !responded {
        becomeLeader()
    }

    log.Printf("<<< Find leader\n")
    return
}

var id uint
var ports []uint
var ended = false
var highest uint = 0
var isLeader = false
var mutex = &sync.Mutex{}

var opt = grpc.WithTransportCredentials(insecure.NewCredentials())
var ctx = context.Background()
var conn *grpc.ClientConn
var leader pb.AuctionClient

func Spawn(_id uint, _ports []uint, seconds uint) {
    id = _id
    ports = _ports
    var n = Node{}

    go func() {
        var grpcServer = grpc.NewServer()

        pb.RegisterAuctionServer(grpcServer, &n)

        var address = fmt.Sprintf("localhost:%d", ports[id])
        log.Printf("Node %d listening on port %s\n", id, address)
        var list, listErr = net.Listen("tcp", address)
        if listErr != nil {
            log.Panicf("Failed to listen: %v", listErr)
        }
        defer list.Close()

        grpcServer.Serve(list)
        log.Printf("A")
        defer grpcServer.Stop()
    }()
}

