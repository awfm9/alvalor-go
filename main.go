// Copyright (c) 2017 The Alvalor Authors
//
// This file is part of Alvalor.
//
// Alvalor is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Alvalor is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with Alvalor.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"time"

	"github.com/dgraph-io/badger"
	"github.com/rs/zerolog"
	"github.com/spf13/pflag"

	"github.com/alvalor/alvalor-go/blockchain"
	"github.com/alvalor/alvalor-go/codec"
	"github.com/alvalor/alvalor-go/finder"
	"github.com/alvalor/alvalor-go/kv"
	"github.com/alvalor/alvalor-go/network"
	"github.com/alvalor/alvalor-go/node"
	"github.com/alvalor/alvalor-go/store"
	"github.com/alvalor/alvalor-go/types"
)

func main() {

	// set up a channel to catch system signals to cleanly shut down
	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt)

	// initialize default configuration
	cfg := DefaultConfig

	// apply the command line parameters to configuration
	pflag.Uint16Var(&cfg.Port, "port", 21517, "listen port for incoming connections")
	pflag.Parse()

	// seed the random generator
	rand.Seed(time.Now().UnixNano())

	// initialize the logger to standard error as output
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"
	zerolog.MessageFieldName = "message"
	zerolog.TimestampFunc = time.Now().UTC
	log := zerolog.New(os.Stderr).With().Timestamp().Logger().Level(zerolog.InfoLevel)

	// determine current user directory
	user, err := user.Current()
	if err != nil {
		log.Fatal().Err(err).Msg("could not get current user")
	}
	dir := filepath.Join(user.HomeDir, ".alvalor")

	// use our efficient capnproto codec for network communication
	codec := codec.NewProto()

	// create channel to pipe messages from network layer to node layer
	sub := make(chan interface{}, 128)

	// initialize the network component to create our p2p network node
	address := fmt.Sprintf("%v:%v", cfg.IP, cfg.Port)
	net := network.New(log, codec, sub,
		network.SetListen(cfg.Listen),
		network.SetAddress(address),
		network.SetMinPeers(4),
		network.SetMaxPeers(16),
	)

	// add own address & bootstrapping nodes
	net.Add(address)
	for _, address := range cfg.Bootstrap {
		net.Add(address)
	}

	// make sure the database directory exists
	dbDir := filepath.Join(dir, "database")
	err = os.MkdirAll(dbDir, os.ModePerm)
	if err != nil {
		log.Fatal().Err(err).Msg("could not create database directory")
	}

	// initialize the key-value database
	opts := badger.DefaultOptions
	opts.Dir = dbDir
	opts.ValueDir = dbDir
	db, err := badger.Open(opts)
	if err != nil {
		log.Fatal().Err(err).Msg("could not open database")
	}

	// create the wrapper around badger & initialize entity stores
	kv := kv.NewBadger(db)
	blocks := store.New(kv, codec, "b")
	txs := store.New(kv, codec, "t")
	chain, err := blockchain.New(kv, kv, blocks, txs)
	if err != nil {
		log.Fatal().Err(err).Msg("could not initialize blockchain")
	}

	// initialize the path finder
	find := finder.New(chain.Header().Hash())

	// initialize the node subscriber
	n := node.New(log, net, chain, find, codec, sub)

	// wait for a stop signal to initialize shutdown
	stats := time.NewTicker(10 * time.Second)
	gen := time.NewTicker(2 * time.Second)

Loop:
	for {
		select {
		case <-sig:
			signal.Stop(sig)
			close(sig)
			break Loop
		case <-stats.C:
			net.Stats()
			n.Stats()
		case <-gen.C:
			tx := generateTransaction()
			n.Submit(tx)
		}
	}

	// shut down the p2p network node
	net.Stop()
}

func generateTransaction() *types.Transaction {

	// determine the composition of the transaction
	numTransfers := rand.Int()%3 + 1
	numFees := rand.Int()%3 + 1
	numData := rand.Int() % 4 * 1024

	// create the transfers
	transfers := make([]*types.Transfer, 0, numTransfers)
	for i := 0; i < numTransfers; i++ {
		transfer := generateTransfer()
		transfers = append(transfers, transfer)
	}

	// create the fees
	fees := make([]*types.Fee, 0, numFees)
	for i := 0; i < numFees; i++ {
		fee := generateFee()
		fees = append(fees, fee)
	}

	// create the data block
	data := make([]byte, numData)
	_, _ = rand.Read(data)

	// initialize transaction
	tx := &types.Transaction{
		Transfers: transfers,
		Fees:      fees,
		Data:      data,
	}

	return tx
}

func generateTransfer() *types.Transfer {
	from := make([]byte, 32)
	_, _ = rand.Read(from)
	to := make([]byte, 32)
	_, _ = rand.Read(to)
	amount := rand.Uint64()
	return &types.Transfer{
		From:   from,
		To:     to,
		Amount: amount,
	}
}

func generateFee() *types.Fee {
	from := make([]byte, 32)
	_, _ = rand.Read(from)
	amount := rand.Uint64()
	return &types.Fee{
		From:   from,
		Amount: amount,
	}
}
