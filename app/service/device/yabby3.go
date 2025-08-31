//go:generate stringer -type=ReasonCode
package device

import (
	"encoding/json"
	"fmt"
)

// TODO: do these field names need to be exported?

// These types are a mix of JSON decoding types (DAOs?) and models. The
// difference between the two (besides the JSON tags) is that I flatten the
// []Fields from the device, with their optional members, into pointers to types
// (thus optional). This works better with Go than having to have switch.(type)
// statements in multiple places. Don't let the device's design leak into the
// business logic.
// I can only get away with this because of the irregular JSON decoding that
// gives me a place to insert logic and make the Reading* elements as I want
// them (pointers).
// One place where this falls down is in the AnalogueReading, where voltages are
// Num1..4, instead of InternalBatteryVoltage, Temperature, LastGSMCQ and
// LoadedVoltage.

type TagTx struct {
	ID      string
	SerNo   int      `json:"SerNo"`
	IMEI    string   `json:"IMEI"`
	ICCID   string   `json:"ICCID"`
	ProdID  int      `json:"ProdId"`
	Fw      string   `json:"FW"`
	Records []Record `json:"Records"`
}

type Record struct {
	ID              string
	SeqNo           int    `json:"SeqNo"`
	Reason          int    `json:"Reason"`
	DateUTC         string `json:"DateUTC"`
	GPSReading      *GPSReading
	GPIOReading     *GPIOReading
	AnalogueReading *AnalogueReading
	TripTypeReading *TripTypeReading
}

type GPSReading struct { // FType0
	Spd     int     `json:"Spd"`
	SpdAcc  int     `json:"SpdAcc"`
	Head    int     `json:"Head"`
	GpsStat int     `json:"GpsStat"`
	GpsUTC  string  `json:"GpsUTC"`
	Lat     float64 `json:"Lat"`
	Long    float64 `json:"Long"`
	Alt     int     `json:"Alt"`
	FType   int     `json:"FType"`
	PosAcc  int     `json:"PosAcc"`
	PDOP    int     `json:"PDOP"`
}

type GPIOReading struct { // FType2
	DIn     int `json:"DIn"`
	DOut    int `json:"DOut"`
	DevStat int `json:"DevStat"`
	FType   int `json:"FType"`
}

type AnalogueReading struct { // FType6
	AnalogueData AnalogueData `json:"AnalogueData"`
	FType        int          `json:"FType"`
}

type AnalogueData struct {
	Num1 int `json:"1"`
	Num3 int `json:"3"`
	Num4 int `json:"4"`
	Num5 int `json:"5"`
}

type TripTypeReading struct { // FType15
	FType int `json:"FType"`
	Tt    int `json:"TT"`
	Trim  int `json:"Trim"`
}

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

const MinReasonCode int = 1
const MaxReasonCode int = 50

// TODO: Is this needed, or can you just typecast an int to ReasonCode
var ReasonMap = map[int]ReasonCode{
	1:  StartOfTrip,
	2:  EndOfTrip,
	3:  ElapsedTime,
	4:  SpeedChange,
	5:  HeadingChange,
	6:  DistanceTravelled,
	7:  MaximumSpeed,
	8:  Stationary,
	9:  DigitalInputChanged,
	10: DigitalOutputChanged,
	11: HeartbeatStatus,
	12: HarshBrake,
	13: HarshAcceleration,
	14: HarshCornering,
	15: ExternalPowerChange,
	16: SystemPowerMonitoring,
	17: DriverIDTagRead,
	18: Overspeed,
	19: FuelSensorRecord,
	20: TowingAlert,
	21: DebugMessage,
	22: SDI12SensorDataRecorded,
	23: Accident,
	24: AccidentData,
	25: SensorValueElapsedTime,
	26: SensorValueChange,
	27: SensorAlarm,
	28: RainGaugeTipped,
	29: TamperAlert,
	30: BLOBNotification,
	31: TimeAndAttendance,
	32: TripRestart,
	33: TagGained,
	34: TagUpdate,
	35: TagLost,
	36: RecoveryModeOn,
	37: RecoveryModeOff,
	38: ImmobiliserOn,
	39: ImmobiliserOff,
	40: GarminFMIStopResponse,
	41: LoneWorkerAlarm,
	42: DeviceCounters,
	43: ConnectedDeviceData,
	44: EnteredGeoFence,
	45: ExitedGeoFence,
	46: HighGEvent,
	47: Reserved,
	48: Duress,
	49: CellTowerConnection,
	50: BluetoothTagData,
}

/*
Hoping to not need these with the Stringer code gen
1     Start of trip
2     End of trip
3     Elapsed time
4     Speed change
5     Heading change
6     Distance travelled
7     Maximum speed (not used)
8     Stationary
9     Digital Input Changed
10    Digital Output Changed
11    Heartbeat / Status
12    Harsh Brake
13    Harsh Acceleration
14    Harsh Cornering
15    External Power Change
16    System power monitoring
17    Driver ID Tag Read
18    Over speed
19   Fuel sensor record
20   Towing Alert (not used)
21   Debug message
22   SDI12 sensor data recorded
23   Accident
24   Accident Data
25   Sensor value elapsed time
26   Sensor value change
27   Sensor alarm
28   Rain Gauge Tipped
29   Tamper Alert
30   BLOB notification (not used)
31   Time and Attendance
32   Trip Restart
33   Tag Gained (not used)
34   Tag Update (not used)
35   Tag Lost (not used)
36   Recovery Mode On
37   Recovery Mode Off
38   Immobiliser On
39   Immobiliser Off
40   Garmin FMI Stop Response
41   Lone Worker Alarm
42   Device Counters
43   Connected Device Data
44   Entered Geo-Fence
45   Exited Geo-Fence
46   High-G Event
47   Reserved
48   Duress
49   Cell Tower Connection
50   Bluetooth Tag Data
*/

func (r *Record) UnmarshalJSON(p []byte) error {
	// The type of the Fields varies.
	// Unmarshal the regular parts of the JSON value
	var rawRecord struct {
		DateUTC string            `json:"DateUTC"`
		Fields  []json.RawMessage `json:"Fields"`
		SeqNo   int               `json:"SeqNo"`
		Reason  int               `json:"Reason"`
	}

	if err := json.Unmarshal(p, &rawRecord); err != nil {
		return fmt.Errorf("custom unmarshaling Record: %w", err)
	}

	r.DateUTC = rawRecord.DateUTC
	r.SeqNo = rawRecord.SeqNo
	r.Reason = rawRecord.Reason

	for _, rawField := range rawRecord.Fields {
		fType := struct {
			FType int `json:"FType"`
		}{}
		if err := json.Unmarshal(rawField, &fType); err != nil {
			return fmt.Errorf("unmarshaling the raw field %v: %w", rawField, err)
		}

		switch fType.FType {
		case 0:
			var f *GPSReading
			if err := json.Unmarshal(rawField, &f); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.GPSReading = f
		case 2:
			var f *GPIOReading
			if err := json.Unmarshal(rawField, &f); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.GPIOReading = f
		case 6:
			var f *AnalogueReading
			if err := json.Unmarshal(rawField, &f); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.AnalogueReading = f
		case 15:
			var f *TripTypeReading
			if err := json.Unmarshal(rawField, &f); err != nil {
				return fmt.Errorf("unmarshaling the field %v: %w", rawField, err)
			}
			r.TripTypeReading = f
		default:
			return fmt.Errorf("unrecognised FType: %v", fType.FType)
		}
	}

	return nil
}

const AnalogueDataFType = 6

var IdToName = map[float64]string{
	810095: "rueger",
	810243: "charlie",
}
