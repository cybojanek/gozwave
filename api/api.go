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

type ZWAPI struct {
	DevicePath string

	con *controller.Controller
}

// Check if api is open
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

func (api *ZWAPI) SerialAPIGetInitData() (*message.SerialAPIGetInitData, error) {
	requestPacket := message.SerialAPIGetInitDataRequest()
	if responsePacket, err := api.blockingRequest(requestPacket); err != nil {
		return nil, err
	} else {
		return message.SerialAPIGetInitDataResponse(responsePacket)
	}
}

func (api *ZWAPI) SerialAPIGetCapabilities() (*message.SerialAPIGetCapabilities, error) {
	requestPacket := message.SerialAPIGetCapabilitiesRequest()
	if responsePacket, err := api.blockingRequest(requestPacket); err != nil {
		return nil, err
	} else {
		return message.SerialAPIGetCapabilitiesResponse(responsePacket)
	}
}

func (api *ZWAPI) ZWGetControllerCapabilities() (*message.ZWGetControllerCapabilities, error) {
	requestPacket := message.ZWGetControllerCapabilitiesRequest()
	if responsePacket, err := api.blockingRequest(requestPacket); err != nil {
		return nil, err
	} else {
		return message.ZWGetControllerCapabilitiesResponse(responsePacket)
	}
}
