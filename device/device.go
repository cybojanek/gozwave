// Package device handles device enumeratio
package device

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

// Basic Type
const (
	BasicTypeController       uint8 = 0x01
	BasicTypeStaticController       = 0x02
	BasicTypeSlave                  = 0x03
	BasicTypeRoutingSlave           = 0x04
)

// Generic Type
const (
	GenericTypeGenericController  uint8 = 0x01
	GenericTypeStaticController         = 0x02
	GenericTypeAVControlPoint           = 0x03
	GenericTypeDisplay                  = 0x04
	GenericTypeNetworkExtender          = 0x05
	GenericTypeAppliance                = 0x06
	GenericTypeSensorNotification       = 0x07
	GenericTypeSwitchThermostat         = 0x08
	GenericTypeWindowCovering           = 0x09
	GenericTypeRepeaterSlave            = 0x0F
	GenericTypeSwitchBinary             = 0x10
	GenericTypeSwitchMultiLevel         = 0x11
	GenericTypeSwitchRemote             = 0x12
	GenericTypeSwitchToggle             = 0x13
	GenericTypeZipNode                  = 0x15
	GenericTypeVentilation              = 0x16
	GenericTypeSecurityPanel            = 0x17
	GenericTypeWallController           = 0x18
	GenericTypeSensorBinary             = 0x20
	GenericTypeSensorMultiLevel         = 0x21
	GenericTypeMeterPulse               = 0x30
	GenericTypeMeter                    = 0x31
	GenericTypeEntryControl             = 0x40
	GenericTypeSemiInteroperable        = 0x50
	GenericTypeSensorAlarm              = 0xA1
	GenericTypeNonInteroperable         = 0xFF
)

// Command Class
const (
	CommandClassNoOperation                 uint8 = 0x01
	CommandClassBinarySwitch                      = 0x25
	CommandClassAllSwitch                         = 0x27
	CommandClassMeter                             = 0x32
	CommandClassColorSwitch                       = 0x33
	CommandClassAssociationGroupInformation       = 0x59
	CommandClassZwavePlusInfo                     = 0x5e
	CommandClassConfiguration                     = 0x70
	CommandClassManufacturerSpecific              = 0x72
	CommandClassFirmwareUpdateMetadata            = 0x73
	CommandClassNodeNamingAndLocation             = 0x77
	CommandClassClock                             = 0x81
	CommandClassWakeup                            = 0x84
	CommandClassAssociation                       = 0x85
	CommandClassVersion                           = 0x86
	CommandClassMark                              = 0xef
)
