package main

import (
	"fmt"
	"math/rand"
	"crypto/sha1"
	"net"
	"encoding/binary"
	"bytes"
	bencode "github.com/jackpal/bencode-go"
)

type NNode struct{
	Nid []byte
	Ip net.IP
	Port int
}

type KTable struct{
	Nid []byte
	nodes []KNode
}

type KNode struct{
	Nid []byte
	Ip net.IP
	Port int
}

type Dht struct {
	BindPort int
	BindIp net.IP
	BindAddr net.UDPAddr

	IsServerWorking bool
	IsClientWorking bool
	Table KTable
	
	Connection *net.UDPConn
}

var (
	BootStrapNodes = []net.UDPAddr { }
	TidLength = 4
	ReJoinDhtInterval = 10
	ThreadNumber = 3
)

func initialLoger(){
	
}

func entropy(length int) string {
	chars := make([]byte, length)
	for i:=0; i<length; i++ {
		chars = append(chars, byte(rand.Intn(255)))
	}
	return string(chars)
}

func randomId() [20]byte{
	return sha1.Sum([]byte(entropy(20)))
}

func inet_ntoa(bytes []byte) net.IP {  
	return net.IPv4(bytes[3],bytes[2],bytes[1],bytes[0])
}

func decodeNodes(nodes []byte) []NNode{
	length := len(nodes)
	maxNodeCount := length / 26
	n := make([]NNode, 0, maxNodeCount)
	if length % 26 != 0 {
		return n
	}
	for i:=0; i<length; i+=26 {
		node := NNode{}
		copy(node.Nid, nodes[i:i+20])
		node.Ip = inet_ntoa(nodes[i+20:i+24])
		portBytes := []byte{nodes[i+26], nodes[i+25]}
		binary.Read(bytes.NewBuffer(portBytes), binary.BigEndian, &node.Port)
		
		n = append(n, node)
	}

	return n
}

func getNeighbor(target [20]byte, end int) [20]byte{
	ret := [20]byte{}
	rnd := randomId()
	for i:=0; i<end; i++ {
		ret[i] = target[i]
	}
	for i:=end; i<20; i++ {
		ret[i] = rnd[i]
	}
	return ret
}

func (v *KTable) put(node KNode){
	v.nodes = append(v.nodes, node)
}

func (v *Dht) Start(){
	v.BindAddr.IP = v.BindIp
	v.BindAddr.Port = v.BindPort

	var err error
	v.Connection, err = net.ListenUDP("udp", &v.BindAddr)
	
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Connect OK!")
	}

	go v.Server()
}

func (v *Dht) joinDht(){
	for address := range BootStrapNodes {
		v.sendFindNode(BootStrapNodes[address], nil)
	}
}

func (v *Dht) Server(){
	v.joinDht()
	
	for {
		if v.IsServerWorking == false {
			break
		}
		
		buff := []byte{}
		length, remoteAddr, err := v.Connection.ReadFromUDP(buff)
		if remoteAddr != nil {
			fmt.Println("Recv From %s", remoteAddr.String())
		}
		if err != nil {
			fmt.Println("Recv Err %s", err.Error())
		}
		if length <= 0 {
			fmt.Println("Recv From 0")
		}
		msg, err := bencode.Decode(bytes.NewBuffer(buff))

		v.onMessage(msg, *remoteAddr)
	}
}

func (v *Dht) onMessage(msg interface{}, address net.UDPAddr){

}

func (v *Dht) sendFindNode(address net.UDPAddr, nid *[20]byte){
	newNid := [20]byte{}
	if nid == nil {
		newNid = v.Table.Nid
	} else {
		newNid = getNeighbor(*nid, 10)
	}

	tid := entropy(TidLength)
	
	

	v.sendKrpc(msg, address)
}

func main() {
	bitTorrentAddr,err := net.ResolveUDPAddr("udp", "router.bittorrent.com:6881")
	transmissionBtAddr,err := net.ResolveUDPAddr("udp", "dht.transmissionbt.com:6881")
	uTorrentAddr,err := net.ResolveUDPAddr("udp", "router.utorrent.com:6881")
	if err != nil {
		fmt.Println(err.Error())
	}

	BootStrapNodes = append(BootStrapNodes, *bitTorrentAddr)
	BootStrapNodes = append(BootStrapNodes, *transmissionBtAddr)
	BootStrapNodes = append(BootStrapNodes, *uTorrentAddr)

	v := Dht{}
	v.Start()
}
