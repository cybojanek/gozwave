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
	versionGet                uint8 = 0x11
	versionReport                   = 0x12
	versionCommandClassGet          = 0x13
	versionCommandClassReport       = 0x14
)

// Version information
type Version struct {
	*Node
}

// GetVersion returns a Version or nil object
func (node *Node) GetVersion() *Version {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassVersion) {
		return &Version{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Get the node version information. Return value is for library, protocol,
// application
func (node *Version) Get() (library uint8, protocol uint16, application uint16, err error) {
	var response *ApplicationCommandData

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassVersion, []uint8{versionGet}, versionReport, nil); err != nil {
		return
	}

	data := response.Command.Data
	if len(data) != 5 {
		err = fmt.Errorf("Unexpected data length: %d != %d", len(data), 5)
		return
	}

	library = data[0]
	protocol = binary.BigEndian.Uint16(data[1:3])
	application = binary.BigEndian.Uint16(data[3:5])

	return
}

////////////////////////////////////////////////////////////////////////////////

// GetCommandClass version for a given command class
func (node *Version) GetCommandClass(commandClass uint8) (uint8, error) {
	// Fail early to avoid long timeout errors
	if !node.supportsCommandClass(commandClass) {
		return 0, fmt.Errorf("Node does not support command class")
	}

	filter := func(response *ApplicationCommandData) bool {
		return len(response.Command.Data) > 0 && response.Command.Data[0] == commandClass
	}

	var response *ApplicationCommandData
	var err error

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassVersion, []uint8{versionCommandClassGet, commandClass},
		versionCommandClassReport, filter); err != nil {
		return 0, err
	}

	data := response.Command.Data
	if len(data) != 2 {
		return 0, fmt.Errorf("Unexpected data length: %d != 2", len(data))
	}

	return data[1], nil
}
