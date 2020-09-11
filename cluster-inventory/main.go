package main

import (
	"fmt"
	"os"

	"github.com/vmware-tanzu/sonobuoy-plugins/cluster-inventory/cmd"
)

func main() {
	err := cmd.NewClusterInventoryCommand().Execute()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
