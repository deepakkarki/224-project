package mydynamotest

import (
    "mydynamo"
    "testing"
    "time"
)

func TestPutW2(t *testing.T){
    t.Logf("Starting PutW2 test")
    cmd := InitDynamoServer("./twoserver.ini")
    ready := make(chan bool)
    go StartDynamoServer(cmd, ready)
    defer KillDynamoServer(cmd)

    time.Sleep(3 * time.Second)
    <-ready

    clientInstance0 := MakeConnectedClient(8080)
    clientInstance1 := MakeConnectedClient(8081)
    clientInstance0.Put(PutFreshContext("s1", []byte("abcde")))
    gotValuePtr := clientInstance1.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestPutW2: Failed to get")
    }
    gotValue := *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("abcde"))){
        t.Fail()
        t.Logf("TestPutW2: Failed to get value")
    }

}

func TestGossip(t *testing.T){
    t.Logf("Starting Gossip test")
    cmd := InitDynamoServer("./myconfig.ini")
    ready := make(chan bool)
    go StartDynamoServer(cmd, ready)
    defer KillDynamoServer(cmd)

    time.Sleep(3 * time.Second)
    <-ready

    clientInstance0 := MakeConnectedClient(8080)
    clientInstance1 := MakeConnectedClient(8081)
    clientInstance0.Put(PutFreshContext("s1", []byte("abcde")))
    clientInstance0.Gossip()
    gotValuePtr := clientInstance1.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestGossip: Failed to get")
    }
    gotValue := *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("abcde"))){
        t.Fail()
        t.Logf("TestGossip: Failed to get value")
    }

}

func TestMultipleKeys(t *testing.T){
    t.Logf("Starting MultipleKeys test")
    cmd := InitDynamoServer("./myconfig.ini")
    ready := make(chan bool)
    go StartDynamoServer(cmd, ready)
    defer KillDynamoServer(cmd)

    time.Sleep(3 * time.Second)
    <-ready

    clientInstance0 := MakeConnectedClient(8080)
    clientInstance1 := MakeConnectedClient(8081)
    clientInstance0.Put(PutFreshContext("s1", []byte("abcde")))
    clientInstance0.Gossip()
    gotValuePtr := clientInstance0.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestMultipleKeys: Failed to get")
    }
    gotValue := *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("abcde"))){
        t.Fail()
        t.Logf("TestMultipleKeys: Failed to get value")
    }

    gotValuePtr = clientInstance1.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestMultipleKeys: Failed to get")
    }
    gotValue = *gotValuePtr

    clientInstance1.Put(mydynamo.NewPutArgs("s1",gotValue.EntryList[0].Context ,[]byte("efghi")))
    gotValuePtr = clientInstance1.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestMultipleKeys: Failed to get")
    }
    gotValue = *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("efghi"))){
        t.Fail()
        t.Logf("TestMultipleKeys: Failed to get value")
    }
}

func TestDynamoPaper(t *testing.T){
    t.Logf("DynamoPaper test")
    cmd := InitDynamoServer("./myconfig.ini")
    ready := make(chan bool)
    go StartDynamoServer(cmd, ready)
    defer KillDynamoServer(cmd)

    time.Sleep(3 * time.Second)
    <-ready

    clientInstance0 := MakeConnectedClient(8080)
    clientInstance1 := MakeConnectedClient(8081)
    clientInstance2 := MakeConnectedClient(8082)

    clientInstance0.Put(PutFreshContext("s1", []byte("abcde")))
    gotValuePtr := clientInstance0.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestDynamoPaper: Failed to get first value")
    }

    gotValue := *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("abcde"))){
        t.Fail()
        t.Logf("TestDynamoPaper: First value doesn't match")
    }
    clientInstance0.Put(mydynamo.NewPutArgs("s1", gotValue.EntryList[0].Context, []byte("bcdef")))
    gotValuePtr = clientInstance0.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestDynamoPaper: Failed to get second value")
    }
    gotValue = *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("bcdef"))){
        t.Fail()
        t.Logf("TestDynamoPaper: Second value doesn't match")
    }

    clientInstance0.Gossip()
    clientInstance1.Put(mydynamo.NewPutArgs("s1", gotValue.EntryList[0].Context, []byte("cdefg")))
    clientInstance2.Put(mydynamo.NewPutArgs("s1", gotValue.EntryList[0].Context, []byte("defgh")))
    gotValuePtr = clientInstance1.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestDynamoPaper: Failed to get third value")
    }
    gotValue = *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("cdefg"))){
        t.Fail()
        t.Logf("TestDynamoPaper: Third value doesn't match")
    }
    gotValuePtr = clientInstance2.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestDynamoPaper: Failed to get fourth value")
    }
    gotValue = *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("defgh"))){
        t.Fail()
        t.Logf("TestDynamoPaper: Fourth value doesn't match")
    }
    clientInstance1.Gossip()
    clientInstance2.Gossip()
    gotValuePtr = clientInstance0.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestDynamoPaper: Failed to get fifth value")
    }
    gotValue = *gotValuePtr
    clockList := make([]mydynamo.VectorClock, 0)
    for _, a := range gotValue.EntryList {
        clockList = append(clockList, a.Context.Clock)
    }
    clockList[0].Combine(clockList)
    combinedClock := clockList[0]
    combinedContext := mydynamo.Context {
        Clock:combinedClock,
    }
    clientInstance0.Put(mydynamo.NewPutArgs("s1", combinedContext, []byte("zyxw")))
    gotValuePtr = clientInstance0.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestDynamoPaper: Failed to get sixth value")
    }
    gotValue = *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("zyxw"))){
        t.Fail()
        t.Logf("TestDynamoPaper: Sixth value doesn't match")
    }

}

