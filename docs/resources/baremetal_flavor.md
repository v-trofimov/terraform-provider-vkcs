---
subcategory: "Baremetal"
layout: "vkcs"
page_title: "vkcs: vkcs_baremetal_flavor"
description: |-
  Manages project-specific properties of an existing VKCS bare metal flavor.
---

# vkcs_baremetal_flavor

This resource manages project-specific properties of an existing bare metal flavor. The flavor itself is not created or deleted.

## Example Usage

```terraform
data "vkcs_baremetal_flavor" "selected" {
  name = "BM_CX301_N_BOND"
}

resource "vkcs_baremetal_flavor" "main" {
  id           = data.vkcs_baremetal_flavor.selected.id
  display_name = "Terraform managed bare metal flavor"
}

data "vkcs_baremetal_flavor" "updated" {
  id = vkcs_baremetal_flavor.main.id
}

output "display_name" {
  value = data.vkcs_baremetal_flavor.updated.display_name
}
```

## Argument Reference

- `display_name` **required** *string* &rarr; The project-specific display name of the flavor.

- `id` **required** *string* &rarr; The UUID of the existing flavor.

- `region` optional *string* &rarr; The region to fetch the bare metal flavor from, defaults to the provider's region.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

- `bond_vlan_capable` *boolean* &rarr; Bond and VLAN capable.

- `cpu_cores` *number* &rarr; CPU core count including hyper-threading.

- `cpu_model` *string* &rarr; The CPU model.

- `name` *string* &rarr; The name of the flavor.

- `ram_size` *number* &rarr; RAM in gigabytes.

- `ssd_size` *number* &rarr; SSD size in gigabytes.

- `hdd_size` *number* &rarr; HDD size in gigabytes.

Destroying the resource clears the project-specific display name. The flavor is not deleted.

## Import

An existing flavor can be imported using its ID:

```shell
terraform import vkcs_baremetal_flavor.main <flavor-id>
```
