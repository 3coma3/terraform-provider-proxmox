package proxmox

import (
	"encoding/json"
	pxapi "github.com/3coma3/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform/helper/schema"
	"strconv"
)

// convert a schema.TypeSet such as net, mp or disk to map[int]map[string]interface{}
func devicesSetToMap(devicesSet *schema.Set) pxapi.VmDevices {
	apiDevicesMap := pxapi.VmDevices{}

	for _, subset := range devicesSet.List() {
		if subsetMap, isMap := subset.(map[string]interface{}); isMap {
			apiDevicesMap[subsetMap["id"].(int)] = subsetMap
		}
	}

	return apiDevicesMap
}

// update a schema.TypeSet with new values from map[int]map[string]interface{}
func updateDevicesSet(
	terraDevicesSet *schema.Set,
	apiDevicesMap pxapi.VmDevices,
) *schema.Set {
	var apiDevicesList []interface{}

	for id, conf := range terraDevicesSet.List() {
		confMap := conf.(map[string]interface{})

		// add missing data in the api map with tf config state
		if _, ok := apiDevicesMap[id]; !ok {
			apiDevicesMap[id] = confMap
		}

		for k, v := range confMap {
			if _, ok := apiDevicesMap[id][k]; !ok {
				apiDevicesMap[id][k] = v
			}
		}

		// update the tf state with the fresh api data
		for k, v := range apiDevicesMap[id] {
			// ignore fields from the api map that don't exist in the schema
			// FIXME: this might not be working correctly as is
			if _, ok := confMap[k]; !ok {
				continue
			}

			// if the schema key is bool and the value is int (comes from the api)
			// convert it to bool as tf expects
			_, keybool := confMap[k].(bool)
			if _, valueint := v.(int); keybool && valueint {
				if bV, err := strconv.ParseBool(strconv.Itoa(v.(int))); err == nil {
					confMap[k] = bV
				}
			} else {
				confMap[k] = v
			}
		}

		apiDevicesList = append(apiDevicesList, confMap)
	}

	// this callback returns an identity token for members of the set
	var hashFunc = func(i interface{}) int {
		if id := i.(map[string]interface{})["id"]; id != nil {
			return id.(int)
		} else {
			b, _ := json.MarshalIndent(i, "", "")
			return schema.HashString(string(b))
		}
	}

	return schema.NewSet(hashFunc, apiDevicesList)
}

// update a schema.TypeSet with new values from map[string]interface{}
func updateDeviceSet(
	terraDevicesSet *schema.Set,
	apiDeviceMap pxapi.VmDevice,
) *schema.Set {
	nestApiDeviceMap := pxapi.VmDevices{}
	for id, _ := range terraDevicesSet.List() {
		nestApiDeviceMap[id] = apiDeviceMap
	}
	return updateDevicesSet(terraDevicesSet, nestApiDeviceMap)
}
