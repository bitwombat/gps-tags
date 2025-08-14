DROP TABLE IF EXISTS gpsReading;
DROP TABLE IF EXISTS gpioReading;
DROP TABLE IF EXISTS analogueReading;
DROP TABLE IF EXISTS tripTypeReading;
DROP TABLE IF EXISTS record;
DROP TABLE IF EXISTS tx;

CREATE TABLE tx (
    ID VARCHAR(100) PRIMARY KEY,
    ProdID INT,
    Fw VARCHAR(100),
    SerNo INT,
    Imei VARCHAR(100),
    Iccid VARCHAR(100),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    json TEXT
);

CREATE TABLE record (
    ID VARCHAR(100) PRIMARY KEY,
    TxID VARCHAR(100) NOT NULL,
    DeviceDateTime DATETIME,
    SeqNo INT,
    Reason INT,
    FOREIGN KEY (TxID) REFERENCES tx(ID)
);

CREATE TABLE gpsReading (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    RecordID VARCHAR(100) NOT NULL,
    Spd INT,
    SpdAcc INT,
    Head INT,
    GpsStat INT,
    GpsUTC DATETIME,
    Lat DOUBLE,
    Lng DOUBLE,
    Alt INT,
    PosAcc INT,
    Pdop INT,
    FOREIGN KEY (RecordID) REFERENCES record(ID)
);

CREATE TABLE gpioReading (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    RecordID VARCHAR(100) NOT NULL,
    DIn INT,
    DOut INT,
    DevStat INT,
    FOREIGN KEY (RecordID) REFERENCES record(ID)
);

CREATE TABLE analogueReading (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    RecordID VARCHAR(100) NOT NULL,
    InternalBatteryVoltage INT,
    Temperature INT,
    LastGSMCQ INT,
    LoadedVoltage INT,
    FOREIGN KEY (RecordID) REFERENCES record(ID)
);

CREATE TABLE tripTypeReading (
    ID INTEGER PRIMARY KEY AUTOINCREMENT,
    RecordID VARCHAR(100) NOT NULL,
    Tt INT,
    Trim INT,
    FOREIGN KEY (RecordID) REFERENCES record(ID)
);

