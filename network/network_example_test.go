package network_test

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
	"github.com/cybojanek/gozwave/network"
	"github.com/cybojanek/gozwave/node"
	"time"
)

func Example() {
	// Create network
	net := network.Network{DevicePath: "/dev/tty.usbmodem1451",
		DebugLogging: true}

	// Open
	if err := net.Open(); err != nil {
		fmt.Printf("Failed to Open: %v", err)
		return
	}

	// Initialize - gets controller information, list of nodes on network
	if err := net.Initialize(); err != nil {
		fmt.Printf("Failed to Initialize: %v", err)
		return
	}

	// Iterate over nodes and refresh them, querying for more specific
	// information about their supported operations
	for _, n := range net.GetNodes() {
		if _, err := n.Load(nil); err != nil {
			fmt.Printf("Node: %d failed to refresh: %v\n", n.ID, err)
		} else {
			fmt.Printf("Node: %+v\n", n)
		}
	}

	// Get specific node
	if n := net.GetNode(5); n != nil {
		fmt.Printf("Node: %d exists!\n", n.ID)

		// Do node work ... from many go routines
		if bs := n.GetBinarySwitch(); bs != nil {
			// Node is a binary switch!
			if err := bs.Off(); err != nil {
				fmt.Printf("Failed to turn off switch: %v\n", err)
			}
			// ...
		}
	} else {
		fmt.Printf("Node: %d does not exist on the network\n", 5)
	}

	// Register for application callbacks (optional)
	applicationCommandChannel := make(chan *node.ApplicationCommandData, 1)
	applicationUpdateChannel := make(chan *node.ApplicationUpdateData, 1)

	for _, n := range net.GetNodes() {
		n.AddApplicationCommandCallbackChannel(applicationCommandChannel)
		n.AddApplicationUpdateCallbackChannel(applicationUpdateChannel)
	}

outer:
	for {
		select {
		case report := <-applicationCommandChannel:
			n := net.GetNode(report.NodeID)
			if n == nil {
				fmt.Printf("Unknown node: %d\n", report.NodeID)
				continue
			}

			switch report.Command.ClassID {
			case node.CommandClassAlarm:
				alarm := n.GetAlarm()
				if alarm == nil {
					fmt.Printf("Node %d is not an alarm\n", n.ID)
					continue
				}

				if alarm.IsReport(report) {
					if isActive, alarmType, err := alarm.ParseReport(report); err != nil {
						fmt.Printf("Failed to parse alarm report: %v\n", err)
					} else {
						fmt.Printf("Node %d alarm status: %v %v\n", n.ID, isActive, alarmType)
					}
				}

			case node.CommandClassBattery:
				battery := n.GetBattery()
				if battery == nil {
					fmt.Printf("Node %d is not a battery\n", n.ID)
					continue
				}

				if battery.IsReport(report) {
					if isLow, value, err := battery.ParseReport(report); err != nil {
						fmt.Printf("Failed to parse battery report: %v\n", err)
					} else {
						fmt.Printf("Node %d battery: isLow? %v, level: %d\n",
							n.ID, isLow, value)
					}
				}

			case node.CommandClassBinarySensor:
				sensor := n.GetBinarySensor()
				if sensor == nil {
					fmt.Printf("Node %d is not a binary sensor\n", n.ID)
					continue
				}

				if sensor.IsReport(report) {
					if active, sensorType, err := sensor.ParseReport(report); err != nil {
						fmt.Printf("Failed to parse binary sensor report: %v\n", err)
					} else {
						fmt.Printf("Active? %v sensorType: %d\n", active, sensorType)
					}
				}

			case node.CommandClassBinarySwitch:
				bs := n.GetBinarySwitch()
				if bs == nil {
					fmt.Printf("Node %d is not a binary switch\n", n.ID)
					continue
				}

				if bs.IsReport(report) {
					if isOn, err := bs.ParseReport(report); err != nil {
						fmt.Printf("Failed to parse binary switch report: %v\n", err)
					} else {
						fmt.Printf("CC Node %d isOn? %v\n", n.ID, isOn)
					}
				}
			}

		case update := <-applicationUpdateChannel:
			n := net.GetNode(update.NodeID)
			if n == nil {
				fmt.Printf("Unknown node: %d\n", update.NodeID)
				continue
			}

			// NOTE: binary switches without association only send a general
			// 		 message that the button was pressed, and not what state
			// 		 it is in
			if bs := n.GetBinarySwitch(); bs != nil {
				if isOn, err := bs.IsOn(); err != nil {
					fmt.Printf("Failed to check switch status: %v\n", err)
				} else {
					fmt.Printf("AU Node %d isOn? %v\n", n.ID, isOn)
				}
			}

		case <-time.After(1 * time.Minute):
			fmt.Printf("No activity for 1 minute - exiting\n")
			break outer
		}
	}

	for _, n := range net.GetNodes() {
		n.RemoveApplicationCommandCallbackChannel(applicationCommandChannel)
		n.RemoveApplicationUpdateCallbackChannel(applicationUpdateChannel)
	}

	// Close API
	if err := net.Close(); err != nil {
		fmt.Printf("Failed to Close: %v", err)
		return
	}
}
