package usbutil

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type UsbStorage struct {
	VendorID  string   // Device Vendor ID
	ProductID string   // Device Product ID
	Serial    string   // Serial Number
	Vendor    string   // Vendor String
	Product   string   // Product string
	Volumes   []Volume // Volumes
}

type Volume struct {
	Name       string
	Label      string
	MountPoint string
	FileSystem string // FAT32, FAT16, EXT4
	Capacity   uint64
	FreeSpace  uint64
}

func ReadFromFile(volumes []string) []UsbStorage {
	var payload UsbStorage
	var storages []UsbStorage

	for _, vol := range volumes {
		metaPath := filepath.Join(vol, ".device_meta")
		metaFile, err := os.Open(metaPath)
		if err == nil {
			err = json.NewDecoder(metaFile).Decode(&payload)
			if err == nil && payload.Product != "" {
				payload.Volumes = []Volume{{MountPoint: vol}}
				storages = append(storages, payload)
			}
		}
		_ = metaFile.Close()
	}

	return storages
}

func formatHex(hex string) string {
	return strings.ToUpper(strings.TrimPrefix(hex, "0x"))
}
