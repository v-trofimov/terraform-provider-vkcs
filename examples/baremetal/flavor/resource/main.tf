data "vkcs_baremetal_flavor" "selected" {
  name = "BM_CX301_N_BOND"
}

# Get the current project-specific display name.
output "display_name_before" {
  value = data.vkcs_baremetal_flavor.selected.display_name
}

# Write a project-specific display name.
resource "vkcs_baremetal_flavor_display_name" "main" {
  id           = data.vkcs_baremetal_flavor.selected.id
  display_name = "Terraform managed bare metal flavor"
}

# Rewrite the display name by changing the value above and running:
# terraform apply

# The data source reads the value written by the resource.
data "vkcs_baremetal_flavor" "updated" {
  id = vkcs_baremetal_flavor_display_name.main.id
}

output "display_name" {
  value = data.vkcs_baremetal_flavor.updated.display_name
}

# Reset the display name by running:
# terraform destroy
# Destroying this resource clears the project-specific display name.
