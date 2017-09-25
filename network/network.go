// Package network manages a network of ZWave nodes through a serial controller.
// All public methods are goroutine safe.
package network

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
	"github.com/cybojanek/gozwave/message"
	"github.com/cybojanek/gozwave/node"
	"github.com/cybojanek/gozwave/packet"
	"log"
	"sync"
)

// Network instance
type Network struct {
	DevicePath   string // Path to ZWave controller
	DebugLogging bool   // Enable debug logging

	mutex                  sync.RWMutex                 // API mutex
	serialController       *controller.SerialController // Controller
	callbackChannel        chan *packet.Packet          // Channel for receiving async controller packets
	stopCallbackHandler    chan int                     // Exit signal channel for callbackHandler
	stoppedCallbackHandler chan int                     // Exit confirmation channel for callbackHandler
	nodexMutex             sync.RWMutex                 // Nodes mutex
	nodes                  map[uint8]*node.Node         // Nodes
	supportedMessageTypes  []uint8                      // Supported message types
}

////////////////////////////////////////////////////////////////////////////////

// isOpen checks if the api is open, must we called with api lock
func (network *Network) isOpen() bool {
	return network.serialController != nil
}

// Open network. goroutine safe.
func (network *Network) Open() error {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	if network.isOpen() {
		return nil
	}

	// Open controller
	serialController := controller.SerialController{DevicePath: network.DevicePath,
		DebugLogging: network.DebugLogging}
	if err := serialController.Open(); err != nil {
		return err
	}

	// Set up callback packet callback handler
	if network.callbackChannel == nil {
		network.callbackChannel = make(chan *packet.Packet, 1)
		network.stopCallbackHandler = make(chan int)
		network.stoppedCallbackHandler = make(chan int)

		serialController.SetCallbackChannel(network.callbackChannel)

		network.nodes = make(map[uint8]*node.Node)
	}

	// Start callback handler
	go network.callbackHandler()

	network.serialController = &serialController
	return nil
}

// Close network. goroutine safe.
func (network *Network) Close() error {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	if !network.isOpen() {
		return nil
	}

	// Close controller
	err := network.serialController.Close()

	// Stop callback handler
	network.stopCallbackHandler <- 0
	<-network.stoppedCallbackHandler

	// Purge unrouted messages
loop:
	for {
		select {
		case packet := <-network.callbackChannel:
			log.Printf("INFO Dropping Close packet: %v", packet)

		default:
			break loop
		}
	}

	network.serialController = nil
	return err
}

////////////////////////////////////////////////////////////////////////////////

// DoRequest sends a request and awaits a response
func (network *Network) DoRequest(request *packet.Packet) (*packet.Packet, error) {
	network.mutex.RLock()
	defer network.mutex.RUnlock()

	if !network.isOpen() {
		return nil, errors.New("API is not open")
	}

	// Check MessageType is supported
	found := false
	if network.supportedMessageTypes != nil {
		for _, x := range network.supportedMessageTypes {
			if x == request.MessageType {
				found = true
				break
			}
		}
	} else {
		// Before initialization assume it is supported
		found = true
	}

	if !found {
		return nil, fmt.Errorf("MessageType 0x%02x not supported by controller",
			request.MessageType)
	}

	return network.serialController.DoRequest(request)
}

////////////////////////////////////////////////////////////////////////////////

// Initialize serial controller and node list. goroutine safe.
func (network *Network) Initialize() error {
	network.mutex.Lock()
	defer network.mutex.Unlock()

	var err error

	// NOTE: only use internal functions to prevent mutex deadlock

	// SerialAPIGetCapabilities
	capabilities, err := network.initialSerialAPIGetCapabilities()
	if err != nil {
		return err
	}
	// Save supported message types
	network.supportedMessageTypes = capabilities.MessageTypes

	// GetVersion
	version, err := network.initialGetVersion()
	if err != nil {
		return err
	}

	// GetMemoryID
	memoryID, err := network.initialGetMemoryID()
	if err != nil {
		return err
	}
	// Check nodeID of controller - should be 0x01
	if memoryID.NodeID != 0x1 {
		return fmt.Errorf("Expected Controller node 0x01 not: 0x%02x", memoryID.NodeID)
	}

	// SerialAPIGetInitData
	initData, err := network.initialSerialAPIGetInitData()
	if err != nil {
		return err
	}

	if network.DebugLogging {
		log.Printf("DEBUG GetVersion: %+v", version)
		log.Printf("DEBUG GetMemoryID: %+v", memoryID)
		log.Printf("DEBUG SerialAPIGetCapabilities: %+v", capabilities)
		log.Printf("DEBUG SerialAPIGetInitData: %+v", initData)
	}

	// Add all known nodes
	for _, id := range initData.Nodes {
		// Don't add controller
		if id == memoryID.NodeID {
			continue
		}
		n, ok := network.nodes[id]
		if !ok {
			n = node.MakeNode(id, network)
			network.nodes[id] = n
		}
	}

	// TODO: remove dead nodes...

	return nil
}

