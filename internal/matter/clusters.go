package matter

// Standard Matter cluster IDs
const (
	ClusterOccupancySensing       uint32 = 0x0406
	ClusterIlluminanceMeasurement uint32 = 0x0400
	ClusterTemperatureMeasurement uint32 = 0x0402
	ClusterPowerSource            uint32 = 0x002F
	ClusterOnOff                  uint32 = 0x0006
)

// Attribute IDs for OccupancySensing
const (
	AttributeOccupancy     uint32 = 0x0000
	AttributeOccupancyType uint32 = 0x0001
)

type ClusterAttribute struct {
	ClusterID   uint32
	AttributeID uint32
	Name        string
	DataType    string // "bool", "uint16", "int16", etc.
}
