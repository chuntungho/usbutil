package usbutil

import (
	"strings"

	"github.com/yusufpapurcu/wmi"
)

type Win32_USBControllerDevice struct {
	Antecedent string // ...Win32_USBController.DeviceID="PCI\\..."
	Dependent  string // ...Win32_PnPEntity.DeviceID="USBSTOR\\..." or "USB\\VID_xxxx&PID_xxx\\xxxx"
}

type Win32_DiskDrive struct {
	DeviceID string //  \\.\PHYSICALDRIVE2
	Model    string // Mass Storage Device USB Device
}

type Win32_DiskPartition struct {
	DeviceID    string // Disk #1, Partition #0
	PNPDeviceID string // 磁盘 #1，分区 #0
}

type Win32_LogicalDisk struct {
	DeviceID   string // F:
	Name       string // F:
	VolumeName string // Windows
	FileSystem string // FAT
	Size       uint64
	FreeSpace  uint64
}

// ListUsbStorage list usb storage devices with volumes
func ListUsbStorage() ([]UsbStorage, error) {
	var storages []UsbStorage
	storages = ReadFromFile(getUsbVolumes())
	if storages != nil {
		return storages, nil
	}

	var dev []Win32_USBControllerDevice
	query := wmi.CreateQuery(&dev, "")
	err := wmi.Query(query, &dev)
	if err != nil {
		return storages, nil
	}

	// USBSTOR dependent follows USB Device dependent with same Antecedent
	for i, v := range dev {
		pnpDevID := strings.Trim(strings.Split(v.Dependent, "=")[1], "\"")
		if strings.HasPrefix(pnpDevID, "USBSTOR") {
			if dev[i-1].Antecedent == v.Antecedent {
				usb := strings.Trim(strings.Split(dev[i-1].Dependent, "=")[1], "\"")
				storage := parseUsbDevice(usb, pnpDevID)
				volumes := getVolumes(pnpDevID)
				if volumes != nil {
					storage.Volumes = volumes

					// populate product name when only one volume
					if len(volumes) == 1 {
						var product string
						if volumes[0].Label == "" {
							product = volumes[0].MountPoint
						} else {
							product = volumes[0].Label + " - (" + volumes[0].MountPoint + ")"
						}
						storage.Product = product
					}

					storages = append(storages, storage)
				}
			}
		}
	}

	return storages, nil
}

func getUsbVolumes() []string {
	var volumes []string
	var disk []Win32_LogicalDisk
	// USB drive type = 2
	_ = wmi.Query(wmi.CreateQuery(&disk, "WHERE DriveType='2'"), &disk)
	for _, d := range disk {
		volumes = append(volumes, d.Name)
	}
	return volumes
}

// get volumes by PnP Device ID
func getVolumes(pnpDevID string) []Volume {
	var volumes []Volume
	var drive []Win32_DiskDrive
	_ = wmi.Query(wmi.CreateQuery(&drive, "WHERE PnPDeviceID='"+pnpDevID+"'"), &drive)
	if len(drive) == 0 {
		return nil
	}

	var par []Win32_DiskPartition
	_ = wmi.Query("ASSOCIATORS OF {Win32_DiskDrive.DeviceID='"+drive[0].DeviceID+"'} WHERE ResultClass = Win32_DiskPartition", &par)
	// each partitions
	for _, p := range par {
		var logic []Win32_LogicalDisk
		_ = wmi.Query("ASSOCIATORS OF {Win32_DiskPartition.DeviceID='"+p.DeviceID+"'} WHERE ResultClass = Win32_LogicalDisk", &logic)
		// only one volume on a partition
		if len(logic) > 0 {
			vol := logic[0]
			volumes = append(volumes, Volume{
				Name:       vol.Name,
				Label:      vol.VolumeName,
				MountPoint: vol.DeviceID,
				FileSystem: vol.FileSystem,
				Capacity:   vol.Size,
				FreeSpace:  vol.FreeSpace,
			})
		}
	}

	return volumes
}

// Parse USB device info from Device ID
// "USB\\VID_xxxx&PID_xxx\\xxxx"
// "USBSTOR\\DISK&VEN_MASS&PROD_STORAGE_DEVICE&REV_1.00\\121220160204&0"
func parseUsbDevice(usb string, usbstor string) UsbStorage {
	vid := strings.Index(usb, "VID_")
	pid := strings.Index(usb, "PID_")
	sn := strings.LastIndex(usb, "\\")
	disk := strings.Split(usbstor, "\\\\")[1]
	info := strings.Split(disk, "&")
	return UsbStorage{
		VendorID:  usb[vid+4 : vid+8],
		ProductID: usb[pid+4 : pid+8],
		Serial:    usb[sn+2:],
		Vendor:    info[1][4:],
		Product:   info[2][5:],
	}
}
