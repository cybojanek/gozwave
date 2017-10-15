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
	if response, err := node.zwSendDataWaitForResponse(
		CommandClassBinarySwitch, []uint8{binarySwitchCommandGet},
		binarySwitchCommandReport); err != nil {
		return false, err
	} else if len(response.Command.Data) != 1 {
		return false, fmt.Errorf("Bad response")
	} else {
		return response.Command.Data[0] != 0, nil
	}
}
