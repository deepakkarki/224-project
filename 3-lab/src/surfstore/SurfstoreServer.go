package surfstore

import (
	"log"
	"fmt"
	"sync"
	"net"
	"net/http"
	"net/rpc"
)

type Server struct {
	BlockStore BlockStoreInterface
	MetaStore  MetaStoreInterface
}

var BlockStoreLock = &sync.Mutex{}
var MetaStoreLock = &sync.Mutex{}

var BS = BlockStore{BlockMap: map[string]Block{}}
var MS = MetaStore{FileMetaMap: map[string]FileMetaData{}}

func (s *Server) GetFileInfoMap(succ *bool, serverFileInfoMap *map[string]FileMetaData) error {
	MetaStoreLock.Lock()
	defer MetaStoreLock.Unlock()
	return s.MetaStore.GetFileInfoMap(succ, serverFileInfoMap)
}

func (s *Server) UpdateFile(fileMetaData *FileMetaData, latestVersion *int) error {
	MetaStoreLock.Lock()
	defer MetaStoreLock.Unlock()
	return s.MetaStore.UpdateFile(fileMetaData, latestVersion)
}

func (s *Server) GetBlock(blockHash string, blockData *Block) error {
	BlockStoreLock.Lock()
	defer BlockStoreLock.Unlock()
	return s.BlockStore.GetBlock(blockHash, blockData)
}

func (s *Server) PutBlock(blockData Block, succ *bool) error {
	BlockStoreLock.Lock()
	defer BlockStoreLock.Unlock()
	return s.BlockStore.PutBlock(blockData, succ)
}

func (s *Server) HasBlocks(blockHashesIn []string, blockHashesOut *[]string) error {
	BlockStoreLock.Lock()
	defer BlockStoreLock.Unlock()
	return s.BlockStore.HasBlocks(blockHashesIn, blockHashesOut)
}

// This line guarantees all method for surfstore are implemented
var _ Surfstore = new(Server)

func NewSurfstoreServer() Server {
	return Server{
		BlockStore: &BS,
		MetaStore:  &MS,
	}
}

func ServeSurfstoreServer(hostAddr string, surfstoreServer Server) error {
	rpc.Register(&surfstoreServer)
	rpc.HandleHTTP()

	l, e := net.Listen("tcp", hostAddr)
	if e != nil {
		log.Fatal("listen error:", e)
	}

	http.Serve(l, nil)
	// fmt.Println("Commands : m => MetaStore, b => BlockStore, q => quit")
	// handleQuery(surfstoreServer)
	return nil
}


func handleQuery(server Server) {
	var s string
	for {
		fmt.Println("=============")
		fmt.Scanln(&s)
		if s[0:1] == "m" {
			PrintMetaStore(MS.FileMetaMap)
		} else if s[0:1] == "b" {
			PrintBlockStore(BS.BlockMap)
		} else if s[0:1] == "q" {
			break
		}
		fmt.Println("")
	}
}

func PrintMetaStore(ms map[string]FileMetaData) {
	fmt.Println("MetaStore")
	for _, filemeta := range ms {
		fmt.Println(filemeta.Filename, filemeta.Version, filemeta.BlockHashList)
	}
}

func PrintBlockStore(bs map[string]Block){
	fmt.Println("BlockStore")
	for hash, blk := range bs {
		fmt.Println(hash, ": ", string(blk.BlockData))
	}
}