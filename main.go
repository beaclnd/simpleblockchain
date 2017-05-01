package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var MY_HTTP_PORT string
var MY_P2P_PORT string

const VERSION = "0.1"

// To store neighboring peers whose identities are their respective tcp addressess.
var listPeers map[string]bool = make(map[string]bool)

// A simple data struct of Block, without any transactions data.
type Block struct {
	Version           string
	PreviousBlockHash string
	TimeStamp         int64
	MetaData          string
}

// The struct of block to show clients
type ShowBlock struct {
	Version      string `json:"version"`
	Hash         string `json:"hash"`
	Height       uint64 `json:"height"`
	TimeStamp    int64  `json:"time_stamp"`
	MetaData     string `json:"meta_data"`
	PreviousHash string `json:"previous_hash"`
	NextHash     string `json:"next_hash"`
}

// Non-persistent blockchain storage, all blocks are stored in a map, the keys are Blocks' respective crypto hash strings.
var BlockChain map[string]Block = make(map[string]Block)

// A map to store orphan blocks, a key of the map is a orphan block's height and hash string.
var orphanBlocks map[OrphanBlockKey]Block = make(map[OrphanBlockKey]Block)

type OrphanBlockKey struct {
	height uint64
	hash   string
}

// A hexcode hash string pointing to the current latest block in the BlockChain
var hashOfLatestBlock string

// To synchronize threads altering sharing resources.
var rwMutex sync.RWMutex

// Message transmited within p2p blockchain network
type Message struct {
	From string // Source peer's address
	To   string // Destination peer's address
	Type string // Message type, could be one of ["HELLO", "RESPONSE_HELLO", "SYNC_GET_BLOCKS", "SYNC_BLOCKS",
	//"BROADCAST_UPDATED", "BROADCAST_BLOCK"]
	Msg      map[string][]byte // Some payload data
	MyHeight uint64            // The max height of blocks held by this peer
}

// To receive a peer informations from a http addPeer request.
type Peer struct {
	Address string
}

// To receive meta data from a http mineNewBlock request.
type BlockMetaData struct {
	MetaData string
}

func getGensisBlock() (block Block) {
	// The genesis block
	block = Block{"0.1", "0", 1493174849, "The Genesis Block of My Naive BlockChain!"}
	return
}

func initBlockChain() {
	block := getGensisBlock()
	hashOfLatestBlock = cryptoHash(&block)
	BlockChain[hashOfLatestBlock] = getGensisBlock()
}

func initHttpServer() {
	http.HandleFunc("/addPeer", addPeer)
	http.HandleFunc("/getPeers", getPeers)
	http.HandleFunc("/mineNewBlock", mineNewBlock)
	http.HandleFunc("/getBlockChain", getBlockChain)

	MY_HTTP_PORT = "9090"
	if http_port := os.Getenv("HTTP_PORT"); http_port != "" {
		MY_HTTP_PORT = http_port
	}
	log.Printf("Http server started on :%s", MY_HTTP_PORT)

	var addr string
	addr = ":" + MY_HTTP_PORT
	log.Fatal(http.ListenAndServe(addr, nil))
}

func initP2PServer() {
	MY_P2P_PORT = "7676"
	if p2p_port := os.Getenv("P2P_PORT"); p2p_port != "" {
		MY_P2P_PORT = p2p_port
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp4", getMyAddress())
	handleErr(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	handleErr(err)
	log.Printf("P2PServer started on :%s", MY_P2P_PORT)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			continue
		}
		go receive(conn)
	}
}

func handleErr(err error) {
	if err != nil {
		log.Println(err)
	}
}

func addPeer(w http.ResponseWriter, r *http.Request) {
	peer := new(Peer)
	if r.Body == nil {
		http.Error(w, "Please send a request with body", 400)
		return
	}
	err := json.NewDecoder(r.Body).Decode(peer)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("A client calls http server to add a peer with address: %s", peer.Address)

	ip, err := net.ResolveTCPAddr("tcp4", peer.Address)
	if err != nil {
		log.Printf("Reveived wrong tcp address: %s", peer.Address)
		w.Write([]byte("Wrong tcp address!"))
		return
	}

	listPeers[ip] = true
	log.Printf("Updated my peer list: %v\n", listPeers)

	msg := new(Message)
	msg.From = getMyAddress()
	msg.To = peer.Address
	msg.Type = "HELLO"
	msg.MyHeight = uint64(len(BlockChain))

	send(msg)
}

