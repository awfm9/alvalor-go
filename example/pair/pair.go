package main

import (
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	network "github.com/veltor/veltor-network"
	"github.com/veltor/veltor-network/proto"
)

func main() {
	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT)
	done := make(chan struct{})
	addr := "127.0.0.1:10000"
	sub := make(chan interface{})
	codec := proto.Codec{}
	log1 := network.NewSimpleLog("A")
	node1 := network.NewNode(
		network.SetLog(log1),
		network.SetCodec(codec),
		network.SetSubscriber(sub),
		network.SetListen(true),
		network.SetAddress(addr),
		network.SetMinPeers(1),
		network.SetMaxPeers(1),
	)
	book := network.NewSimpleBook()
	book.Add("127.0.0.1:10000")
	log2 := network.NewSimpleLog("B")
	node2 := network.NewNode(
		network.SetLog(log2),
		network.SetCodec(codec),
		network.SetBook(book),
		network.SetSubscriber(sub),
		network.SetMinPeers(1),
		network.SetMaxPeers(1),
	)
	go func() {
	Loop:
		for {
			select {
			case <-done:
				break Loop
			case <-time.After(2 * time.Second):
				msg := strconv.FormatUint(uint64(rand.Uint32()), 10)
				log.Printf("sending message: %v", msg)
				err := node2.Send("127.0.0.1:10000", msg)
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
	_ = node1
	_ = node2
	os.Exit(0)
}
