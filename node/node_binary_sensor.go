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
	binarySensorCommandGet    uint8 = 0x02
	binarySensorCommandReport       = 0x03
)

// Binary Sensor Type
const (
	BinarySensorTypeGeneral        uint8 = 0x01
	BinarySensorTypeSmoke                = 0x02
	BinarySensorTypeCarbonMonoxide       = 0x03
	BinarySensorTypeCarbonDioxide        = 0x04
	BinarySensorTypeHeat                 = 0x05
	BinarySensorTypeWater                = 0x06
	BinarySensorTypeFreeze               = 0x07
	BinarySensorTypeTamper               = 0x08
	BinarySensorTypeAUX                  = 0x09
	BinarySensorTypeDoorWindow           = 0x0a
	BinarySensorTypeTilt                 = 0x0b
	BinarySensorTypeMotion               = 0x0c
	BinarySensorTypeGlassBreak           = 0x0d
	BinarySensorTypeFirstSupported       = 0xff
)

// BinarySensor information
type BinarySensor struct {
	*Node
}

// GetBinarySensor returns a BinarySensor or nil object
func (node *Node) GetBinarySensor() *BinarySensor {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassBinarySensor) {
		return &BinarySensor{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// IsActive queries the sensor
func (node *BinarySensor) IsActive() (bool, error) {
	var response *ApplicationCommandData
	var err error

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassBinarySensor, []uint8{binarySensorCommandGet},
		binarySensorCommandReport, nil); err != nil {
		return false, err
	}

	isActive, _, err := node.ParseReport(response)
	return isActive, err
}

// IsReport checks if the report is a ParseReport
func (node *BinarySensor) IsReport(report *ApplicationCommandData) bool {
	return report.Command.ID == binarySensorCommandReport
}

// ParseReport of status
func (node *BinarySensor) ParseReport(report *ApplicationCommandData) (isActive bool, sensorType uint8, err error) {
	if report.Command.ClassID != CommandClassBinarySensor {
		err = fmt.Errorf("Bad Report Command Class ID: 0x%02x != 0x%02x",
			report.Command.ClassID, CommandClassBinarySensor)
		return
	}

	if report.Command.ID != binarySensorCommandReport {
		err = fmt.Errorf("Bad Report Command ID 0x%02x != 0x%02x",
			report.Command.ID, binarySensorCommandReport)
		return
	}

	data := report.Command.Data
	if len(data) != 1 && len(data) != 2 {
		err = fmt.Errorf("Bad Report Data length %d != 1 and %d != 2", len(data), len(data))
		return
	}

	isActive = data[0] == 0xff
	sensorType = BinarySensorTypeGeneral
	if len(data) > 1 {
		sensorType = data[1]
	}
	return
}

////////////////////////////////////////////////////////////////////////////////

// IsActiveV2 queries the sensor for the given type
func (node *BinarySensor) IsActiveV2(sensorType uint8) (bool, error) {
	var response *ApplicationCommandData
	var err error

	filter := func(response *ApplicationCommandData) bool {
		return len(response.Command.Data) > 1 && response.Command.Data[1] == sensorType
	}

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassBinarySensor, []uint8{binarySensorCommandGet, sensorType},
		binarySensorCommandReport, filter); err != nil {
		return false, err
	}

	isActive, _, err := node.ParseReport(response)
	return isActive, err
}