func getPeers(w http.ResponseWriter, r *http.Request) {
	listpeers_bytes, err := json.MarshalIndent(listPeers, "", "    ")
	handleErr(err)
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	w.Write(listpeers_bytes)
	w.Write([]byte("\n"))
}

// Not to actually mine a block like bitcoin core, just to simulate a mining process.
func mineNewBlock(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Please send a request with body", 400)
		return
	}
	metaData := new(BlockMetaData)
	err := json.NewDecoder(r.Body).Decode(metaData)
	if err != nil {
		log.Println(err)
		return
	}

	log.Printf("A client calls http server to mine a new block with metadata: %s", metaData.MetaData)

	// To mine a new block.
	block := new(Block)
	rwMutex.Lock()
	block.Version = VERSION
	block.PreviousBlockHash = hashOfLatestBlock
	block.TimeStamp = time.Now().Unix()
	block.MetaData = metaData.MetaData

	hashOfLatestBlock = cryptoHash(block)
	BlockChain[hashOfLatestBlock] = *block
	rwMutex.Unlock()

	log.Printf("Updated BlockChian after mining a new block, result BlockChain: \n%v", BlockChain)

	broadcastBlock(block)
}

// To get current blockchain and show it to a client.
func getBlockChain(w http.ResponseWriter, r *http.Request) {
	rwMutex.RLock()
	number_of_blocks := len(BlockChain)
	show_blocks := make([]ShowBlock, number_of_blocks)
	hash := hashOfLatestBlock
	var next_hash string
	for i := 0; i < number_of_blocks; i = i + 1 {
		var show_block ShowBlock
		var block Block = BlockChain[hash]
		show_block.Version = block.Version
		show_block.Hash = hash
		show_block.Height = uint64(number_of_blocks - i - 1)
		show_block.TimeStamp = block.TimeStamp
		show_block.MetaData = block.MetaData
		show_block.PreviousHash = block.PreviousBlockHash
		show_block.NextHash = next_hash

		next_hash = hash
		hash = BlockChain[hash].PreviousBlockHash
		show_blocks[i] = show_block
	}
	rwMutex.RUnlock()

	blockchain_bytes, err := json.MarshalIndent(show_blocks, "", "    ")
	handleErr(err)
	w.Header().Set("Content-type", "application/json; charset=utf-8")
	w.Write(blockchain_bytes)
	w.Write([]byte("\n"))
}

func getMyAddress() (addr string) {
	name, err := os.Hostname()
	handleErr(err)
	address, err := net.ResolveIPAddr("ip", name)
	handleErr(err)

	addr = address.String() + ":" + MY_P2P_PORT
	return
}

func createConnection(addr string) (conn net.Conn) {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", addr)
	handleErr(err)
	conn, err = net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		log.Println(err)
		return nil
	}
	return conn
}

func cryptoHash(block *Block) (hash string) {
	buf := make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(block.TimeStamp))
	s := [][]byte{[]byte(block.PreviousBlockHash), buf, []byte(block.MetaData)}
	my_bytes := bytes.Join(s, []byte(""))

	// To get hash of a given block by hashing the bytes array twice through sha256 algorithm.
	h := sha256.New()
	h.Write(my_bytes)
	h.Write(h.Sum(nil))
	hash = hex.EncodeToString(h.Sum(nil))
	return
}

func broadcastBlock(block *Block) {
	msg := new(Message)
	msg.From = getMyAddress()
	msg.Type = "BROADCAST_BLOCK"
	msg.MyHeight = uint64(len(BlockChain))
	msg.Msg = make(map[string][]byte)
	my_bytes, err := json.Marshal(block)
	handleErr(err)
	msg.Msg["broadcast_block"] = my_bytes

	for to, _ := range listPeers {
		msg.To = to
		send(msg)
	}
}

func receive(conn net.Conn) {
	defer conn.Close()

	dec := json.NewDecoder(conn)
	msg := new(Message)
	err := dec.Decode(msg)
	if err != nil {
		log.Printf("Decoding error: %v", err)
		return
	}
	log.Printf("Received a p2p message: \n%v", *msg)

	handleMessage(msg)
}

