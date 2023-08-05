package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
)

func main() {
	client := newClient()
	defer client.Close()

	account := func(i int) string {
		return fmt.Sprintf("accts/%d", i)
	}

	// Setup accounts.
	totalAccounts := 5
	for i := 0; i < totalAccounts; i++ {
		k := account(i)
		if _, err := client.Put(context.Background(), k, "100"); err != nil {
			log.Fatal(err)
		}
	}

	exchange := func(stm concurrency.STM) error {
		from, to := rand.Intn(totalAccounts), rand.Intn(totalAccounts)
		if from == to {
			// Nothing to do
			return nil
		}

		fromK, toK := account(from), account(to)
		fromV, toV := stm.Get(fromK), stm.Get(toK)
		fromInt, err := strconv.Atoi(fromV)
		if err != nil {
			return err
		}

		toInt, err := strconv.Atoi(toV)
		if err != nil {
			return err
		}

		// Transfer amount.
		xfer := fromInt / 2
		fromInt, toInt = fromInt-xfer, toInt+xfer

		// Write back.
		stm.Put(fromK, fmt.Sprint(fromInt))
		stm.Put(toK, fmt.Sprint(toInt))
		return nil
	}

	// Concurrently exchange values between accounts.
	var wg sync.WaitGroup
	wg.Add(10)

	for i := 0; i < 10; i++ {
		go func() {
			defer wg.Done()

			if _, err := concurrency.NewSTM(client, exchange); err != nil {
				log.Fatal(err)
			}
		}()
	}

	wg.Wait()

	// Confirm account sum matches sum from beginning.
	sum := 0
	accts, err := client.Get(context.Background(), "accts/", clientv3.WithPrefix())
	if err != nil {
		log.Fatal(err)
	}

	for _, kv := range accts.Kvs {
		n, err := strconv.Atoi(string(kv.Value))
		if err != nil {
			panic(err)
		}
		sum += n
	}

	fmt.Println("Account sum is:", sum)
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
