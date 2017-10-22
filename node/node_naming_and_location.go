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
	namingSet      uint8 = 0x01
	namingGet            = 0x02
	namingReport         = 0x03
	locationSet          = 0x04
	locationGet          = 0x05
	locationReport       = 0x06
)

const (
	encodingASCII         uint8 = 0x00
	encodingExtendedASCII       = 0x01
	encodingUTF16               = 0x02
)

const (
	maxNameLength     uint8 = 16
	maxLocationLength       = 16
)

// NamingAndLocation information
type NamingAndLocation struct {
	*Node
}

// GetNamingAndLocation returns a NamingAndLocation or nil object
func (node *Node) GetNamingAndLocation() *NamingAndLocation {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassNodeNamingAndLocation) {
		return &NamingAndLocation{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

func extractString(bytes []uint8) (string, error) {
	// Check length
	if len(bytes) < 1 {
		return "", fmt.Errorf("Empty reply")
	}

	encoding := bytes[0]
	stringBytes := bytes[1:]

	switch encoding {
	case encodingASCII:
		fallthrough
	case encodingExtendedASCII:
		// FIXME: is extended ascii the same thing?
		return string(stringBytes), nil
	case encodingUTF16:
		return "", fmt.Errorf("UTF16 encoding is not supported")
	default:
		return "", fmt.Errorf("Unknown encoding type: 0x%02x", encoding)
	}
}

////////////////////////////////////////////////////////////////////////////////

// GetName of the node
func (node *NamingAndLocation) GetName() (string, error) {
	var response *ApplicationCommandData
	var err error

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassNodeNamingAndLocation, []uint8{namingGet}, namingReport,
		nil); err != nil {
		return "", err
	}

	return extractString(response.Command.Data)
}

// SetName of the node
func (node *NamingAndLocation) SetName(name string) error {
	stringBytes := []byte(name)

	if len(stringBytes) > int(maxNameLength) {
		return fmt.Errorf("Name is too long: max is %d bytes", maxNameLength)
	}

	data := make([]byte, 2+len(stringBytes))
	data[0] = namingSet
	data[1] = encodingASCII
	copy(data[2:], stringBytes)

	return node.zwSendDataRequest(CommandClassNodeNamingAndLocation, data)
}

////////////////////////////////////////////////////////////////////////////////

// GetLocation of the node
func (node *NamingAndLocation) GetLocation() (string, error) {
	var response *ApplicationCommandData
	var err error

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassNodeNamingAndLocation, []uint8{locationGet}, locationReport,
		nil); err != nil {
		return "", err
	}

	return extractString(response.Command.Data)
}

// SetLocation of the node
func (node *NamingAndLocation) SetLocation(location string) error {
	stringBytes := []byte(location)

	if len(stringBytes) > int(maxLocationLength) {
		return fmt.Errorf("Location is too long: max is %d bytes", maxLocationLength)
	}

	data := make([]byte, 2+len(stringBytes))
	data[0] = locationSet
	data[1] = encodingASCII
	copy(data[2:], stringBytes)

	return node.zwSendDataRequest(CommandClassNodeNamingAndLocation, data)
}