func send(msg *Message) {
	conn := createConnection(msg.To)
	if conn == nil {
		return
	}
	enc := json.NewEncoder(conn)
	enc.Encode(msg)
	log.Printf("Sent a p2p message: \n%v", *msg)
}

func handleMessage(msg *Message) {
	switch msg.Type {
	case "HELLO":
		handleHello(msg)
	case "RESPONSE_HELLO":
		if msg.MyHeight > uint64(len(BlockChain)) {
			syncGetBlocks()
		}
	case "BROADCAST_BLOCK":
		handleBroadcastBlock(msg)
	case "BROADCAST_UPDATED":
		if msg.MyHeight > uint64(len(BlockChain)) {
			syncGetBlocks()
		}
	case "SYNC_GET_BLOCKS":
		handleSyncGetBlocks(msg)
	case "SYNC_BLOCKS":
		handleSyncBlocks(msg)
	}
}

func handleHello(msg *Message) {
	listPeers[msg.From] = true
	log.Printf("Updated my peers list after receiving a hello message: \n%v", listPeers)

	msg.To = msg.From
	msg.From = getMyAddress()
	msg.Type = "RESPONSE_HELLO"

	if msg.MyHeight > uint64(len(BlockChain)) { // A sender has a higher block height, which means a longer BlockChain.
		go syncGetBlocks()
	}

	msg.MyHeight = uint64(len(BlockChain))
	send(msg)
}

func handleBroadcastBlock(msg *Message) {
	defer rwMutex.Unlock()
	rwMutex.Lock()
	if msg.MyHeight <= uint64(len(BlockChain)) { // To avoid endless broadcast loop, refuse lagged block as well.
		return
	}

	broadcast_block := new(Block)
	json.Unmarshal(msg.Msg["broadcast_block"], broadcast_block)

	if msg.MyHeight > uint64(len(BlockChain)+1) { // Can't tell a new block's validity, then just store it into orphan block map.
		orphan_block_key := OrphanBlockKey{msg.MyHeight, cryptoHash(broadcast_block)}
		orphanBlocks[orphan_block_key] = *broadcast_block
		go syncGetBlocks()
		return
	}

	if broadcast_block.PreviousBlockHash == hashOfLatestBlock { //The new block is validity.
		hashOfLatestBlock = cryptoHash(broadcast_block)
		BlockChain[hashOfLatestBlock] = *broadcast_block
		log.Printf("Updated BlockChain after receiving a new broadcast_blcok, result BlockChain: \n%v", BlockChain)
		broadcastBlock(broadcast_block)
	}
}

func handleSyncGetBlocks(msg *Message) {
	defer rwMutex.RUnlock()
	rwMutex.RLock()

	if msg.MyHeight < uint64(len(BlockChain)) { // This peer's blockchain is larger than a requesting peer's, then response it with blocks needed
		hash := hashOfLatestBlock
		my_bytes, err := json.Marshal(hash)
		handleErr(err)
		msg.Msg["hash_of_latest_block"] = my_bytes

		height := uint64(len(BlockChain))
		count := height - msg.MyHeight
		blocks := make(map[string]Block)
		for i := count; i > 0; i = i - 1 {
			blocks[hash] = BlockChain[hash]
			hash = BlockChain[hash].PreviousBlockHash
		}
		my_bytes, err = json.Marshal(blocks)
		handleErr(err)
		msg.Msg["sync_blocks"] = my_bytes

		msg.To = msg.From
		msg.From = getMyAddress()
		msg.Type = "SYNC_BLOCKS"
		msg.MyHeight = height

		send(msg)
	}
}

