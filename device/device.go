// Package device handles device enumeratio
package device

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
