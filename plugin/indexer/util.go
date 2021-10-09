package main

func uintMin(a, b uint) uint {
	if a < b {
		return a
	}
	return b
}

func int64ToBool(i int64) bool {
	if i == 0 {
		return false
	}
	return true
}
