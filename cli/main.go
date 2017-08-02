// Copyright © 2017 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"github.com/spf13/cobra"
	"fmt"
	"os"
	"github.com/jskirchmeier/gifting"
	"time"
	"errors"
)


var dataStore *gifting.DataStore

// rootCmd represents the base command when called without any sub commands
var rootCmd = &cobra.Command{
	Use:   "gifting command pathToDataFile",
	Example:   "gifting [solve,report] pathToDataFile",
	Short: "Calculate a new year of gifting",
	Long:  `Replaces drawing a name from a hat for groups that gift to only one person within that group.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return errors.New("path to data file required")
		}
		ds, err := gifting.ReadAsML(args[0])
		dataStore = ds
		return err
	},
}

var reportCmd = &cobra.Command{
	Use:   "report pathToDataFile",
	Short: "Produce gifting report for the desired year",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		// data store should have been created by rootCmd.PersistentPreRunE
		year,_ := cmd.Flags().GetInt("year")
		dataStore.GiftReport(year)
	},
}

var solveCmd = &cobra.Command{
	Use:   "solve pathToDataFile",
	Short: "Calculate the optimum gifting pairs for the desired year",
	Long: ``,
	Run: func(cmd *cobra.Command, args []string) {
		saveFile,_ := cmd.Flags().GetString("save")

		solver := gifting.Solver{}
		solver.SetData(dataStore)
		solver.Solve()
		solver.Statistics()
		if saveFile != "" {
			// save the file
			dataStore.SaveAsXML(saveFile)
		}
	},
}




func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(os.Stderr, err)
	}
}

func init() {
	rootCmd.AddCommand(solveCmd)
	rootCmd.AddCommand(reportCmd)

	reportCmd.Flags().IntP("year", "y", time.Now().Year(), "year to use, defaults to current year")
	solveCmd.Flags().StringP("save", "s", "", "name of file to save to, if not specified results not saved")
}
