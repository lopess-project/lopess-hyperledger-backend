package main

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"testing"
	"time"
)

func TestPMCalculation(t *testing.T) {
	b := []byte{10, 2}
	pmValue := calculatePMValueFromBytes(b[0], b[1])
	if pmValue != 52.2 {
		t.Errorf("PM calculation was incorrect, got: %g, want: %g.", pmValue, 52.2)
	}
}

func TestTempCalculation(t *testing.T) {
	expectedResult := float32(35.1)
	b1 := byte(95)
	b2 := byte(1)
	output := calculateTempFromBytes(b1, b2)
	if output != expectedResult {
		t.Errorf("Temp calculation was incorrect, got: %g, want: %g.", output, expectedResult)
	}
}

func TestHumidityCalculation(t *testing.T) {
	expectedResult := float32(65.2)
	b1 := byte(140)
	b2 := byte(2)
	output := calculateHumidityFromBytes(b1, b2)
	if output != expectedResult {
		t.Errorf("Humidity calculation was incorrect, got: %g, want: %g.", output, expectedResult)
	}
}

func TestLatitudeConversion(t *testing.T) {
	expectedResult := "049°00'33624\"N"
	b := []byte{'0', '4', '9', '0', '0', '3', '3', '6', '2', '4', 'N'}
	output := calculateLatitudeFromCharBytes(b)
	if output != expectedResult {
		t.Errorf("Latitude conversion was incorrect, got: %s, want: %s", output, expectedResult)
	}
}

func TestLongtitudeConversion(t *testing.T) {
	expectedResult := "0008°25'31116\"E"
	b := []byte{'0', '0', '0', '8', '2', '5', '3', '1', '1', '1', '6', 'E'}
	output := calculateLongtitudeFromCharBytes(b)
	if output != expectedResult {
		t.Errorf("Longtitude conversion was incorrect, got: %s, want: %s", output, expectedResult)
	}
}

func TestTimeToDateConversion(t *testing.T) {
	expectedResult := "2019-07-06 19:43:00 +0200 CEST"
	b := []byte{'1', '7', '4', '3', '0', '0'}
	current_time, _ := time.Parse("2006-01-02 15:04:05", "2019-07-06 17:45:02")
	output := convertTimestampToDate(b, current_time)
	if output.Local().String() != expectedResult {
		t.Errorf("Time conversion was incorrect, got: %s, want: %s", output.Local().String(), expectedResult)
	}

	edgeCaseExpectation := "2019-07-06 23:59:59 +0000 UTC"
	edgeCase := []byte{'2', '3', '5', '9', '5', '9'}
	timeWithLatency, _ := time.Parse("2006-01-02 15:04:05", "2019-07-07 00:01:02")
	edgeCaseOutput := convertTimestampToDate(edgeCase, timeWithLatency)
	if edgeCaseOutput.UTC().String() != edgeCaseExpectation {
		t.Errorf("Time conversion for edge cas was incorrect, got: %s, want: %s", edgeCaseOutput.UTC().String(), expectedResult)
	}
}

func TestDecoding(t *testing.T) {
	/*** expected Values ***/
	expectedTimeResult := "2019-07-06 19:43:00 +0200 CEST"
	expectedLatitudeResult := "049°00'33624\"N"
	expectedLongtitudeResult := "0008°25'31116\"E"
	/*** test ***/
	encodedString := "qgABoAFwhAQ+AVjJqBFIAIQAAtcAOACmAq4AMDYwMDE1MDQ5MDA1OTU2ME4wMDA4MjU1MTY2MEU="
	sig := "3xRcKIUbrdehGRbcZOWfFd01z6n6yge3Zsnl9J/L2plU2HiAhAw2RQdG91nmEYsvN005oyf1wy1PaRH09v4oAQ=="
	signature, err := base64.StdEncoding.DecodeString(sig)
	input, err := base64.StdEncoding.DecodeString(encodedString)
	if err != nil {
		t.Errorf("base64 decoding failed.")
	}
	if input[0] != 170 {
		t.Errorf("header not correct.")
	}
	deviceId := binary.BigEndian.Uint16(input[1:3])
	if deviceId != 1 {
		t.Errorf("conversion from byte 2-3 to integer failed")
	}
	device := DeviceInfo{PublicKey: "pQBakw2oxXklWGruTdMVnbbNsNG+nsojdlusAiaRVLU", EncodingScheme: 0, Owner: "org1", ValidationFlag: true}
	uuidBytes := []byte{128, 23, 72, 1, 33, 112, 114, 72, 196, 96, 18, 136, 161, 84, 49, 63}
	expectedUUID := hex.EncodeToString(uuidBytes)
	data, txId := decodeMessageWithDefaultEncodingScheme(input, signature, device, 1)
	if txId == "" {
		t.Errorf("Signature Verification failed")
	}
	if txId != expectedUUID {
		t.Errorf("Decoded UUID was not correct, got: %s, want: %s", txId, expectedUUID)
	}
	if data.DeviceId != "Device1" {
		t.Errorf("Decoded device ID was not correct, got: %s, want: Device1", data.DeviceId)
	}
	if data.Pm10 != 5.4 {
		t.Errorf("Decoded Pm10 value was not correct, got: %g, want: 5.4", data.Pm10)
	}
	if data.Pm25 != 4.3 {
		t.Errorf("Decoded Pm10 value was not correct, got: %g, want: 4.3", data.Pm25)
	}
	if data.Humidity != 44.4 {
		t.Errorf("Decoded Humidity value was not correct, got: %g, want: 44.4", data.Humidity)
	}
	if data.Temp != 4.3 {
		t.Errorf("Decoded Temp value was not correct, got: %g, want: 4.3", data.Temp)
	}
	if data.Latitude != expectedLatitudeResult {
		t.Errorf("Decoded Latitude value was not correct, got: %s, want: %s", data.Latitude, expectedLatitudeResult)
	}
	if data.Longtitude != expectedLongtitudeResult {
		t.Errorf("Decoded Longtitude value was not correct, got: %s, want: %s", data.Longtitude, expectedLongtitudeResult)
	}
	if data.Timestamp.Local().String() != expectedTimeResult {
		t.Errorf("Decoded Timestamp value was not correct, got: %s, want: %s", data.Timestamp.Local().String(), expectedTimeResult)
	}

}
