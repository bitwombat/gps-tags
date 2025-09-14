//go:generate stringer -type=ReasonCode
package model

type ReasonCode int

const (
	StartOfTrip             ReasonCode = 1
	EndOfTrip               ReasonCode = 2
	ElapsedTime             ReasonCode = 3
	SpeedChange             ReasonCode = 4
	HeadingChange           ReasonCode = 5
	DistanceTravelled       ReasonCode = 6
	MaximumSpeed            ReasonCode = 7
	Stationary              ReasonCode = 8
	DigitalInputChanged     ReasonCode = 9
	DigitalOutputChanged    ReasonCode = 10
	HeartbeatStatus         ReasonCode = 11
	HarshBrake              ReasonCode = 12
	HarshAcceleration       ReasonCode = 13
	HarshCornering          ReasonCode = 14
	ExternalPowerChange     ReasonCode = 15
	SystemPowerMonitoring   ReasonCode = 16
	DriverIDTagRead         ReasonCode = 17
	Overspeed               ReasonCode = 18
	FuelSensorRecord        ReasonCode = 19
	TowingAlert             ReasonCode = 20
	DebugMessage            ReasonCode = 21
	SDI12SensorDataRecorded ReasonCode = 22
	Accident                ReasonCode = 23
	AccidentData            ReasonCode = 24
	SensorValueElapsedTime  ReasonCode = 25
	SensorValueChange       ReasonCode = 26
	SensorAlarm             ReasonCode = 27
	RainGaugeTipped         ReasonCode = 28
	TamperAlert             ReasonCode = 29
	BLOBNotification        ReasonCode = 30
	TimeAndAttendance       ReasonCode = 31
	TripRestart             ReasonCode = 32
	TagGained               ReasonCode = 33
	TagUpdate               ReasonCode = 34
	TagLost                 ReasonCode = 35
	RecoveryModeOn          ReasonCode = 36
	RecoveryModeOff         ReasonCode = 37
	ImmobiliserOn           ReasonCode = 38
	ImmobiliserOff          ReasonCode = 39
	GarminFMIStopResponse   ReasonCode = 40
	LoneWorkerAlarm         ReasonCode = 41
	DeviceCounters          ReasonCode = 42
	ConnectedDeviceData     ReasonCode = 43
	EnteredGeoFence         ReasonCode = 44
	ExitedGeoFence          ReasonCode = 45
	HighGEvent              ReasonCode = 46
	Reserved                ReasonCode = 47
	Duress                  ReasonCode = 48
	CellTowerConnection     ReasonCode = 49
	BluetoothTagData        ReasonCode = 50
)

const (
	MinReasonCode int = 1
	MaxReasonCode int = 50
)
