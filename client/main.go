/*
Copyright © 2019 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package main

import (
	"log"

	pb "github.com/asgaines/blockchain/protogo/blockchain"
	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "cli",
	Short: "CLI for Domain Blocklist gRPC service",
}

func main() {
	pb.NodeClientCommand.Short = "Blockchain gRPC client"
	cmd.AddCommand(pb.NodeClientCommand)

	if err := cmd.Execute(); err != nil {
		log.Fatalf("Failed running command: %s", err)
	}
}
