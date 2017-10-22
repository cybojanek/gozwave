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
	"github.com/cybojanek/gozwave/message"
)

const (
	multiLevelSensorCommandGetSensorTypes    uint8 = 0x01
	multiLevelSensorCommandReportSensorTypes       = 0x02
	multiLevelSensorCommandGetScaleTypes           = 0x03
	multiLevelSensorCommandGet                     = 0x04
	multiLevelSensorCommandReport                  = 0x05
	multiLevelSensorCommandReportScaleTypes        = 0x06
)

// Multi Level Sensor Type
const (
	MultiLevelSensorTypeTemperature            uint8 = 0x01
	MultiLevelSensorTypeGeneral                      = 0x02
	MultiLevelSensorTypeLuminance                    = 0x03
	MultiLevelSensorTypePower                        = 0x04
	MultiLevelSensorTypeRelativeHumidity             = 0x05
	MultiLevelSensorTypeVelocity                     = 0x06
	MultiLevelSensorTypeDirection                    = 0x07
	MultiLevelSensorTypeAtmosphericPressure          = 0x08
	MultiLevelSensorTypeBarometricPressure           = 0x09
	MultiLevelSensorTypeSolarRadiation               = 0x0a
	MultiLevelSensorTypeDewPoint                     = 0x0b
	MultiLevelSensorTypeRainRate                     = 0x0c
	MultiLevelSensorTypeTideLevel                    = 0x0d
	MultiLevelSensorTypeWeight                       = 0x0e
	MultiLevelSensorTypeVoltage                      = 0x0f
	MultiLevelSensorTypeCurrent                      = 0x10
	MultiLevelSensorTypeCO2                          = 0x11
	MultiLevelSensorTypeAirFlow                      = 0x12
	MultiLevelSensorTypeTankCapacity                 = 0x13
	MultiLevelSensorTypeDistance                     = 0x14
	MultiLevelSensorTypeAnglePosition                = 0x15
	MultiLevelSensorTypeRotation                     = 0x16
	MultiLevelSensorTypeWaterTemperature             = 0x17
	MultiLevelSensorTypeSoilTemperature              = 0x18
	MultiLevelSensorTypeSeismicIntensity             = 0x19
	MultiLevelSensorTypeSeismicMagnitude             = 0x1a
	MultiLevelSensorTypeUltraviolet                  = 0x1b
	MultiLevelSensorTypeElectricalResistivity        = 0x1c
	MultiLevelSensorTypeElectricalConductivity       = 0x1d
	MultiLevelSensorTypeLoudness                     = 0x1e
	MultiLevelSensorTypeMoisture                     = 0x1f
)

