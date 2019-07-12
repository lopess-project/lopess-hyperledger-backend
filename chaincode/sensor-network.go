/*
 * Licensed to the Apache Software Foundation (ASF) under one
 * or more contributor license agreements.  See the NOTICE file
 * distributed with this work for additional information
 * regarding copyright ownership.  The ASF licenses this file
 * to you under the Apache License, Version 2.0 (the
 * "License"); you may not use this file except in compliance
 * with the License.  You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing,
 * software distributed under the License is distributed on an
 * "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
 * KIND, either express or implied.  See the License for the
 * specific language governing permissions and limitations
 * under the License.
 */

/*
 * The sample smart contract for documentation topic:
 * Writing Your First Blockchain Application
 */

package main

/* Imports
 * 5 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 1 library for verifying eddsa signatures
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
	"golang.org/x/crypto/ed25519"
)

// Define the Smart Contract structure
type SmartContract struct {
}

// Define the sensor data structure, with 8 properties.  Structure tags are used by encoding/json library
type SensorData struct {
	DeviceId   string    `json:"deviceId"`
	Pm10       float32   `json:"pm10"`
	Pm25       float32   `json:"pm25"`
	Temp       float32   `json:"temp"`
	Humidity   float32   `json:"humidity"`
	Timestamp  time.Time `json:"timestamp"`
	Longtitude string    `json:"longtitude"`
	Latitude   string    `json:"latitude"`
}

// Define the devince info structure, with 4 properties.  Structure tags are used by encoding/json library
type DeviceInfo struct {
	PublicKey      string `json:"pubKey"`
	EncodingScheme int    `json:"code"`
	Owner          string `json:"owner"`
	ValidationFlag bool   `json:"valid"`
}

/*
 * The Init method is called when the Smart Contract is instantiated by the blockchain network
 * Best practice is to have any Ledger initialization in separate function -- see initLedger()
 */
func (s *SmartContract) Init(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

/*
 * The Invoke method is called as a result of an application request to run the Smart Contract "fabcar"
 * The calling application program has also specified the particular smart contract function to be called, with arguments
 */
func (s *SmartContract) Invoke(APIstub shim.ChaincodeStubInterface) sc.Response {

	// Retrieve the requested Smart Contract function and arguments
	function, args := APIstub.GetFunctionAndParameters()
	// Route to the appropriate handler function to interact with the ledger appropriately
	if function == "registerDevice" {
		return s.registerDevice(APIstub, args)
	} else if function == "revokeDevice" {
		return s.revokeDevice(APIstub, args)
	} else if function == "registerMeasurement" {
		return s.registerMeasurement(APIstub, args)
	} else if function == "getMeasurementRecords" {
		return s.getMeasurementRecords(APIstub)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	} else if function == "getDeviceRecords" {
		return s.getDeviceRecords(APIstub)
	} else if function == "getAllRecords" {
		return s.getAllRecords(APIstub)
	}
	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	devices := []DeviceInfo{
		DeviceInfo{PublicKey: "+sBjAJFDE5c1iML63cPh6MPRUDPQgS3Kcwcvg+NkPFI", EncodingScheme: 0, Owner: "org1", ValidationFlag: true},
		DeviceInfo{PublicKey: "RakaJDXqkmm0YzwKxTo4BVVko5T/7oElNdP2FGrUHu8", EncodingScheme: 0, Owner: "org2", ValidationFlag: true},
		DeviceInfo{PublicKey: "IQ1BO5vN3mcdQz6ZyV1f77uMJnpbJOL1IMqNUiJENeU", EncodingScheme: 0, Owner: "org1", ValidationFlag: true},
		DeviceInfo{PublicKey: "cOGzNiLH0e2C7WstGQfZk3CRdDSwR3yt58OeTc7f+V0", EncodingScheme: 0, Owner: "org2", ValidationFlag: true},
		DeviceInfo{PublicKey: "CQatsesQKp+qRTQPsAVTQdg6JBDsIIp9iaCgsWPxPUo", EncodingScheme: 0, Owner: "org1", ValidationFlag: true},
	}
	i := 0
	for i < len(devices) {
		fmt.Println("i is ", i)
		deviceAsBytes, _ := json.Marshal(devices[i])
		deviceIdAsString := "DEVICE" + strconv.Itoa(i+1)
		APIstub.PutState(deviceIdAsString, deviceAsBytes)
		fmt.Println("Added", devices[i])
		i = i + 1
	}
	return shim.Success(nil)
}

func (s *SmartContract) registerDevice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	startKey := "DEVICE1"
	endKey := "DEVICE9999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)

	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()
	i := 1
	for resultsIterator.HasNext() {
		resultsIterator.Next()
		i = i + 1
	}
	vflag, err := strconv.ParseBool(args[3])
	scheme, err := strconv.Atoi(args[1])

	var data = DeviceInfo{PublicKey: args[0], EncodingScheme: scheme, Owner: args[2], ValidationFlag: vflag}
	dataAsBytes, _ := json.Marshal(data)
	APIstub.PutState("DEVICE%s"+strconv.Itoa(i), dataAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) revokeDevice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	id := "DEVICE%s" + args[0]
	deviceAsBytes, _ := APIstub.GetState(id)
	device := DeviceInfo{}
	json.Unmarshal(deviceAsBytes, &device)
	device.ValidationFlag = false
	deviceAsBytes, _ = json.Marshal(device)
	APIstub.PutState(id, deviceAsBytes)
	return shim.Success(nil)
}

