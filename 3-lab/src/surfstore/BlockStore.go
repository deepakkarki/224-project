package surfstore

import (
	"errors"
)

type BlockStore struct {
	BlockMap map[string]Block
}

func (bs *BlockStore) GetBlock(blockHash string, blockData *Block) error {
	block, ok := bs.BlockMap[blockHash]
	if !ok {
		return errors.New("Block does not exist")
	}
	*blockData = block
	return nil
}

func (bs *BlockStore) PutBlock(block Block, succ *bool) error {
	hash := HexHash(block.BlockData[0:block.BlockSize])
	bs.BlockMap[hash] = block
	*succ = true
	return nil
}

// Given a list of hashes "blockHashesIn", returns a list containing
// the subset of in that are stored in the BlockStore "bs"
func (bs *BlockStore) HasBlocks(blockHashesIn []string, blockHashesOut *[]string) error {
	hashes := []string{}

	for _, hash := range blockHashesIn {
		if _, ok := bs.BlockMap[hash]; ok {
			hashes = append(hashes, hash)
		}
	}
	*blockHashesOut = hashes

	return nil
}

// This line guarantees all method for BlockStore are implemented
var _ BlockStoreInterface = new(BlockStore)
