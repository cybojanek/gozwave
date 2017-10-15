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
	"github.com/cybojanek/gozwave/message"
)

const (
	meterCommandGet    uint8 = 0x01
	meterCommandReport       = 0x02
)

// Meter Type
const (
	MeterTypeElectric uint8 = 0x01
	MeterTypeGas            = 0x02
	MeterTypeWater          = 0x03
	MeterTypeHeating        = 0x04
	MeterTypeCooling        = 0x05
)

// Meter Scale
const (
	MeterScaleElectricKWH         uint8 = 0x00
	MeterScaleElectricKVAH              = 0x01
	MeterScaleElectricW                 = 0x02
	MeterScaleElectricPulseCount        = 0x03
	MeterScaleElectricV                 = 0x04
	MeterScaleElectricA                 = 0x05
	MeterScaleElectricPowerFactor       = 0x06
	MeterScaleElectricMST               = 0x07

	MeterScaleGasCubicMeters = 0x00
	MeterScaleGasCubicFeet   = 0x01
	MeterScaleGasPulseCount  = 0x03
	MeterScaleGasMST         = 0x07

	MeterScaleWaterCubicMeters     = 0x00
	MeterScaleWaterCubicFeet       = 0x01
	MeterScaleWaterCubicUSGallons  = 0x02
	MeterScaleWaterCubicPulseCount = 0x03
	MeterScaleWaterMST             = 0x07

	MeterScaleHeatingKWH = 0x00

	MeterScaleCoolingKWH = 0x00
)

// Rate Type
const (
	RateTypeNone   uint8 = 0x00
	RateTypeImport       = 0x01
	RateTypeExport       = 0x02
)

// MeterResult information
type MeterResult struct {
	MeterType     uint8
	MeterScale    uint8
	RateType      uint8
	Value         float32
	DeltaTime     uint16
	PreviousValue float32
}

// Meter information
type Meter struct {
	*Node
}

// GetMeter returns a Meter or nil object
func (node *Node) GetMeter() *Meter {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassMeter) {
		return &Meter{node}
	}

	return nil
}

// Get current value. Supports both V1 and V2.
func (node *Meter) Get() (*MeterResult, error) {
	var response *applicationCommandData
	var err error
	var result MeterResult

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassMeter, []uint8{meterCommandGet},
		meterCommandReport); err != nil {
		return nil, err
	}

	data := response.Command.Data

	if len(data) < 3 {
		return nil, fmt.Errorf("Bad response, data too short %d < %d", len(data))
	}

	// Process MeterType
	// NOTE: this handles both V1, V2
	meterType := data[0] & 0x1f
	switch meterType {
	case MeterTypeElectric, MeterTypeGas, MeterTypeWater, MeterTypeHeating, MeterTypeCooling:
		result.MeterType = meterType
	default:
		return nil, fmt.Errorf("Unknown meter type: 0x%02x", meterType)
	}

	// Process RateType
	// NOTE: this will be 0 for V1
	rateType := (data[0] >> 5) & 0x3
	switch rateType {
	case RateTypeNone, RateTypeImport, RateTypeExport:
		result.RateType = rateType
	default:
		return nil, fmt.Errorf("Unknown rate type: 0x%02x", rateType)
	}

	// Number of decimal points
	precision := uint8((data[1] >> 5) & 0x7)

	// What is it reporting KWH, KVAH, etc...
	scale := uint8((data[1] >> 3) & 0x3)
	switch meterType {
	case MeterTypeElectric:
		switch scale {
		case MeterScaleElectricKWH, MeterScaleElectricKVAH, MeterScaleElectricW, MeterScaleElectricPulseCount, MeterScaleElectricV, MeterScaleElectricA, MeterScaleElectricPowerFactor, MeterScaleElectricMST:
			result.MeterType = meterType
		default:
			return nil, fmt.Errorf("Unknown scale 0x%02x for type Electric", scale)
		}
	case MeterTypeGas:
		switch scale {

		case MeterScaleGasCubicMeters, MeterScaleGasCubicFeet, MeterScaleGasPulseCount, MeterScaleGasMST:
			result.MeterType = meterType
		default:
			return nil, fmt.Errorf("Unknown scale 0x%02x for type Gas", scale)
		}
	case MeterTypeWater:
		switch scale {
		case MeterScaleWaterCubicMeters, MeterScaleWaterCubicFeet, MeterScaleWaterCubicUSGallons, MeterScaleWaterCubicPulseCount, MeterScaleWaterMST:
			result.MeterType = meterType
		default:
			return nil, fmt.Errorf("Unknown scale 0x%02x for type Water", scale)
		}
	case MeterTypeHeating:
		switch scale {
		case MeterScaleHeatingKWH:
			result.MeterType = meterType
		default:
			return nil, fmt.Errorf("Unknown scale 0x%02x for type Heating", scale)
		}
	case MeterTypeCooling:
		switch scale {
		case MeterScaleCoolingKWH:
			result.MeterType = meterType
		default:
			return nil, fmt.Errorf("Unknown scale 0x%02x for type Cooling", scale)
		}
	}

	// Bytes of data
	size := uint8(data[1] & 0x7)
	if len(data) < int(size)+2 {
		return nil, fmt.Errorf("Bad size %d < %d", len(data), size+2)
	}

	// Offset into array
	offset := 2

	// Decode Value
	var value float32
	if value, err = message.DecodeFloat(data[offset:offset+int(size)], precision); err != nil {
		return nil, err
	}
	result.Value = float32(value)
	offset += int(size)

	// If we have more bytes, check for DeltaTime and PreviousValue
	if len(data) > offset {

		// Check we have enough for DeltaTime
		if len(data) < offset+2 {
			return nil, fmt.Errorf("Expected Delta Time, bad size: %d < %d",
				len(data), offset+2)
		}
		// Decode DeltaTime
		result.DeltaTime = binary.BigEndian.Uint16(data[offset : offset+2])
		offset += 2
	}

	// Check for PreviousValue
	if result.DeltaTime > 0 && len(data) > offset {
		// If we have more bytes, try to decode previous value
		if len(data) < offset+int(size) {
			return nil, fmt.Errorf("Expected Previous Value, bad size: %d < %d",
				len(data), offset+int(size))
		}
		if value, err = message.DecodeFloat(data[offset:offset+int(size)], precision); err != nil {
			return nil, err
		}
		result.PreviousValue = value
		offset += int(size)
	}

	// NOTE: it looks like there might be 4 extra 0 bytes? Why?

	return &result, nil
}