func handleSyncBlocks(msg *Message) {
	defer rwMutex.Unlock()
	rwMutex.Lock()

	if msg.MyHeight > uint64(len(BlockChain)) {
		var hash string
		var blocks map[string]Block
		var is_getting_longer_chain bool
		json.Unmarshal(msg.Msg["hash_of_latest_block"], &hash)
		hash_used_to_add_block := hash
		json.Unmarshal(msg.Msg["sync_blocks"], &blocks)
		json.Unmarshal(msg.Msg["is_getting_longer_chain"], &is_getting_longer_chain)

		// To Check whether the given blocks are valid.
		count := msg.MyHeight - (uint64)(len(BlockChain))
		if is_getting_longer_chain {
			count = uint64(len(blocks))
		}
		for i := count; i > 0; i = i - 1 {
			_, has := blocks[hash]
			if !has {
				log.Println("Warning: refuse to sync valid blocks")
				return
			}
			block := blocks[hash]
			if cryptoHash(&block) != hash {
				log.Println("Warning: refuse to sync valid blocks")
				return
			}
			hash = blocks[hash].PreviousBlockHash
		}

		if is_getting_longer_chain { // That means this peer is requesting a longer chain to override the current shorter fork.
			genesis_block := getGensisBlock()
			if hash == cryptoHash(&genesis_block) { // If two genesis blocks can match, then just replace the current BlockChain.
				BlockChain = blocks
				BlockChain[cryptoHash(&genesis_block)] = genesis_block
				hashOfLatestBlock = hash_used_to_add_block

				log.Printf("Updated Blockchain after cutting off a shorter fork by repacing it with a longer chain, result BlockChain: \n%v", BlockChain)

				go broadcastUpdated()
				disposeOrphanBlocks()
			} else {
				log.Println("Warning: refuse to replace BlockChain with a valid chain based on a different genesis block")
			}
		} else {
			if hash == hashOfLatestBlock { // Subsequent blocks are valid.
				hashOfLatestBlock = hash_used_to_add_block
				for i := count; i > 0; i = i - 1 {
					BlockChain[hash_used_to_add_block] = blocks[hash_used_to_add_block]
					hash_used_to_add_block = blocks[hash_used_to_add_block].PreviousBlockHash
				}

				log.Printf("Updated BlockChain after receiving sync_blocks, result BlockChain: \n%v", BlockChain)

				go broadcastUpdated()
				disposeOrphanBlocks()
			} else { // Maybe my chain contains a shorter fork.
				log.Println("Maybe my BlockChain contains a shorter fork, trying to sync")
				syncGetLongerChain(msg.From)
			}
		}
	}
}

func syncGetBlocks() {
	msg := new(Message)
	msg.From = getMyAddress()
	msg.Type = "SYNC_GET_BLOCKS"
	msg.MyHeight = uint64(len(BlockChain))
	msg.Msg = make(map[string][]byte)

	for to, _ := range listPeers {
		msg.To = to
		send(msg)
	}
	time.Sleep(1 * time.Second)
}

func syncGetLongerChain(to string) {
	msg := new(Message)
	msg.From = getMyAddress()
	msg.To = to
	msg.Type = "SYNC_GET_BLOCKS"
	msg.MyHeight = 1
	msg.Msg = make(map[string][]byte)
	is_getting_longer_chain := true
	my_bytes, err := json.Marshal(is_getting_longer_chain)
	handleErr(err)
	msg.Msg["is_getting_longer_chain"] = my_bytes

	send(msg)
}

func broadcastUpdated() {
	msg := new(Message)
	msg.From = getMyAddress()
	msg.Type = "BROADCAST_UPDATED"
	msg.MyHeight = uint64(len(BlockChain))

	for to, _ := range listPeers {
		msg.To = to
		send(msg)
	}
}

func disposeOrphanBlocks() { // To retrieve orphan blocks, and check their validities.
	for orphan_block_key, orphan_block := range orphanBlocks {
		if orphan_block_key.height <= uint64(len(BlockChain)) {
			_, has := BlockChain[cryptoHash(&orphan_block)]
			if has { // The orphan_block is valid.
				go broadcastBlock(&orphan_block)
			}
			delete(orphanBlocks, orphan_block_key)
		} else if orphan_block_key.height == uint64(len(BlockChain)+1) {
			if orphan_block.PreviousBlockHash == hashOfLatestBlock { // Valid successor
				hashOfLatestBlock = cryptoHash(&orphan_block)
				BlockChain[hashOfLatestBlock] = orphan_block
				go broadcastBlock(&orphan_block)
			}
			delete(orphanBlocks, orphan_block_key)
		}
	}
}

func main() {
	initBlockChain()
	go initHttpServer()
	initP2PServer()
}
