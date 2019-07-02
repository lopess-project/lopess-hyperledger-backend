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

// Define the sensor data structure, with 4 properties.  Structure tags are used by encoding/json library
type SensorData struct {
	DeviceId  string  `json:"deviceId"`
	Timestamp string  `json:"timestamp"`
	Pm10      float32 `json:"pm10"`
	Pm25      float32 `json:"pm25"`
}

// Define the devince info structure, with 3 properties.  Structure tags are used by encoding/json library
type DeviceInfo struct {
	PublicKey      string `json:"pubKey"`
	EncodingScheme int    `json:"code"`
	Owner          string `json:"owner"`
	ValidationFlag bool   `json:"valid"`
}

/*
 * The Init method is called when the Smart Contract "fabcar" is instantiated by the blockchain network
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
	}
	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	devices := []DeviceInfo{
		DeviceInfo{PublicKey: "+sBjAJFDE5c1iML63cPh6MPRUDPQgS3Kcwcvg+NkPFI", EncodingScheme: 0, Owner: "org1", ValidationFlag: true},
		DeviceInfo{PublicKey: "39b23651ccc0b2c21c40161d54cad2ae663e48c6ee8124da188378ce1fa60d69", EncodingScheme: 0, Owner: "org2", ValidationFlag: true},
		DeviceInfo{PublicKey: "d912e6d6e200f8e0db86ec7d173604b30a1d87188df052a4c7dd2ba7fc0b4f37", EncodingScheme: 0, Owner: "org1", ValidationFlag: true},
		DeviceInfo{PublicKey: "9a39c7a3e10f23a94c28aa8688a8e0dd0ce9e86f37675670c64d0f529e147a40", EncodingScheme: 0, Owner: "org2", ValidationFlag: true},
		DeviceInfo{PublicKey: "202fb11920d8450cd99d3761ec3e5f139474a5694f7031108f45bcf093881b14", EncodingScheme: 0, Owner: "org1", ValidationFlag: true},
	}
	i := 0
	for i < len(devices) {
		fmt.Println("i is ", i)
		deviceAsBytes, _ := json.Marshal(devices[i])
		APIstub.PutState("Device%s"+strconv.Itoa(i+1), deviceAsBytes)
		fmt.Println("Added", devices[i])
		i = i + 1
	}
	return shim.Success(nil)
}

func (s *SmartContract) registerDevice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 3 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}

	startKey := "Device1"
	endKey := "Device9999"

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
	vflag, err := strconv.ParseBool(args[2])

	var data = DeviceInfo{PublicKey: args[0], Owner: args[1], ValidationFlag: vflag}
	dataAsBytes, _ := json.Marshal(data)
	APIstub.PutState("Device%s"+strconv.Itoa(i), dataAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) revokeDevice(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {
	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 2")
	}
	id := "Device%s" + args[0]
	deviceAsBytes, _ := APIstub.GetState(id)
	device := DeviceInfo{}
	json.Unmarshal(deviceAsBytes, &device)
	device.ValidationFlag = false
	deviceAsBytes, _ = json.Marshal(device)
	APIstub.PutState(id, deviceAsBytes)
	return shim.Success(nil)
}

func (s *SmartContract) registerMeasurement(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 1 {
		return shim.Error("Incorrect number of arguments. Expecting 1.")
	}

	b, err := base64.StdEncoding.DecodeString(args[0])
	if err != nil {
		return shim.Error("Decoding from base64 to bytes failed.")
	}
	// parse input string based on decoding scheme

	if b[0] != 170 {
		return shim.Error("Incorrect header format. Expecting start byte 10101010.")
	}
	// get decoding scheme from device Id and decode accordingly
	deviceId := binary.BigEndian.Uint16(b[1:3])
	deviceAsBytes, _ := APIstub.GetState(strconv.Itoa(int(deviceId)))
	device := DeviceInfo{}
	json.Unmarshal(deviceAsBytes, &device)
	if device.ValidationFlag == false {
		return shim.Error("Device has been revoked. Transaction aborted")
	} else {
		data := SensorData{}
		txId := ""
		enc := device.EncodingScheme
		switch enc {
		case 0:
			data, txId = decodeMessageWithDefaultEncodingScheme(b, device, deviceId)
		case 1:
			data, txId = decodeMessageWithAlternateEncodingScheme(b, device, deviceId)
		default:
			data, txId = decodeMessageWithDefaultEncodingScheme(b, device, deviceId)
		}
		if (data == SensorData{} || txId == "") {
			return shim.Error("Error occured while decoding the message. Either decoding from hex to bytes threw the error or the signature is not valid.")
		}
		dataAsBytes, _ := json.Marshal(data)
		APIstub.PutState(txId, dataAsBytes)
	}
	return shim.Success(nil)
}

func decodeMessageWithDefaultEncodingScheme(b []byte, device DeviceInfo, deviceId uint16) (SensorData, string) {
	/* Default Encoding:
	** Byte 1:		Header: 10101010
	** Byte 2-3:	Device Id: (1-65535)
	** Byte 4-19:	UUID of the transaction
	** Byte 20:		LowByte Pm10
	** Byte 21:		HighByte Pm10
	** Byte 22:		LowByte Pm25
	** Byte 23:		HighByte Pm25
	** Byte 24-87:	Signature
	 */
	pubKeyFromDevice, err := base64.RawStdEncoding.DecodeString(device.PublicKey)
	if err != nil {
		return SensorData{}, "encoding failure..."
	}
	msg := []byte(b[:23])
	signature := []byte(b[23:87])
	verification := ed25519.Verify(pubKeyFromDevice, msg, signature)
	if !verification {
		return SensorData{}, ""
	}
	uuid := []byte(b[3:19])
	txId := hex.EncodeToString(uuid)
	deviceIdStr := "Device" + strconv.Itoa(int(deviceId))
	pm10 := calculatePMValueFromBytes(b[19], b[20])
	pm25 := calculatePMValueFromBytes(b[21], b[22])
	var data = SensorData{DeviceId: deviceIdStr, Timestamp: "00:00:00", Pm10: pm10, Pm25: pm25}
	return data, txId
}

func decodeMessageWithAlternateEncodingScheme(b []byte, device DeviceInfo, deviceId uint16) (SensorData, string) {
	//todo
	return SensorData{}, ""
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
		if strings.HasPrefix(queryResponse.Key, "Device") == false {
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

	fmt.Printf("- queryAllData:\n%s\n", buffer.String())

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

// expects 4 byte input
func convertMillisToTime(b []byte) time.Time {
	m := binary.BigEndian.Uint32(b)
	return time.Unix(0, int64(m)*int64(time.Millisecond))
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new  Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
