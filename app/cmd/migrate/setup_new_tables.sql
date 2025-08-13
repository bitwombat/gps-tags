DROP TABLE IF EXISTS `gpsReading`;

DROP TABLE IF EXISTS `gpioReading`;

DROP TABLE IF EXISTS `analogueReading`;

DROP TABLE IF EXISTS `tripTypeReading`;

DROP TABLE IF EXISTS `record`;

DROP TABLE IF EXISTS `tx`;

CREATE TABLE
    `tx` (
        `ID` VARCHAR(100) PRIMARY KEY,
        `ProdID` INT,
        `Fw` VARCHAR(100),
        `SerNo` INT,
        `Imei` VARCHAR(100),
        `Iccid` VARCHAR(100),
        `created_at` DATETIME DEFAULT CURRENT_TIMESTAMP,
        `json` TEXT
    );

CREATE TABLE
    `record` (
        `ID` VARCHAR(100) PRIMARY KEY,
        `TxID` VARCHAR(100) NOT NULL,
        `DeviceDateTime` DATETIME,
        `SeqNo` INT,
        `Reason` INT,
        CONSTRAINT `c_rcd` FOREIGN KEY (`TxID`) REFERENCES `tx` (`ID`)
    );

CREATE TABLE
    `gpsReading` (
        `ID` INT PRIMARY KEY AUTO_INCREMENT,
        `RecordID` VARCHAR(100) NOT NULL,
        `Spd` INT,
        `SpdAcc` INT,
        `Head` INT,
        `GpsStat` INT,
        `GpsUTC` DATETIME,
        `Lat` DOUBLE,
        `Lng` DOUBLE,
        `Alt` INT,
        `PosAcc` INT,
        `Pdop` INT,
        CONSTRAINT `c_gps` FOREIGN KEY (`RecordID`) REFERENCES `record` (`ID`)
    );

CREATE TABLE
    `gpioReading` (
        `ID` INT PRIMARY KEY AUTO_INCREMENT,
        `RecordID` VARCHAR(100) NOT NULL,
        `DIn` INT,
        `DOut` INT,
        `DevStat` INT,
        CONSTRAINT `c_gpio` FOREIGN KEY (`RecordID`) REFERENCES `record` (`ID`)
    );

CREATE TABLE
    `analogueReading` (
        `ID` INT PRIMARY KEY AUTO_INCREMENT,
        `RecordID` VARCHAR(100) NOT NULL,
        `InternalBatteryVoltage` INT,
        `Temperature` INT,
        `LastGSMCQ` INT,
        `LoadedVoltage` INT,
        CONSTRAINT `c_analogue` FOREIGN KEY (`RecordID`) REFERENCES `record` (`ID`)
    );

CREATE TABLE
    `tripTypeReading` (
        `ID` INT PRIMARY KEY AUTO_INCREMENT,
        `RecordID` VARCHAR(100) NOT NULL,
        `Tt` INT,
        `Trim` INT,
        CONSTRAINT `c_triptype` FOREIGN KEY (`RecordID`) REFERENCES `record` (`ID`)
    );
