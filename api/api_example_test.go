package api_test

/*
Copyright (C) 2017 Jan Kasiak

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

import (
	"fmt"
	"github.com/cybojanek/gozwave/api"
)

func Example() {
	// Create API device
	zwapi := api.ZWAPI{DevicePath: "/dev/tty.usbmodem1451",
		DebugLogging: true}

	// Open API device
	if err := zwapi.Open(); err != nil {
		fmt.Printf("Failed to Open: %v", err)
		return
	}

	// Initialize - gets controller information, list of nodes on network
	if err := zwapi.Initialize(); err != nil {
		fmt.Printf("Failed to Initialize: %v", err)
		return
	}

	// Iterate over nodes and refresh them getting more specific information
	// about their supported operations
	for _, node := range zwapi.GetNodes() {
		if err := node.Refresh(); err != nil {
			fmt.Printf("Node: %d failed to refresh: %v\n", node.ID, err)
		} else {
			fmt.Printf("Node: %+v\n", node)
		}
	}

	// Get specific node
	if node := zwapi.GetNode(5); node != nil {
		fmt.Printf("Node: %d exists!\n", node.ID)

		// Do node work ... from many go routines
		if bs := node.GetBinarySwitch(); bs != nil {
			// Node is a binary switch!
			if err := bs.Off(); err != nil {
				fmt.Printf("Failed to turn off switch: %v\n", err)
			}
			// ...
		}
	} else {
		fmt.Printf("Node: %d does not exist on the network\n", node.ID)
	}

	// Do more work ... from many go routines

	// Close API
	if err := zwapi.Close(); err != nil {
		fmt.Printf("Failed to Close: %v", err)
		return
	}
}
