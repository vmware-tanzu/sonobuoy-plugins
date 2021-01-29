package qos

func meetsMinimum(test, minimum string) bool {
	hierarchy := make(map[string]int)

	hierarchy["Guaranteed"] = 3
	hierarchy["Burstable"] = 2
	hierarchy["BestEffort"] = 1

	if hierarchy[test] >= hierarchy[minimum] {
		return true
	}
	return false
}
