package api

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
	"github.com/cybojanek/gozwave/packet"
	"log"
	"sync"
)

var (
	// ErrNodeNotFound is returned in case of node not found
	ErrNodeNotFound = errors.New("Node not found")
	// DefaultTransmitOptions for ZW Send Data commands
	DefaultTransmitOptions = (message.TransmitOptionACK |
		message.TransmitOptionAutoRoute | message.TransmitOptionExplore)
)

// ZWAPI instance
type ZWAPI struct {
	DevicePath   string // Path to ZWave controller
	DebugLogging bool   // Enable debug logging

	mutex                 sync.RWMutex           // API mutex
	homeID                uint32                 // ID of ZWave network
	supportedMessageTypes []uint8                // Supported message types
	nodes                 map[uint8]*Node        // Nodes
	nodexMutex            sync.RWMutex           // Nodes mutex
	con                   *controller.Controller // Cntroller
	defaultChannel        chan *packet.Packet    // Channel for receiving async controller packets
	stopDefaultHandler    chan int               // Exit signal channel for defaultHandler
	stoppedDefaultHandler chan int               // Exit confirmation channel for defaultHandler
}

// Open api. goroutine safe.
func (api *ZWAPI) Open() error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	if api.isOpen() {
		return nil
	}

	// Open controller
	con := controller.Controller{DevicePath: api.DevicePath}
	con.DebugLogging = api.DebugLogging
	if err := con.Open(); err != nil {
		return err
	}

	// Set up default packet callback handler
	if api.defaultChannel == nil {
		api.defaultChannel = make(chan *packet.Packet, 1)
		api.stopDefaultHandler = make(chan int)
		api.stoppedDefaultHandler = make(chan int)

		con.SetCallbackChannel(&api.defaultChannel)

		api.nodes = make(map[uint8]*Node)
	}

	// Start callback handler
	go api.defaultHandler()

	api.con = &con
	return nil
}

// Close api. goroutine safe.
func (api *ZWAPI) Close() error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	if !api.isOpen() {
		return nil
	}

	// Close controller
	err := api.con.Close()

	// Stop default handler
	api.stopDefaultHandler <- 0
	<-api.stoppedDefaultHandler

	// Purge unrouted messages
loop:
	for {
		select {
		case packet := <-api.defaultChannel:
			log.Printf("INFO Dropping Close packet: %v", packet)

		default:
			break loop
		}
	}

	api.con = nil
	return err
}

// Initialize serial controller and node list. goroutine safe.
func (api *ZWAPI) Initialize() error {
	api.mutex.Lock()
	defer api.mutex.Unlock()

	var err error

	// NOTE: only use internal functions to prevent mutex deadlock

	// SerialAPIGetCapabilities
	capabilities, err := api.internalSerialAPIGetCapabilities()
	if err != nil {
		return err
	}
	// Save supported message types
	api.supportedMessageTypes = capabilities.MessageTypes

	// GetVersion
	version, err := api.internalGetVersion()
	if err != nil {
		return err
	}

	// GetMemoryID
	memoryID, err := api.internalGetMemoryID()
	if err != nil {
		return err
	}
	// Check nodeID of controller - should be 0x01
	if memoryID.NodeID != 0x1 {
		return fmt.Errorf("Expected Controller node 0x01 not: 0x%02x", memoryID.NodeID)
	}
	// Save HomeID
	api.homeID = memoryID.HomeID

	// SerialAPIGetInitData
	initData, err := api.internalSerialAPIGetInitData()
	if err != nil {
		return err
	}

	if api.DebugLogging {
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
		node, ok := api.nodes[id]
		if !ok {
			node = &Node{ID: id, api: api}
			api.nodes[id] = node
		}
	}

	// TODO: remove dead nodes...

	return nil
}

// GetNodes returns a list copy of all the nodes. goroutine safe.
func (api *ZWAPI) GetNodes() []*Node {
	api.mutex.RLock()
	defer api.mutex.RUnlock()

	nodeList := make([]*Node, len(api.nodes))

	i := 0
	for _, v := range api.nodes {
		nodeList[i] = v
		i++
	}

	return nodeList
}

// GetNode returns the node or nil if doesn't exist. goroutine safe.
func (api *ZWAPI) GetNode(nodeID uint8) *Node {
	api.mutex.RLock()
	defer api.mutex.RUnlock()

	node, ok := api.nodes[nodeID]
	if ok {
		return node
	}
	return nil
}

