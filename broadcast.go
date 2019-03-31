package main

import (
	"bytes"
        "context"
//	"crypto/rand"
	"fmt"
	"time"

	host "github.com/libp2p/go-libp2p-host"
	libp2p "github.com/libp2p/go-libp2p"
//        crypto	"github.com/libp2p/go-libp2p-crypto"
	pstore "github.com/libp2p/go-libp2p-peerstore"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
//	multiaddr "github.com/multiformats/go-multiaddr"
)

const gossipSubID = "/meshsub/1.0.0"


/*
func newHost (port int) (host.Host) {
	r := rand.Reader
        prvKey, _, err := crypto.GenerateKeyPairWithReader(crypto.RSA, 2048, r)
	if err != nil {
		panic(err)
	}

	// 0.0.0.0 will listen on any interface device.
	sourceMultiAddr, _ := multiaddr.NewMultiaddr(fmt.Sprintf("/ip4/0.0.0.0/tcp/%d", port))

	// libp2p.New constructs a new libp2p Host.
	// Other options can be added here.
	host, err := libp2p.New(
		context.Background(),
		libp2p.ListenAddrs(sourceMultiAddr),
		libp2p.Identity(prvKey),
	)

	if err != nil {
		panic(err)
	}
        return host
} */

func main() {

	//golog.SetAllLoggers(gologging.DEBUG) // Change to DEBUG for extra info
	h1 := newHost(2001)
	h2 := newHost(2002)
	h3 := newHost(2003)
	fmt.Printf("host 1: \n\t-Addr:%s\n\t-ID: %s\n", h1.Addrs()[0], h1.ID().Pretty())
	fmt.Printf("host 2: \n\t-Addr:%s\n\t-ID: %s\n", h2.Addrs()[0], h2.ID().Pretty())
	fmt.Printf("host 3: \n\t-Addr:%s\n\t-ID: %s\n", h3.Addrs()[0], h3.ID().Pretty())

	time.Sleep(100 * time.Millisecond)

	// add h1 to h2's store
	h2.Peerstore().AddAddr(h1.ID(), h1.Addrs()[0], pstore.PermanentAddrTTL)
	// add h2 to h1's store
	h1.Peerstore().AddAddr(h2.ID(), h2.Addrs()[0], pstore.PermanentAddrTTL)
	// add h2 to h3's store
	h3.Peerstore().AddAddr(h2.ID(), h2.Addrs()[0], pstore.PermanentAddrTTL)
	// add h3 to h2's store
	h2.Peerstore().AddAddr(h3.ID(), h3.Addrs()[0], pstore.PermanentAddrTTL)

	// ---- gossip sub part
	topic := "random"
	opts := pubsub.WithMessageSigning(false)
	g1, err := pubsub.NewGossipSub(context.Background(), h1, opts)
	requireNil(err)
	g2, err := pubsub.NewGossipSub(context.Background(), h2, opts)
	requireNil(err)
	g3, err := pubsub.NewGossipSub(context.Background(), h3, opts)
	requireNil(err)
	//s1, err := g2.Subscribe(topic)
	//requireNil(err)
	s2, err := g2.Subscribe(topic)
	requireNil(err)
	s3, err := g3.Subscribe(topic)
	requireNil(err)

        // 1 connect to 2 and 2 connect to 3
	err = h1.Connect(context.Background(), h2.Peerstore().PeerInfo(h2.ID()))
	requireNil(err)
	err = h2.Connect(context.Background(), h3.Peerstore().PeerInfo(h3.ID()))
	requireNil(err)

	time.Sleep(1 * time.Second)
	// publish and read
	msg := []byte("Hello World")
	requireNil(g1.Publish(topic, msg))

	pbMsg, err := s2.Next(context.Background())
	requireNil(err)
	checkEqual(msg, pbMsg.Data)
	fmt.Println(" GOSSIPING WORKS #1")

	pbMsg, err = s3.Next(context.Background())
	requireNil(err)
	checkEqual(msg, pbMsg.Data)
	fmt.Println(" GOSSIPING WORKS #2")
}


func checkEqual(exp, rcvd []byte) {
	if !bytes.Equal(exp, rcvd) {
		panic("not equal")
	}
}

func requireNil(err error) {
	if err != nil {
		panic(err)
	}
}

func newHost(port int) host.Host {
	opts := []libp2p.Option{
		libp2p.ListenAddrStrings(fmt.Sprintf("/ip4/127.0.0.1/tcp/%d", port)),
		libp2p.DisableRelay(),
	}
	basicHost, err := libp2p.New(context.Background(), opts...)
	if err != nil {
		panic(err)
	}
	return basicHost
}
