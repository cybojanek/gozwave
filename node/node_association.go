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
	associationCommandSet             uint8 = 0x01
	associationCommandGet                   = 0x02
	associationCommandReport                = 0x03
	associationCommandRemove                = 0x04
	associationCommandGroupingsGet          = 0x05
	associationCommandGroupingsReport       = 0x06
)

// Association information
type Association struct {
	*Node
}

// GetAssociation returns a Association or nil object
func (node *Node) GetAssociation() *Association {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassAssociation) {
		return &Association{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Add adds the nodes to the association group
func (node *Association) Add(association uint8, nodes []uint8) error {
	data := make([]uint8, 2+len(nodes))
	data[0] = associationCommandSet
	data[1] = association
	for i, b := range nodes {
		data[i+2] = b
	}
	return node.zwSendDataRequest(CommandClassAssociation, data)
}

// Remove removes the nodes from the association group
func (node *Association) Remove(association uint8, nodes []uint8) error {
	data := make([]uint8, 2+len(nodes))
	data[0] = associationCommandRemove
	data[1] = association
	for i, b := range nodes {
		data[i+2] = b
	}
	return node.zwSendDataRequest(CommandClassAssociation, data)
}

// RemoveAllFromAssociation removes all nodes from the association
func (node *Association) RemoveAllFromAssociation(association uint8) error {
	return node.Remove(association, []uint8{})
}

// RemoveFromAllAssociations removes nodes from all associations. Only V2.
func (node *Association) RemoveFromAllAssociations(nodes []uint8) error {
	return node.Remove(0, nodes)
}

// RemoveAll removes all nodes from all associations. Only V2.
func (node *Association) RemoveAll() error {
	return node.Remove(0, []uint8{})
}

////////////////////////////////////////////////////////////////////////////////

// Get gets the nodes in the association group
func (node *Association) Get(association uint8) (maxNodes uint8, nodes []uint8, err error) {
	var response *ApplicationCommandData

	filter := func(response *ApplicationCommandData) bool {
		return len(response.Command.Data) > 0 && response.Command.Data[0] == association
	}

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassAssociation, []uint8{associationCommandGet},
		associationCommandReport, filter); err != nil {
		return
	}

	data := response.Command.Data
	if len(data) < 3 {
		err = fmt.Errorf("Response is too short %d < 3", len(data))
		return
	}

	maxNodes = data[1]
	// TODO: add support for reports to follow data[2]
	nodes = make([]uint8, len(data)-3)

	for i, b := range data[3:] {
		nodes[i] = b
	}

	return
}

////////////////////////////////////////////////////////////////////////////////

// GetSupported gets the number of supported association groups
func (node *Association) GetSupported(association uint8) (uint8, error) {
	var response *ApplicationCommandData
	var err error

	filter := func(response *ApplicationCommandData) bool {
		return len(response.Command.Data) > 0 && response.Command.Data[0] == association
	}

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassAssociation, []uint8{associationCommandGroupingsGet},
		associationCommandGroupingsReport, filter); err != nil {
		return 0, err
	}

	data := response.Command.Data
	if len(data) != 1 {
		return 0, fmt.Errorf("Response has bad length %d != 1", len(data))
	}

	return data[0], nil
}
