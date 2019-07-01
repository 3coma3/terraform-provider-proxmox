package proxmox

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	pxapi "github.com/3coma3/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceVmQemu() *schema.Resource {
	*pxapi.Debug = true
	return &schema.Resource{
		Create: resourceVmQemuCreate,
		Read:   resourceVmQemuRead,
		Update: resourceVmQemuUpdate,
		Delete: resourceVmQemuDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"desc": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			"target_node": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"onboot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"agent": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  0,
			},
			"iso": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"clone": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ostype": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "l26",
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if new == "l26" {
						return len(d.Get("clone").(string)) > 0 // the cloned source may have a different os, which we shoud leave alone
					}
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  512,
			},
			"cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"sockets": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"net": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
						"model": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"macaddr": &schema.Schema{
							// TODO: Find a way to set MAC address in .tf config.
							Type:     schema.TypeString,
							Optional: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if new == "" {
									return true // macaddr auto-generates and its ok
								}
								return strings.TrimSpace(old) == strings.TrimSpace(new)
							},
						},
						"bridge": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "nat",
						},
						"tag": &schema.Schema{
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "VLAN tag.",
							Default:     -1,
						},
						"firewall": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"rate": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  -1,
						},
						"queues": &schema.Schema{
							Type:     schema.TypeInt,
							Optional: true,
							Default:  -1,
						},
						"link_down": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"disk": &schema.Schema{
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": &schema.Schema{
							Type:     schema.TypeInt,
							Required: true,
						},
						"type": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"storage": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"storage_type": &schema.Schema{
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "dir",
							Description: "One of PVE types as described: https://pve.proxmox.com/wiki/Storage",
						},
						"size": &schema.Schema{
							Type:     schema.TypeString,
							Required: true,
						},
						"format": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "raw",
						},
						"cache": &schema.Schema{
							Type:     schema.TypeString,
							Optional: true,
							Default:  "none",
						},
						"backup": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"iothread": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"replicate": &schema.Schema{
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"preprovision_ostype": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"preprovision_netconfig": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			"ssh_forward_ip": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ssh_user": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"ssh_private_key": {
				Type:      schema.TypeString,
				Optional:  true,
				Sensitive: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			"force_create": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ci_wait": { // how long to wait before provision
				Type:     schema.TypeInt,
				Optional: true,
				Default:  30,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "" {
						return true // old empty ok
					}
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			"ciuser": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"cipassword": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"searchdomain": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"nameserver": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sshkeys": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return strings.TrimSpace(old) == strings.TrimSpace(new)
				},
			},
			"ipconfig0": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ipconfig1": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"preprovision": {
				Type:          schema.TypeBool,
				Optional:      true,
				Default:       true,
				ConflictsWith: []string{"ssh_forward_ip", "ssh_user", "ssh_private_key", "preprovision_ostype", "preprovision_netconfig"},
			},
		},
	}
}

var rxIPconfig = regexp.MustCompile("ip6?=([0-9a-fA-F:\\.]+)")

