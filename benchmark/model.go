package benchmark

// TemperaturePoint
type TemperaturePoint struct {
	SensorId    uint32  // we have 1,000,000 sensors
	Timestamp   uint32 // unix timestamp in seconds, but we have nothing before 1970 and have a year 21xx something problem
	Temperature int8   // for real world temperatures
}
