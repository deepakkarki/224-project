package mydynamotest

import (
	"mydynamo"
	"testing"
)

type TS map[string]uint


// Test the collapse function
func TestCollapseObjects(t *testing.T) {
	t.Logf("Starting TestCollapseObjects")
	// l1= [{x:2, y:2}, {x:2, y:1, z:1}], l2= [{x:1, y:1, z:1}, {x:3, y:3}]
	//then CollapseObjects returns [{x:2, y:1, z:1}, {x:3, y:3}]

	o1 := MakeObjectEntry(TS{"x":2, "y":2})
	o2 := MakeObjectEntry(TS{"x":2, "y":1, "z":1})
	o3 := MakeObjectEntry(TS{"x":1, "y":1, "z":1})
	o4 := MakeObjectEntry(TS{"x":3, "y":3})

	l1 := []mydynamo.ObjectEntry{o1, o2}
	l2 := []mydynamo.ObjectEntry{o3, o4}
	lr := []mydynamo.ObjectEntry{o2, o4}

	res, count := mydynamo.CollapseObjects(l1, l2)
	t.Logf("Collapsed Entries: %v", res)

	if !(ObjectEquals(res[0], lr[0]) && ObjectEquals(res[1], lr[1])){
		t.Fail()
		t.Logf("Object Entries were not equal")
		t.Logf("\tExpected %v", lr)
		t.Logf("\tGot      %v", res)
	}

	if count != 1 {
		t.Fail()
		t.Logf("Count: %d. Expected value : 1", count)
	}

	//TODO: Might need to test a few more cases
}