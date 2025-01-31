package proxmox

import (
	"fmt"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"

	pxapi "github.com/3coma3/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceVmQemu() *schema.Resource {
	*pxapi.Debug = true
	return &schema.Resource{
		Create: resourceVmQemuCreate,
		Read:   resourceVmQemuRead,
		Update: resourceVmQemuUpdate,
		Delete: ResourceVmDelete,
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
				Default:  "1",
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
				Type:       schema.TypeSet,
				Optional:   true,
				ConfigMode: schema.SchemaConfigModeAttr,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"model": {
							Type:     schema.TypeString,
							Required: true,
						},
						"macaddr": {
							// TODO: Find a way to set MAC address in .tf config.
							Type:     schema.TypeString,
							Optional: true,
							Computed: true,
						},
						"bridge": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "nat",
						},
						"tag": {
							Type:        schema.TypeInt,
							Optional:    true,
							Description: "VLAN tag.",
							Default:     -1,
						},
						"firewall": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"rate": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  -1,
						},
						"queues": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  -1,
						},
						"link_down": {
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
			"status": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

var rxIPconfig = regexp.MustCompile("ip6?=([0-9a-fA-F:\\.]+)")

func resourceVmQemuCreate(d *schema.ResourceData, meta interface{}) (err error) {
	var (
		vmid      int
		vm, src   *pxapi.Vm
		node      *pxapi.Node
		qemuDisks = devicesSetToMap(d.Get("disk").(*schema.Set))
		config    = &pxapi.ConfigQemu{
			Name:        d.Get("name").(string),
			Description: d.Get("desc").(string),
			Onboot:      d.Get("onboot").(bool),
			Agent:       d.Get("agent").(string),
			Ostype:      d.Get("ostype").(string),
			Memory:      d.Get("memory").(int),
			Cores:       d.Get("cores").(int),
			Sockets:     d.Get("sockets").(int),
			Iso:         d.Get("iso").(string),

			// Cloud-init.
			CIuser:       d.Get("ciuser").(string),
			CIpassword:   d.Get("cipassword").(string),
			Searchdomain: d.Get("searchdomain").(string),
			Nameserver:   d.Get("nameserver").(string),
			Sshkeys:      d.Get("sshkeys").(string),
			Ipconfig0:    d.Get("ipconfig0").(string),
			Ipconfig1:    d.Get("ipconfig1").(string),
			Disk:         qemuDisks,
			Net:          devicesSetToMap(d.Get("net").(*schema.Set)),
		}
		newstatus = d.Get("status").(string)
		pconf     = meta.(*providerConfiguration)
	)

	pmParallelBegin(pconf)
	// TODO: check interaction with mutex, beware this might mean going back to
	// explicit client passing
	pconf.Client.Set()

	log.Print("[DEBUG] checking for duplicate name")
	vm, _ = pxapi.FindVm(config.Name)

	if vm != nil {
		if !d.Get("force_create").(bool) {
			err = fmt.Errorf("Duplicate VM name (%s) with vmId: %d. Set force_create=true to recycle", config.Name, vm.Id())
			goto End
		}

		if vm.Node().Name() != node.Name() {
			err = fmt.Errorf("Duplicate VM name (%s) with vmId: %d on different target_node=%s", config.Name, vm.Id(), vm.Node())
			goto End
		}
	}

	if vmid, err = nextVmId(pconf); err != nil {
		goto End
	}
	vm = pxapi.NewVm(vmid)

	if node, err = pxapi.FindNode(d.Get("target_node").(string)); err != nil {
		goto End
	}
	vm.SetNode(node)

	// check if ISO or clone
	if d.Get("clone").(string) != "" {
		if src, err = pxapi.FindVm(d.Get("clone").(string)); err != nil {
			goto End
		}
		log.Print("[DEBUG] cloning VM")

		cloneParams := map[string]interface{}{
			"name": config.Name,
		}

		if _, err = src.Clone(vm.Id(), cloneParams); err != nil {
			goto End
		}

		// give sometime to proxmox to catchup
		time.Sleep(5 * time.Second)

		if err = prepareDiskSize(vm, qemuDisks); err != nil {
			goto End
		}

	} else if config.Iso != "" {
		log.Print("[DEBUG] create VM from iso at node " + vm.Node().Name() + ", vmid " + strconv.Itoa(vm.Id()) + " type " + vm.Type())
		if err = config.CreateVm(vm); err != nil {
			goto End
		}
	} else {
		return fmt.Errorf("Either clone or iso must be set")
	}

	// give sometime to proxmox to catchup
	time.Sleep(5 * time.Second)

	if newstatus != "" {
		if _, err = vm.SetStatus(newstatus); err != nil {
			goto End
		}
	}

	if err = initConnInfo(d, pconf, vm, config); err != nil {
		goto End
	}

	// a non-blank ID tells Terraform that a resource was created
	d.SetId(resourceId(vm))

	// Apply pre-provision if enabled.
	// preprovision(d, pconf, vm, true)

End:
	pmParallelEnd(pconf)

	if d.Id() == "" {
		log.Printf("An error ocurred at creation, and the resource Id is null, signaling destruction. Returning err now.")
		return err
	}

	return resourceVmQemuRead(d, meta)
}

func resourceVmQemuRead(d *schema.ResourceData, meta interface{}) (err error) {
	var (
		vmid   int
		vm     *pxapi.Vm
		config *pxapi.ConfigQemu
	)

	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)
	pconf.Client.Set()

	if _, _, vmid, err = parseResourceId(d.Id()); err != nil {
		d.SetId("")
		goto End
	}

	vm = pxapi.NewVm(vmid)

	if config, err = pxapi.NewConfigQemuFromApi(vm); err != nil {
		d.SetId("")
		goto End
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
	d.Set("ciuser", config.CIuser)
	d.Set("cipassword", config.CIpassword)
	d.Set("searchdomain", config.Searchdomain)
	d.Set("nameserver", config.Nameserver)
	d.Set("sshkeys", config.Sshkeys)
	d.Set("ipconfig0", config.Ipconfig0)
	d.Set("ipconfig1", config.Ipconfig1)

	if err = d.Set("net", updateDevicesSet(d.Get("net").(*schema.Set), config.Net)); err != nil {
		goto End
	}
	err = d.Set("disk", updateDevicesSet(d.Get("disk").(*schema.Set), config.Disk))

End:
	pmParallelEnd(pconf)
	return
}

func resourceVmQemuUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	var (
		vmid      int
		vm        *pxapi.Vm
		config    *pxapi.ConfigQemu
		qemuDisks pxapi.VmDevices

		pconf     = meta.(*providerConfiguration)
		newstatus = d.Get("status").(string)
	)

	pmParallelBegin(pconf)
	// TODO: check interaction with mutex
	// beware this might mean needing to go back to explicit client passing
	pconf.Client.Set()

	if _, _, vmid, err = parseResourceId(d.Id()); err != nil {
		d.SetId("")
		goto End
	}

	vm = pxapi.NewVm(vmid)

	if config, err = pxapi.NewConfigQemuFromApi(vm); err != nil {
		d.SetId("")
		goto End
	}

	config.Name = d.Get("name").(string)
	config.Description = d.Get("desc").(string)
	config.Onboot = d.Get("onboot").(bool)
	config.Agent = d.Get("agent").(string)
	config.Memory = d.Get("memory").(int)
	config.Cores = d.Get("cores").(int)
	config.Sockets = d.Get("sockets").(int)
	config.Ostype = d.Get("ostype").(string)
	config.CIuser = d.Get("ciuser").(string)
	config.CIpassword = d.Get("cipassword").(string)
	config.Searchdomain = d.Get("searchdomain").(string)
	config.Nameserver = d.Get("nameserver").(string)
	config.Sshkeys = d.Get("sshkeys").(string)
	config.Ipconfig0 = d.Get("ipconfig0").(string)
	config.Ipconfig1 = d.Get("ipconfig1").(string)

	qemuDisks = devicesSetToMap(d.Get("disk").(*schema.Set))
	config.Disk = qemuDisks
	config.Net = devicesSetToMap(d.Get("net").(*schema.Set))

	if err = config.UpdateConfig(vm); err != nil {
		goto End
	}

	// give sometime to proxmox to catchup
	time.Sleep(5 * time.Second)

	prepareDiskSize(vm, qemuDisks)

	// give sometime to proxmox to catchup
	time.Sleep(5 * time.Second)

	if newstatus != "" {
		if _, err = vm.SetStatus(newstatus); err != nil {
			goto End
		}
	}

	if err = initConnInfo(d, pconf, vm, config); err != nil {
		goto End
	}

	// Apply pre-provision if enabled.
	// preprovision(d, pconf, vm, false)

	// give sometime to bootup
	time.Sleep(9 * time.Second)

End:
	pmParallelEnd(pconf)

	if d.Id() == "" {
		log.Printf("An error ocurred at update. Returning err now.")
		return err
	}

	return resourceVmQemuRead(d, meta)
}

func ResourceVmDelete(d *schema.ResourceData, meta interface{}) (err error) {
	var (
		vmid int
		vm   *pxapi.Vm
	)

	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)
	pconf.Client.Set()

	if _, _, vmid, err = parseResourceId(d.Id()); err != nil {
		d.SetId("")
		goto End
	}

	vm = pxapi.NewVm(vmid)

	if err = vm.Check(); err != nil {
		goto End
	}

	if _, err = vm.Shutdown(); err != nil {
		goto End
	}

	// give sometime to proxmox to catchup
	time.Sleep(2 * time.Second)
	_, err = vm.Delete()

End:
	pmParallelEnd(pconf)
	return
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
