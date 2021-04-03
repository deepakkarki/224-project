package surfstore

import (
	"net/rpc"
)

type RPCClient struct {
	ServerAddr string
	BaseDir    string
	BlockSize  int
}

const INDEX_FILE = "index.txt"

func (surfClient *RPCClient) makeRPC(fName string, args, reply interface{}) error{
	conn, e := rpc.DialHTTP("tcp", surfClient.ServerAddr)
	if e != nil {
		return e
	}
	rpcName := "Server." + fName //name spacing, since rpc.Register is called on type Server
	e = conn.Call(rpcName, args, reply) // call the RPC server
	if e != nil {
		conn.Close()
		return e
	}

	return conn.Close()
}

func (surfClient *RPCClient) GetBlock(blockHash string, block *Block) error {
	return surfClient.makeRPC("GetBlock", blockHash, block)
}

func (surfClient *RPCClient) PutBlock(block Block, succ *bool) error {
	return surfClient.makeRPC("PutBlock", block, succ)
}

func (surfClient *RPCClient) HasBlocks(blockHashesIn []string, blockHashesOut *[]string) error {
	return surfClient.makeRPC("HasBlocks", blockHashesIn, blockHashesOut)
}

func (surfClient *RPCClient) GetFileInfoMap(succ *bool, serverFileInfoMap *map[string]FileMetaData) error {
	return surfClient.makeRPC("GetFileInfoMap", succ, serverFileInfoMap)
}

func (surfClient *RPCClient) UpdateFile(fileMetaData *FileMetaData, latestVersion *int) error {
	return surfClient.makeRPC("UpdateFile", fileMetaData, latestVersion)
}

var _ Surfstore = new(RPCClient)

// Create an Surfstore RPC client
func NewSurfstoreRPCClient(hostPort, baseDir string, blockSize int) RPCClient {
	dl := len(baseDir)
	if string(baseDir[dl-1]) == "/"{
		baseDir = baseDir[0:dl-1]
	}
	return RPCClient{
		ServerAddr: hostPort,
		BaseDir:    baseDir,
		BlockSize:  blockSize,
	}
}