func resourceVmQemuCreate(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)

	// TODO: check interaction with mutex
	// beware this might mean needing to go back to explicit client passing
	pconf.Client.Set()

	vmName := d.Get("name").(string)
	networks := d.Get("network").(*schema.Set)
	qemuNetworks := DevicesSetToMap(networks)
	disks := d.Get("disk").(*schema.Set)
	qemuDisks := DevicesSetToMap(disks)

	config := pxapi.ConfigQemu{
		Name:        vmName,
		Description: d.Get("desc").(string),
		Onboot:      d.Get("onboot").(bool),
		Agent:       d.Get("agent").(string),
		Ostype:      d.Get("ostype").(string),
		Memory:      d.Get("memory").(int),
		Cores:       d.Get("cores").(int),
		Sockets:     d.Get("sockets").(int),
		Net:         devicesSetToMap(d.Get("net").(*schema.Set)),
		Disk:        qemuDisks,
		// Cloud-init.
		CIuser:       d.Get("ciuser").(string),
		CIpassword:   d.Get("cipassword").(string),
		Searchdomain: d.Get("searchdomain").(string),
		Nameserver:   d.Get("nameserver").(string),
		Sshkeys:      d.Get("sshkeys").(string),
		Ipconfig0:    d.Get("ipconfig0").(string),
		Ipconfig1:    d.Get("ipconfig1").(string),
	}
	log.Print("[DEBUG] checking for duplicate name")
	dupVm, _ := pxapi.FindVm(vmName)

	forceCreate := d.Get("force_create").(bool)
	targetNode := d.Get("target_node").(string)

	if dupVm != nil && forceCreate {
		pmParallelEnd(pconf)
		return fmt.Errorf("Duplicate VM name (%s) with vmId: %d. Set force_create=false to recycle", vmName, dupVm.Id())
	} else if dupVm != nil && dupVm.Node().Name() != targetNode {
		pmParallelEnd(pconf)
		return fmt.Errorf("Duplicate VM name (%s) with vmId: %d on different target_node=%s", vmName, dupVm.Id(), dupVm.Node())
	}

	vm := dupVm

	if vm == nil {
		// get unique id
		nextid, err := nextVmId(pconf)
		if err != nil {
			pmParallelEnd(pconf)
			return err
		}
		vm = pxapi.NewVm(nextid)

		n, _ := pxapi.FindNode(targetNode)
		vm.SetNode(n)

		// check if ISO or clone
		if d.Get("clone").(string) != "" {
			sourceVm, err := pxapi.FindVm(d.Get("clone").(string))
			if err != nil {
				pmParallelEnd(pconf)
				return err
			}
			log.Print("[DEBUG] cloning VM")

			cloneParams := map[string]interface{}{}

			_, err = sourceVm.Clone(vm.Id(), cloneParams)
			if err != nil {
				pmParallelEnd(pconf)
				return err
			}

			// give sometime to proxmox to catchup
			time.Sleep(5 * time.Second)

			err = prepareDiskSize(vm, qemuDisks)
			if err != nil {
				pmParallelEnd(pconf)
				return err
			}

		} else if d.Get("iso").(string) != "" {
			log.Print("[DEBUG] create VM from iso at node " + vm.Node().Name() + ", vmid " + strconv.Itoa(vm.Id()) + " type " + vm.Type())
			config.Iso = d.Get("iso").(string)
			err := config.CreateVm(vm)
			if err != nil {
				pmParallelEnd(pconf)
				return err
			}
		} else {
			return fmt.Errorf("Either clone or iso must be set")
		}
	} else {
		log.Printf("[DEBUG] recycling VM: %d", vm.Id())

		vm.Stop()

		err := config.UpdateConfig(vm)
		if err != nil {
			pmParallelEnd(pconf)
			return err
		}

		// give sometime to proxmox to catchup
		time.Sleep(5 * time.Second)

		err = prepareDiskSize(vm, qemuDisks)
		if err != nil {
			pmParallelEnd(pconf)
			return err
		}
	}
	d.SetId(resourceId(vm))

	// give sometime to proxmox to catchup
	time.Sleep(5 * time.Second)

	log.Print("[DEBUG] starting VM")
	_, err := vm.Start()
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}

	err = initConnInfo(d, pconf, vm, &config)
	if err != nil {
		return err
	}

	// Apply pre-provision if enabled.
	preprovision(d, pconf, vm, true)

	return nil
}

func resourceVmQemuUpdate(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)

	pconf.Client.Set()

	_, _, vmID, err := parseResourceId(d.Id())
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}

	vm := pxapi.NewVm(vmID)
	if err = vm.Check(); err != nil {
		pmParallelEnd(pconf)
		return err
	}
	configDisksSet := d.Get("disk").(*schema.Set)
	qemuDisks := DevicesSetToMap(configDisksSet)
	configNetworksSet := d.Get("network").(*schema.Set)
	qemuNetworks := DevicesSetToMap(configNetworksSet)

	config := pxapi.ConfigQemu{
		Name:        d.Get("name").(string),
		Description: d.Get("desc").(string),
		Onboot:      d.Get("onboot").(bool),
		Agent:       d.Get("agent").(string),
		Memory:      d.Get("memory").(int),
		Cores:       d.Get("cores").(int),
		Sockets:     d.Get("sockets").(int),
		Ostype:      d.Get("ostype").(string),
		Net:         devicesSetToMap(d.Get("net").(*schema.Set)),
		Disk:        qemuDisks,
		// Cloud-init.
		CIuser:       d.Get("ciuser").(string),
		CIpassword:   d.Get("cipassword").(string),
		Searchdomain: d.Get("searchdomain").(string),
		Nameserver:   d.Get("nameserver").(string),
		Sshkeys:      d.Get("sshkeys").(string),
		Ipconfig0:    d.Get("ipconfig0").(string),
		Ipconfig1:    d.Get("ipconfig1").(string),
	}

	err = config.UpdateConfig(vm)
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}

	// give sometime to proxmox to catchup
	time.Sleep(5 * time.Second)

	prepareDiskSize(vm, qemuDisks)

	// give sometime to proxmox to catchup
	time.Sleep(5 * time.Second)

	// Start VM only if it wasn't running.
	vmStatus, err := vm.GetStatus()
	if err == nil && vmStatus["status"] == "stopped" {
		log.Print("[DEBUG] starting VM")
		_, err = vm.Start()
	} else if err != nil {
		pmParallelEnd(pconf)
		return err
	}

	err = initConnInfo(d, pconf, vm, &config)
	if err != nil {
		return err
	}

	// Apply pre-provision if enabled.
	preprovision(d, pconf, vm, false)

	// give sometime to bootup
	time.Sleep(9 * time.Second)
	return nil
}

