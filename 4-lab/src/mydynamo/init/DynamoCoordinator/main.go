package main

import (
	"fmt"
	"github.com/go-ini/ini"
	"log"
	"mydynamo"
	"net/rpc"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {
	var err error
	/*-----------------------------*/
	// When the input argument is less than 1
	if len(os.Args) != mydynamo.ARG_COUNT {
		log.Println(mydynamo.USAGE_STRING)
		os.Exit(mydynamo.EX_USAGE)
	}

	// Load the configuration file
	configFilePath := os.Args[mydynamo.CONFIG_FILE_INDEX]
	configContent, err := ini.Load(configFilePath)
	if err != nil {
		log.Println(err)
		log.Println("Failed to load config file:", configFilePath)
		log.Println(mydynamo.USAGE_STRING)
		os.Exit(mydynamo.EX_CONFIG)
	}

	// Load the detailed configuration from section "mydynamo"
	dynamoConfigs := configContent.Section(mydynamo.MYDYNAMO)

	serverPort, err := dynamoConfigs.Key(mydynamo.SERVER_PORT).Int()
	r_value, err := dynamoConfigs.Key(mydynamo.R_VALUE).Int()
	w_value, err := dynamoConfigs.Key(mydynamo.W_VALUE).Int()
	cluster_size, err := dynamoConfigs.Key(mydynamo.CLUSTER_SIZE).Int()
	if err != nil {
		log.Println(err)
		log.Println("Failed to load config file, field is wrong type:", configFilePath)
		log.Println(mydynamo.USAGE_STRING)
		os.Exit(mydynamo.EX_CONFIG)
	}
	fmt.Println("Done loading configurations")

	//keep a list of servers so we can communicate with them
	serverList := make([]mydynamo.DynamoServer, 0)

	//spin up a dynamo cluster
	dynamoNodeList := make([]mydynamo.DynamoNode, 0)

	//Use a waitgroup to ensure that we don't exit this goroutine until all servers have exited
	wg := new(sync.WaitGroup)
	wg.Add(cluster_size)
	for idx := 0; idx < cluster_size; idx++ {

		//Create a server instance
		serverInstance := mydynamo.NewDynamoServer(w_value, r_value, "localhost", strconv.Itoa(serverPort+idx), strconv.Itoa(idx))
		serverList = append(serverList, serverInstance)

		//Create an anonymous function in a goroutine that starts the server
		go func() {
			log.Fatal(mydynamo.ServeDynamoServer(serverInstance))
			wg.Done()
		}()
		nodeInfo := mydynamo.DynamoNode{
			Address: "localhost",
			Port:    strconv.Itoa(serverPort + idx),
		}
		dynamoNodeList = append(dynamoNodeList, nodeInfo)
	}
	time.Sleep(1 * time.Second)

	//Create a duplicate of dynamoNodeList that we can rotate
	//so that each node has a distinct preference list
	nodePreferenceList := dynamoNodeList

	//Send the preference list to all servers
	for _, info := range dynamoNodeList {
		var empty mydynamo.Empty
		c, err := rpc.DialHTTP("tcp", info.Address+":"+info.Port)
		if err != nil {
			log.Println("Failed to send preference list")
		} else {
			err2 := c.Call("MyDynamo.SendPreferenceList", nodePreferenceList, &empty)
			if err2 != nil {
				log.Println("Failed to send preference list")
			}
		}
		nodePreferenceList = mydynamo.RotateServerList(nodePreferenceList)
	}
	/*---------------------------------------------*/

	//wait for all servers to finish
	wg.Wait()
}