func TestInvalidPut(t *testing.T){
    t.Logf("Starting repeated Put test")
    cmd := InitDynamoServer("./myconfig.ini")
    ready := make(chan bool)
    go StartDynamoServer(cmd, ready)
    defer KillDynamoServer(cmd)

    time.Sleep(3 * time.Second)
    <-ready
    clientInstance := MakeConnectedClient(8080)

    clientInstance.Put(PutFreshContext("s1", []byte("abcde")))
    clientInstance.Put(PutFreshContext("s1", []byte("efghi")))
    gotValue := clientInstance.Get("s1")
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("abcde"))){
        t.Fail()
        t.Logf("TestInvalidPut: Got wrong value")
    }
}

func TestGossipW2(t *testing.T){
    t.Logf("Starting GossipW2 test")
    cmd := InitDynamoServer("./twoserver.ini")
    ready := make(chan bool)
    go StartDynamoServer(cmd, ready)
    defer KillDynamoServer(cmd)

    time.Sleep(3 * time.Second)
    <-ready

    clientInstance0 := MakeConnectedClient(8080)
    clientInstance1 := MakeConnectedClient(8081)
    clientInstance0.Put(PutFreshContext("s1", []byte("abcde")))
    clientInstance0.Gossip()
    gotValuePtr := clientInstance1.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestGossipW2: Failed to get first element")
    }
    gotValue := *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("abcde"))){
        t.Fail()
        t.Logf("TestGossipW2: Failed to get value")
    }
    clientInstance1.Put(mydynamo.NewPutArgs("s1",gotValue.EntryList[0].Context, []byte("efghi")))

    gotValuePtr = clientInstance1.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestGossipW2: Failed to get")
    }
    gotValue = *gotValuePtr

    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("efghi"))){
        t.Fail()
        t.Logf("GossipW2: Failed to get value")
    }
    gotValuePtr = clientInstance0.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestGossipW2: Failed to get")
    }
    gotValue = *gotValuePtr

    if(len(gotValue.EntryList) != 1) || !valuesEqual(gotValue.EntryList[0].Value, []byte("efghi")){
        t.Fail()
        t.Logf("GossipW2: Failed to get value")
    }

}

func TestReplaceMultipleVersions(t *testing.T){
    t.Logf("Starting ReplaceMultipleVersions test")
    cmd := InitDynamoServer("./myconfig.ini")
    ready := make(chan bool)
    go StartDynamoServer(cmd, ready)
    defer KillDynamoServer(cmd)

    time.Sleep(3 * time.Second)
    <-ready

    clientInstance0 := MakeConnectedClient(8080)
    clientInstance1 := MakeConnectedClient(8081)
    clientInstance0.Put(PutFreshContext("s1", []byte("abcde")))
    clientInstance1.Put(PutFreshContext("s1", []byte("efghi")))
    clientInstance0.Gossip()
    gotValuePtr := clientInstance1.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestReplaceMultipleVersions: Failed to get")
    }

    gotValue := *gotValuePtr
    clockList := make([]mydynamo.VectorClock, 0)
    for _, a := range gotValue.EntryList {
        clockList = append(clockList, a.Context.Clock)
    }
    clockList[0].Combine(clockList)
    combinedClock := clockList[0]
    combinedContext := mydynamo.Context {
        Clock:combinedClock,
    }
    clientInstance1.Put(mydynamo.NewPutArgs("s1", combinedContext, []byte("zxyw")))
    gotValuePtr = nil
    gotValuePtr = clientInstance1.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestReplaceMultipleVersions: Failed to get")
    }


    gotValue = *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("zxyw"))){
        t.Fail()
        t.Logf("testReplaceMultipleVersions: Values don't match")
    }


}

func TestConsistent(t *testing.T){
    t.Logf("Starting Consistent test")
    cmd := InitDynamoServer("./consistent.ini")
    ready := make(chan bool)
    go StartDynamoServer(cmd, ready)
    defer KillDynamoServer(cmd)

    time.Sleep(3 * time.Second)
    <-ready

    clientInstance0 := MakeConnectedClient(8080)
    clientInstance1 := MakeConnectedClient(8081)
    clientInstance2 := MakeConnectedClient(8082)
    clientInstance3 := MakeConnectedClient(8083)
    clientInstance4 := MakeConnectedClient(8084)

    clientInstance0.Put(PutFreshContext("s1", []byte("abcde")))
    gotValuePtr := clientInstance1.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestConsistent: Failed to get")
    }
    gotValue := *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("abcde"))){
        t.Fail()
        t.Logf("TestConsistent: Failed to get value")
    }


    clientInstance3.Put(mydynamo.NewPutArgs("s1", gotValue.EntryList[0].Context, []byte("zyxw")))
    clientInstance0.Crash(3)
    clientInstance1.Crash(3)
    clientInstance4.Crash(3)
    gotValuePtr = clientInstance2.Get("s1")
    if gotValuePtr == nil {
        t.Fail()
        t.Logf("TestConsistent: Failed to get")
    }
    gotValue = *gotValuePtr
    if(len(gotValue.EntryList) != 1 || !valuesEqual(gotValue.EntryList[0].Value, []byte("zyxw"))){
        t.Fail()
        t.Logf("TestConsistent: Failed to get value")
    }
}
