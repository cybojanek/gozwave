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
	BinarySensorTypeGeneralPurpose uint8 = 0x01
	BinarySensorTypeSmoke                = 0x02
	BinarySensorTypeCO                   = 0x03
	BinarySensorTypeCO2                  = 0x04
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

// IsIdle queries the sensor
func (node *BinarySensor) IsIdle() (bool, error) {
	if response, err := node.zwSendDataWaitForResponse(
		CommandClassBinarySensor, []uint8{binarySensorCommandGet},
		binarySensorCommandReport); err != nil {
		return false, err
	} else if len(response.Command.Data) != 1 {
		return false, fmt.Errorf("Bad response")
	} else {
		return response.Command.Data[0] == 0, nil
	}
}

// IsIdleV2 queries the sensor for the given type
func (node *BinarySensor) IsIdleV2(sensorType uint8) (bool, error) {
	if response, err := node.zwSendDataWaitForResponse(
		CommandClassBinarySensor, []uint8{binarySensorCommandGet, sensorType},
		binarySensorCommandReport); err != nil {
		return false, err
	} else if len(response.Command.Data) != 2 || response.Command.Data[1] != sensorType {
		return false, fmt.Errorf("Bad response")
	} else {
		return response.Command.Data[0] == 0, nil
	}
}