// Multi Level Sensor Type Scales
const (
	MultiLevelSensorScaleTemperatureCelcius                    uint8 = 0x00
	MultiLevelSensorScaleTemperatureFahrenheit                       = 0x01
	MultiLevelSensorScaleGeneralPercentage                           = 0x00
	MultiLevelSensorScaleGeneralAbsolute                             = 0x01
	MultiLevelSensorScaleLuminancePercentage                         = 0x00
	MultiLevelSensorScaleLuminanceLUX                                = 0x01
	MultiLevelSensorScalePowerWatts                                  = 0x00
	MultiLevelSensorScalePowerBTUPerHour                             = 0x01
	MultiLevelSensorScaleRelativeHumidityPercentage                  = 0x00
	MultiLevelSensorScaleRelativeHumidityAbsolute                    = 0x01
	MultiLevelSensorScaleVelocityMetersPerSecond                     = 0x00
	MultiLevelSensorScaleVelocityMilesPerHour                        = 0x01
	MultiLevelSensorScaleDirectionAbsolute                           = 0x00
	MultiLevelSensorScaleAtmosphericPressureKiloPascals              = 0x00
	MultiLevelSensorScaleAtmosphericPressureInchesOfMercury          = 0x01
	MultiLevelSensorScaleSolarRadiationWattsPerMeterSquare           = 0x00
	MultiLevelSensorScaleDewPointCelcius                             = 0x00
	MultiLevelSensorScaleDewPointFahrenheit                          = 0x01
	MultiLevelSensorScaleRainRateMillimetersPerHour                  = 0x00
	MultiLevelSensorScaleRainRateInchesPerHour                       = 0x01
	MultiLevelSensorScaleTideLevelMeters                             = 0x00
	MultiLevelSensorScaleTideLevelFeet                               = 0x01
	MultiLevelSensorScaleWeightKilograms                             = 0x00
	MultiLevelSensorScaleWeightPounds                                = 0x01
	MultiLevelSensorScaleVoltageVolts                                = 0x00
	MultiLevelSensorScaleVoltageMilliVolts                           = 0x01
	MultiLevelSensorScaleCurrentAmps                                 = 0x00
	MultiLevelSensorScaleCurrentMilliAmps                            = 0x01
	MultiLevelSensorScaleCO2PartsPerMillion                          = 0x00
	MultiLevelSensorScaleAirflowMetersCubedPerHour                   = 0x00
	MultiLevelSensorScaleAirflowCubicFeetPerMinute                   = 0x01
	MultiLevelSensorScaleTankCapacityLiters                          = 0x00
	MultiLevelSensorScaleTankCapacityCubicMeters                     = 0x01
	MultiLevelSensorScaleTankCapacityGallons                         = 0x02
	MultiLevelSensorScaleDistanceMeters                              = 0x00
	MultiLevelSensorScaleDistanceCentimeters                         = 0x01
	MultiLevelSensorScaleDistanceFeet                                = 0x02
	MultiLevelSensorScaleAnglePositionPercentage                     = 0x00
	MultiLevelSensorScaleAnglePositionDegreesNorth                   = 0x01
	MultiLevelSensorScaleAnglePositionDegreesSouth                   = 0x02
	MultiLevelSensorScaleRotationRPM                                 = 0x00
	MultiLevelSensorScaleRotationHZ                                  = 0x01
	MultiLevelSensorScaleWaterTemperatureCelcius                     = 0x00
	MultiLevelSensorScaleWaterTemperatureFahrenheit                  = 0x01
	MultiLevelSensorScaleSoilTemperatureCelcius                      = 0x00
	MultiLevelSensorScaleSoilTemperatureFahrenheit                   = 0x01
	MultiLevelSensorScaleSeismicIntensityMercalli                    = 0x00
	MultiLevelSensorScaleSeismicIntensityEUMacroseismic              = 0x01
	MultiLevelSensorScaleSeismicIntensityLiedu                       = 0x02
	MultiLevelSensorScaleSeismicIntensityShindo                      = 0x03
	MultiLevelSensorScaleSeismicMagnitudeLocal                       = 0x00
	MultiLevelSensorScaleSeismicMagnitudeMoment                      = 0x01
	MultiLevelSensorScaleSeismicMagnitudeSurfaceWave                 = 0x02
	MultiLevelSensorScaleSeismicMagnitudeBodyWave                    = 0x03
	MultiLevelSensorScaleUltravioletAbsolute                         = 0x00
	MultiLevelSensorScaleElectricalResistivityOHM                    = 0x00
	MultiLevelSensorScaleElectricalConductivitySiemensPerMeter       = 0x00
	MultiLevelSensorScaleLoudnessDecibel                             = 0x00
	MultiLevelSensorScaleLoudnessDecibelWeighting                    = 0x01
	MultiLevelSensorScaleMoisturePercentage                          = 0x00
	MultiLevelSensorScaleMoistureContent                             = 0x01
	MultiLevelSensorScaleMoistureKiloOhms                            = 0x02
	MultiLevelSensorScaleMoistureWaterActivity                       = 0x03
)

// MultiLevelSensorResult information
type MultiLevelSensorResult struct {
	SensorType  uint8
	SensorScale uint8
	Value       float32
}

// MultiLevelSensor information
type MultiLevelSensor struct {
	*Node
}

