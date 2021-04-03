package surfstore

import (
	"os"
	"io"
	"fmt"
	"bytes"
	"bufio"
	"strings"
	"strconv"
	"io/ioutil"
	"crypto/sha256"
)


func HexHash(b []byte) string {
	hash := fmt.Sprintf("%x", sha256.Sum256(b))
	return hash[0:8] //truncate - easy to debug
}

func isEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

func isDeleted(m FileMetaData) bool {
	hashes := m.BlockHashList
	if (len(hashes) == 1) && (hashes[0] == "0") {
		return true
	}
	return false
}

// return indexes of elements in A but not in B. For example,
// A = [1, 2, 3, 5], B = [4, 3, 1]; return = [1, 3] (=> A[1], A[3] are not in B)
func indexDiff(a, b []string) []int {
	m := make(map[string]bool)
	diff := []int{}

	for _, item := range b {
		  m[item] = true
	}

	for i, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, i)
		}
	}
	return diff
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// Parses the index.txt at data dir and generates the map
func (client *RPCClient) getLocalIndex() map[string]FileMetaData {
	localIndex := map[string]FileMetaData{}
	path := client.BaseDir + "/" + INDEX_FILE

	if fileExists(client.BaseDir + "/" + INDEX_FILE) {
		content, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		lines := bytes.Split(content, []byte("\n"))
		for _, line := range lines {
			if len(line) < 2 {
				continue //ignore "" and " " lines
			}
			pLine := bytes.Split(line, []byte(","))
			fname := string(pLine[0])
			v, _ := strconv.Atoi(string(pLine[1]))
			hList := strings.Split(string(pLine[2]), " ")
			metaData := FileMetaData{Filename: fname, Version: v, BlockHashList: hList}
			localIndex[fname] = metaData
		}
	} else { // if file doesn't exist, create it
		f, err := os.Create(path)
		if err != nil {
			panic(err)
		}
		f.Close()
	}

	return localIndex
}

func (client *RPCClient) setLocalIndex(index map[string]FileMetaData){
	var builder strings.Builder
	if len(index) == 0 {
		return
	}
	for _, meta := range index {
		builder.WriteString(meta.Filename + ",")
		builder.WriteString(strconv.Itoa(meta.Version) + ",")
		builder.WriteString(strings.Join(meta.BlockHashList, " "))
		builder.WriteString("\n")
	}
	path := client.BaseDir + "/" + INDEX_FILE
	f, err := os.OpenFile(path, os.O_RDWR|os.O_TRUNC, 0744)
	if err != nil {
		panic("Unable to update local index after sync.")
	}
	defer f.Close()
	data := []byte(builder.String())
	f.Write(data[0:len(data)-1]) // remove trailing "\n"
	f.Sync()
}

func (client *RPCClient) getRemoteIndex() map[string]FileMetaData {
	var remoteIndex map[string]FileMetaData
	_ignore := true
	client.GetFileInfoMap(&_ignore, &remoteIndex)
	return remoteIndex
}


