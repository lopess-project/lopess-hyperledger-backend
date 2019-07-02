package main

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestPMCalculation(t *testing.T) {
	b := []byte{10, 2}
	pmValue := calculatePMValueFromBytes(b[0], b[1])
	if pmValue != 52.2 {
		t.Errorf("Sum was incorrect, got: %g, want: %g.", pmValue, 52.2)
	}
}

func TestDecoding(t *testing.T) {
	/*
	*** Expected UUID = 180242166196611978143232941166141113120
	*** Expected Pm10 = 2.9
	*** Expected Pm25 = 2.6
	 */
	encodedString := "qgABgBdIASFwckjEYBKIoVQxPzYAKwBUuZyFoRG+NuvdDpziAN5UWmwfCBPa+JrY94NFEG+4K4/624uP3jNrEOxFjYxTlYoNVyboJqE09i46tMP2LLMJAA=="
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
	data, txId := decodeMessageWithDefaultEncodingScheme(input, device, 1)
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

}
