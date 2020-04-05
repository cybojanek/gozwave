// Package node represents a ZWave node. All public methods are goroutine safe.
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
	"encoding/json"
	"errors"
	"fmt"
	"github.com/cybojanek/gozwave/controller"
	"github.com/cybojanek/gozwave/message"
	"log"
	"sync"
	"time"
)

var (
	// ErrNodeNotFound is returned in case of node not found
	ErrNodeNotFound = errors.New("Node not found")
	// DefaultTransmitOptions for ZW Send Data commands
	DefaultTransmitOptions = (message.TransmitOptionACK |
		message.TransmitOptionAutoRoute | message.TransmitOptionExplore)
)

const responseTimeout = (10 * time.Second)

// Node information
type Node struct {
	ID uint8

	CommandClasses        []uint8 // List of supported command classes
	ControlCommandClasses []uint8 // List of control command classes
	Listening             bool    // Is node actively listening
	DeviceClass           struct {
		Basic    uint8 // Basic Device Class
		Generic  uint8 // Generic Device Class
		Specific uint8 // Specific Device Class
	}
	Manufacturer struct {
		ID uint16 // Manufacturer ID
	}
	Product struct {
		ID   uint16 // Product ID
		Type uint16 // Product Type
	}

	network controller.Controller // Reference to parent network
	mutex   sync.RWMutex          // Node mutex

	keyCallbacks                map[uint16]map[chan *ApplicationCommandData]chan *ApplicationCommandData
	applicationCommandCallbacks map[chan *ApplicationCommandData]chan *ApplicationCommandData
	applicationUpdateCallbacks  map[chan *ApplicationUpdateData]chan *ApplicationUpdateData
}

// ApplicationCommandData information
type ApplicationCommandData struct {
	Status  uint8 // ??
	NodeID  uint8 // Source NodeID
	Command struct {
		ClassID uint8   // Command Class ID
		ID      uint8   // Command Class Subcommand ID
		Data    []uint8 // Command data
	}
}

// ApplicationUpdateData information
type ApplicationUpdateData struct {
	Status uint8   // One of message.ZWApplicationUpdateState
	NodeID uint8   // Source NodeID
	Data   []uint8 // Update data
}

type applicationCallbackFilter func(response *ApplicationCommandData) bool

// MakeNode makes a new node
func MakeNode(nodeID uint8, controller controller.Controller) *Node {
	return &Node{ID: nodeID, network: controller}
}

// commandClassIDsToMapKey returns the 16 bit key to use for callback maps
func commandClassIDsToMapKey(commandClassID uint8, commandID uint8) uint16 {
	return (uint16(commandClassID) << 8) | (uint16(commandID))
}

const NODE_CACHE_VERSION = "d55bb26e8f524ae2"

type nodeCache struct {
	Version               string
	NodeProtocolInfo      message.ZWGetNodeProtocolInfo
	ApplicationUpdateData ApplicationUpdateData

	Manufacturer struct {
		ID uint16
	}
	Product struct {
		ID   uint16
		Type uint16
	}
}

