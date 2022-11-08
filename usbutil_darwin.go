package usbutil

import (
	"bytes"
	"os/exec"
	"strings"

	"howett.net/plist"
)

// ListUsbStorage list usb storage devices with volumes
// Result stucture like below:
/* [{
	"_dataType": "SPUSBDataType",
	"_items": [{
		"host_controller": "AppleUSBXHCIWPT",
		"_items": [{
		"_name": "...",
		"manufacturer": "Kingston",
		"product_id": "0x...",
		"vendor_id": "0x... (Kingston)",
		"serial_num": "xxx",
		"Media": [{
			"_name": "xxx",
			"size": "1G",
			"size_in_bytes": 999,
			"volumes": [{"_name": "xx",
				"file_system": "MS-DOS FAT32",
				"mount_point": "/Volumes/UDISK",
				"size_in_bytes": 999
				"free_space_in_bytes": 999,
				}]
			}]
	}
]}]}]
*/
func ListUsbStorage() ([]UsbStorage, error) {
	b, err := exec.Command("system_profiler", "-xml", "SPUSBDataType").Output()
	if err != nil {
		return nil, err
	}

	d := plist.NewDecoder(bytes.NewReader(b))
	var payload []map[string]interface{}
	err = d.Decode(&payload)
	if err != nil {
		return nil, err
	}

	results, ok := payload[0]["_items"]
	if !ok {
		return nil, nil
	}

	var storages []UsbStorage
	for _, result := range results.([]interface{}) {
		devices, ok := result.(map[string]interface{})["_items"]
		if !ok {
			continue
		}
		
		for _, device := range devices.([]interface{}) {
			usbStorage := getUsbStorage(device)
			if usbStorage != nil {
				storages = append(storages, *usbStorage)
			}
		}
	}

	return storages, nil
}

func getUsbStorage(device interface{}) *UsbStorage {
	dev := device.(map[string]interface{})
	usbStorage := UsbStorage{
		VendorID:  formatHex(getString(dev, "vendor_id")),
		Vendor:    getString(dev, "manufacturer"),
		ProductID: formatHex(getString(dev, "product_id")),
		Product:   getString(dev, "_name"),
		Serial:    getString(dev, "serial_num"),
	}
	// vendor_id: 0x1212 (xxx)
	if strings.Contains(usbStorage.VendorID, " ") {
		usbStorage.VendorID = strings.Split(usbStorage.VendorID, " ")[0]
	}

	media, isUsb := dev["Media"]
	if !isUsb {
		return nil
	}

	var volumes []Volume
	for _, m := range media.([]interface{}) {
		volumes = append(volumes, getVolumes(m)...)
	}
	usbStorage.Volumes = volumes

	return &usbStorage
}

func getVolumes(m interface{}) []Volume {
	var volumes []Volume
	if vols, ok := m.(map[string]interface{})["volumes"]; ok {
		for _, vol := range vols.([]interface{}) {
			volMap := vol.(map[string]interface{})
			volumes = append(volumes, Volume{
				Name:       getString(volMap, "_name"),
				Label:      getString(volMap, "_name"),
				MountPoint: getString(volMap, "mount_point"),
				FileSystem: getString(volMap, "file_system"),
				Capacity:   volMap["size_in_bytes"].(uint64),
				FreeSpace:  volMap["size_in_bytes"].(uint64),
			})
		}
	}
	return volumes
}

func getString(dev map[string]interface{}, key string) string {
	if val, ok := dev[key]; ok {
		return val.(string)
	} else {
		return ""
	}
}
