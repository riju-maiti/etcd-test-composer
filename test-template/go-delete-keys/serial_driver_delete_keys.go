package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/antithesishq/antithesis-sdk-go/assert"
	"github.com/antithesishq/antithesis-sdk-go/random"
	clientv3 "go.etcd.io/etcd/client/v3"
)

func Connect() *clientv3.Client {
	hosts := [][]string{[]string{"etcd0:2379"}, []string{"etcd1:2379"}, []string{"etcd2:2379"}}
	host := random.RandomChoice(hosts)
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   host,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to connect to etcd: %v", err)
		// Antithesis Assertion: client should always be able to connect to an etcd host
		assert.Unreachable("Client failed to connect to an etcd host", map[string]interface{}{"host": host, "error": err})
		os.Exit(1)
	}
	return cli
}

func DeleteKeys() {
	ctx := context.Background()

	// Connect to an etcd node
	cli := Connect()

	// Get all keys
	resp, err := cli.Get(ctx, "", clientv3.WithPrefix())

	// Antithesis Assertion: sometimes get requests are successful. A failed request is OK since we expect them to happen.
	assert.Sometimes(err == nil, "Client got all keys", map[string]interface{}{"error": err})
	cli.Close()

	if err != nil {
		log.Printf("Client failed to get all keys: %v", err)
		os.Exit(0)
	}

	// Choose half of the keys
	var keys []string
	for _, k := range resp.Kvs {
		keys = append(keys, string(k.Key))
	}
	half := len(keys) / 2
	half_keys := keys[:half]

	// Connect to a new etcd node
	cli = Connect()

	// Delete half of the keys chosen
	var deleted_keys []string
	for _, k := range half_keys {
		_, err := cli.Delete(ctx, k)
		assert.Sometimes(err == nil, "Client deleted a key", map[string]interface{}{"error": err})
		if err != nil {
			log.Printf("Failed to delete key %s: %v", k, err)
		} else {
			log.Printf("Successfully deleted key %v", k)
			deleted_keys = append(deleted_keys, k)
		}
	}

	// Close the client
	cli.Close()

	// Connect to a new client
	cli = Connect()

	// Check to see if those keys were deleted / exist
	for _, k := range deleted_keys {
		resp, err := cli.Get(ctx, k)
		assert.Sometimes(err == nil, "Client got a key", map[string]interface{}{"error": err})
		if err != nil {
			log.Printf("Client failed to get key %s: %v", k, err)
			continue
		}
		assert.Always(resp.Count == 0, "Key was deleted correctly", map[string]interface{}{"key": k})
	}

	// Close the client
	cli.Close()

	assert.Reachable("Completion of a key deleting check", nil)
	log.Printf("Completion of a key deleting check")
}

func main() {
	DeleteKeys()
}
