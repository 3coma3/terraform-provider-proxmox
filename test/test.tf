provider "proxmox" {
    pm_tls_insecure = true
    pm_api_url      = "https://10.40.0.147:8006/api2/json"
    pm_user         = "root@pam"
}

# resource "proxmox_vm_qemu" "test-tfcreate" {
#    name = "test-tfcreate"
#    desc = "A test for using terraform and cloudinit"
#
#    # Node name has to be the same name as within the cluster
#    # this might not include the FQDN
#    target_node = "pve"
#
#    # The template name to clone this vm from
#    # clone = "linux-cloudinit-template"
#    iso = "local:iso/uccorelinux.iso"
#
#    # Activate QEMU agent for this VM
#    agent = "enabled=1"
#
#    cores = "2"
#    sockets = "2"
#    memory = "2048"
#
#    # Setup the disk. The id has to be unique
#    disk {
#        id = 0
#        size = "8G"
#        type = "virtio"
#        storage = "local-lvm"
#        storage_type = "lvm"
#        iothread = true
#    }
#
#    # Setup the network interface and assign a vlan tag: 256
#    net {
#        id = 0
#        model = "virtio"
#        bridge = "vmbr0"
#        tag = 256
#    }
#
#    preprovision = false
#    # preprovision_ostype = "cloud-init"
#    # Setup the ip address using cloud-init.
#    # Keep in mind to use the CIDR notation for the ip.
#    # ipconfig0 = "ip=192.168.10.20/24,gw=192.168.10.1"
#
#    # sshkeys = <<EOF
#    # ssh-rsa 9182739187293817293817293871== user@pc
#    # EOF
#}

resource "proxmox_vm_lxc" "test-tfcreatelxc" {
    target_node = "pve"
    hostname = "test-tfcreatelxc"

    ostemplate  = "local:vztmpl/ubuntu-16.04-standard_16.04-1_amd64.tar.gz"
    ostype      = "ubuntu"
    password    = "rootroot"

    cores       = 1
    memory      = 4096

    net {
        id     = 0
        name   = "eth0"
        bridge = "vmbr0"
        ip     = "dhcp"
        ip6    = "dhcp"
    }

    rootfs {
        storage = "local-lvm"
        size    = "8G"
        acl     = true
    }
}
