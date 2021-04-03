package mydynamotest

import (
	"mydynamo"
	"testing"
)

// Create 2 vector clocks, check for equality
func TestBasicVectorClock(t *testing.T) {
	t.Logf("Starting TestBasicVectorClock")

	//create two vector clocks
	clock1 := mydynamo.NewVectorClock()
	clock2 := mydynamo.PresetVectorClock(map[string]uint{})

	//Test for equality
	if !clock1.Equals(clock2) {
		t.Fail()
		t.Logf("Vector Clocks were not equal")
	}

}

func TestIncrement(t *testing.T) {
	t.Logf("Starting TestIncrement")

	//create two vector clocks
	clock := mydynamo.NewVectorClock()
	clock.Increment("id1")
	clock.Increment("id2")
	clock.Increment("id2")

	ts := map[string]uint{"id1" : 1, "id2" : 2}
	pclock := mydynamo.PresetVectorClock(ts)

	//Test for equality
	if !clock.Equals(pclock) {
		t.Fail()
		t.Logf("Vector Clocks were not equal")
	}
}


func TestLessThan(t *testing.T) {
	t.Logf("Starting TestLessThan")

	ts1 := map[string]uint{}
	ts2 := map[string]uint{"id1" : 1}
	ts3 := map[string]uint{"id1" : 2}
	ts4 := map[string]uint{"id1" : 1, "id2" : 2}

	c1 := mydynamo.PresetVectorClock(ts1)
	c2 := mydynamo.PresetVectorClock(ts2)
	c3 := mydynamo.PresetVectorClock(ts3)
	c4 := mydynamo.PresetVectorClock(ts4)

	if !c1.LessThan(c2) {
		t.Fail()
		t.Logf("Vector Clock '%v' was not LessThan '%v'", c1, c2)
	}

	if !c2.LessThan(c3) {
		t.Fail()
		t.Logf("Vector Clock '%v' was not LessThan '%v'", c2, c3)
	}

	if c3.LessThan(c4) {
		t.Fail()
		t.Logf("Vector Clock '%v' was LessThan '%v'", c3, c4)
	}

	if !c2.LessThan(c4) {
		t.Fail()
		t.Logf("Vector Clock '%v' was not LessThan '%v'", c2, c4)
	}

	if c4.LessThan(c4) {
		t.Fail()
		t.Logf("Vector Clock '%v' was LessThan '%v'", c4, c4)
	}
}


func TestConcurrent(t *testing.T) {
	t.Logf("Starting TestConcurrent")

	ts1 := map[string]uint{}
	ts2 := map[string]uint{"id1" : 1}
	ts3 := map[string]uint{"id2" : 2}
	ts4 := map[string]uint{"id1" : 2, "id2" : 1}

	c1 := mydynamo.PresetVectorClock(ts1)
	c2 := mydynamo.PresetVectorClock(ts2)
	c3 := mydynamo.PresetVectorClock(ts3)
	c4 := mydynamo.PresetVectorClock(ts4)

	if c1.Concurrent(c2) {
		t.Fail()
		t.Logf("Vector Clock '%v' was Concurrent to '%v'", c1, c2)
	}

	if c2.Concurrent(c4) {
		t.Fail()
		t.Logf("Vector Clock '%v' was Concurrent to '%v'", c2, c4)
	}

	if !c2.Concurrent(c3) {
		t.Fail()
		t.Logf("Vector Clock '%v' was not Concurrent to '%v'", c2, c3)
	}

	if !c3.Concurrent(c4) {
		t.Fail()
		t.Logf("Vector Clock '%v' was not Concurrent to '%v'", c3, c4)
	}

}


func TestCombine(t *testing.T) {
	t.Logf("Starting TestCombine")
	var ts1, ts2, ts3, tsr map[string]uint
	var c1, c2, c3, cr mydynamo.VectorClock

	// collapse between causal values
	ts1 = map[string]uint{}
	ts2 = map[string]uint{"id1" : 2}
	c1 = mydynamo.PresetVectorClock(ts1)
	c2 = mydynamo.PresetVectorClock(ts2)

	c1.Combine([]mydynamo.VectorClock{c2})
	if !c1.Equals(c2) {
		t.Fail()
		t.Logf("Vector Clocks were not equal")
		t.Logf("\tExpected %v  Got %v", cr, c1)
	}

	// multiple keys
	ts1 = map[string]uint{"id1" : 1}
	ts2 = map[string]uint{"id2" : 2}
	tsr = map[string]uint{"id1" : 1, "id2" : 2}
	c1 = mydynamo.PresetVectorClock(ts1)
	c2 = mydynamo.PresetVectorClock(ts2)
	cr = mydynamo.PresetVectorClock(tsr)

	c1.Combine([]mydynamo.VectorClock{c2})
	if !c1.Equals(cr) {
		t.Fail()
		t.Logf("Vector Clocks were not equal")
		t.Logf("\tExpected %v  Got %v", cr, c1)
	}

	// multiple clocks
	ts1 = map[string]uint{"id1" : 1, "id2": 3}
	ts2 = map[string]uint{"id2" : 2, "id3": 2}
	ts3 = map[string]uint{"id1" : 4, "id4": 1}
	tsr = map[string]uint{"id1" : 4, "id2" : 3, "id3": 2, "id4": 1}
	c1 = mydynamo.PresetVectorClock(ts1)
	c2 = mydynamo.PresetVectorClock(ts2)
	c3 = mydynamo.PresetVectorClock(ts3)
	cr = mydynamo.PresetVectorClock(tsr)

	c1.Combine([]mydynamo.VectorClock{c2, c3})
	if !c1.Equals(cr) {
		t.Fail()
		t.Logf("Vector Clocks were not equal")
		t.Logf("\tExpected %v  Got %v", cr, c1)
	}
}
