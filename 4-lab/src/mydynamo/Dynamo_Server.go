package mydynamo

import (
	"log"
	"net"
	"sync"
	"time"
	"net/http"
	"net/rpc"
	"errors"
)

// for each key the server might contain a number of values, each tagged with
// a 'context' (viz a VectorClock). ObjectEntry => value + vector clock tag
type Store map[string][]ObjectEntry

type DynamoServer struct {
	wValue         int           //Number of nodes to write to on each Put
	rValue         int           //Number of nodes to read from on each Get
	preferenceList []DynamoNode  //Order of preference of nodes to perform Get/Put on
	selfNode       DynamoNode    //This node's address and port info
	nodeID         string        //ID of this node
	store          Store         //The actual in-memory data store
	hasCrashed     bool          //Flag to track if node is in crashed state
	mtxCrash       sync.Mutex    //Mutex for hasCrashed
	mtxStore       sync.RWMutex  //Mutex for store
}

func (s *DynamoServer) SendPreferenceList(incomingList []DynamoNode, _ *Empty) error {
	s.preferenceList = incomingList
	return nil
}

// Forces server to gossip
func (s *DynamoServer) Gossip(_ Empty, _ *Empty) error {
	if s.hasCrashed {
		return errors.New("Server Crash")
	}

	for _, node := range s.preferenceList {
		if node == s.selfNode {
			continue
		}

		s.mtxStore.RLock()
		RemoteS_Gossip(node, s.store)
		s.mtxStore.RUnlock()
	}
	return nil
}

//Makes server unavailable for some seconds
func (s *DynamoServer) Crash(seconds int, success *bool) error {
	s.mtxCrash.Lock()
	defer s.mtxCrash.Unlock()

	if s.hasCrashed {
		*success = false
		return errors.New("Server Crash")
	}

	s.hasCrashed = true
	go func(){
		time.Sleep(time.Duration(seconds) * time.Second)
		s.hasCrashed = false
	}()

	*success = true
	return nil
}

// Put a file to this server and W other servers
func (s *DynamoServer) Put(pa PutArgs, result *bool) error {
	inserted := false

	if s.hasCrashed {
		*result = inserted
		return errors.New("Server Crash")
	}

	// increment the clock
	pa.Context.Clock.Increment(s.nodeID)
	obj := ObjectEntry{Value: pa.Value, Context: pa.Context}

	// acquire the write lock for the store
	s.mtxStore.Lock()
	defer s.mtxStore.Unlock()

	newList, count := CollapseObjects(s.store[pa.Key], []ObjectEntry{obj})
	if count == 0 { // incoming put has made 0 updates
		*result = inserted
		return errors.New("Object version out of date")
	}

	s.store[pa.Key] = newList
	inserted = true

	log.Println("PUT", s.selfNode, newList)

	// write to upto s.rValue nodes in the preferenceList
	ws := 0 // number of writes that were successful
	for _, node := range s.preferenceList {
		if node == s.selfNode {
			ws++ // we've written this value to s.store already
			continue
		}
		if ws >= s.wValue {
			break // done writing to write quorum
		}
		// make the call to node
		err := RemoteS_Put(node, S_PutArgs{Key: pa.Key, EntryList: newList})
		if err == nil { // Success!
			ws++
		}
	}

	*result = inserted
	return nil
}

//Get a file from this server, matched with R other servers
func (s *DynamoServer) Get(key string, result *DynamoResult) error {
	if s.hasCrashed {
		*result = DynamoResult{}
		return errors.New("Server Crash")
	}

	// resulting slice of results from read quorum
	allVals := []ObjectEntry{}

	s.mtxStore.RLock()
	defer s.mtxStore.RUnlock()

	// get value from s.store
	if val, ok := s.store[key]; ok {
		allVals = val
	}

	rs := 0 // number of reads that were successful
	for _, node := range s.preferenceList {
		if node == s.selfNode {
			rs++ // we've got the value from s.store already
			continue
		}
		if rs >= s.rValue {
			break // done reading read quorum
		}

		// make the call to node
		vals, err := RemoteS_Get(node, key)
		if err == nil { // Success!
			// collapse to ensure only concurrent values exist
			allVals, _ = CollapseObjects(allVals, vals)
			rs++
		}
	}

	// if key did not exist anywhere, return err
	if len(allVals) == 0 {
		*result = DynamoResult{EntryList: []ObjectEntry{}}
		log.Println("Error GET", s.selfNode, "No such value")
		return errors.New("No such value")
	}

	log.Println("GET", s.selfNode, allVals)
	*result = DynamoResult{ EntryList: allVals }
	return nil
}


