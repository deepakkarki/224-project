package mydynamo

// VectorClock is a multipart timestamp.
// maps node-ids to the "timestamp" of each node
type VectorClock struct {
	Timestamp  map[string]uint
}

//Creates a new VectorClock
func NewVectorClock() VectorClock {
	v := VectorClock{Timestamp: map[string]uint{}}
	return v
}

//Creates a vector clock based on the provided timestamp
func PresetVectorClock(ts map[string]uint) VectorClock {
	v := VectorClock{Timestamp: ts}
	return v
}

//Returns a copy of the timestamp of the clock
func (s *VectorClock) GetTimestamp() map[string]uint {
	ts := map[string]uint{}

	for k, v := range s.Timestamp {
		ts[k] = v
	}
	return ts
}

//Increments this VectorClock at the element associated with nodeId
func (s *VectorClock) Increment(nodeId string) {
	s.Timestamp[nodeId] += 1
}

//Changes this VectorClock to be causally descended from all VectorClocks in clocks
func (s *VectorClock) Combine(clocks []VectorClock) {
	ts := map[string]uint{}

	for k, v := range s.Timestamp {
		ts[k] = v
	}

	for _, clock := range clocks {
		for k, v := range clock.Timestamp {
			if v > ts[k] {
				ts[k] = v
			}
		}
	}
	s.Timestamp = ts

}

//Returns true if the other VectorClock is causally descended from this one
func (s VectorClock) LessThan(otherClock VectorClock) bool {
	// if 's' is longer, no way it is an ancestor to otherClock
	if len(s.Timestamp) > len(otherClock.Timestamp) {
		return false
	}

	if s.Equals(otherClock) {
		return false
	}

	// for each node-id (k) in 's' check if it exists in otherClock
	for k, vs := range s.Timestamp {
		if vo, ok := otherClock.Timestamp[k]; ok {
			if vs > vo {
				// otherClock can't be decendent of 's' if OC has lesser value
				return false
			}
		} else {
			// if the id doesn't exist, otherClock can't be decendent of 's'
			return false
		}
	}
	return true
}

//Tests if two VectorClocks are equal
func (s *VectorClock) Equals(otherClock VectorClock) bool {
	if len(s.Timestamp) != len(otherClock.Timestamp) {
		return false
	}

	// for each node-id (k) in 's' check if it exists in otherClock
	for k, vs := range s.Timestamp {
		if vo, ok := otherClock.Timestamp[k]; ok {
			if vs != vo {
				return false
			}
		} else {
			// if the id doesn't exist, clocks can't be equal
			return false
		}
	}
	return true
}

//Returns true if neither VectorClock is causally descended from the other
func (s VectorClock) Concurrent(otherClock VectorClock) bool {
	// vector clocks can either be => equal, greater-than, less-than, or concurrent
	if !(s.Equals(otherClock) || s.LessThan(otherClock) || otherClock.LessThan(s)){
		return true //if not any other option
	}
	return false
}

