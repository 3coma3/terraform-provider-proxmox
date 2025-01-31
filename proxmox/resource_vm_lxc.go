package proxmox

import (
	"encoding/json"
	"fmt"
	pxapi "github.com/3coma3/proxmox-api-go/proxmox"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"log"
	"reflect"
	"strconv"
	"time"
)

func resourceVmLxc() *schema.Resource {
	*pxapi.Debug = true
	return &schema.Resource{
		Create: resourceVmLxcCreate,
		Read:   resourceVmLxcRead,
		Update: resourceVmLxcUpdate,
		Delete: ResourceVmDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"arch": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "amd64",
			},
			"cmode": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "tty",
			},
			"console": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"clone": {
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"ostemplate"},
			},
			"cores": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1,
			},
			"cpuunits": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  1024,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"hostname": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"memory": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  512,
			},
			"mp": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"volume": {
							Type:     schema.TypeString,
							Required: true,
						},
						"mp": {
							Type:     schema.TypeString,
							Required: true,
						},
						"acl": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"backup": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"quota": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"replicate": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"ro": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"shared": {
							Type:     schema.TypeBool,
							Optional: true,
						},
						"size": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
			"nameserver": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"net": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"bridge": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"firewall": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"gw": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"hwaddr": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"gw6": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"ip6": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"mtu": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"rate": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"tag": {
							Type:     schema.TypeInt,
							Optional: true,
							Default:  0,
						},
						"trunks": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "",
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "veth",
						},
					},
				},
			},
			"onboot": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ostype": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"ostemplate": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"clone"},
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Id() != ""
				},
			},
			"password": {
				Type:     schema.TypeString,
				Optional: true,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					return d.Id() != ""
				},
			},
			"protection": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"rootfs": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"storage": {
							Type:     schema.TypeString,
							Required: true,
						},
						"size": {
							Type:     schema.TypeString,
							Required: true,
						},
						"acl": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"quota": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"replicate": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"ro": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
						"shared": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  false,
						},
					},
				},
			},
			"searchdomain": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
			"startup": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"sshkeys": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"swap": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  512,
			},
			"tty": {
				Type:     schema.TypeInt,
				Optional: true,
				Default:  2,
			},
			"unprivileged": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"target_node": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
		},
	}
}

