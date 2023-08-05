package main

import (
	"fmt"
	"log"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

// etcd comes with the concurrency package.
func main() {
	client := newClient()
	defer client.Close()

	sess, err := concurrency.NewSession(client)
	if err != nil {
		log.Fatalf("Error starting new session: %s", err)
	}
	defer sess.Close()

	locker := concurrency.NewLocker(sess, "/my/key")
	locker.Lock()

	fmt.Println("Locked")

	// Simulate critical session.
	time.Sleep(5 * time.Second)

	locker.Unlock()
	fmt.Println("Unlocked")
}

func newClient() *clientv3.Client {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"localhost:12379"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Error connecting to etcd: %s", err)
	}

	return client
}
