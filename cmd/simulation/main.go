package main

import (
	"bytes"
	"fmt"
	"io"
	"kademlia/internal/adapters"
	"kademlia/internal/core/entities"
	"kademlia/internal/core/kademlia"
	"kademlia/internal/logging"
	"math/rand/v2"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

func runNode(node kademlia.Kademlia, knownContact *entities.Address) {
	err := node.Run(knownContact)
	if err != nil {
		panic("failed to run node")
	}
}

func nodeAddressFromIndex(index int) entities.Address {
	return entities.Address{
		IP:   "127.0.0.1",
		Port: 8000 + index,
	}
}

func waitUntilReady(nodes []kademlia.Kademlia) {
	timeout := time.After(20 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		var ready atomic.Bool
		ready.Store(true)

		var waitGroup sync.WaitGroup
		for i := range nodes {
			waitGroup.Go(func() {
				_, err := nodes[0].Ping(nodeAddressFromIndex(i))
				if err != nil {
					ready.Store(false)
				}
			})
		}
		waitGroup.Wait()

		if ready.Load() {
			return
		}

		select {
		case <-timeout:
			panic("timed out waiting for all nodes to start")
		case <-ticker.C:
		}
	}
}

func experiment1(numberNodes int) int {
	var out bytes.Buffer
	logging.InitLogger(&out)

	network := adapters.NewMockNetworkAdapter(0.0)
	nodes := make([]kademlia.Kademlia, numberNodes)

	for i := range numberNodes {
		nodes[i] = kademlia.NewKademlia(nodeAddressFromIndex(i), network)
	}

	go runNode(nodes[0], nil)
	time.Sleep(100 * time.Millisecond)
	defer nodes[0].Quit()

	address := nodeAddressFromIndex(0)
	for i := 1; i < numberNodes; i++ {
		go runNode(nodes[i], &address)
		defer nodes[i].Quit()
	}

	waitUntilReady(nodes)

	for range 1000 {
		_, _, err := nodes[rand.IntN(numberNodes)].GetValue(
			entities.NewRandomKademliaID().String(),
		)
		if err != nil {
			panic("failed to get")
		}
	}

	count := 0
	for line := range strings.SplitSeq(out.String(), "\n") {
		if strings.Contains(line, `"msg":"Receive FIND_VALUE RPC"`) {
			count++
		}
	}
	return count
}

func experiment2(numberNodes int, packetLossPercentage float32) float64 {
	const numberKeys = 200
	const numberLookups = 1000

	logging.InitLogger(io.Discard)

	network := adapters.NewMockNetworkAdapter(0.0)
	nodes := make([]kademlia.Kademlia, numberNodes)

	for i := range numberNodes {
		nodes[i] = kademlia.NewKademlia(nodeAddressFromIndex(i), network)
	}

	go runNode(nodes[0], nil)
	time.Sleep(100 * time.Millisecond)
	defer nodes[0].Quit()

	address := nodeAddressFromIndex(0)
	for i := 1; i < numberNodes; i++ {
		go runNode(nodes[i], &address)
		defer nodes[i].Quit()
	}

	waitUntilReady(nodes)

	keys := make([]string, numberKeys)
	values := make([]string, numberKeys)
	var putGroup sync.WaitGroup
	for i := range numberKeys {
		putGroup.Go(func() {
			values[i] = fmt.Sprintf("data %d", i)
			key, err := nodes[rand.IntN(numberNodes)].Put(values[i])
			if err != nil {
				panic("failed to put")
			}
			keys[i] = key.String()
		})
	}
	putGroup.Wait()

	network.SetPacketLossPercentage(packetLossPercentage)

	var successes atomic.Int64
	var getGroup sync.WaitGroup
	for range numberLookups {
		getGroup.Go(func() {
			k := rand.IntN(numberKeys)
			data, _, err := nodes[rand.IntN(numberNodes)].GetValue(keys[k])
			if err == nil && data != nil && *data == values[k] {
				successes.Add(1)
			}
		})
	}
	getGroup.Wait()

	return float64(successes.Load()) / numberLookups
}

func main() {
	// fmt.Printf("N;Probes\n")
	// for _, numberNodes := range []int{2, 5, 10, 20, 50, 100, 200, 500, 1000} {
	// 	fmt.Printf("%d;%d\n", numberNodes, experiment1(numberNodes))
	// }

	fmt.Printf("P;SuccessRate\n")
	for _, p := range []float32{0.0, 0.1, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 1.0} {
		fmt.Printf("%.2f;%.4f\n", p, experiment2(100, p))
	}
}
