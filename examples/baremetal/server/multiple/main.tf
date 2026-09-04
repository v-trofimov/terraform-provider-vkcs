variable "server_count" {
  type        = number
  description = "Number of bare metal servers to rent."
  default     = 2
}

resource "vkcs_compute_keypair" "generated_key" {
  name = "baremetal-multiple-tf-example"
}

data "vkcs_baremetal_flavor" "server" {
  name = "BM_CX301_N_BOND"
}

data "vkcs_baremetal_os" "ubuntu" {
  name      = "ubuntu"
  version   = "22.04"
  raid_type = "RAID1"
}

resource "vkcs_networking_network" "server" {
  name        = "baremetal-multiple-tf-example"
  description = "Network for the bare metal servers example"
}

resource "vkcs_networking_subnet" "server" {
  name       = "baremetal-multiple-tf-example"
  network_id = vkcs_networking_network.server.id
  cidr       = "192.168.210.0/24"
}

resource "vkcs_baremetal_server" "server" {
  count             = var.server_count
  name              = "baremetal-server-${count.index + 1}"
  availability_zone = "ME1"
  flavor_id         = data.vkcs_baremetal_flavor.server.id
  os_id             = data.vkcs_baremetal_os.ubuntu.id
  key_pair          = vkcs_compute_keypair.generated_key.name
  raid_type         = "RAID1"

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

output "server_ids" {
  value = vkcs_baremetal_server.server[*].id
}

output "server_names" {
  value = vkcs_baremetal_server.server[*].name
}