func resourceVmQemuRead(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)
	pconf.Client.Set()

	_, _, vmId, err := parseResourceId(d.Id())
	if err != nil {
		pmParallelEnd(pconf)
		d.SetId("")
		return err
	}

	vm := pxapi.NewVm(vmId)
	if err = vm.Check(); err != nil {
		pmParallelEnd(pconf)
		return err
	}

	config, err := pxapi.NewConfigQemuFromApi(vm)
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}

	d.SetId(resourceId(vm))
	d.Set("target_node", vm.Node().Name())
	d.Set("name", config.Name)
	d.Set("desc", config.Description)
	d.Set("onboot", config.Onboot)
	d.Set("agent", config.Agent)
	d.Set("memory", config.Memory)
	d.Set("cores", config.Cores)
	d.Set("sockets", config.Sockets)
	d.Set("ostype", config.Ostype)
	// Cloud-init.
	d.Set("ciuser", config.CIuser)
	d.Set("cipassword", config.CIpassword)
	d.Set("searchdomain", config.Searchdomain)
	d.Set("nameserver", config.Nameserver)
	d.Set("sshkeys", config.Sshkeys)
	d.Set("ipconfig0", config.Ipconfig0)
	d.Set("ipconfig1", config.Ipconfig1)

	// Disks.
	configDisksSet := d.Get("disk").(*schema.Set)
	activeDisksSet := UpdateDevicesSet(configDisksSet, config.QemuDisks)
	d.Set("disk", activeDisksSet)

	// Networks.
	configNetworksSet := d.Get("network").(*schema.Set)
	activeNetworksSet := UpdateDevicesSet(configNetworksSet, config.QemuNetworks)
	d.Set("network", activeNetworksSet)

	pmParallelEnd(pconf)
	return nil
}

func resourceVmQemuDelete(d *schema.ResourceData, meta interface{}) error {
	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)

	pconf.Client.Set()

	_, _, vmId, err := parseResourceId(d.Id())
	if err != nil {
		pmParallelEnd(pconf)
		d.SetId("")
		return err
	}

	vm := pxapi.NewVm(vmId)
	if err = vm.Check(); err != nil {
		pmParallelEnd(pconf)
		return err
	}

	_, err = vm.Stop()
	if err != nil {
		pmParallelEnd(pconf)
		return err
	}
	// give sometime to proxmox to catchup
	time.Sleep(2 * time.Second)
	_, err = vm.Delete()
	pmParallelEnd(pconf)
	return err
}

