package cache

import (
	"fmt"
	"net/rpc"
)

const (
	Address = "/tmp/passman.sock" // Socket path
	Network = "unix"              // Unix domain socket
	ChecksumSize = 256
)

type Checksum [ChecksumSize]byte

type SumKeyPair struct {
	Sum Checksum
	Key []byte
}

type KeyReply struct {
	Key       []byte
	Available bool
}

type Cache map[Checksum][]byte

func (c Cache) RequestKey(sum Checksum, reply *KeyReply) error {
	reply.Key, reply.Available = c[sum]
	return nil
}

func (c Cache) SetKey(pair SumKeyPair, reply *bool) error {
	_, *reply = c[pair.Sum]
	c[pair.Sum] = pair.Key
	return nil
}

func CacheKey(key []byte) error {
	client, err := rpc.Dial(Network, Address)
	if err != nil {
		return err
	}
	var sum Checksum
	var ok bool
	fmt.Printf("caching key (key=%s keyp=%p)\n", key, &key)
	client.Call("Cache.SetKey", SumKeyPair{sum, key}, &ok)
	fmt.Println("call completed")
	return nil
}

func GetKey() (reply *KeyReply, err error) {
	reply = new(KeyReply)
	client, err := rpc.Dial(Network, Address)
	if err != nil {
		return
	}

	var sum Checksum
	client.Call("Cache.RequestKey", sum, reply)
	return
}