func (s *SmartContract) registerMeasurement(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 2 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}

	b, err := base64.StdEncoding.DecodeString(args[0])

	if err != nil {
		return shim.Error("Decoding from base64 to bytes failed.")
	}

	b2, err1 := base64.StdEncoding.DecodeString(args[1])

	if err1 != nil {
		return shim.Error("Decoding from base64 to bytes failed.")
	}
	// parse input string based on decoding scheme

	if b[0] != 170 {
		return shim.Error("Incorrect header format. Expecting start byte 10101010.")
	}
	// get decoding scheme from device Id and decode accordingly
	deviceId := binary.BigEndian.Uint16(b[1:3])
	deviceIdAsString := "DEVICE" + strconv.Itoa(int(deviceId))
	deviceAsBytes, _ := APIstub.GetState(deviceIdAsString)
	device := DeviceInfo{}
	json.Unmarshal(deviceAsBytes, &device)
	if device.ValidationFlag == false {
		return shim.Error("Device has been revoked. Transaction aborted. DeviceId was "+deviceIdAsString)
	} else {
		data := SensorData{}
		txId := ""
		enc := device.EncodingScheme
		switch enc {
		case 0:
			data, txId = decodeMessageWithDefaultEncodingScheme(b, b2, device, deviceId)
		case 1:
			data, txId = decodeMessageWithAlternateEncodingScheme(b, b2, device, deviceId)
		default:
			data, txId = decodeMessageWithDefaultEncodingScheme(b, b2, device, deviceId)
		}
		if (data == SensorData{} || txId == "") {
			return shim.Error("Error occured while decoding the message. Either decoding from hex to bytes threw the error or the signature is not valid.")
		}
		dataAsBytes, _ := json.Marshal(data)
		APIstub.PutState(txId, dataAsBytes)
	}
	return shim.Success(nil)
}

