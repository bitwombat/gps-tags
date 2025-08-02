DROP TABLE IF EXISTS `gpsReading`;

DROP TABLE IF EXISTS `GPIOReading`;

DROP TABLE IF EXISTS `AnalogueReading`;

DROP TABLE IF EXISTS `TripTypeReading`;

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
        `ID` INT PRIMARY KEY AUTO_INCREMENT,
        `TxID` VARCHAR(100) NOT NULL,
        `ProdID` INT,
        `DeviceDateTime` DATETIME,
        `SeqNo` INT,
        `Reason` INT,
        CONSTRAINT `c_rcd` FOREIGN KEY (`TxID`) REFERENCES `tx` (`ID`)
    );

CREATE TABLE
    `gpsReading` (
        `ID` INT PRIMARY KEY AUTO_INCREMENT,
        `RecordID` INT NOT NULL,
        `Spd` INT,
        `SpdAcc` INT,
        `Head` INT,
        `GpsStat` INT,
        `GpsUTC` DATETIME,
        `Lat` FLOAT,
        `Long` FLOAT,
        `Alt` INT,
        `PosAcc` INT,
        `Pdop` INT,
        CONSTRAINT `c_gps` FOREIGN KEY (`RecordID`) REFERENCES `record` (`ID`)
    );

CREATE TABLE
    `gpioReading` (
        `ID` INT PRIMARY KEY AUTO_INCREMENT,
        `RecordID` INT NOT NULL,
        `DIn` INT,
        `DOut` INT,
        `DevStat` INT,
        CONSTRAINT `c_gpio` FOREIGN KEY (`RecordID`) REFERENCES `record` (`ID`)
    );

CREATE TABLE
    `analogueReading` (
        `ID` INT PRIMARY KEY AUTO_INCREMENT,
        `RecordID` INT NOT NULL,
        `InternalBatteryVoltage` INT,
        `Temperature` INT,
        `LastGSMCQ` INT,
        `LoadedVoltage` INT,
        CONSTRAINT `c_analogue` FOREIGN KEY (`RecordID`) REFERENCES `record` (`ID`)
    );

CREATE TABLE
    `tripTypeReading` (
        `ID` INT PRIMARY KEY AUTO_INCREMENT,
        `RecordID` INT NOT NULL,
        `Tt` INT,
        `Trim` INT,
        CONSTRAINT `c_triptype` FOREIGN KEY (`RecordID`) REFERENCES `record` (`ID`)
    );