// Load the node information: Listening, DeviceClass, CommandClasses.
func (node *Node) Load(cacheBytes []byte) ([]byte, error) {
	// Try to unmarshal old cache.
	var oldCache *nodeCache
	if cacheBytes != nil {
		oldCache = &nodeCache{}
		if err := json.Unmarshal(cacheBytes, oldCache); err != nil {
			oldCache = nil
		} else if oldCache.Version != NODE_CACHE_VERSION {
			oldCache = nil
		}
	}

	// Create new cache object
	newCache := nodeCache{Version: NODE_CACHE_VERSION}

	// Acquire exclusive lock, since we'll be updating fields
	node.mutex.Lock()

	// Contact controller to get device description
	if oldCache == nil {
		nodeProtocolInfo, err := node.zWGetNodeProtocolInfo()
		if err != nil {
			node.mutex.Unlock()
			return nil, err
		}
		newCache.NodeProtocolInfo = *nodeProtocolInfo
	} else {
		newCache.NodeProtocolInfo = oldCache.NodeProtocolInfo
	}

	// Update fields
	nodeProtocolInfo := newCache.NodeProtocolInfo
	node.Listening = nodeProtocolInfo.Capabilities.Listening
	node.DeviceClass.Basic = nodeProtocolInfo.DeviceClass.Basic
	node.DeviceClass.Generic = nodeProtocolInfo.DeviceClass.Generic
	node.DeviceClass.Specific = nodeProtocolInfo.DeviceClass.Specific

	// If it's a listening device, we can issue more commands to it
	if node.Listening {
		node.mutex.Unlock()

		// Create channel for receiving node info.
		channel := make(chan *ApplicationUpdateData, 1)
		if oldCache == nil {
			// Add channel to application callbacks.
			node.AddApplicationUpdateCallbackChannel(channel)
			defer node.RemoveApplicationUpdateCallbackChannel(channel)

			node.mutex.Lock()
			if err := node.zWRequestNodeInfo(); err != nil {
				node.mutex.Unlock()
				return nil, err
			}
			node.mutex.Unlock()
		} else {
			// Send cache data to channel.
			channel <- &oldCache.ApplicationUpdateData
		}

		end := time.Now().Add(responseTimeout)
	outer:
		for {
			now := time.Now()
			timeLeft := end.Sub(now)

			if now.After(end) {
				return nil, fmt.Errorf("Timed out waiting for node info data")
			}

			select {
			case response := <-channel:
				if response.Status != message.ZWApplicationUpdateStateReceived {
					continue
				}
				newCache.ApplicationUpdateData = *response

				data := response.Data
				if len(data) < 3 {
					return nil, fmt.Errorf(
						"ZWApplicationUpdateStateReceived too short: %d < 3",
						len(response.Data))
				}

				// Lock again because we're updating
				node.mutex.Lock()

				// NOTE: zWGetNodeProtocolInfo also does device class, but does not do
				//       command classes
				// Update DeviceClass
				node.DeviceClass.Basic = data[0]
				node.DeviceClass.Generic = data[1]
				node.DeviceClass.Specific = data[2]

				// Update CommandClasses
				node.CommandClasses = []uint8{}
				node.ControlCommandClasses = []uint8{}

				// NOTE: CommandClasses before CommandClassMark are those supported by
				//       the Node, while the CommandClasses after CommandClassMark are
				//       those which the Node can control
				afterMark := false
				for _, x := range data[3:] {
					if !afterMark && x == CommandClassMark {
						afterMark = true
					} else if !afterMark {
						node.CommandClasses = append(node.CommandClasses, x)
					} else { // afterMark
						node.ControlCommandClasses = append(node.ControlCommandClasses, x)
					}
				}

				node.mutex.Unlock()

				break outer

			case <-time.After(timeLeft):
				return nil, fmt.Errorf("Timed out waiting for response")
			}
		}

		// Check if we can get manufacturer information
		if manuf := node.GetManufacturerSpecific(); manuf != nil {
			if oldCache == nil {
				manufacturerID, productType, productID, err := manuf.Get()
				if err != nil {
					return nil, err
				}

				newCache.Manufacturer.ID = manufacturerID
				newCache.Product.Type = productType
				newCache.Product.ID = productID
			} else {
				newCache.Manufacturer.ID = oldCache.Manufacturer.ID
				newCache.Product.Type = oldCache.Product.Type
				newCache.Product.ID = oldCache.Product.ID
			}

			node.mutex.Lock()
			node.Manufacturer.ID = newCache.Manufacturer.ID
			node.Product.Type = newCache.Product.Type
			node.Product.ID = newCache.Product.ID
			node.mutex.Unlock()
		}
	} else {
		// Can't fill anything in
		node.mutex.Unlock()
	}

	cacheBytes, err := json.Marshal(newCache)
	if err != nil {
		return nil, err
	}

	return cacheBytes, nil
}

// RefreshWithIDs using the manufacturer information and local database
func (node *Node) RefreshWithIDs(manufacturerID uint16, productID uint16, typeID uint16) error {
	// TODO: implement
	return nil
}

