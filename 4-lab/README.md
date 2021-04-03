# Project Milestone 4: DynamoDB

This is the starter code for Module 4: DynamoDB. Before you get started, you will want to understand the following things
- Vector clocks
- Quorums

Additionally, you will want to read and understand this paper: [DynamoDB paper](https://www.allthingsdistributed.com/files/amazon-dynamo-sosp2007.pdf)

## Data Types
The starter code defines the following (long) list of types for your usage in `Dynamo_Types.go`:
```golang
//Placeholder type for RPC functions that don't need an argument list or a return value
type Empty struct {}

//Context associated with some value
type Context struct {
    Clock VectorClock
}

//Information needed to connect to a DynamoNOde
type DynamoNode struct {
    Address string
    Port    string
}

//A single value, as well as the Context associated with it
type ObjectEntry struct {
    Context Context
    Value  []byte
}

//Result of a Get operation, a list of ObjectEntry structs
type DynamoResult struct {
    EntryList []ObjectEntry
}

//Arguments required for a Put operation: the key, the context, and the value
type PutArgs struct {
    Key     string
    Context Context
    Value  []byte
}
```
These types are intended to be used for the RPC interfaces, so please *do not modify them*. However, feel free to add other types to the `Dynamo_Types.go` file if you feel the need.

## Vector Clocks
A skeleton implementation of a vector clock is in the file `Dynamo_VectorClock.go` **You must implement the methods that have `panic("todo")` as their body. You must also fill in the VectorClock struct at the location specified by `//todo`** Feel free to add helper methods as you see fit.

## Dynamo Nodes
A skeleton implementation of a Dynamo node is in the file `Dynamo_Server.go`. This file defines an RPC interface for a Dynamo node. As a result, please *do not modify the signatures of the given functions and methods*. 
**You must implement the methods that have `panic("todo")` as their body.** Feel free to add more methods or functions as you see fit, but keep in mind to expose them to the RPC interface, they need to be of the appropriate signature.
Additionally, feel free to add members to the DynamoServer struct as you see fit, but remember to initialize them in `NewDynamoServer()` if the members need initialization.

## Dynamo Client
An RPC client has been given to you in the file `Dynamo_Client.go`. You should not need to modify this file, but 
make sure to familiarize yourself with it, as you will use it extensively in testing.

## Utility Functions
A couple utility functions have been provided in the file `Dynamo_Utils.go`. These functions may be helpful when you are writing your code. Feel free to add more functions as you need to this file.

## Setup
You will need to setup your runtime environment variables so that you can build your code and also use the executables that will be generated.
1. If you are using a Mac, open `~/.bash_profile` or if you are using a unix/linux machine, open `~/.bashrc`. Then add the following:
```
export GOPATH=<path to starter code>
export PATH=$PATH:$GOPATH/bin
```
2. Run `source ~/.bash_profile` or `source ~/.bashrc`

## Usage
### Building the code
To build the code, run 
```
./build.sh
```
This should generate `bin/DynamoCoordinator` and `bin/DynamoClient`

### Running the code
To start up a set of nodes, run
```
./run-server.sh [config file]
```
where `config file` is a .ini file. We have provided `myconfig.ini` as an example config file for you to use.
To run your server in the background, you can use
```
nohup ./run-server.sh [config file] &
```
This will start your server in the background and append the output to a file called `nohup.out`

To run your client, run
```
./run-client.sh
```

### Unit Testing
To test your code, navigate to `src/mydynamotest/` and run
```
go test
```

To run a specific test in your code, run

```
go test -run [testname]
```
**Keep in mind that although `go test` recompiles your testing files, it does not recompile all of your code. If you made modifications to any file in `src/mydynamo`, you will have to run `build.sh` again**

Two tests have been provided for you, in `src/mydynamotest/basic_test.go` and `src/mydynamotest/vectorclock_test.go`. Follow the format of these tests to create your own unit tests.

For more information on testing, visit the [Go documentation](https://golang.org/pkg/testing/)
