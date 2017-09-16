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
	"github.com/cybojanek/gozwave/controller"
	"github.com/cybojanek/gozwave/message"
	"github.com/cybojanek/gozwave/packet"
)

var (
	// ErrNodeNotFound is returned in case of node not found
	ErrNodeNotFound = errors.New("Node not found")
)

// ZWAPI instance
type ZWAPI struct {
	DevicePath string

	con *controller.Controller
}

// IsOpen checks if the API is open and initialized
func (api *ZWAPI) IsOpen() bool {
	return api.con != nil
}

// Open api
func (api *ZWAPI) Open() error {
	if api.IsOpen() {
		return nil
	}

	con := controller.Controller{DevicePath: api.DevicePath}
	if err := con.Open(); err != nil {
		return err
	}

	api.con = &con
	return nil
}

// Close api
func (api *ZWAPI) Close() error {
	if !api.IsOpen() {
		return nil
	}

	if err := api.con.Close(); err != nil {
		return err
	}

	api.con = nil
	return nil
}

// Sending a blocking request and wait for a reply
func (api *ZWAPI) blockingRequest(request *packet.Packet) (*packet.Packet, error) {
	if !api.IsOpen() {
		return nil, errors.New("API is not open")
	}
	return api.con.BlockingRequest(request)
}

// SerialAPIGetInitData gets the message.SerialAPIGetInitData information
func (api *ZWAPI) SerialAPIGetInitData() (*message.SerialAPIGetInitData, error) {
	requestPacket := message.SerialAPIGetInitDataRequest()
	responsePacket, err := api.blockingRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.SerialAPIGetInitDataResponse(responsePacket)
}

// SerialAPIGetCapabilities gets the message.SerialAPIGetCapabilities information
func (api *ZWAPI) SerialAPIGetCapabilities() (*message.SerialAPIGetCapabilities, error) {
	requestPacket := message.SerialAPIGetCapabilitiesRequest()
	responsePacket, err := api.blockingRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.SerialAPIGetCapabilitiesResponse(responsePacket)
}

// ZWGetControllerCapabilities gets the message.ZWGetControllerCapabilities
// information
func (api *ZWAPI) ZWGetControllerCapabilities() (*message.ZWGetControllerCapabilities, error) {
	requestPacket := message.ZWGetControllerCapabilitiesRequest()
	responsePacket, err := api.blockingRequest(requestPacket)
	if err != nil {
		return nil, err
	}
	return message.ZWGetControllerCapabilitiesResponse(responsePacket)
}

// ZWGetNodeProtocolInfo gets the message.ZWGetNodeProtocolInfo information
// for a requested node
func (api *ZWAPI) ZWGetNodeProtocolInfo(nodeID uint8) (*message.ZWGetNodeProtocolInfo, error) {
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
