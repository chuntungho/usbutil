package usbutil

import (
	"fmt"
	"testing"
)

func TestListUsbStorage(t *testing.T) {
	storage, err := ListUsbStorage()
	if err == nil {
		for _, usbStorage := range storage {
			fmt.Println("storage", usbStorage)
		}
	}
}

func TestFormatHex(t *testing.T) {
	s := formatHex("0x121d")
	fmt.Println(s)
}