// Increase disk size if original disk was smaller than new disk.
func prepareDiskSize(
	vm *pxapi.Vm,
	diskConfMap pxapi.VmDevices,
) error {
	clonedConfig, err := pxapi.NewConfigQemuFromApi(vm)
	if err != nil {
		return err
	}
	//log.Printf("%s", clonedConfig)
	for diskID, diskConf := range diskConfMap {
		diskName := fmt.Sprintf("%v%v", diskConf["type"], diskID)

		diskSize := diskSizeGB(diskConf["size"])

		if _, diskExists := clonedConfig.Disk[diskID]; !diskExists {
			return err
		}

		clonedDiskSize := diskSizeGB(clonedConfig.Disk[diskID]["size"])

		if err != nil {
			return err
		}

		if diskSize > clonedDiskSize {
			log.Print("[DEBUG] resizing disk " + diskName)
			_, err = vm.ResizeDisk(diskName, strconv.Itoa(int(diskSize)))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func diskSizeGB(dcSize interface{}) float64 {
	var diskSize float64
	// TODO support other units M/G/K
	switch dcSize.(type) {
	case string:
		diskSizeGB := dcSize.(string)
		diskSize, _ = strconv.ParseFloat(strings.Trim(diskSizeGB, "G"), 64)
	case float64:
		diskSize = dcSize.(float64)
	}
	return diskSize
}

// Converting from schema.TypeSet to map of id and conf for each device,
// which will be sent to Proxmox API.
func DevicesSetToMap(devicesSet *schema.Set) pxapi.QemuDevices {

	devicesMap := pxapi.VmDevices{}

	for _, set := range devicesSet.List() {
		setMap, isMap := set.(map[string]interface{})
		if isMap {
			setID := setMap["id"].(int)
			devicesMap[setID] = setMap
		}
	}
	return devicesMap
}

// Update schema.TypeSet with new values comes from Proxmox API.
// TODO: Maybe it's better to create a new Set instead add to current one.
func UpdateDevicesSet(
	devicesSet *schema.Set,
	devicesMap pxapi.VmDevices,
) *schema.Set {

	configDevicesMap := DevicesSetToMap(devicesSet)

	activeDevicesMap := updateDevicesDefaults(devicesMap, configDevicesMap)

	for _, setConf := range devicesSet.List() {
		devicesSet.Remove(setConf)
		setConfMap := setConf.(map[string]interface{})
		deviceID := setConfMap["id"].(int)
		// Value type should be one of types allowed by Terraform schema types.
		for key, value := range activeDevicesMap[deviceID] {
			// This nested switch is used for nested config like in `net[n]`,
			// where Proxmox uses `key=<0|1>` in string" at the same time
			// a boolean could be used in ".tf" files.
			switch setConfMap[key].(type) {
			case bool:
				switch value.(type) {
				// If the key is bool and value is int (which comes from Proxmox API),
				// should be converted to bool (as in ".tf" conf).
				case int:
					sValue := strconv.Itoa(value.(int))
					bValue, err := strconv.ParseBool(sValue)
					if err == nil {
						setConfMap[key] = bValue
					}
				// If value is bool, which comes from Terraform conf, add it directly.
				case bool:
					setConfMap[key] = value
				}
			// Anything else will be added as it is.
			default:
				setConfMap[key] = value
			}
			devicesSet.Add(setConfMap)
		}
	}

	return devicesSet
}

// Because default values are not stored in Proxmox, so the API returns only active values.
// So to prevent Terraform doing unnecessary diffs, this function reads default values
// from Terraform itself, and fill empty fields.
func updateDevicesDefaults(
	activeDevicesMap pxapi.VmDevices,
	configDevicesMap pxapi.VmDevices,
) pxapi.VmDevices {

	for deviceID, deviceConf := range configDevicesMap {
		if _, ok := activeDevicesMap[deviceID]; !ok {
			activeDevicesMap[deviceID] = configDevicesMap[deviceID]
		}
		for key, value := range deviceConf {
			if _, ok := activeDevicesMap[deviceID][key]; !ok {
				activeDevicesMap[deviceID][key] = value
			}
		}
	}
	return activeDevicesMap
}

func initConnInfo(
	d *schema.ResourceData,
	pconf *providerConfiguration,
	vm *pxapi.Vm,
	config *pxapi.ConfigQemu) error {

	sshPort := "22"
	sshHost := ""
	var err error
	if config.HasCloudInit() {
		if d.Get("ssh_forward_ip") != nil {
			sshHost = d.Get("ssh_forward_ip").(string)
		}
		if sshHost == "" {
			// parse IP address out of ipconfig0
			ipMatch := rxIPconfig.FindStringSubmatch(d.Get("ipconfig0").(string))
			sshHost = ipMatch[1]
		}
	} else {
		log.Print("[DEBUG] setting up SSH forward")
		sshPort, err = vm.SshForwardUsernet()
		if err != nil {
			pmParallelEnd(pconf)
			return err
		}
		sshHost = d.Get("ssh_forward_ip").(string)
	}

	// Done with proxmox API, end parallel and do the SSH things
	pmParallelEnd(pconf)

	client := pxapi.GetClient()

	d.SetConnInfo(map[string]string{
		"type":            "ssh",
		"host":            sshHost,
		"port":            sshPort,
		"user":            d.Get("ssh_user").(string),
		"private_key":     d.Get("ssh_private_key").(string),
		"pm_api_url":      client.ApiUrl,
		"pm_user":         client.Username,
		"pm_password":     client.Password,
		"pm_tls_insecure": "true", // TODO - pass pm_tls_insecure state around, but if we made it this far, default insecure
	})
	return nil
}

// Internal pre-provision.
func preprovision(
	d *schema.ResourceData,
	pconf *providerConfiguration,
	vm *pxapi.Vm,
	systemPreProvision bool,
) error {

	if d.Get("preprovision").(bool) {

		if systemPreProvision {
			switch d.Get("os_type").(string) {

			case "ubuntu":
				// give sometime to bootup
				time.Sleep(9 * time.Second)
				err := preProvisionUbuntu(d)
				if err != nil {
					return err
				}

			case "centos":
				// give sometime to bootup
				time.Sleep(9 * time.Second)
				err := preProvisionCentos(d)
				if err != nil {
					return err
				}

			case "cloud-init":
				// wait for OS too boot awhile...
				log.Print("[DEBUG] sleeping for OS bootup...")
				time.Sleep(time.Duration(d.Get("ci_wait").(int)) * time.Second)

			default:
				return fmt.Errorf("Unknown os_type: %s", d.Get("os_type").(string))
			}
		}
	}
	return nil
}
