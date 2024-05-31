---
page_title: "Upgrading minor version with the NetApp Ontap Provider"
subcategory: ""
description: |-
---

# Updating minor version With The NetApp ONTAP Provider

Before getting started, you will need:
* ONTAP 9.6 or later
* Terraform 1.4 or later

This Provide will work with on-prem ONTAP system and Amazon FSx for NetApp ONTAP.

## Overview
This guide will walk you though 
* Installing the latest version of NetApp ONTAP Provider
* Resuming on running NetApp ONTAP Provider

## Install The Latest Version of NetApp ONTAP Provider
Please go to the [Terraform Registry](https://registry.terraform.io/providers/NetApp/netapp-ontap/latest) to get the latest provider configuration, and copy that in to a file called `provider.tf` in the directory you just created. 
During `Terraform init` Terraform will download the provider and any required plugins.
You should have something that looks like this


```terraform
terraform {
  required_providers {
    netapp-ontap = {
      source = "NetApp/netapp-ontap"
      version = "1.0.0"
    }
  }
}

provider "netapp-ontap" {
  # Configuration options
}
```

### Resuming On Running NetApp ONTAP Provider
Now run `terraform init` to initialize the provider and download the required plugins. 
This will download the NetApp ONTAP Provider and any required plugins.

Then run `terraform plan` to get a preview.

```bash 
$ terraform plan
var.password
  Enter a value: 

var.username
  Enter a value: admin

var.validate_certs
  Enter a value: false


Terraform used the selected providers to generate the following execution plan.
Resource actions are indicated with the following symbols:
  + create
```

Please go to the [Building Infrastructure Guide](https://github.com/NetApp/terraform-provider-netapp-ontap/blob/integration/main/docs/guides/getting-starting.md#building-infrastructure) to get more information on Building Infrastructure for NetApp ONTAP Provider.
