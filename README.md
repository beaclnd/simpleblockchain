# simpleBlockChain
It's an extremely simple and naive blockchain implementation based on golang, providing some limited functionalities ignoring transaction data, practical consensus protocol, persistent storage and so on, which ought to be included into a practical blockchain platform.

## structure
http server + p2p server

The program provids a http server for clients to interact with, offering http request interfaces
*addPeer*, *getPeers*, *mineNewBlock* and *getBlockChain*. With these request interfaces, clients can interact with a peer in a blockchain network, which is consisted of different nodes within a P2P network.

Utilizing tcp sockets, the program builds a simple P2P network, where peers can send or receive p2p messages to or from their respective neighbor peers. 

## functionalities
A peer can manually dicsovery a peer through a client's http request.
A client can get the peer list held by a peer.
A peer can create a new block based on the current blockchain state, through a client's http request.
A client can get the blockchain state of any peer.

All peers's blockchain data can keep same state eventually.

## notes
Non-persistent storage           -------  The blockchain data are just stored in memory in runtime.<br>
No transaction data              -------  Without any transaction data within a block.<br>
Non-automitic P2P peer discovery -------  With the http request interface *addPeer*, clients can Manually discovery a neighboring peer.<br>
Non-practical block mining       -------  Clients can call http server to imitate a block mining process, then get a new block to broadcast within the p2p net.<br>
Non-practical consensus protocal -------  To let peers' blockchain states keep same, a longer chain will eventually override all peers' chains.
