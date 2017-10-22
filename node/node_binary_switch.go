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
	binarySwitchCommandSet    uint8 = 0x01
	binarySwitchCommandGet          = 0x02
	binarySwitchCommandReport       = 0x03
)

// BinarySwitch information
type BinarySwitch struct {
	*Node
}

// GetBinarySwitch returns a BinarySwitch or nil object
func (node *Node) GetBinarySwitch() *BinarySwitch {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassBinarySwitch) {
		return &BinarySwitch{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// On turns the switch on
func (node *BinarySwitch) On() error {
	return node.zwSendDataRequest(CommandClassBinarySwitch,
		[]uint8{binarySwitchCommandSet, 0xff})
}

// Off turns the switch off
func (node *BinarySwitch) Off() error {
	return node.zwSendDataRequest(CommandClassBinarySwitch,
		[]uint8{binarySwitchCommandSet, 0x00})
}

// IsOn queries the switch to check current status
func (node *BinarySwitch) IsOn() (bool, error) {
	var response *ApplicationCommandData
	var err error

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassBinarySwitch, []uint8{binarySwitchCommandGet},
		binarySwitchCommandReport, nil); err != nil {
		return false, err
	}

	return node.ParseReport(response)
}

// IsReport checks if the report is a ParseReport
func (node *BinarySwitch) IsReport(report *ApplicationCommandData) bool {
	return report.Command.ID == binarySwitchCommandReport
}

// ParseReport of status
func (node *BinarySwitch) ParseReport(report *ApplicationCommandData) (bool, error) {
	if report.Command.ClassID != CommandClassBinarySwitch {
		return false, fmt.Errorf("Bad Report Command Class ID: 0x%02x != 0x%02x",
			report.Command.ClassID, CommandClassBinarySwitch)
	}

	if report.Command.ID != binarySwitchCommandReport {
		return false, fmt.Errorf("Bad Report Command ID 0x%02x != 0x%02x",
			report.Command.ID, binarySwitchCommandReport)
	}

	data := report.Command.Data
	if len(data) != 1 {
		return false, fmt.Errorf("Bad Report Data length %d != 1", len(data))
	}

	return data[0] != 0, nil
}
