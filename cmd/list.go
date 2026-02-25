package cmd

import (
	"devserve/internal"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List processes",
	RunE: func(cmd *cobra.Command, args []string) error {
		req := &internal.Request{
			Action: "list",
		}
		resp, err := internal.Send(req)
		if err != nil {
			return err
		}
		if !resp.OK {
			return errors.New(resp.Error)
		}
		fmt.Println(resp.Data)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
}
