package node

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
)

const (
	batteryCommandGet    uint8 = 0x02
	batteryCommandReport       = 0x03
)

// Battery information
type Battery struct {
	*Node
}

// GetBattery returns a Battery or nil object
func (node *Node) GetBattery() *Battery {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassBattery) {
		return &Battery{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Get queries node to check if the battery is low
func (node *Battery) Get() (isLow bool, level uint8, err error) {
	var response *ApplicationCommandData

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassBattery, []uint8{batteryCommandGet},
		batteryCommandReport, nil); err != nil {
		return false, 0, err
	}

	return node.ParseReport(response)
}

// IsReport checks if the report is a ParseReport
func (node *Battery) IsReport(report *ApplicationCommandData) bool {
	return report.Command.ID == batteryCommandReport
}

// ParseReport of status
func (node *Battery) ParseReport(report *ApplicationCommandData) (isLow bool, level uint8, err error) {
	if report.Command.ClassID != CommandClassBattery {
		return false, 0, fmt.Errorf("Bad Report Command Class ID: 0x%02x != 0x%02x",
			report.Command.ClassID, CommandClassBattery)
	}

	if report.Command.ID != batteryCommandReport {
		return false, 0, fmt.Errorf("Bad Report Command ID 0x%02x != 0x%02x",
			report.Command.ID, batteryCommandReport)
	}

	data := report.Command.Data
	if len(data) != 1 {
		return false, 0, fmt.Errorf("Bad Report Data length %d != 1", len(data))
	}

	// Level is 0-100 or 255
	level = data[0]
	isLow = data[0] == 0xff
	if isLow {
		level = 0
	}

	return
}