// Callback for asynchronous messages
func (api *ZWAPI) defaultHandler() {
	for {
		select {

		case packet := <-api.defaultChannel:
			if api.DebugLogging {
				log.Printf("DEBUG defaultHandler received packet: %s", packet)
			}

			// Route based on Message Type
			switch packet.MessageType {

			case message.MessageTypeApplicationCommandHandler:
				response, err := message.ApplicationCommandHandlerResponse(packet)
				if err != nil {
					log.Printf("ERROR defaultHandler decoding ApplicationCommandHandler: %v", err)
				} else {
					node := api.GetNode(response.NodeID)
					if node == nil {
						log.Printf("INFO defaultHandler ApplicationCommandHandler no node: %d for %+v",
							response.NodeID, response)
					} else {
						go func() {
							node.applicationCommandHandler(response)
						}()
					}
				}

			case message.MessageTypeZWApplicationUpdate:
				response, err := message.ZWApplicationUpdateResponse(packet)
				if err != nil {
					log.Printf("ERROR defaultHandler decoding ZWApplicationUpdate: %v", err)
				} else if response.Status == message.ZWApplicationUpdateStateReceived {
					node := api.GetNode(response.NodeID)
					if node == nil {
						log.Printf("INFO defaultHandler ZWApplicationUpdate no node: %d for %+v",
							response.NodeID, response)
					} else {
						go func() {
							node.applicationUpdate(response)
						}()
					}
				} else {
					log.Printf("INFO defaultHandler ZWApplicationUpdate non State Received: %d for %+v",
						response.NodeID, response)
				}

			default:
				log.Printf("INFO defaultHandler unhandled MessageType: 0x%02x",
					packet.MessageType)
			}

		case <-api.stopDefaultHandler:
			api.stoppedDefaultHandler <- 0
			return
		}
	}
}

// isOpen checks if the api is open, should we called with api lock
func (api *ZWAPI) isOpen() bool {
	return api.con != nil
}

// Sending a blocking request and wait for a reply
func (api *ZWAPI) blockingRequest(request *packet.Packet) (*packet.Packet, error) {
	api.mutex.RLock()
	defer api.mutex.RUnlock()

	if !api.isOpen() {
		return nil, errors.New("API is not open")
	}

	// Check MessageType is supported
	found := false
	if api.supportedMessageTypes != nil {
		for _, x := range api.supportedMessageTypes {
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

	return api.con.BlockingRequest(request)
}

// getVersion gets the message.GetVersion information
func (api *ZWAPI) internalGetVersion() (*message.GetVersion, error) {
	requestPacket := message.GetVersionRequest()
	responsePacket, err := api.con.BlockingRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.GetVersionResponse(responsePacket)
}

// getMemoryID gets thes message.MemoryGetID information
func (api *ZWAPI) internalGetMemoryID() (*message.MemoryGetID, error) {
	requestPacket := message.MemoryGetIDRequest()
	responsePacket, err := api.con.BlockingRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.MemoryGetIDResponse(responsePacket)
}

// serialAPIGetCapabilities gets thes message.SerialAPIGetCapabilities information
func (api *ZWAPI) internalSerialAPIGetCapabilities() (*message.SerialAPIGetCapabilities, error) {
	requestPacket := message.SerialAPIGetCapabilitiesRequest()
	responsePacket, err := api.con.BlockingRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.SerialAPIGetCapabilitiesResponse(responsePacket)
}

// serialAPIGetInitData gets the message.SerialAPIGetInitData information
func (api *ZWAPI) internalSerialAPIGetInitData() (*message.SerialAPIGetInitData, error) {
	requestPacket := message.SerialAPIGetInitDataRequest()
	responsePacket, err := api.con.BlockingRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.SerialAPIGetInitDataResponse(responsePacket)
}

// zWGetControllerCapabilities gets the message.ZWGetControllerCapabilities
// information
func (api *ZWAPI) internalZWGetControllerCapabilities() (*message.ZWGetControllerCapabilities, error) {
	requestPacket := message.ZWGetControllerCapabilitiesRequest()
	responsePacket, err := api.con.BlockingRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.ZWGetControllerCapabilitiesResponse(responsePacket)
}

// zWGetNodeProtocolInfo gets the message.ZWGetNodeProtocolInfo information
// for a requested node. Returns ErrNodeNotFound if the request node could not
// be found by the controller.
func (api *ZWAPI) zWGetNodeProtocolInfo(nodeID uint8) (*message.ZWGetNodeProtocolInfo, error) {
	requestPacket, err := message.ZWGetNodeProtocolInfoRequest(nodeID)
	if err != nil {
		return nil, err
	}
	responsePacket, err := api.blockingRequest(requestPacket)
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

// zWSendData sends the ZWSendData request to a given node
func (api *ZWAPI) zWSendData(nodeID uint8, commandClass uint8, payload []uint8) error {
	requestPacket, err := message.ZWSendDataRequest(nodeID, commandClass, payload,
		DefaultTransmitOptions, 0x00)
	if err != nil {
		return err
	}
	responsePacket, err := api.blockingRequest(requestPacket)
	if err != nil {
		return err
	}
	responseMessage, err := message.ZWSendDataResponse(responsePacket)
	if err != nil {
		return err
	}

	if responseMessage.Status != message.TransmitCompleteOK {
		return fmt.Errorf("Bad ZWSendData reply: 0x%02x", responseMessage.Status)
	}

	if api.DebugLogging {
		log.Printf("DEBUG ZWSendData: %+v", responseMessage)
	}
	return nil
}

// zWRequestNodeInfo gets the message.ZWRequestNodeInfo information for a
// requested node. Returns ErrNodeNotFound if the request node could not be
// found by the controller.
func (api *ZWAPI) zWRequestNodeInfo(nodeID uint8) error {
	requestPacket, err := message.ZWRequestNodeInfoRequest(nodeID)
	if err != nil {
		return err
	}
	responsePacket, err := api.blockingRequest(requestPacket)
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