func decodeMessageWithDefaultEncodingScheme(b, b2 []byte, device DeviceInfo, deviceId uint16) (SensorData, string) {
	/* Default Encoding: (Byte Array starts counting at posistion 0)
	** Byte 1:		Header: 10101010
	** Byte 2-3:	Device Id: (1-65535)
	** Byte 4-19:	UUID of the transaction
	** Byte 20:		LowByte Pm10
	** Byte 21:		HighByte Pm10
	** Byte 22:		LowByte Pm25
	** Byte 23:		HighByte Pm25
	** Byte 24:		LowByte Humidity
	** Byte 25:		HighByte Humidity
	** Byte 26:		LowByte Temp
	** Byte 27:		HighByte Temp
	** Byte 28-33:	Timestamp hh:mm:ss
	** Byte 34-44:	Latitude
	** Byte 45-56:  Longtitude
	** Byte 57-120:	Signature
	 */
	pubKeyFromDevice, err := base64.RawStdEncoding.DecodeString(device.PublicKey)
	if err != nil {
		return SensorData{}, "encoding failure..."
	}
	msg := []byte(b[:56])
	signature := b2
	verification := ed25519.Verify(pubKeyFromDevice, msg, signature)
	if !verification {
		return SensorData{}, ""
	}
	uuid := []byte(b[3:19])
	txId := hex.EncodeToString(uuid)
	deviceIdStr := "DEVICE" + strconv.Itoa(int(deviceId))
	pm10 := calculatePMValueFromBytes(b[19], b[20])
	pm25 := calculatePMValueFromBytes(b[21], b[22])
	humidity := calculateHumidityFromBytes(b[23], b[24])
	temp := calculateTempFromBytes(b[25], b[26])
	timestamp := convertTimestampToDate(b[27:33], time.Now().UTC())
	latitude := calculateLatitudeFromCharBytes(b[33:44])
	longtitude := calculateLongtitudeFromCharBytes(b[44:56])
	var data = SensorData{DeviceId: deviceIdStr, Timestamp: timestamp, Pm10: pm10, Pm25: pm25, Humidity: humidity, Temp: temp, Latitude: latitude, Longtitude: longtitude}
	return data, txId
}

func decodeMessageWithAlternateEncodingScheme(b, b2 []byte, device DeviceInfo, deviceId uint16) (SensorData, string) {
	//todo
	return SensorData{}, ""
}

