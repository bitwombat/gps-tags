package model

// This is a copy of ../device/yabby3.go, because the device's structure is an
// OK starting point for our model (and db schema). A model that is nearly a
// copy is a good idea for better separation (the yabby can change with minimal
// impact, new devices can be supported with no impact to higher levels), and
// also so that the cmd/migration utility can use it without confusingly having to
// convert its type to yabby types.
//
// Changes are:
// - json tags removed
// - FType removed
// - Analogue structure flattened and Num* replaced with meaningful names.

type TagTx struct {
	ID      string
	SerNo   int
	IMEI    string
	ICCID   string
	ProdID  int
	Fw      string
	Records []Record
}

type Record struct {
	ID              string
	SeqNo           int
	Reason          ReasonCode
	DateUTC         Time
	GPSReading      *GPSReading
	GPIOReading     *GPIOReading
	AnalogueReading *AnalogueReading
	TripTypeReading *TripTypeReading
}

type GPSReading struct {
	Spd     int
	SpdAcc  int
	Head    int
	GpsStat int
	GpsUTC  Time
	Lat     float64
	Long    float64
	Alt     int
	PosAcc  int
	PDOP    int
}

type GPIOReading struct {
	DIn     int
	DOut    int
	DevStat int
}

type AnalogueReading struct {
	InternalBatteryVoltage int
	Temperature            int
	LastGSMCQ              int
	LoadedVoltage          int
}

type TripTypeReading struct {
	Tt   int
	Trim int
}

var SerNoToName = map[int]string{
	810095: "rueger",
	810243: "charlie",
}