// ApplicationCommandHandler function
func (node *Node) ApplicationCommandHandler(command *message.ApplicationCommand) {
	log.Printf("DEBUG ApplicationCommandHandler: node: %d command: %+v", node.ID, command)

	node.mutex.Lock()
	defer node.mutex.Unlock()

	if len(command.Body) < 2 {
		log.Printf("ERROR ApplicationCommandHandler: command is too short: %d", len(command.Body))
		return
	}

	// Extract command information
	commandClassID := command.Body[0]
	commandID := command.Body[1]
	commandData := command.Body[2:len(command.Body)]

	// Compute lookup key
	key := commandClassIDsToMapKey(commandClassID, commandID)

	for i := 0; i < 2; i++ {
		var callbacks map[chan *ApplicationCommandData]chan *ApplicationCommandData
		var ok bool

		// Choose map depending on loop
		ok = false
		switch i {
		case 0:
			callbacks, ok = node.keyCallbacks[key]

		case 1:
			callbacks = node.applicationCommandCallbacks
			ok = callbacks != nil

		default:

		}

		// No map
		if !ok {
			continue
		}

		for _, channel := range callbacks {
			// Create copy for channel callback
			data := ApplicationCommandData{Status: command.Status, NodeID: command.NodeID}
			data.Command.ClassID = commandClassID
			data.Command.ID = commandID
			data.Command.Data = make([]uint8, len(commandData))
			copy(data.Command.Data, commandData)

			go func() {
				// Send to channel
				channel <- &data
			}()
		}
	}
}

// ApplicationUpdateHandler function
func (node *Node) ApplicationUpdateHandler(update *message.ZWApplicationUpdate) {
	log.Printf("DEBUG ApplicationUpdateHandler: node: %d update: %+v", node.ID, update)

	node.mutex.Lock()
	defer node.mutex.Unlock()

	// Send to callbacks
	for channel := range node.applicationUpdateCallbacks {
		data := ApplicationUpdateData{Status: update.Status, NodeID: node.ID}
		data.Data = make([]uint8, len(update.Body))
		copy(data.Data, update.Body)

		go func() {
			channel <- &data
		}()
	}
}

////////////////////////////////////////////////////////////////////////////////

// Check if node supports a command class.
// Assumption: caller holds node lock
func (node *Node) supportsCommandClass(commandClass uint8) bool {
	for _, x := range node.CommandClasses {
		if x == commandClass {
			return true
		}
	}
	return false
}

// getKeyedApplicationCommandCallbackChannel returns a channel in which the next application
// update result will be sent to
// Assumption: caller holds node lock
func (node *Node) getKeyedApplicationCommandCallbackChannel(commandClassID uint8, commandID uint8) chan *ApplicationCommandData {
	// make with 1 to not block
	channel := make(chan *ApplicationCommandData, 1)
	key := commandClassIDsToMapKey(commandClassID, commandID)

	// Make map if it does not exist
	if node.keyCallbacks == nil {
		node.keyCallbacks = make(
			map[uint16]map[chan *ApplicationCommandData]chan *ApplicationCommandData)
	}

	// Create channel map if it does not exist
	if node.keyCallbacks[key] == nil {
		node.keyCallbacks[key] = make(
			map[chan *ApplicationCommandData]chan *ApplicationCommandData)
	}

	node.keyCallbacks[key][channel] = channel
	return channel
}

// removeKeyedApplicationCallbackChannel removes the channel from future callbacks
// Assumption: caller holds node lock
func (node *Node) removeKeyedApplicationCallbackChannel(channel chan *ApplicationCommandData) {
	for key := range node.keyCallbacks {
		delete(node.keyCallbacks[key], channel)
	}

	close(channel)
}

// AddApplicationCommandCallbackChannel add the report callback channel
func (node *Node) AddApplicationCommandCallbackChannel(channel chan *ApplicationCommandData) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.applicationCommandCallbacks == nil {
		node.applicationCommandCallbacks = make(
			map[chan *ApplicationCommandData]chan *ApplicationCommandData)
	}

	node.applicationCommandCallbacks[channel] = channel
}

// RemoveApplicationCommandCallbackChannel add the report callback channel
func (node *Node) RemoveApplicationCommandCallbackChannel(channel chan *ApplicationCommandData) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.applicationCommandCallbacks == nil {
		node.applicationCommandCallbacks = make(
			map[chan *ApplicationCommandData]chan *ApplicationCommandData)
	}

	delete(node.applicationCommandCallbacks, channel)
}

// AddApplicationUpdateCallbackChannel add the report callback channel
func (node *Node) AddApplicationUpdateCallbackChannel(channel chan *ApplicationUpdateData) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.applicationUpdateCallbacks == nil {
		node.applicationUpdateCallbacks = make(
			map[chan *ApplicationUpdateData]chan *ApplicationUpdateData)
	}

	node.applicationUpdateCallbacks[channel] = channel
}

