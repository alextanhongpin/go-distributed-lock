package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

func main() {
	client := newClient()
	defer client.Close()

	ctx := context.Background()
	lockKey := "my-key"
	lockVal := "my-val"

	dl := DistributedLock{
		Key:        lockKey,
		Value:      lockVal,
		etcdClient: client,
	}

	if err := checkKV(ctx, client, lockKey); err != nil {
		fmt.Printf("Error getting value: %s", err)
		os.Exit(1)
	}

	err := dl.Lock(ctx, 10)
	if err != nil {
		fmt.Printf("Error acquiring lock: %s", err)
		os.Exit(1)
	}

	if err := checkKV(ctx, client, lockKey); err != nil {
		fmt.Printf("Error getting value: %s", err)
		os.Exit(1)
	}

	// Simulate a critical section.
	time.Sleep(5 * time.Second)

	err = dl.Unlock(ctx)
	if err != nil {
		fmt.Printf("Error releasing lock: %s", err)
		os.Exit(1)
	}

	if err := checkKV(ctx, client, lockKey); err != nil {
		fmt.Printf("Error getting value: %s", err)
		os.Exit(1)
	}
}

func newClient() *clientv3.Client {
	endpoints := []string{"localhost:12379"}

	cfg := clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	}

	client, err := clientv3.New(cfg)
	if err != nil {
		fmt.Printf("Error connecting to etcd: %s", err)
		os.Exit(1)
	}

	return client
}

type DistributedLock struct {
	Key        string
	Value      string
	LeaseID    clientv3.LeaseID
	etcdClient *clientv3.Client
}

func (dl *DistributedLock) Lock(ctx context.Context, ttl int64) error {
	// Create a new lease with the specified TTL.
	lease, err := dl.etcdClient.Grant(ctx, ttl)
	if err != nil {
		return err
	}

	// Put the lock key-value pair into etcd with the lease attached.
	_, err = dl.etcdClient.Put(ctx, dl.Key, dl.Value, clientv3.WithLease(lease.ID))
	if err != nil {
		return err
	}

	// TODO: clientv3.KeepAlive

	dl.LeaseID = lease.ID
	log.Printf("Lock acquired: %s", dl.Key)

	return nil
}

func (dl *DistributedLock) Unlock(ctx context.Context) error {
	// Delete the lock key-value pair from etcd.
	// NOTE: The kv will be deleted automatically after the lease is revoked..
	/*
		_, err := dl.etcdClient.Delete(ctx, dl.Key)
		if err != nil {
			return err
		}
	*/

	// Revoke the lease.
	_, err := dl.etcdClient.Revoke(ctx, dl.LeaseID)
	if err != nil {
		return err
	}

	log.Printf("Lock released: %s", dl.Key)

	return nil
}

func checkKV(ctx context.Context, client *clientv3.Client, key string) error {
	val, err := client.Get(ctx, key)
	if err != nil {
		return err
	}

	if len(val.Kvs) == 0 {
		fmt.Println("no KV")
	}
	for _, kv := range val.Kvs {
		fmt.Printf("got kv: key=%s value=%s\n", kv.Key, kv.Value)
	}

	return nil
}