// GetMultiLevelSensor returns a MultiLevelSensor or nil object
func (node *Node) GetMultiLevelSensor() *MultiLevelSensor {
	node.mutex.Lock()
	defer node.mutex.Unlock()

	if node.supportsCommandClass(CommandClassMultiLevelSensor) {
		return &MultiLevelSensor{node}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Get queries the sensor, expects a V1-4 reply
func (node *MultiLevelSensor) Get() (*MultiLevelSensorResult, error) {
	var response *ApplicationCommandData
	var err error

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassMultiLevelSensor, []uint8{multiLevelSensorCommandGet},
		multiLevelSensorCommandReport, nil); err != nil {
		return nil, err
	}

	return node.ParseReport(response)
}

// IsReport checks if the report is a ParseReport
func (node *MultiLevelSensor) IsReport(report *ApplicationCommandData) bool {
	return report.Command.ID == multiLevelSensorCommandReport
}

// ParseReport of status
func (node *MultiLevelSensor) ParseReport(report *ApplicationCommandData) (*MultiLevelSensorResult, error) {
	var result MultiLevelSensorResult
	var err error

	if report.Command.ClassID != CommandClassMultiLevelSensor {
		return nil, fmt.Errorf("Bad Report Command Class ID: 0x%02x != 0x%02x",
			report.Command.ClassID, CommandClassMultiLevelSensor)
	}

	if report.Command.ID != multiLevelSensorCommandReport {
		return nil, fmt.Errorf("Bad Report Command ID 0x%02x != 0x%02x",
			report.Command.ID, multiLevelSensorCommandReport)
	}

	data := report.Command.Data
	if len(data) < 3 {
		return nil, fmt.Errorf("Bad Report Data length %d < 1", len(data))
	}

	// Sensor Type
	sensorType := data[0]
	result.SensorType = sensorType

	// Precision, Scale, Size
	precision := (data[1] >> 5) & 0x7
	scale := (data[1] >> 3) & 0x3
	size := data[1] & 0x7

	result.SensorScale = scale

	// Offset into array
	offset := 2

	// Decode Value
	var value float32
	if value, err = message.DecodeFloat(data[offset:offset+int(size)], precision); err != nil {
		return nil, err
	}
	result.Value = float32(value)
	offset += int(size)

	return &result, nil
}

////////////////////////////////////////////////////////////////////////////////

// GetSupportedSensorTypes queries the sensor to get the list of supported sensor types
func (node *MultiLevelSensor) GetSupportedSensorTypes() ([]uint8, error) {
	var response *ApplicationCommandData
	var err error

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassMultiLevelSensor, []uint8{multiLevelSensorCommandGetSensorTypes},
		multiLevelSensorCommandReportSensorTypes, nil); err != nil {
		return nil, err
	}

	// Loop over bit mask
	var sensors []uint8
	sensorType := uint8(1)
	for _, b := range response.Command.Data {
		for i := uint32(0); i < 8; i++ {
			if (b & (1 << i)) != 0 {
				sensors = append(sensors, sensorType)
			}
			sensorType++
		}
	}

	return sensors, nil
}

// GetSupportedScaleTypes queries the sensor to get the list of supported sensor
// scale types for the given sensor type
func (node *MultiLevelSensor) GetSupportedScaleTypes(sensorType uint8) ([]uint8, error) {
	var response *ApplicationCommandData
	var err error

	filter := func(response *ApplicationCommandData) bool {
		return len(response.Command.Data) > 0 && response.Command.Data[0] == sensorType
	}

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassMultiLevelSensor, []uint8{multiLevelSensorCommandGetScaleTypes},
		multiLevelSensorCommandReportScaleTypes, filter); err != nil {
		return nil, err
	}

	data := response.Command.Data
	if len(data) != 2 {
		return nil, fmt.Errorf("Response size mismatch %d != 2", len(data))
	}

	var scaleIndices []uint8
	for i := uint8(0); i < 4; i++ {
		if ((1 << i) & data[1]) != 0 {
			scaleIndices = append(scaleIndices, i)
		}
	}

	return scaleIndices, nil
}

////////////////////////////////////////////////////////////////////////////////

// GetV5 queries the sensor, expects a V5-V11 response
func (node *MultiLevelSensor) GetV5(sensorType uint8) (*MultiLevelSensorResult, error) {
	var response *ApplicationCommandData
	var err error

	filter := func(response *ApplicationCommandData) bool {
		return len(response.Command.Data) > 0 && response.Command.Data[0] == sensorType
	}

	if response, err = node.zwSendDataWaitForResponse(
		CommandClassMultiLevelSensor, []uint8{multiLevelSensorCommandGet, sensorType},
		multiLevelSensorCommandReport, filter); err != nil {
		return nil, err
	}

	return node.ParseReport(response)
}
