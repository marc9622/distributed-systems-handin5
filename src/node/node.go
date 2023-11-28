package node

import (
	"fmt"
	"log"
	"net"

	pb "github.com/marc9622/distributed-systems-handin5/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Node struct {
    pb.UnimplementedActionServer
    port     uint
    allPorts []uint
}

func Spawn(port uint, allPorts []uint) Node {
    if len(allPorts) <= 1 {
        log.Fatalf("No nodes to connect to")
    }
    
    var node = Node {
        port: port,
        allPorts: allPorts,
    }

    // Setting up gRPC server
    go func() {
        var grpcServer = grpc.NewServer()

        pb.RegisterAuctionServer(grpcServer, &node)

        var list, listErr = net.Listen("tcp", fmt.Sprintf("localhost:%d", port));
        if listErr != nil {
            log.Fatalf("Failed to listen: %v", listErr)
        }
        defer list.Close()
    }()

    // Setting up gRPC client
    go func() {
        var opt = grpc.WithTransportCredentials(insecure.NewCredentials())
        var _ = opt
    }()

    return node
}

func (n *Node) Bid(amount uint) (bool, error) {
    return true, nil
}

func (n *Node) Result() (uint, error) {
    return 0, nil
}

func (n *Node) End() error {
    return nil
}