// RemoveApplicationUpdateCallbackChannel add the report callback channel
func (node *Node) RemoveApplicationUpdateCallbackChannel(channel chan *ApplicationUpdateData) {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.applicationUpdateCallbacks == nil {
		node.applicationUpdateCallbacks = make(
			map[chan *ApplicationUpdateData]chan *ApplicationUpdateData)
	}

	delete(node.applicationUpdateCallbacks, channel)
}

////////////////////////////////////////////////////////////////////////////////

// zWGetNodeProtocolInfo gets the message.ZWGetNodeProtocolInfo information
// for a requested node. Returns ErrNodeNotFound if the request node could not
// be found by the controller.
func (node *Node) zWGetNodeProtocolInfo() (*message.ZWGetNodeProtocolInfo, error) {
	requestPacket, err := message.ZWGetNodeProtocolInfoRequest(node.ID)
	if err != nil {
		return nil, err
	}
	responsePacket, err := node.network.DoRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	responseMessage, err := message.ZWGetNodeProtocolInfoResponse(responsePacket)
	if err != nil {
		return responseMessage, err
	}

	// Check node exists
	if responseMessage.DeviceClass.Generic == 0 {
		return nil, ErrNodeNotFound
	}
	return responseMessage, nil
}

// zWRequestNodeInfo gets the message.ZWRequestNodeInfo information for a
// requested node.
func (node *Node) zWRequestNodeInfo() error {
	requestPacket, err := message.ZWRequestNodeInfoRequest(node.ID)
	if err != nil {
		return err
	}
	responsePacket, err := node.network.DoRequest(requestPacket)
	if err != nil {
		return err
	}
	responseMessage, err := message.ZWRequestNodeInfoResponse(responsePacket)
	if err != nil {
		return err
	}

	if responseMessage.Status != 1 {
		return fmt.Errorf("Bad ZWRequestNodeInfo reply: %d", responseMessage.Status)
	}

	return nil
}

// zWSendData sends the ZWSendData request to a given node
func (node *Node) zWSendData(commandClass uint8, payload []uint8) error {
	requestPacket, err := message.ZWSendDataRequest(node.ID, commandClass, payload,
		DefaultTransmitOptions, 0x00)
	if err != nil {
		return err
	}
	responsePacket, err := node.network.DoRequest(requestPacket)
	if err != nil {
		return err
	}
	responseMessage, err := message.ZWSendDataResponse(responsePacket)
	if err != nil {
		return err
	}

	if responseMessage.Status != message.TransmitCompleteOK {
		return fmt.Errorf("ZWSendData failed to contact node: 0x%02x",
			responseMessage.Status)
	}

	return nil
}

// zwSendDataRequest sends the ZWSendData request
func (node *Node) zwSendDataRequest(commandClass uint8, data []uint8) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	if err := node.zWSendData(commandClass, data); err != nil {
		return err
	}
	return nil
}

// zwSendDataWaitForResponse sends the ZWSendData request, and awaits the
// ApplicationCommandUpdate for the specified command, and can additionally wait
// until optional filter returns true
func (node *Node) zwSendDataWaitForResponse(commandClass uint8,
	data []uint8, command uint8, filter applicationCallbackFilter) (*ApplicationCommandData, error) {
	node.mutex.Lock()

	channel := node.getKeyedApplicationCommandCallbackChannel(commandClass, command)
	defer func() {
		node.mutex.Lock()
		node.removeKeyedApplicationCallbackChannel(channel)
		node.mutex.Unlock()
	}()

	if err := node.zWSendData(commandClass, data); err != nil {
		node.mutex.Unlock()
		return nil, err
	}
	node.mutex.Unlock()

	end := time.Now().Add(responseTimeout)
	for {
		now := time.Now()
		timeLeft := end.Sub(now)

		if now.After(end) {
			return nil, fmt.Errorf("Timed out waiting for response")
		}

		select {
		case response := <-channel:
			if filter != nil && !filter(response) {
				continue
			}
			return response, nil

		case <-time.After(timeLeft):
			return nil, fmt.Errorf("Timed out waiting for response")
		}
	}
}
