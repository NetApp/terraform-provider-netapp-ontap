---
page_title: "ONTAP: Cluster Schedule"
subcategory: "Cluster"
description: |-
  Cluster schedule resource
---

# Resource Cluster Schedule

Create/Modify/Delete a job schedule in a cluster.

## Supported Platforms
* On-perm ONTAP system 9.6 or higher
* Amazon FSx for NetApp ONTAP

## Example Usage

```terraform
# Create a job schedule using cron type
resource "netapp-ontap_cluster_schedule_resource" "cs_example1" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "cs_test_cron"
  cron = {
    minutes = [1, 2, 3, 4]
    hours = [10]
    days = [1, 2]
    months = [6, 7]
    weekdays = [1, 3, 4]
  }
}

# Create a job schedule using interval type
resource "netapp-ontap_cluster_schedule_resource" "cs_example2" {
  # required to know which system to interface with
  cx_profile_name = "cluster4"
  name = "cs_test_interval"
  interval = "PT8M30S"
}
```

## Argument Reference

### Required

- `cx_profile_name` (String) Connection profile name
- `name` (String) The name of the cluster schedule

### Optional

- `cron` (Attributes) (see [below for nested schema](#nestedatt--cron))
- `interval` (String) Cluster schedule interval

### Read-Only

- `id` (String) Cluster/Job schedule identifier

<a id="nestedatt--cron"></a>
### Nested Schema for `cron`

Optional:

- `days` (Set of Number) List of cluster schedule days
- `hours` (Set of Number) List of cluster schedule hours
- `minutes` (Set of Number) List of cluster schedule minutes
- `months` (Set of Number) List of cluster schedule months
- `weekdays` (Set of Number) List of cluster schedule weekdays

## Import
This Resource supports import, which allows you to import existing cluster job schedule into the state of this resoruce.
Import require a unique ID composed of the schedule job name and cx_profile_name, separated by a comma.

 id = `name`,`cx_profile_name`

 ### Terraform Import

 For example
 ```shell
  terraform import netapp-ontap_cluster_schedule_resource.example job1,cluster4
 ```

!> The terraform import CLI command can only import resources into the state. Importing via the CLI does not generate configuration. If you want to generate the accompanying configuration for imported resources, use the import block instead.

### Terrafomr Import Block
This requires Terraform 1.5 or higher, and will auto create the configuration for you

First create the block
```terraform
import {
  to = netapp-ontap_cluster_schedule_resource.example.schedulejob_import
  id = "job1,cluster4"
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
# __generated__ by Terraform from "job1,cluster4"
resource "netapp-ontap_cluster_schedule_resource.example" "schedulejob_import" {
  cx_profile_name = "cluster4"
  name       = "job1"
}
```
