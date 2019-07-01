package proxmox

import (
	pxapi "github.com/3coma3/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform/helper/schema"
	"strconv"
)

func UpdateDeviceConfDefaults(
	activeDeviceConf pxapi.VmDevice,
	defaultDeviceConf *schema.Set,
) *schema.Set {
	defaultDeviceConfMap := defaultDeviceConf.List()[0].(map[string]interface{})
	for key, _ := range defaultDeviceConfMap {
		if deviceConfigValue, ok := activeDeviceConf[key]; ok {
			defaultDeviceConfMap[key] = deviceConfigValue
			switch deviceConfigValue.(type) {
			case int:
				sValue := strconv.Itoa(deviceConfigValue.(int))
				bValue, err := strconv.ParseBool(sValue)
				if err == nil {
					defaultDeviceConfMap[key] = bValue
				}
			default:
				defaultDeviceConfMap[key] = deviceConfigValue
			}
		}
	}
	defaultDeviceConf.Remove(defaultDeviceConf.List()[0])
	defaultDeviceConf.Add(defaultDeviceConfMap)
	return defaultDeviceConf
}
