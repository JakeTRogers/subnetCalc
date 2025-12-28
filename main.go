/*
Copyright © 2025 Jake Rogers <code@supportoss.org>
*/
package main

import (
	"os"

	"github.com/JakeTRogers/subnetCalc/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
