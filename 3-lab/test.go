package main

import (
	"fmt"
	"os"
	"strconv"
	"crypto/sha256"
	"./surfstore/src/surfstore"
)

func printBlock(b surfstore.Block){
	fmt.Println("size:", b.BlockSize)
	fmt.Println("data:", string(b.BlockData))
}

func main() {
	if len(os.Args) < 4 {
		fmt.Println("Usage: ./run-client host:port baseDir blockSize")
		os.Exit(1)
	}

	hostPort := os.Args[1]
	baseDir := os.Args[2]
	blockSize, err := strconv.Atoi(os.Args[3])
	if err != nil {
		fmt.Println("Usage: ./run-client host:port baseDir blockSize")
	}

	rpcClient := surfstore.NewSurfstoreRPCClient(hostPort, baseDir, blockSize)
	_i := false
	_j := 0

	//test PutBlock (put 10 blocks)
	for i :=0; i < 10; i++ {
		data := []byte(fmt.Sprintf("%d. Hello world", i))
		blk := surfstore.Block{BlockData: data, BlockSize: len(data)}
		err := rpcClient.PutBlock(blk, &_i)
		if err != nil {
			fmt.Println("PutBlock error:", err)
		} else {
			fmt.Println("uploaded block #", i)
		}
	}

	// GetBlock
	blk := surfstore.Block{}
	byteHash := sha256.Sum256([]byte("1. Hello world"))
	hash := string(byteHash[:])
	err = rpcClient.GetBlock(hash, &blk)
	if err != nil {
		fmt.Println("GetBlock error:", err)
	} else{
		printBlock(blk)
	}

	// non existing block
	err = rpcClient.GetBlock("yo", &blk)
	if err != nil {
		fmt.Println("GetBlock error:", err)
	} else{
		printBlock(blk)
	}

	// HasBlocks()
	hashes := []string{}
	sHashes := []string{}
	for i := 5; i <15; i++{
		data := []byte(fmt.Sprintf("%d. Hello world", i))
		bHash := sha256.Sum256(data)
		hashes = append(hashes, string(bHash[:]))
	}
	err = rpcClient.HasBlocks(hashes, &sHashes)
	if err != nil {
		fmt.Println("HasBlocks error:", err)
	} else{
		fmt.Println("Server hashes:", len(sHashes))
	}

	// UpdateFile
	hList := []string{"b0", "b1", "b2", "b3"}
	fm := surfstore.FileMetaData{Filename:"t1", Version:2, BlockHashList: hList}
	err = rpcClient.UpdateFile(&fm, &_j)
	if err != nil{
		fmt.Println("Error in UpdateFile:", err)
	}
	// GetFileInfoMap
	m := map[string]surfstore.FileMetaData{}
	rpcClient.GetFileInfoMap(&_i, &m)
	fmt.Println(m)
}
