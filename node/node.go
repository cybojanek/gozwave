// Package node represents a ZWave node attached to a controller
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
	"errors"
	"fmt"
	"github.com/cybojanek/gozwave/controller"
	"github.com/cybojanek/gozwave/device"
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

	commandClasses        []uint8 // List of supported command classes
	controlCommandClasses []uint8 // List of control command classes
	listening             bool    // Is node actively listening
	deviceClass           struct {
		basic    uint8 // Basic Device Class
		generic  uint8 // Generic Device Class
		specific uint8 // Specific Device Class
	}
	manufacturer struct {
		id uint16 // Manufacturer ID
	}
	product struct {
		id  uint16 // Product ID
		typ uint16 // Product Type
	}

	api   controller.RequestProcessor // Reference to parent api
	mutex sync.RWMutex                // Node mutex

	requestNodeInfoComplete chan int // Temporary channel used during refresh

	oneShotCallbacks map[uint16]map[chan *ApplicationCommandData]chan *ApplicationCommandData // One shot ZWSenData callbacks
	longCallbacks    map[uint16]map[chan *ApplicationCommandData]chan *ApplicationCommandData // Long ZWSendData callbacks
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

// MakeNode makes a new node
func MakeNode(nodeID uint8, controller controller.RequestProcessor) *Node {
	return &Node{ID: nodeID, api: controller}
}

// commandClassIDsToMapKey returns the 16 bit key to use for callback maps
func commandClassIDsToMapKey(commandClassID uint8, commandID uint8) uint16 {
	return (uint16(commandClassID) << 8) | (uint16(commandID))
}

// Refresh the node information: Listening, DeviceClass, CommandClasses.
func (node *Node) Refresh() error {
	// Acquire exclusive lock, since we'll be updating fields
	node.mutex.Lock()

	// Contact controller to get device description
	nodeProtocolInfo, err := node.zWGetNodeProtocolInfo()
	if err != nil {
		node.mutex.Unlock()
		return err
	}

	// Update fields
	node.listening = nodeProtocolInfo.Capabilities.Listening
	node.deviceClass.basic = nodeProtocolInfo.DeviceClass.Basic
	node.deviceClass.generic = nodeProtocolInfo.DeviceClass.Generic
	node.deviceClass.specific = nodeProtocolInfo.DeviceClass.Specific

	// If it's a listening device, we can issue more commands to it
	if node.listening {
		// Fill supported command classes
		channelA := make(chan int, 1)
		// FIXME: this is not goroutine safe and multiple Refresh could stall...
		node.requestNodeInfoComplete = channelA
		if err := node.zWRequestNodeInfo(); err != nil {
			node.mutex.Unlock()
			return err
		}

		// Unlock to allow for update
		node.mutex.Unlock()

		select {
		case <-channelA:
			// Finished getting command classes

		case <-time.After(responseTimeout):
			return fmt.Errorf("Timed out waiting for command classes")
		}

		// Check if we can get manufacturer information
		if manuf := node.GetManufacturerSpecific(); manuf != nil {
			manufacturerID, productType, productID, err := manuf.Report()
			if err != nil {
				return err
			}
			node.mutex.Lock()
			node.manufacturer.id = manufacturerID
			node.product.typ = productType
			node.product.id = productID
			node.mutex.Unlock()
		}
	} else {
		// Fill supported command classes based on device class
		// TODO:
		node.mutex.Unlock()
	}

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

	// Loop over oneShotCallbacks and longCallbacks
	for i := 0; i < 2; i++ {
		var callbacks map[chan *ApplicationCommandData]chan *ApplicationCommandData
		var ok bool

		// Choose map depending on loop
		if i == 0 {
			callbacks, ok = node.oneShotCallbacks[key]
		} else {
			callbacks, ok = node.longCallbacks[key]
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

			// If its a one shot callback, then delete the channel
			if i == 0 {
				delete(callbacks, channel)
			}
		}
	}
}

