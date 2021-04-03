package mydynamo

import (
	"fmt"
	"log"
	"net/rpc"
)

type RPCClient struct {
	ServerAddr string
	rpcConn    *rpc.Client
}

//Removes the RPC connection associated with this client
func (dynamoClient *RPCClient) CleanConn() {
	var e error
	if dynamoClient.rpcConn != nil {
		e = dynamoClient.rpcConn.Close()
		if e != nil {
			fmt.Println("CleanConnError", e)
		}
	}
	dynamoClient.rpcConn = nil
	return
}

//Establishes an RPC connection to the server this Client is associated with
func (dynamoClient *RPCClient) RpcConnect() error {
	if dynamoClient.rpcConn != nil {
		return nil
	}

	var e error
	dynamoClient.rpcConn, e = rpc.DialHTTP("tcp", dynamoClient.ServerAddr)
	if e != nil {
		dynamoClient.rpcConn = nil
	}

	return e
}

//Removes and re-establishes an RPC connection to the server
func (dynamoClient *RPCClient) CleanAndConn() error {
	var e error
	if dynamoClient.rpcConn != nil {
		e = dynamoClient.rpcConn.Close()
		if e != nil {
			fmt.Println("CleanConnError", e)
		}
	}
	dynamoClient.rpcConn = nil

	dynamoClient.rpcConn, e = rpc.DialHTTP("tcp", dynamoClient.ServerAddr)
	if e != nil {
		dynamoClient.rpcConn = nil
	}

	return e
}

//Puts a value to the server.
func (dynamoClient *RPCClient) Put(value PutArgs) bool {
	var result bool
	if dynamoClient.rpcConn == nil {
		return false
	}
	err := dynamoClient.rpcConn.Call("MyDynamo.Put", value, &result)
	if err != nil {
		log.Println(err)
		return false
	}
	return result
}

//Gets a value from a server.
func (dynamoClient *RPCClient) Get(key string) *DynamoResult {
	var result DynamoResult
	if dynamoClient.rpcConn == nil {
		return nil
	}
	err := dynamoClient.rpcConn.Call("MyDynamo.Get", key, &result)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &result
}

func (dynamoClient *RPCClient) S_Get(key string) (*DynamoResult, error) {
	var result DynamoResult
	if dynamoClient.rpcConn == nil {
		return &result, nil
	}
	err := dynamoClient.rpcConn.Call("MyDynamo.S_Get", key, &result)
	if err != nil {
		return &result, err
	}
	return &result, nil
}

func (dynamoClient *RPCClient) S_Put(pa S_PutArgs) error {
	if dynamoClient.rpcConn == nil {
		return nil
	}
	b := false
	err := dynamoClient.rpcConn.Call("MyDynamo.S_Put", pa, &b)
	if err != nil {
		return err
	}
	return nil
}

func (dynamoClient *RPCClient) S_Gossip(store Store) error {
	if dynamoClient.rpcConn == nil {
		return nil
	}
	emt := Empty{}
	err := dynamoClient.rpcConn.Call("MyDynamo.S_Gossip", store, &emt)
	if err != nil {
		return err
	}
	return nil
}


//Emulates a crash on the server this client is connected to
func (dynamoClient *RPCClient) Crash(seconds int) bool {
	if dynamoClient.rpcConn == nil {
		return false
	}
	var success bool
	err := dynamoClient.rpcConn.Call("MyDynamo.Crash", seconds, &success)
	if err != nil {
		log.Println(err)
		return false
	}
	return success
}

//Instructs the server this client is connected to gossip
func (dynamoClient *RPCClient) Gossip() {
	if dynamoClient.rpcConn == nil {
		return
	}
	var v Empty
	err := dynamoClient.rpcConn.Call("MyDynamo.Gossip", v, &v)
	if err != nil {
		log.Println(err)
		return
	}
}

//Creates a new DynamoRPCClient
func NewDynamoRPCClient(serverAddr string) *RPCClient {
	return &RPCClient{
		ServerAddr: serverAddr,
		rpcConn:    nil,
	}
}

func ConnectToNode(node DynamoNode) (*RPCClient, error) {
	serverAddr := node.Address +":"+ node.Port
	r := RPCClient{
		ServerAddr: serverAddr,
		rpcConn:    nil,
	}
	err := r.RpcConnect()
	return &r, err
}

func RemoteS_Get(node DynamoNode, key string) ([]ObjectEntry, error){
	client, err := ConnectToNode(node)
	if err != nil {
		log.Println("Could not connect To node", node)
		return []ObjectEntry{}, err
	}
	defer client.CleanConn()

	res, err := client.S_Get(key)
	if err != nil {
		log.Println("Could not execute RPC S_GET on node:", node)
		log.Println("An error occured:", err)
		return []ObjectEntry{}, err
	}

	return res.EntryList, nil
}

func RemoteS_Put(node DynamoNode, pa S_PutArgs) error {
	client, err := ConnectToNode(node)
	if err != nil {
		log.Println("Could not connect To node", node)
		return err
	}
	defer client.CleanConn()

	err = client.S_Put(pa)
	if err != nil {
		log.Println("Could not execute RPC S_PUT on node:", node)
		log.Println("An error occured:", err)
		return err
	}

	return nil
}

func RemoteS_Gossip(node DynamoNode, store Store) error {
	client, err := ConnectToNode(node)
	if err != nil {
		log.Println("Could not connect To node", node)
		return err
	}
	defer client.CleanConn()

	err = client.S_Gossip(store)
	if err != nil {
		log.Println("Could not execute RPC S_Gossip on node:", node)
		log.Println("An error occured:", err)
		return err
	}

	return nil
}