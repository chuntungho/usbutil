package usbutil

import (
	"bufio"
	"bytes"
	"encoding/json"
	"os/exec"
	"strconv"
	"strings"
)

type BlockDevice struct {
	Name        string
	Label       string
	FileSystem  string   `json:"fstype"`
	FsVersion   string   `json:"fsver"`
	Avail       string   `json:"fsavail"`
	Kind        string   `json:"type"`
	MountPoints []string `json:"mountpoints"`
	Children    []BlockDevice
}

// ListUsbStorage list usb storage devices with volumes
func ListUsbStorage() ([]UsbStorage, error) {
	out, err := exec.Command("lsblk", "-p", "-f", "-b", "--json").Output()
	if err != nil {
		return nil, err
	}

	var result map[string][]BlockDevice
	err = json.Unmarshal(out, &result)
	if err != nil {
		return nil, err
	}

	var storages []UsbStorage
	if devices, ok := result["blockdevices"]; ok {
		for _, device := range devices {
			usbStorage := getUsbStorage(device.Name)
			if usbStorage != nil {
				var volumes []Volume
				for _, part := range device.Children {
					freeSpace, _ := strconv.Atoi(part.Avail)
					volumes = append(volumes, Volume{
						Name:       part.Name,
						Label:      part.Label,
						FileSystem: part.FileSystem,
						FreeSpace:  uint64(freeSpace),
						MountPoint: part.MountPoints[0],
					})
				}
				usbStorage.Volumes = volumes

				storages = append(storages, *usbStorage)
			}
		}
	}

	return storages, nil
}

func getUsbStorage(device string) *UsbStorage {
	b, err := exec.Command("udevadm", "info", "-q", "property", "-n", device).Output()
	if err != nil {
		return nil
	}

	usbStorage, isUsb := UsbStorage{}, false
	scanner := bufio.NewScanner(bytes.NewReader(b))
	for scanner.Scan() {
		line := scanner.Text()
		s := strings.Split(line, "=")
		switch s[0] {
		case "ID_VENDOR_ID":
			usbStorage.VendorID = formatHex(s[1])
		case "ID_VENDOR_ENC":
			usbStorage.Vendor = strings.Trim(strings.Replace(s[1], "\\x20", " ", -1), " ")
		case "ID_MODEL_ID":
			usbStorage.ProductID = formatHex(s[1])
		case "ID_MODEL_ENC":
			usbStorage.Product = strings.Trim(strings.Replace(s[1], "\\x20", " ", -1), " ")
		case "ID_SERIAL_SHORT":
			usbStorage.Serial = s[1]
		case "ID_USB_DRIVER":
			isUsb = (s[1] == "usb-storage")
		}
	}

	if isUsb {
		return &usbStorage
	} else {
		return nil
	}
}