// getVersion gets the message.GetVersion information
// Assumption: called only from Initialize
func (network *Network) initialGetVersion() (*message.GetVersion, error) {
	requestPacket := message.GetVersionRequest()
	responsePacket, err := network.serialController.DoRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.GetVersionResponse(responsePacket)
}

// getMemoryID gets thes message.MemoryGetID information
// Assumption: called only from Initialize
func (network *Network) initialGetMemoryID() (*message.MemoryGetID, error) {
	requestPacket := message.MemoryGetIDRequest()
	responsePacket, err := network.serialController.DoRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.MemoryGetIDResponse(responsePacket)
}

// serialAPIGetCapabilities gets thes message.SerialAPIGetCapabilities information
// Assumption: called only from Initialize
func (network *Network) initialSerialAPIGetCapabilities() (*message.SerialAPIGetCapabilities, error) {
	requestPacket := message.SerialAPIGetCapabilitiesRequest()
	responsePacket, err := network.serialController.DoRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.SerialAPIGetCapabilitiesResponse(responsePacket)
}

// serialAPIGetInitData gets the message.SerialAPIGetInitData information
// Assumption: called only from Initialize
func (network *Network) initialSerialAPIGetInitData() (*message.SerialAPIGetInitData, error) {
	requestPacket := message.SerialAPIGetInitDataRequest()
	responsePacket, err := network.serialController.DoRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.SerialAPIGetInitDataResponse(responsePacket)
}

// zWGetControllerCapabilities gets the message.ZWGetControllerCapabilities
// information
// Assumption: called only from Initialize
func (network *Network) initialZWGetControllerCapabilities() (*message.ZWGetControllerCapabilities, error) {
	requestPacket := message.ZWGetControllerCapabilitiesRequest()
	responsePacket, err := network.serialController.DoRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.ZWGetControllerCapabilitiesResponse(responsePacket)
}

////////////////////////////////////////////////////////////////////////////////

// GetNode returns the node or nil if doesn't exist. goroutine safe.
func (network *Network) GetNode(nodeID uint8) *node.Node {
	network.mutex.RLock()
	defer network.mutex.RUnlock()

	node, ok := network.nodes[nodeID]
	if ok {
		return node
	}
	return nil
}

// GetNodes returns a list copy of all the nodes. goroutine safe.
func (network *Network) GetNodes() []*node.Node {
	network.mutex.RLock()
	defer network.mutex.RUnlock()

	nodeList := make([]*node.Node, len(network.nodes))

	i := 0
	for _, v := range network.nodes {
		nodeList[i] = v
		i++
	}

	return nodeList
}

////////////////////////////////////////////////////////////////////////////////

// Callback for asynchronous messages
func (network *Network) callbackHandler() {
	for {
		select {

		case packet := <-network.callbackChannel:
			if network.DebugLogging {
				log.Printf("DEBUG callbackHandler received packet: %s", packet)
			}

			// Route based on Message Type
			switch packet.MessageType {

			case message.MessageTypeApplicationCommand:
				if response, err := message.ApplicationCommandResponse(packet); err != nil {
					log.Printf("ERROR callbackHandler decoding ApplicationCommand: %v", err)
				} else if node := network.GetNode(response.NodeID); node == nil {
					log.Printf("INFO callbackHandler ApplicationCommand no node: %d for %+v",
						response.NodeID, response)
				} else {
					go func() {
						node.ApplicationCommandHandler(response)
					}()
				}

			case message.MessageTypeZWApplicationUpdate:
				if response, err := message.ZWApplicationUpdateResponse(packet); err != nil {
					log.Printf("ERROR callbackHandler decoding ZWApplicationUpdate: %v", err)
				} else if response.Status == message.ZWApplicationUpdateStateReceived {
					if node := network.GetNode(response.NodeID); node == nil {
						log.Printf("INFO callbackHandler ZWApplicationUpdate no node: %d for %+v",
							response.NodeID, response)
					} else {
						go func() {
							node.ApplicationUpdateHandler(response)
						}()
					}
				} else {
					log.Printf("INFO callbackHandler ZWApplicationUpdate non State Received: %d for %+v",
						response.NodeID, response)
				}

			default:
				log.Printf("INFO callbackHandler unhandled MessageType: 0x%02x",
					packet.MessageType)
			}

		case <-network.stopCallbackHandler:
			network.stoppedCallbackHandler <- 0
			return
		}
	}
}
