variable "os_version" {
  type        = string
  description = "Ubuntu version to reinstall. Changing it reprovisions the existing server and destroys local-disk data."
  default     = "22.04"
}

variable "raid_type" {
  type        = string
  description = "RAID mode used while reinstalling the OS. Changing it reprovisions the existing server and destroys local-disk data."
  default     = "RAID1"
}

# WARNING: reprovisioning reinstalls the OS on this existing server and erases
# data on its local disks. Back up the server before changing these settings.

resource "vkcs_compute_keypair" "generated_key" {
  name = "baremetal-reprovision-tf-example"
}

data "vkcs_baremetal_flavor" "server" {
  name = "BM_CX301_N_BOND"
}

data "vkcs_baremetal_os" "ubuntu" {
  name      = "ubuntu"
  version   = var.os_version
  raid_type = var.raid_type
}

resource "vkcs_networking_network" "server" {
  name        = "baremetal-reprovision-tf-example"
  description = "Network for the bare metal reprovision example"
}

resource "vkcs_networking_subnet" "server" {
  name       = "baremetal-reprovision-tf-example"
  network_id = vkcs_networking_network.server.id
  cidr       = "192.168.211.0/24"
}

resource "vkcs_baremetal_server" "server" {
  name              = "baremetal-reprovision-tf-example"
  availability_zone = "ME1"
  flavor_id         = data.vkcs_baremetal_flavor.server.id
  os_id             = data.vkcs_baremetal_os.ubuntu.id
  key_pair          = vkcs_compute_keypair.generated_key.name
  raid_type         = var.raid_type

  # Changing os_version, raid_type, user_data, or key_pair and applying the
  # change reinstalls the OS on this server and destroys local-disk data.
  user_data = templatefile("${path.module}/cloud-init-user-data.yaml", {
    public_key = vkcs_compute_keypair.generated_key.public_key
  })

  nic {
    name = "nic0"
    vlan {
      native     = true
      network_id = vkcs_networking_network.server.id
      subnet_id  = vkcs_networking_subnet.server.id
    }
  }
}

output "server_id" {
  value = vkcs_baremetal_server.server.id
}

output "os_version" {
  value = var.os_version
}