func ClientSync(client RPCClient) {

	// localIndex : index.txt, state of local data after last sync
	localIndex := client.getLocalIndex()

	// remoteIndex : matadata describing state of data on the server currently
	remoteIndex := client.getRemoteIndex()

	// get the files currently in the data dir
	files, err := ioutil.ReadDir(client.BaseDir)
	if err != nil {
		panic(err)
	}

	// Deal with actions on existing files. Possible actions -
	// 	1. Need to pull : if remoteIndex.ver > localIndex.ver
	// 	2. Need to push : file not in localIndex OR (localHashLish != fileHashList)
	// 	3. No action reqd. : otherwise do nothing
	for _, file := range files {
		fname := file.Name()
		if file.IsDir() || (fname == INDEX_FILE) {
			continue // don't handle these cases
		}
		fmt.Println("\n=========")
		fmt.Println("Dealing with file:", fname)
		path := client.BaseDir + "/" + fname

		// if fname is not in the index, remote and localMeta are initialised
		// to FileMetaData{} zero value => {"", 0, []string{}}
		remoteMeta := remoteIndex[fname]
		localMeta := localIndex[fname]

		if (remoteMeta.Version > localMeta.Version){ //need to do a pull
			processPull(client, remoteMeta)
			localIndex[fname] = remoteMeta
		} else { // might need to do a push - new file or update
			oldHashList := localMeta.BlockHashList
			newHashList := getHashList(path, client.BlockSize)

			if !isEqual(oldHashList, newHashList){ // need to do a push
				currentMeta := FileMetaData{fname, localMeta.Version+1, newHashList}
				err := processPush(client, localMeta, currentMeta)
				if err != nil { // version error => server has an updated version of the file
					// refetch updated index
					fmt.Println("push failed, pulling instead")
					remoteIndex = client.getRemoteIndex()
					remoteMeta = remoteIndex[fname]
					processPull(client, remoteMeta)
					localIndex[fname] = remoteMeta
				} else { //Success! update localIndex variable
					localIndex[fname] = currentMeta
					remoteIndex[fname] = currentMeta
				}
			}
		}
	}

	// Now deal with deleted files (local) and new files (on server)
	fileMap := map[string]bool{}
	for _, file := range files{
		fileMap[file.Name()] = true
	}

	// case : in local (with non-zero hash) but not in current => local deletion
	for _, m := range localIndex {
		if !isDeleted(m) && !fileMap[m.Filename] {
			// file has been deleted after last sync
			remoteMeta := remoteIndex[m.Filename]
			fmt.Println("Dealing with non existant file:", m.Filename)
			if remoteMeta.Version > m.Version {
				//there is new version on server - get that instead of deleting
				processPull(client, remoteMeta)
				localIndex[m.Filename] = remoteMeta
			} else {
				currentMeta := FileMetaData{m.Filename, m.Version+1, []string{"0"}}
				_v := 0
				err = client.UpdateFile(&currentMeta, &_v)
				if err != nil { //deleted file has newer version
					remoteIndex = client.getRemoteIndex()
					remoteMeta := remoteIndex[m.Filename]
					processPull(client, remoteMeta)
					localIndex[m.Filename] = remoteMeta
				} else { //Success! update localIndex variable
					localIndex[m.Filename] = currentMeta
					remoteIndex[m.Filename] = currentMeta
				}
			}
		}
	}

	// case : in remote but not in current => new file (server) or deleted file recreated
	for _, r := range remoteIndex {
		if !isDeleted(r) && !fileMap[r.Filename] {
			fmt.Println("Dealing with non existant file:", r.Filename)
			processPull(client, r)
			localIndex[r.Filename] = r
		}
	}

	// write back to the local index file
	client.setLocalIndex(localIndex)
}

// At this point we know we have to update the local file (create/overwrite)
func processPull(client RPCClient, remoteMeta FileMetaData) {
	//recreate the file if it exists
	path := client.BaseDir + "/" + remoteMeta.Filename
	fmt.Println("processPull", path)

	if (isDeleted(remoteMeta) && fileExists(path)) {
		os.Remove(path)
		return
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0744)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	for _, hash := range remoteMeta.BlockHashList {
		block := Block{}
		err := client.GetBlock(hash, &block)
		if err != nil {
			panic(err)
		}
		buf := block.BlockData[0:block.BlockSize]
		f.Write(buf)
	}
	f.Sync()
}

func processPush(client RPCClient, oldMeta, newMeta FileMetaData) error {
	fmt.Println("processPush", oldMeta.Filename)
	// find what chunks I need to upload
	diffIndex := indexDiff(newMeta.BlockHashList, oldMeta.BlockHashList)

	hashes := make([]string, 0, len(diffIndex))
	for _, i := range diffIndex {
		hashes = append(hashes, newMeta.BlockHashList[i])
	}
	serverHashList := []string{} //has the hashes server already has
	client.HasBlocks(hashes, &serverHashList)

	// turn list into map for quick lookup
	serverMap := map[string]bool{}
	for _, s := range serverHashList {
		serverMap[s] = true
	}
	// keep track of blocks sent to avoid duplicates
	sentMap := map[string]bool{}

	path := client.BaseDir + "/" + newMeta.Filename
	f, err := os.Open(path)
	if err != nil {
		panic("Can't open file: " + path)
	}
	defer f.Close()
	buf := make([]byte, client.BlockSize)

	//upload data to BlockStore
	for _, i := range diffIndex {
		hash := newMeta.BlockHashList[i]
		_, ok1 := serverMap[hash]
		_, ok2 := sentMap[hash]
		if !ok1 && !ok2 { // not on server & not already sent
			f.Seek(int64(client.BlockSize*i), 0) // go to the i'th block from beginning
			n , _ := f.Read(buf)

			block := Block{BlockData: buf[0:n], BlockSize: n}
			_res := false
			err := client.PutBlock(block, &_res)
			if err != nil {
				panic("Couldn't upload block of file: " + path)
			}
			sentMap[hash] = true
		}
	}

	// update remote index
	_v := 1
	fmt.Println("processPush update", newMeta)
	err = client.UpdateFile(&newMeta, &_v)
	return err
}

func getHashList(path string, blockSize int) []string{
	hash := []string{}
	buf := make([]byte, blockSize)

	f, err := os.Open(path)
	if err != nil {
		panic("Cannot read the file")
	}
	defer f.Close()

	r := bufio.NewReader(f)
	for {
		n, err := r.Read(buf)
		if err == io.EOF {
			break
		}
		hash = append(hash, HexHash(buf[0:n]))
	}

	return hash
}
