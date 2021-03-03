package internal

import "strings"

func IsSonobouyPod(podName string) bool {
	if podName == "sonobuoy" || strings.HasPrefix(podName, "sonobuoy-reliability-scanner-job") {
		return true
	}
	return false
}
