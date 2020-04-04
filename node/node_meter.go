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
	meterCommandGet             uint8 = 0x01
	meterCommandReport                = 0x02
	meterCommandSupportedGet          = 0x03
	meterCommandSupportedReport       = 0x04
	meterCommandReset                 = 0x05
)

// Meter Type
const (
	MeterTypeElectric uint8 = 0x01
	MeterTypeGas            = 0x02
	MeterTypeWater          = 0x03
	MeterTypeHeating        = 0x04
	MeterTypeCooling        = 0x05
)

// Meter Scale ELectric
const (
	MeterScaleElectricKWH         uint8 = 0x00
	MeterScaleElectricKVAH              = 0x01
	MeterScaleElectricW                 = 0x02
	MeterScaleElectricPulseCount        = 0x03
	MeterScaleElectricV                 = 0x04
	MeterScaleElectricA                 = 0x05
	MeterScaleElectricPowerFactor       = 0x06
	MeterScaleElectricMST               = 0x07
)

// Meter Scale Gas
const (
	MeterScaleGasCubicMeters uint8 = 0x00
	MeterScaleGasCubicFeet         = 0x01
	MeterScaleGasPulseCount        = 0x03
	MeterScaleGasMST               = 0x07
)

// Meter Scale Water
const (
	MeterScaleWaterCubicMeters     uint8 = 0x00
	MeterScaleWaterCubicFeet             = 0x01
	MeterScaleWaterCubicUSGallons        = 0x02
	MeterScaleWaterCubicPulseCount       = 0x03
	MeterScaleWaterMST                   = 0x07
)

// Meter Scale Heating
const (
	MeterScaleHeatingKWH uint8 = 0x00
)

// Meter Scale Cooling
const (
	MeterScaleCoolingKWH uint8 = 0x00
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

////////////////////////////////////////////////////////////////////////////////

// Get current value
func (node *Meter) Get() (*MeterResult, error) {
	var response *ApplicationCommandData
	var err error

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassMeter, []uint8{meterCommandGet},
		meterCommandReport, nil); err != nil {
		return nil, err
	}

	return node.ParseReport(response)
}

// IsReport checks if the report is a ParseReport
func (node *Meter) IsReport(report *ApplicationCommandData) bool {
	return report.Command.ID == meterCommandReport
}

// ParseReport of status
func (node *Meter) ParseReport(report *ApplicationCommandData) (*MeterResult, error) {
	var result MeterResult
	var err error

	if report.Command.ClassID != CommandClassMeter {
		return nil, fmt.Errorf("Bad Report Command Class ID: 0x%02x != 0x%02x",
			report.Command.ClassID, CommandClassMeter)
	}

	if report.Command.ID != meterCommandReport {
		return nil, fmt.Errorf("Bad Report Command ID 0x%02x != 0x%02x",
			report.Command.ID, meterCommandReport)
	}

	data := report.Command.Data
	if len(data) < 3 {
		return nil, fmt.Errorf("Bad response, data too short %d < 3", len(data))
	}

	// Process MeterType
	// NOTE: this handles both V1, V2, V3
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
	// Handle V3
	if (data[0] & 0x80) != 0 {
		scale |= 4
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

	// Handle V4
	if len(data) > offset && scale == 0x7 {
		// Actual scale is in Scale2 byte
		scale = data[offset]
		offset++
	}
	result.MeterScale = scale

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

	return &result, nil
}

////////////////////////////////////////////////////////////////////////////////

// GetV2 the current value in the requested scale
func (node *Meter) GetV2(scaleType uint8) (*MeterResult, error) {
	if (scaleType & 0x3) != scaleType {
		return nil, fmt.Errorf("Scale out of range [0, 3]")
	}

	var response *ApplicationCommandData
	var err error

	filter := func(response *ApplicationCommandData) bool {
		if result, err := node.ParseReport(response); err == nil && result.MeterScale == scaleType {
			return true
		}
		return false
	}

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassMeter, []uint8{meterCommandGet, scaleType << 3},
		meterCommandReport, filter); err != nil {
		return nil, err
	}

	return node.ParseReport(response)
}

// GetSupported queries the meter to get the supported operations information
func (node *Meter) GetSupported() (canReset bool, rateType uint8, meterType uint8, scales []uint8, err error) {
	var response *ApplicationCommandData

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassMeter, []uint8{meterCommandSupportedGet},
		meterCommandSupportedReport, nil); err != nil {
		return
	}

	data := response.Command.Data
	if len(data) < 2 {
		err = fmt.Errorf("Response size mismatch %d < 2", len(data))
		return
	}

	canReset = (data[0] & 0x80) != 0
	// Rate Type is V4 extension
	rateType = (data[0] >> 5) & 0x3
	meterType = (data[0] & 0x1f)

	// Get first 7 bits
	scaleType := uint8(0)
	for i := uint32(0); i < 7; i++ {
		if (data[1] & (1 << i)) != 0 {
			scales = append(scales, scaleType)
		}
		scaleType++
	}

	// FIXME: looks like there are still trom trailing bytes in V4?

	// Handle V4 extension
	if (data[1] & 0x80) != 0 {
		scaleBytes := data[2:]
		if len(scaleBytes) == 0 {
			err = fmt.Errorf("Scales bytes missing")
			return
		} else if len(scaleBytes)-1 != int(scaleBytes[0]) {
			err = fmt.Errorf("Scales bytes size mismatch %d != %d",
				len(scaleBytes)-1, scaleBytes[0])
			return
		}

		for _, b := range scaleBytes[1:] {
			for i := uint32(0); i < 8; i++ {
				if (b & (1 << i)) != 0 {
					scales = append(scales, scaleType)
				}
				scaleType++
			}
		}

	} else if len(data) != 2 {
		err = fmt.Errorf("Response size mismatch %d != 2", len(data))
		return
	}

	return
}

// Reset meter
func (node *Meter) Reset() error {
	return node.zwSendDataRequest(CommandClassMeter, []uint8{meterCommandReset})
}

////////////////////////////////////////////////////////////////////////////////

// GetV3 the current value in the requested scale
func (node *Meter) GetV3(scaleType uint8) (*MeterResult, error) {
	if (scaleType & 0x7) != scaleType {
		return nil, fmt.Errorf("Scale out of range [0, 7]")
	}

	var response *ApplicationCommandData
	var err error

	filter := func(response *ApplicationCommandData) bool {
		if result, err := node.ParseReport(response); err == nil && result.MeterScale == scaleType {
			return true
		}
		return false
	}

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassMeter, []uint8{meterCommandGet, scaleType << 3},
		meterCommandReport, filter); err != nil {
		return nil, err
	}

	return node.ParseReport(response)
}

////////////////////////////////////////////////////////////////////////////////

// GetV4 the current value in the requested scale
func (node *Meter) GetV4(scaleType uint8, rateType uint8) (*MeterResult, error) {
	if (rateType & 0x3) != rateType {
		return nil, fmt.Errorf("Scale out of range [0, 3]")
	}

	var response *ApplicationCommandData
	var err error

	filter := func(response *ApplicationCommandData) bool {
		if result, err := node.ParseReport(response); err == nil && result.MeterScale == scaleType && result.RateType == rateType {
			return true
		}
		return false
	}

	data := []uint8{meterCommandGet, rateType << 6}
	if scaleType <= 0x7 {
		// FIXME: is this required?
		data[1] |= (scaleType << 3)
	} else {
		data[1] |= (0x7 << 3)
		data = append(data, scaleType)
	}

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassMeter, data, meterCommandReport, filter); err != nil {
		return nil, err
	}

	return node.ParseReport(response)
}
