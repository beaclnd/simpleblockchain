# SimpleBlockChain
It's an extremely simple and naive blockchain implementation based on golang, providing some limited functionalities ignoring transaction data, practical consensus protocol, persistent storage and so on, which ought to be included into a practical blockchain platform. However, I think it's benefit for someone at least myselft to learn blockchain.

## Structure
http server + p2p server

The program provids a http server for clients to interact with, offering http request interfaces
*addPeer*, *getPeers*, *mineNewBlock* and *getBlockChain*. With these request interfaces, clients can interact with a peer in a blockchain network, which consist of different nodes within a P2P network.

Utilizing tcp sockets, the program builds a simple P2P network, where peers can send or receive p2p messages to or from their respective neighbor peers. 

## Functionalities
- A peer can manually dicsovery another peer through a client's http request.
- A client can get the peer list held by a peer.
- A peer can create a new block based on the current blockchain state, through a client's http request.
- A client can get the blockchain state of any peer.
- All peers's blockchain data state can keep same eventually, inspite of any fork derivated from different block mining at the same time.

## Notes
- Non-persistent storage           -------  The blockchain data are just stored in memory in runtime.
- No transaction data              -------  Without any transaction data within a block.
- Non-automitic P2P peer discovery -------  With the http request interface *addPeer*, clients can Manually discovery a neighboring peer.
- Non-practical block mining       -------  Clients can call http server to imitate a block mining process, then get a new block to broadcast within the p2p net.
- Non-practical consensus protocal -------  To let peers' blockchain states keep same, a longer chain will eventually override all peers' chains.

## QuickStart
We can utilize Docker to run multiple containers simulating multiple computer nodes, with which we can imitate a p2p network.
With an assumption of having installed [*docker*](https://www.docker.com/), [*docker-compose*](https://docs.docker.com/compose/) and [*curl*](https://curl.haxx.se/) ,following instructions will set up a simple blockchain network.

Start up 3 seperate peers(described in *Dockerfile* and *docker-compose.ymal*):
```bash
docker-compose up
```
(The standard output will show you what happened when following instructions are executed.)
There will be 3 seperate peers: peer0, peer1 and peer2.

Check peer lists held by peer0, peer1 and peer2 respectively:
```bash
curl localhost:9000/getPeers
curl localhost:9001/getPeers
curl localhost:9002/getPeers
```
Check every peer's initial blockchain state:
```bash
curl localhost:9000/getBlockChain
curl localhost:9001/getBlockChain
curl localhost:9002/getBlockChain
```
Let peer1 connect with peer0:
```bash
curl -H "Content-type:application/json" -X POST -d '{"Address": "peer0:7676"}' localhost:9001/addPeer
```
Now you can check peer lists again.

Now, peer0 is connecting with peer 1, peer2 is seperated from them.
Let peer0 or peer1 create a new block：
```bash
curl -H "Content-type:application/json" -X POST -d '{"MetaData": "My BlockChain #CH1 #BL1"}' localhost:9000/mineNewBlock
```
Now，check each peer's blockchain state. You could notice that peer0 and peer1 hold the same blockchain data, containing a new block created with the metadata you input.

Let peer2 create a new block:
```bash
curl -H "Content-type:application/json" -X POST -d '{"MetaData": "My BlockChain #CH2 #BL1"}' localhost:9002/mineNewBlock
```
You could check peer2's blockchain state again.

Let peer2 connect with peer1:
```bash
curl -H "Content-type:application/json" -X POST -d '{"Address": "peer1:7676"}' localhost:9002/addPeer
```
Check each peer's blockchain state again. You could find that peer0 and peer1 hold the same blockchain, while peer2 hold a different one.

Now, let peer2 create two new blocks:
```bash
curl -H "Content-type:application/json" -X POST -d '{"MetaData": "My BlockChain #CH2 #BL2"}' localhost:9002/mineNewBlock
curl -H "Content-type:application/json" -X POST -d '{"MetaData": "My BlockChain #CH2 #BL3"}' localhost:9002/mineNewBlock
```
Now, check each peer's blockchain state, you could notice that the three peers share the same blockchain through dropping a shorter fork.









