package client

import (
	"context"
	"fmt"
	"time"

	pb "github.com/ranjan42/grassdb/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Client struct {
	peers []string
}

func NewClient(peers []string) *Client {
	return &Client{peers: peers}
}

func (c *Client) Set(key, value string) error {
	for _, peer := range c.peers {
		conn, err := grpc.NewClient(peer, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue // Try next peer
		}
		defer conn.Close()

		client := pb.NewDatabaseClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		resp, err := client.Set(ctx, &pb.SetRequest{Key: key, Value: value})
		if err == nil {
			if !resp.Success {
				// Not leader or other error, try next?
				// Actually server returns Success=false if not leader.
				// We should ideally check LeaderId redirect here.
				if resp.Error == "Not Leader" {
					continue
				}
				return fmt.Errorf("server error: %s", resp.Error)
			}
			return nil // Success
		}
		// RPC error (network, etc), try next
	}
	return fmt.Errorf("failed to set key on any node")
}

func (c *Client) Get(key string) (string, bool, error) {
	for _, peer := range c.peers {
		conn, err := grpc.NewClient(peer, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue
		}
		defer conn.Close()

		client := pb.NewDatabaseClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		resp, err := client.Get(ctx, &pb.GetRequest{Key: key})
		if err == nil {
			return resp.Value, resp.Found, nil
		}
	}
	return "", false, fmt.Errorf("failed to get key from any node")
}

func (c *Client) TakeSnapshot() error {
	for _, peer := range c.peers {
		conn, err := grpc.NewClient(peer, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			continue
		}
		defer conn.Close()

		client := pb.NewDatabaseClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_, err = client.TakeSnapshot(ctx, &pb.TakeSnapshotRequest{})
		if err == nil {
			return nil
		}
		// Try next if error (maybe not leader)
	}
	return fmt.Errorf("failed to take snapshot on any node")
}
