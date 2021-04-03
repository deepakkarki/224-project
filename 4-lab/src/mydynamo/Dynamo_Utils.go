package mydynamo

/*
 function takes in two list of objects (with no causaiblity b/w any two objs
 of the same list) and returns a thrid list containing non-causal values.
 i.e. l1, l2 have only concurrent values, so does the return values.

 It also returns the number of elements in l2 that have made it to the result

 Eg. [{x:2, y:2}, {x:2, y:1, z:1}] and [{x:1, y:1, z:1}, {x:3, y:3}]
 then CollapseObjects returns [{x:3, y:3}, {x:2, y:1, z:1}]
*/
func CollapseObjects(l1, l2 []ObjectEntry) ([]ObjectEntry, int) {
	res := []ObjectEntry{}
	count := 0

	//NOTE: vector clocks can either be => equal, greater-than, less-than, or concurrent

	// for each element in l1, see if it's less-than any element in l2
	for _, o1 := range l1 {
		canAdd := true
		for _, o2 := range l2 {
			if o1.Context.Clock.LessThan(o2.Context.Clock) {
				// as far as o1 !< o2, it will be in the result
				canAdd = false
				break
			}
		}
		if canAdd {
			res = append(res, o1)
		}
	}

  // for each element in l2, see if it's greater-than or concurrent to each
  // element in l1. Implies, o2 shouldn't be '<' or '=' to any o1
	for _, o2 := range l2 {
		canAdd := true
		for _, o1 := range l1 {
			clk1 := o1.Context.Clock
			clk2 := o2.Context.Clock
			if clk2.LessThan(clk1) || clk2.Equals(clk1) {
				canAdd = false
				break
			}
		}
		if canAdd {
			res = append(res, o2)
			count++
		}
	}

	return res, count
}

func isContextEmpty(ctx Context) bool {
	return len(ctx.Clock.Timestamp) == 0
}

//Removes an element at the specified index from a list of ObjectEntry structs
func remove(list []ObjectEntry, index int) []ObjectEntry {
	return append(list[:index], list[index+1:]...)
}

//Returns true if the specified list of ints contains the specified item
func contains(list []int, item int) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

//Rotates a preference list by one, so that we can give each node a unique preference list
func RotateServerList(list []DynamoNode) []DynamoNode {
	return append(list[1:], list[0])
}

//Creates a new Context with the specified Vector Clock
func NewContext(vClock VectorClock) Context {
	return Context{
		Clock: vClock,
	}
}

//Creates a new PutArgs struct with the specified members.
func NewPutArgs(key string, context Context, value []byte) PutArgs {
	return PutArgs{
		Key:     key,
		Context: context,
		Value:   value,
	}
}

//Creates a new DynamoNode struct with the specified members
func NewDynamoNode(addr string, port string) DynamoNode {
	return DynamoNode{
		Address: addr,
		Port:    port,
	}
}
