package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	network "github.com/veltor/veltor-network"
)

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	done := make(chan struct{})
	var nodes []*network.Node
	sub := make(chan interface{})
	for i := 0; i < 16; i++ {
		book := network.NewSimpleBook([]string{"127.0.0.1:10000"})
		addr := fmt.Sprintf("127.0.0.1:%v", 10000)
		log1 := network.NewSimpleLog(fmt.Sprintf("node-%v", i))
		node := network.NewNode(
			network.SetLog(log1),
			network.SetBook(book),
			network.SetSubscriber(sub),
			network.SetListen(true),
			network.SetAddress(addr),
			network.SetMinPeers(4),
			network.SetMaxPeers(16),
		)
		nodes = append(nodes, node)
	}
	go func() {
	Loop:
		for {
			select {
			case <-done:
				break Loop
			case <-time.After(time.Millisecond * 100):
				msg := strconv.FormatUint(uint64(rand.Uint32()), 10)
				log.Printf("sending message: %v", msg)
				node := nodes[rand.Int()%len(nodes)]
				err := node.Send("127.0.0.1:10000", msg)
				if err != nil {
					log.Printf("message send failed: %v", err)
				}
			case packet := <-sub:
				msg := packet.(*network.Packet).Message.(string)
				log.Printf("received message: %v", msg)
			}
		}
	}()
	<-sig
	close(done)
	for _, node := range nodes {
		_ = node
	}
	os.Exit(0)
}
