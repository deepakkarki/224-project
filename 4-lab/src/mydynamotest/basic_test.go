package mydynamotest

import (
	"testing"
	"time"
)

func TestBasicPut(t *testing.T) {
	t.Logf("Starting basic Put test")

	//Test initialization
	//Note that in the code below, dynamo servers will use the config file located in src/mydynamotest
	cmd := InitDynamoServer("./myconfig.ini")
	ready := make(chan bool)

	//starts the Dynamo nodes, and get ready to kill them when done with the test
	go StartDynamoServer(cmd, ready)
	defer KillDynamoServer(cmd)

	//Wait for the nodes to finish spinning up.
	time.Sleep(3 * time.Second)
	<-ready

	//Create a client that connects to the first server
	//This assumes that the config file specifies 8080 as the starting port
	clientInstance := MakeConnectedClient(8080)

	//Put a value on key "s1"
	clientInstance.Put(PutFreshContext("s1", []byte("abcde")))

	//Get the value back, and check if we successfully retrieved the correct value
	gotValuePtr := clientInstance.Get("s1")
	if gotValuePtr == nil {
		t.Fail()
		t.Logf("TestBasicPut: Returned nil")
	}
	gotValue := *gotValuePtr
	if len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("abcde")) {
		t.Fail()
		t.Logf("TestBasicPut: Failed to get value")
	}
}
