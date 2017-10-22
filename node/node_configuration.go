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

// NOTE: If a value does not exist, then a single byte value of 0 might be returned

import (
	"encoding/binary"
	"fmt"
)

const (
	configurationSet    uint8 = 0x04
	configurationGet          = 0x05
	configurationReport       = 0x06
)

// Configuration information
type Configuration struct {
	*Node
}

// GetConfiguration returns a Configuration or nil object
func (node *Node) GetConfiguration() *Configuration {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassConfiguration) {
		return &Configuration{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Internal function to get parameter value, with expected size
func (node *Configuration) getValue(parameter uint8, size uint8) ([]uint8, error) {
	// Check size
	if size != 1 && size != 2 && size != 4 {
		return nil, fmt.Errorf("Bad request size: %d", size)
	}

	filter := func(response *ApplicationCommandData) bool {
		return len(response.Command.Data) > 1 && response.Command.Data[0] == parameter
	}

	// Issue request
	var response *ApplicationCommandData
	var err error
	if response, err = node.zwSendDataWaitForResponse(
		CommandClassConfiguration, []uint8{configurationGet, parameter},
		configurationReport, filter); err != nil {
		return nil, err
	}

	// Check response
	data := response.Command.Data
	if len(data) != 2+int(size) {
		return nil, fmt.Errorf("Unexpected data length: %d != %d, value might not exist",
			len(data), 2+size)
	}

	if data[1] != size {
		// Validate size
		return nil, fmt.Errorf("Bad size: 0x%02x != 0x%02x", data[1], size)
	}

	// Return data
	return data[2 : 2+size], nil
}

////////////////////////////////////////////////////////////////////////////////

// GetBool returns the boolean configuration value of the parameter
func (node *Configuration) GetBool(parameter uint8) (bool, error) {
	var value []uint8
	var err error

	if value, err = node.getValue(parameter, 1); err != nil {
		return false, err
	}
	return value[0] != 0, nil
}

// GetByte returns the boolean configuration value of the parameter
func (node *Configuration) GetByte(parameter uint8) (uint8, error) {
	var value []uint8
	var err error

	if value, err = node.getValue(parameter, 1); err != nil {
		return 0, err
	}
	return value[0], nil

}

// GetShort returns the boolean configuration value of the parameter
func (node *Configuration) GetShort(parameter uint8) (uint16, error) {
	var value []uint8
	var err error

	if value, err = node.getValue(parameter, 2); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint16(value), nil
}

// GetInt returns the boolean configuration value of the parameter
func (node *Configuration) GetInt(parameter uint8) (uint32, error) {
	var value []uint8
	var err error

	if value, err = node.getValue(parameter, 4); err != nil {
		return 0, err
	}
	return binary.BigEndian.Uint32(value), nil
}

////////////////////////////////////////////////////////////////////////////////

// SetBool sets the boolean value
func (node *Configuration) SetBool(parameter uint8, value bool) error {
	v := uint8(0)
	if value {
		v = 1
	}
	return node.zwSendDataRequest(CommandClassConfiguration,
		[]uint8{configurationSet, parameter, 1, v})
}

// SetByte sets the byte value
func (node *Configuration) SetByte(parameter uint8, value uint8) error {
	return node.zwSendDataRequest(CommandClassConfiguration,
		[]uint8{configurationSet, parameter, 1, value})
}

// SetShort sets the short value
func (node *Configuration) SetShort(parameter uint8, value uint16) error {
	return node.zwSendDataRequest(CommandClassConfiguration,
		[]uint8{configurationSet, parameter, 2, uint8((value >> 8) & (0xff)),
			uint8(value & 0xff)})
}

// SetInt sets the int value
func (node *Configuration) SetInt(parameter uint8, value uint32) error {
	return node.zwSendDataRequest(CommandClassConfiguration,
		[]uint8{configurationSet, parameter, 4, uint8((value >> 24) & (0xff)),
			uint8((value >> 16) & (0xff)), uint8((value >> 8) & (0xff)),
			uint8(value & 0xff)})
}