func (s *SmartContract) getAllRecords(APIstub shim.ChaincodeStubInterface) sc.Response {
	resultsIterator, err := APIstub.GetStateByRange("","")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		fmt.Printf("Retrieved Key: %s\n",queryResponse.Key)
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		buffer.WriteString("{\"Key\":")
		buffer.WriteString("\"")
		buffer.WriteString(queryResponse.Key)
		buffer.WriteString("\"")

		buffer.WriteString(", \"Record\":")
		// Record is a JSON object, so we write as-is
		buffer.WriteString(string(queryResponse.Value))
		buffer.WriteString("}")
		bArrayMemberAlreadyWritten = true
	}
	buffer.WriteString("]")

	fmt.Printf("- queryAllRecords:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) getMeasurementRecords(APIstub shim.ChaincodeStubInterface) sc.Response {

	resultsIterator, err := APIstub.GetStateByRange("", "")
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		if strings.HasPrefix(queryResponse.Key, "DEVICE") == false {
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Record\":")
			// Record is a JSON object, so we write as-is
			buffer.WriteString(string(queryResponse.Value))
			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		}
	}
	buffer.WriteString("]")

	fmt.Printf("- queryMeasurementData:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func (s *SmartContract) getDeviceRecords(APIstub shim.ChaincodeStubInterface) sc.Response {

	startKey := "DEVICE1"
	endKey := "DEVICE999"

	resultsIterator, err := APIstub.GetStateByRange(startKey, endKey)
	if err != nil {
		return shim.Error(err.Error())
	}
	defer resultsIterator.Close()

	// buffer is a JSON array containing QueryResults
	var buffer bytes.Buffer
	buffer.WriteString("[")

	bArrayMemberAlreadyWritten := false
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return shim.Error(err.Error())
		}
		// Add a comma before array members, suppress it for the first array member
		if bArrayMemberAlreadyWritten == true {
			buffer.WriteString(",")
		}
		if strings.HasPrefix(queryResponse.Key, "DEVICE") == true {
			buffer.WriteString("{\"Key\":")
			buffer.WriteString("\"")
			buffer.WriteString(queryResponse.Key)
			buffer.WriteString("\"")

			buffer.WriteString(", \"Record\":")
			// Record is a JSON object, so we write as-is
			buffer.WriteString(string(queryResponse.Value))
			buffer.WriteString("}")
			bArrayMemberAlreadyWritten = true
		}
	}
	buffer.WriteString("]")

	fmt.Printf("- queryDeviceData:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

// from https://cdn-reichelt.de/documents/datenblatt/X200/SDS011-DATASHEET.pdf
// expects two bytes as input: first is the low byte, second the high byte
func calculatePMValueFromBytes(b1, b2 byte) float32 {
	lowByte := binary.BigEndian.Uint16([]byte{0, b1})
	highByte := binary.BigEndian.Uint16([]byte{b2, 0})
	if lowByte > 9999 {
		lowByte = 9999
	}
	if highByte > 9999 {
		highByte = 9999
	}
	f := float32(highByte+lowByte) / 10.0
	return f
}

func calculateTempFromBytes(b1, b2 byte) float32 {
	b2value := b2 & 0x7F
	lowByte := binary.BigEndian.Uint16([]byte{0, b1})
	highByte := binary.BigEndian.Uint16([]byte{b2value, 0})

	t := float32(highByte+lowByte) / 10.0
	b2check := b2 & 0x80
	if b2check != 0 {
		t = t * (-1)
	}
	return t
}

func calculateHumidityFromBytes(b1, b2 byte) float32 {
	lowByte := binary.BigEndian.Uint16([]byte{0, b1})
	highByte := binary.BigEndian.Uint16([]byte{b2, 0})
	h := float32(highByte+lowByte) / 10.0
	return h
}

func calculateLatitudeFromCharBytes(b []byte) string {
	degree := string(b[:3])
	hours := string(b[3:5])
	seconds := string(b[5:10])
	orientation := string(b[10])

	var str strings.Builder
	str.WriteString(degree)
	str.WriteString("°")
	str.WriteString(hours)
	str.WriteString("'")
	str.WriteString(seconds)
	str.WriteString("\"")
	str.WriteString(orientation)

	return str.String()
}

func calculateLongtitudeFromCharBytes(b []byte) string {
	degree := string(b[:4])
	hours := string(b[4:6])
	seconds := string(b[6:11])
	orientation := string(b[11])

	var str strings.Builder
	str.WriteString(degree)
	str.WriteString("°")
	str.WriteString(hours)
	str.WriteString("'")
	str.WriteString(seconds)
	str.WriteString("\"")
	str.WriteString(orientation)

	return str.String()
}

// expects 6 byte input + current_time := time.Now().UTC()
func convertTimestampToDate(b []byte, current_time time.Time) time.Time {
	hh := string(b[:2])
	mm := string(b[2:4])
	ss := string(b[4:6])

	hours, err := strconv.Atoi(hh)
	minutes, err := strconv.Atoi(mm)
	seconds, err := strconv.Atoi(ss)

	timeNow := current_time.Format("15:04:05")
	dateNow := current_time.Format("2006-01-02")

	//get date --- assuming that latency is < 1h
	if hh == "23" {
		if timeNow[:2] != hh {
			newDate := current_time.AddDate(0, 0, -1)
			dateNow = newDate.Format("2006-01-02")
		}
	}

	year, err := strconv.Atoi(string(dateNow[:4]))
	month, err := strconv.Atoi(string(dateNow[5:7]))
	day, err := strconv.Atoi(string(dateNow[8:10]))

	if err != nil {
		fmt.Println("Something went wrong when converting time to Date.")
	}

	return time.Date(year, time.Month(month), day, hours, minutes, seconds, 0, time.UTC)
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new  Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
