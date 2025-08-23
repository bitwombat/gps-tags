package device

type TagTx struct {
	SerNo   int      `json:"SerNo"`
	Imei    string   `json:"IMEI"`
	Iccid   string   `json:"ICCID"`
	ProdID  int      `json:"ProdId"`
	Fw      string   `json:"FW"`
	Records []Record `json:"Records"`
}

type Record struct {
	SeqNo   int     `json:"SeqNo"`
	Reason  int     `json:"Reason"`
	DateUTC string  `json:"DateUTC"`
	Fields  []Field `json:"Fields"`
}

type Field struct {
	GpsUTC       string       `json:"GpsUTC,omitempty"`
	Lat          float64      `json:"Lat,omitempty"`
	Long         float64      `json:"Long,omitempty"`
	Alt          int          `json:"Alt,omitempty"`
	Spd          int          `json:"Spd,omitempty"`
	SpdAcc       int          `json:"SpdAcc,omitempty"`
	Head         int          `json:"Head,omitempty"`
	Pdop         int          `json:"PDOP,omitempty"`
	PosAcc       int          `json:"PosAcc,omitempty"`
	GpsStat      int          `json:"GpsStat,omitempty"`
	FType        int          `json:"FType"`
	DIn          int          `json:"DIn,omitempty"`
	DOut         int          `json:"DOut,omitempty"`
	DevStat      int          `json:"DevStat,omitempty"`
	AnalogueData AnalogueData `json:"AnalogueData,omitempty"`
}

type AnalogueData struct {
	Num1 int `json:"1"`
	Num3 int `json:"3"`
	Num4 int `json:"4"`
	Num5 int `json:"5"`
}

const AnalogueDataFType = 6

var IdToName = map[float64]string{
	810095: "rueger",
	810243: "charlie",
}
