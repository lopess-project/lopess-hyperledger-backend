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
 * 4 utility libraries for formatting, handling bytes, reading and writing JSON, and string manipulation
 * 2 specific Hyperledger Fabric specific libraries for Smart Contracts
 */
import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/hyperledger/fabric/core/chaincode/shim"
	sc "github.com/hyperledger/fabric/protos/peer"
)

// Define the Smart Contract structure
type SmartContract struct {
}

type WindDirection int

const (
	North WindDirection = iota + 1
	North_West
	North_East
	South
	South_West
	South_East
	East
	West
)

func (direction WindDirection) String() string {
	names := [...]string{
		"North",
		"North_West",
		"North_East",
		"South",
		"South_West",
		"South_East",
		"East",
		"West"}

	if direction < North || direction > West {
		return "Unknown"
	}

	return names[direction]

}

// Define the car structure, with 4 properties.  Structure tags are used by encoding/json library
type SensorData struct {
	Id            string        `json:"id"`
	Timestamp     time.Time     `json:"timestamp"`
	Windstrength  float32       `json:"windstrength"`
	Winddirection WindDirection `json:"winddirection"`
	Finedustvalue float32       `json:"finedustvalue"`
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
	if function == "updateData" {
		return s.updateData(APIstub, args)
	} else if function == "queryData" {
		return s.queryData(APIstub)
	} else if function == "initLedger" {
		return s.initLedger(APIstub)
	}

	return shim.Error("Invalid Smart Contract function name.")
}

func (s *SmartContract) initLedger(APIstub shim.ChaincodeStubInterface) sc.Response {
	return shim.Success(nil)
}

func (s *SmartContract) updateData(APIstub shim.ChaincodeStubInterface, args []string) sc.Response {

	if len(args) != 4 {
		return shim.Error("Incorrect number of arguments. Expecting 4")
	}
	id := APIstub.GetTxID()
	timestamp, err := convertStringToTime(args[0])
	windstrength, err1 := strconv.ParseFloat(args[1], 32)
	winddirection, err2 := strconv.ParseInt(args[2], 10, 32)
	finedust, err3 := strconv.ParseFloat(args[3], 32)

	if err != nil || err1 != nil || err2 != nil || err3 != nil {
		return shim.Error("Error during parsing of string arguments. Check format of input")
	}

	var data = SensorData{Id: id, Timestamp: timestamp, Windstrength: float32(windstrength), Winddirection: WindDirection(int(winddirection)), Finedustvalue: float32(finedust)}
	dataAsBytes, _ := json.Marshal(data)
	APIstub.PutState(id, dataAsBytes)

	return shim.Success(nil)
}

func (s *SmartContract) queryData(APIstub shim.ChaincodeStubInterface) sc.Response {

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

	fmt.Printf("- queryAllData:\n%s\n", buffer.String())

	return shim.Success(buffer.Bytes())
}

func convertStringToTime(timestamp string) (time.Time, error) {
	layout := "2006-01-02T15:04:05"
	t, err := time.Parse(layout, timestamp)

	return t, err
}

// The main function is only relevant in unit test mode. Only included here for completeness.
func main() {

	// Create a new Smart Contract
	err := shim.Start(new(SmartContract))
	if err != nil {
		fmt.Printf("Error creating new Smart Contract: %s", err)
	}
}
