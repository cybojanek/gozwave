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
	"encoding/binary"
	"fmt"
)

const (
	manufacturerSpecificCommandGet    uint8 = 0x04
	manufacturerSpecificCommandReport       = 0x05
)

// ManufacturerSpecific information
type ManufacturerSpecific struct {
	*Node
}

// GetManufacturerSpecific returns a ManufacturerSpecific or nil object
func (node *Node) GetManufacturerSpecific() *ManufacturerSpecific {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassManufacturerSpecific) {
		return &ManufacturerSpecific{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Get manufacturer and product information
func (node *ManufacturerSpecific) Get() (manufacturerID uint16, productType uint16, productID uint16, err error) {
	var response *ApplicationCommandData

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassManufacturerSpecific, []uint8{manufacturerSpecificCommandGet},
		manufacturerSpecificCommandReport, nil); err != nil {
		return
	}

	data := response.Command.Data
	if len(data) != 6 {
		err = fmt.Errorf("Bad data length: %d != 6", len(data))
		return
	}

	manufacturerID = binary.BigEndian.Uint16(data[0:2])
	productType = binary.BigEndian.Uint16(data[2:4])
	productID = binary.BigEndian.Uint16(data[4:6])
	return
}
