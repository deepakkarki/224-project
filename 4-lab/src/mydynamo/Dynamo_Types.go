package mydynamo
import "fmt"

//Placeholder type for RPC functions that don't need an argument list or a return value
type Empty struct{}

//Context associated with some value
type Context struct {
	Clock VectorClock
}

//Information needed to connect to a DynamoNOde
type DynamoNode struct {
	Address string
	Port    string
}

func (n DynamoNode) String() string {
	return fmt.Sprintf("%v:%v", n.Address, n.Port)
}

//A single value, as well as the Context associated with it
type ObjectEntry struct {
	Context Context
	Value   []byte
}

func (o ObjectEntry) String() string {
	return fmt.Sprintf("(%v, %s)", o.Context.Clock.Timestamp, string(o.Value))
}

//Result of a Get operation, a list of ObjectEntry structs
type DynamoResult struct {
	EntryList []ObjectEntry
}

//Arguments required for a Put operation: the key, the context, and the value
type PutArgs struct {
	Key     string
	Context Context
	Value   []byte
}

//Arguments required for a S_Put operation: key, list of value-clock pairs
type S_PutArgs struct {
	Key     string
	EntryList []ObjectEntry
}