// ApplicationUpdateHandler function
func (node *Node) ApplicationUpdateHandler(update *message.ZWApplicationUpdate) {
	log.Printf("DEBUG ApplicationUpdateHandler: node: %d update: %+v", node.ID, update)

	node.mutex.Lock()
	defer node.mutex.Unlock()

	switch update.Status {
	case message.ZWApplicationUpdateStateReceived:
		if len(update.Body) < 4 {
			log.Printf("ERROR body message.ZWApplicationUpdateStateReceived too short: %d",
				len(update.Body))
			break
		}
		// NOTE: zWGetNodeProtocolInfo also does device class, but does not do
		// 		 command classes
		// Update DeviceClass
		node.deviceClass.basic = update.Body[0]
		node.deviceClass.generic = update.Body[1]
		node.deviceClass.specific = update.Body[2]

		// Update CommandClasses
		node.commandClasses = []uint8{}
		node.controlCommandClasses = []uint8{}

		// NOTE: CommandClasses before CommandClassMark are those supported by
		//       the Node, while the CommandClasses after CommandClassMark are
		//       those which the Node can control
		afterMark := false
		for _, x := range update.Body[3:len(update.Body)] {
			if !afterMark && x == device.CommandClassMark {
				afterMark = true
			} else if !afterMark {
				node.commandClasses = append(node.commandClasses, x)
			} else { // afterMark
				node.controlCommandClasses = append(node.controlCommandClasses, x)
			}
		}

		// Notify requestNodeInfoComplete
		if node.requestNodeInfoComplete != nil {
			node.requestNodeInfoComplete <- 1
			node.requestNodeInfoComplete = nil
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

// Check if node supports a command class.
// Assumption: caller holds node lock
func (node *Node) supportsCommandClass(commandClass uint8) bool {
	for _, x := range node.commandClasses {
		if x == commandClass {
			return true
		}
	}
	return false
}

// getOneShotCallbackChannel returns a channel in which the next application
// update result will be sent to
// Assumption: caller holds node lock
func (node *Node) getOneShotCallbackChannel(commandClassID uint8, commandID uint8) chan *ApplicationCommandData {
	// make with 1 to not block
	channel := make(chan *ApplicationCommandData, 1)
	key := commandClassIDsToMapKey(commandClassID, commandID)

	// Make map if it does not exist
	if node.oneShotCallbacks == nil {
		node.oneShotCallbacks = make(
			map[uint16]map[chan *ApplicationCommandData]chan *ApplicationCommandData)
	}

	// Create channel map if it does not exist
	if node.oneShotCallbacks[key] == nil {
		node.oneShotCallbacks[key] = make(
			map[chan *ApplicationCommandData]chan *ApplicationCommandData)
	}

	node.oneShotCallbacks[key][channel] = channel
	return channel
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
	responsePacket, err := node.api.BlockingRequest(requestPacket)
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
	responsePacket, err := node.api.BlockingRequest(requestPacket)
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
	responsePacket, err := node.api.BlockingRequest(requestPacket)
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

func (node *Node) zwSendDataRequest(commandClass uint8, data []uint8) error {
	node.mutex.Lock()
	defer node.mutex.Unlock()
	if err := node.zWSendData(commandClass, data); err != nil {
		return err
	}
	return nil
}

func (node *Node) zwSendDataWaitForResponse(commandClass uint8, data []uint8, command uint8) (*ApplicationCommandData, error) {
	node.mutex.Lock()

	channel := node.getOneShotCallbackChannel(commandClass, command)

	if err := node.zWSendData(commandClass, data); err != nil {
		node.mutex.Unlock()
		return nil, err
	}
	node.mutex.Unlock()

	select {
	case response := <-channel:
		return response, nil

	case <-time.After(responseTimeout):
		return nil, fmt.Errorf("Timed out waiting for response")
	}
}
