---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "netapp-ontap_protocols_san_lun-maps_resource Resource - terraform-provider-netapp-ontap"
subcategory: "nas"
description: |-
  ProtocolsSanLunMaps resource
---

# netapp-ontap_protocols_san_lun-maps_resource (Resource)

Create/Delete a protocols_san_lun-maps resource

### Related ONTAP commands
* lun mapping create
* lun mapping delete

## Supported Platforms
* On-perm ONTAP system 9.6 or higher
* Amazon FSx for NetApp ONTAP

## Example Usage
```
# Create a protocols_san_lun
resource "netapp-ontap_protocols_san_lun-maps_resource" "protocols_san_lun" {
  # required to know which system to interface with
  cx_profile_name = "cluster2"
  svm = {
    name = "test"
  }
  lun = {
    name = "/vol/lunTest/test"
  }
  igroup = {
    name = "test"
  }
  logical_unit_number = 1
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `cx_profile_name` (String) Connection profile name
- `igroup` (Attributes) SVM details for ProtocolsSanLunMaps (see [below for nested schema](#nestedatt--igroup))
- `lun` (Attributes) SVM details for ProtocolsSanLunMaps (see [below for nested schema](#nestedatt--lun))
- `svm` (Attributes) SVM details for ProtocolsSanLunMaps (see [below for nested schema](#nestedatt--svm))

### Optional

- `logical_unit_number` (Number) If no value is provided, ONTAP assigns the lowest available value

### Read-Only

- `id` (String) ProtocolsSanLunMaps igroup and lun UUID

<a id="nestedatt--igroup"></a>
### Nested Schema for `igroup`

Required:

- `name` (String) name of the igroup


<a id="nestedatt--lun"></a>
### Nested Schema for `lun`

Required:

- `name` (String) name of the lun


<a id="nestedatt--svm"></a>
### Nested Schema for `svm`

Required:

- `name` (String) name of the SVM

## Import
This resource supports import, which allows you to import existing protocols_san_lun-maps into the state of this resource.
Import require a unique ID composed of the protocols_san_lun-maps svm_name, igroup_name, lun_name and connection profile, separated by a comma.

id = `destination_path`,`cx_profile_name`

### Terraform Import

For example
```shell
 terraform import netapp-ontap_protocols_san_lun-maps_resource.example svm_name,igroup_name,lun_name,cluster5
```
!> The terraform import CLI command can only import resources into the state. Importing via the CLI does not generate configuration. If you want to generate the accompanying configuration for imported resources, use the import block instead.

### Terrafomr Import Block
This requires Terraform 1.5 or higher, and will auto create the configuration for you

First create the block
```terraform
import {
  to = netapp-ontap_protocols_san_lun-maps_resource.protocols_san_lun_import
  id = "svm_name,igroup_name,lun_name,cluster5"
}
```
Next run, this will auto create the configuration for you
```shell
terraform plan -generate-config-out=generated.tf
```
This will generate a file called generated.tf, which will contain the configuration for the imported resource
```terraform
# __generated__ by Terraform
# Please review these resources and move them into your main configuration files.
# __generated__ by Terraform from "svm_name,igroup_name,lun_name,cluster5"
resource "netapp-ontap_protocols_san_lun-maps_resource" "protocols_san_lun_import" {
  cx_profile_name = "cluster4"
  id = "abcd"
  igroup = {
    name = "acc_test"
  }
  logical_unit_number = 0
  lun = {
    name = "/vol/lunTest/ACC-import-lun"
  }
  svm = {
    name = "carchi-test"
  }
}
```