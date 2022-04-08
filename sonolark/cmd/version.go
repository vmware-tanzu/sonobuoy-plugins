package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/vmware-tanzu/sonobuoy-plugins/sonolark/lib/buildinfo"
)

func NewCmdVersion() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print Sonolark version",
		Run: func(cmd *cobra.Command, args []string) {
			printVersion(os.Stdout)
		},
		Args: cobra.ExactArgs(0),
	}

	return cmd
}

func printVersion(w io.Writer) {
	b, err := json.Marshal(buildinfo.Info)
	if err != nil {
		panic(err)
	}
	fmt.Fprintf(w, string(b))
}