func resourceVmLxcCreate(d *schema.ResourceData, meta interface{}) (err error) {
	var (
		vmid   int
		vm     *pxapi.Vm
		node   *pxapi.Node
		config = pxapi.NewConfigLxc()

		pconf     = meta.(*providerConfiguration)
		newstatus = d.Get("status").(string)
	)

	pmParallelBegin(pconf)
	// TODO: check interaction with mutex
	// beware this might mean needing to go back to explicit client passing
	pconf.Client.Set()

	config.Hostname = d.Get("hostname").(string)
	config.Ostemplate = d.Get("ostemplate").(string)
	config.Arch = d.Get("arch").(string)
	config.Cmode = d.Get("cmode").(string)
	config.Console = d.Get("console").(bool)
	config.Cores = d.Get("cores").(int)
	config.Cpuunits = d.Get("cpuunits").(int)
	config.Description = d.Get("description").(string)
	config.Memory = d.Get("memory").(int)
	config.Nameserver = d.Get("nameserver").(string)
	config.Onboot = d.Get("onboot").(bool)
	config.Ostype = d.Get("ostype").(string)
	config.Ostemplate = d.Get("ostemplate").(string)
	config.Password = d.Get("password").(string)
	config.Protection = d.Get("protection").(bool)
	config.Searchdomain = d.Get("searchdomain").(string)
	config.Sshkeys = d.Get("sshkeys").(string)
	config.Startup = d.Get("startup").(string)
	config.Swap = d.Get("swap").(int)
	config.Tty = d.Get("tty").(int)
	config.Unprivileged = d.Get("unprivileged").(bool)

	config.Rootfs = d.Get("rootfs").(*schema.Set).List()[0].(map[string]interface{})
	config.Mp = devicesSetToMap(d.Get("mp").(*schema.Set))
	config.Net = devicesSetToMap(d.Get("net").(*schema.Set))

	log.Print("[DEBUG] checking for duplicate name")
	vm, _ = pxapi.FindVm(config.Hostname)

	if vm != nil {
		if !d.Get("force_create").(bool) {
			err = fmt.Errorf("Duplicate VM name (%s) with vmId: %d. Set force_create=true to recycle", config.Hostname, vm.Id())
			goto End
		}

		if vm.Node().Name() != node.Name() {
			err = fmt.Errorf("Duplicate VM name (%s) with vmId: %d on different target_node=%s", config.Hostname, vm.Id(), vm.Node())
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

	if err = config.CreateVm(vm); err != nil {
		goto End
	}

	// give sometime to proxmox to catchup
	time.Sleep(5 * time.Second)

	if newstatus != "" {
		if _, err = vm.SetStatus(newstatus); err != nil {
			goto End
		}
	}

	// a non-blank ID tells Terraform that a resource was created
	d.SetId(resourceId(vm))

End:
	pmParallelEnd(pconf)

	if d.Id() == "" {
		log.Printf("An error ocurred at creation, and the resource Id is null, signaling destruction. Returning err now.")
		return err
	}

	return resourceVmLxcRead(d, meta)
}

func resourceVmLxcRead(d *schema.ResourceData, meta interface{}) (err error) {
	var (
		vmid   int
		vm     *pxapi.Vm
		config *pxapi.ConfigLxc
	)

	pconf := meta.(*providerConfiguration)
	pmParallelBegin(pconf)
	pconf.Client.Set()

	if _, _, vmid, err = parseResourceId(d.Id()); err != nil {
		d.SetId("")
		goto End
	}

	vm = pxapi.NewVm(vmid)

	if config, err = pxapi.NewConfigLxcFromApi(vm); err != nil {
		d.SetId("")
		goto End
	}

	log.Println("DEBUG A")
	printSet(d, "net")

	d.SetId(resourceId(vm))

	d.Set("target_node", vm.Node().Name())
	d.Set("arch", config.Arch)
	d.Set("cmode", config.Cmode)
	d.Set("console", config.Console)
	d.Set("cores", config.Cores)
	d.Set("cpuunits", config.Cpuunits)
	d.Set("description", config.Description)
	d.Set("hostname", config.Hostname)
	d.Set("memory", config.Memory)
	d.Set("nameserver", config.Nameserver)
	d.Set("onboot", config.Onboot)
	d.Set("ostemplate", config.Ostemplate)
	d.Set("ostype", config.Ostype)
	d.Set("password", config.Password)
	d.Set("searchdomain", config.Searchdomain)
	d.Set("sshkeys", config.Sshkeys)
	d.Set("startup", config.Startup)
	d.Set("swap", config.Swap)
	d.Set("tty", config.Tty)
	d.Set("unprivileged", config.Unprivileged)

	if err = d.Set("net", updateDevicesSet(d.Get("net").(*schema.Set), config.Net)); err != nil {
		goto End
	}
	if err = d.Set("mp", updateDevicesSet(d.Get("mp").(*schema.Set), config.Mp)); err != nil {
		goto End
	}
	err = d.Set("rootfs", updateDeviceSet(d.Get("rootfs").(*schema.Set), config.Rootfs))

	log.Println("DEBUG B")
	printSet(d, "net")

End:
	pmParallelEnd(pconf)
	return
}

func resourceVmLxcUpdate(d *schema.ResourceData, meta interface{}) (err error) {
	var (
		vmid   int
		vm     *pxapi.Vm
		config *pxapi.ConfigLxc

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

	if config, err = pxapi.NewConfigLxcFromApi(vm); err != nil {
		d.SetId("")
		goto End
	}

	config.Ostemplate = d.Get("ostemplate").(string)
	config.Arch = d.Get("arch").(string)
	config.Cmode = d.Get("cmode").(string)
	config.Console = d.Get("console").(bool)
	config.Cores = d.Get("cores").(int)
	config.Cpuunits = d.Get("cpuunits").(int)
	config.Description = d.Get("description").(string)
	config.Hostname = d.Get("hostname").(string)
	config.Memory = d.Get("memory").(int)
	config.Nameserver = d.Get("nameserver").(string)
	config.Onboot = d.Get("onboot").(bool)
	config.Ostype = d.Get("ostype").(string)
	config.Ostemplate = d.Get("ostemplate").(string)
	config.Password = d.Get("password").(string)
	config.Protection = d.Get("protection").(bool)
	config.Searchdomain = d.Get("searchdomain").(string)
	config.Sshkeys = d.Get("sshkeys").(string)
	config.Startup = d.Get("startup").(string)
	config.Swap = d.Get("swap").(int)
	config.Tty = d.Get("tty").(int)

	config.Rootfs = d.Get("rootfs").(*schema.Set).List()[0].(map[string]interface{})
	config.Mp = devicesSetToMap(d.Get("mp").(*schema.Set))
	config.Net = devicesSetToMap(d.Get("net").(*schema.Set))

	if err = config.UpdateConfig(vm); err != nil {
		goto End
	}

	// give sometime to proxmox to catchup
	time.Sleep(5 * time.Second)

	if newstatus != "" {
		if _, err = vm.SetStatus(newstatus); err != nil {
			goto End
		}
	}

End:
	pmParallelEnd(pconf)

	if d.Id() == "" {
		log.Printf("An error ocurred at update. Returning err now.")
		return err
	}

	return resourceVmLxcRead(d, meta)
}

// to debug nested sets
func printSet(d *schema.ResourceData, s string) {
	set := d.Get(s).(*schema.Set)

	log.Println("DEBUG: printSet -- " + s + " has len " + strconv.Itoa(set.Len()))

	for id, conf := range set.List() {
		log.Println("element " + strconv.Itoa(id) + " has type :")
		log.Println(reflect.TypeOf(conf))

		s, err := json.MarshalIndent(conf, "", "  ")
		if err != nil {
			log.Println("error:", err)
		}
		log.Println("the contents of element " + strconv.Itoa(id) + " are " + string(s))
	}
}

// to debug non nested sets
func printMap(m pxapi.VmDevices) {
	log.Println("DEBUG: printMap -- ")
	s, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		log.Println("error:", err)
	}
	log.Println("the contents are " + string(s))
}
