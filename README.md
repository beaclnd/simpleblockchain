# SimpleBlockChain
It's an extremely simple and naive blockchain implementation based on golang, providing some limited functionalities ignoring transaction data, practical consensus protocol, persistent storage and so on, which ought to be included into a practical blockchain platform.

## Structure
http server + p2p server

The program provids a http server for clients to interact with, offering http request interfaces
*addPeer*, *getPeers*, *mineNewBlock* and *getBlockChain*. With these request interfaces, clients can interact with a peer in a blockchain network, which is consisted of different nodes within a P2P network.

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
We can utilize Docker to run multi containers simulating multi computer peers, with which we can imitate a p2p network.
With an assumption of having installed [*docker*](https://www.docker.com/), [*docker-compose*](https://docs.docker.com/compose/) and [*curl*](https://curl.haxx.se/) following instructions will set up a simple blockchain network.

Start up 3 seperate peers(described in *Dockerfile* and *docker-compose.ymal*):
```bash
docker-compose up
```
There will be 3 seperate peers: peer0, peer1 and peer2.

Check peer lists held by peer0, peer1 and peer2 respectively:
```bash
curl localhost:9000/getPeers
curl localhost:9001/getPeers
curl localhost:9002/getPeers
```
Let peer1 connect with peer0:
```bash
curl -H "Content-type:application/json" -X POST -d '{"Address": "peer0:7676"}' localhost:9001/addPeer
```
Now you can check peer lists again.