//Other server puts a value while it's replicating or gossiping, i.e. the other
// node updates this server with the key-values it has
func (s *DynamoServer) S_Put(pa S_PutArgs, result *bool) error {
	inserted := false

	if s.hasCrashed {
		*result = inserted
		return errors.New("Server Crash")
	}

	s.mtxStore.Lock()
	defer s.mtxStore.Unlock()

	newList, count := CollapseObjects(s.store[pa.Key], pa.EntryList)
	log.Println("S_PUT", s.selfNode, newList)
	if count > 0 { // incoming put has made an update
		s.store[pa.Key] = newList
		inserted = true
	}

	*result = inserted
	return nil
}

//Other server tries to get value from this server on Get() (called by client)
// if this node is in it's top 'rValue' elements of it's preferenceList
func (s *DynamoServer) S_Get(key string, result *DynamoResult) error {
	// this function is the normal get, minus the calls to S_Get()
	if s.hasCrashed {
		*result = DynamoResult{EntryList: []ObjectEntry{}}
		return errors.New("Server Crash")
	}

	s.mtxStore.RLock()
	defer s.mtxStore.RUnlock()

	if _, ok := s.store[key]; !ok {
		*result = DynamoResult{}
		return errors.New("No such value")
	}

	log.Println("S_GET", s.selfNode, s.store[key])
	*result = DynamoResult{ EntryList: s.store[key] }
	return nil
}


func (s *DynamoServer) S_Gossip(store Store, e *Empty) error {
	*e = Empty{}
	if s.hasCrashed {
		return errors.New("Server Crash")
	}

	s.mtxStore.Lock()
	defer s.mtxStore.Unlock()

	// for each key, update the list of ObjectEntries (clock-data pairs)
	for k, v := range store {
		combined, n := CollapseObjects(s.store[k], v)
		if n > 0 {
			s.store[k] = combined
		}
	}
	return nil
}


/* Belows are functions that implement server boot up and initialization */
func NewDynamoServer(w int, r int, hostAddr string, hostPort string, id string) DynamoServer {
	preferenceList := make([]DynamoNode, 0)
	selfNodeInfo := DynamoNode{
		Address: hostAddr,
		Port:    hostPort,
	}
	s := Store{}
	crashLock := sync.Mutex{}
	storeLock := sync.RWMutex{}
	return DynamoServer{
		wValue:         w,
		rValue:         r,
		preferenceList: preferenceList,
		selfNode:       selfNodeInfo,
		nodeID:         id,
		store:          s,
		hasCrashed:     false,
		mtxCrash:       crashLock,
		mtxStore:       storeLock,
	}
}

func ServeDynamoServer(dynamoServer DynamoServer) error {
	rpcServer := rpc.NewServer()
	e := rpcServer.RegisterName("MyDynamo", &dynamoServer)
	if e != nil {
		log.Println(DYNAMO_SERVER, "Server Can't start During Name Registration")
		return e
	}

	log.Println(DYNAMO_SERVER, "Successfully Registered the RPC Interfaces")

	l, e := net.Listen("tcp", dynamoServer.selfNode.Address+":"+dynamoServer.selfNode.Port)
	if e != nil {
		log.Println(DYNAMO_SERVER, "Server Can't start During Port Listening")
		return e
	}

	log.Println(DYNAMO_SERVER, "Successfully Listening to Target Port ", dynamoServer.selfNode.Address+":"+dynamoServer.selfNode.Port)
	log.Println(DYNAMO_SERVER, "Serving Server Now")

	return http.Serve(l, rpcServer)
}
