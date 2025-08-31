CREATE TABLE tx (
    ID TEXT PRIMARY KEY,
    ProdID INTEGER,
    Fw TEXT,
    SerNo INTEGER,
    Imei TEXT,
    Iccid TEXT,
    CreatedAt TEXT,
    json TEXT
);

CREATE TABLE record (
    ID TEXT PRIMARY KEY,
    TxID TEXT NOT NULL,
    DeviceDateTime TEXT,
    SeqNo INTEGER,
    Reason INTEGER,
    FOREIGN KEY (TxID) REFERENCES tx(ID)
);

CREATE TABLE gpsReading (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    RecordID TEXT NOT NULL,
    Spd INTEGER,
    SpdAcc INTEGER,
    Head INTEGER,
    GpsStat INTEGER,
    GpsUTC TEXT,
    Lat REAL,
    Lng REAL,
    Alt INTEGER,
    PosAcc INTEGER,
    Pdop INTEGER,
    FOREIGN KEY (RecordID) REFERENCES record(ID)
);

CREATE TABLE gpioReading (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    RecordID TEXT NOT NULL,
    DIn INTEGER,
    DOut INTEGER,
    DevStat INTEGER,
    FOREIGN KEY (RecordID) REFERENCES record(ID)
);

CREATE TABLE analogueReading (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    RecordID TEXT NOT NULL,
    InternalBatteryVoltage INTEGER,
    Temperature INTEGER,
    LastGSMCQ INTEGER,
    LoadedVoltage INTEGER,
    FOREIGN KEY (RecordID) REFERENCES record(ID)
);

CREATE TABLE tripTypeReading (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    RecordID TEXT NOT NULL,
    Tt INTEGER,
    Trim INTEGER,
    FOREIGN KEY (RecordID) REFERENCES record(ID)
);

CREATE INDEX tx_idx ON tx (createdAt);
CREATE INDEX rec_idx ON record (TxID);
CREATE INDEX gps_idx ON gpsReading (RecordID);
CREATE INDEX gpi_idx ON gpioReading (RecordID);
CREATE INDEX ana_idx ON analogueReading (RecordID);
CREATE INDEX tri_idx ON tripTypeReading (RecordID);

PRAGMA main.INTEGRITY_CHECK;

