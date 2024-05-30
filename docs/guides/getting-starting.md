---
page_title: "Getting started with the NetApp Ontap Provider"
subcategory: ""
description: |-
---

# Getting Started With The NetApp ONTAP Provider

Before getting started, you will need:
* ONTAP 9.6 or later
* Terraform 1.4 or later

This Provide will work with on-prem ONTAP system and Amazon FSx for NetApp ONTAP.

## Overview
This guide will walk you though 
* Installing Terraform
* Installing the NetApp ONTAP Provider
* Creating a connection profile
* Creating a volume
* Destroying a volume
* Handling resource dependencies

## Install Terraform
Please follow the instructions on the [Terraform website](https://learn.hashicorp.com/tutorials/terraform/install-cli) to install Terraform.

## Install The NetApp ONTAP Provider
Now that you have installed Terraform, you can install the NetApp ONTAP Provider.
First make a new directory for your Terraform configuration and change into that directory.

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

## Building Infrastructure

### Create A Connection Profile
Next we need to create a connection profile. This is a configuration file that tells the provider how to connect to your ONTAP system.
In your `provider.tf` file, add the following configuration:

* name - A name for the connection profile
* hostname - The hostname or IP address of the ONTAP system
* username - The username to use to connect to the ONTAP system
* password - The password to use to connect to the ONTAP system
* validate_certs - Whether to validate the SSL certificate of the ONTAP system

Using var.variable will prompt for the value when you run `terraform apply`. Based off the information in the Variables File section.
For Password and secrets, we recommend using Vault. Please see the [Inject secrets into Terraform using the Vault provider](https://developer.hashicorp.com/terraform/tutorials/secrets/secrets-vault) documentation for more information.

```terraform
  connection_profiles = [
    {
      name = "cluster1"
      hostname = "********219"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
  ]
 ```
This is all you'll need for your `provider.tf` to connect to your ONTAP system.

### Variables File
In terraform we can create a variables file to store variables that we want to be prompted for when we run `terraform apply`.
To do this create a file called `variables.tf` in the same directory as your `provider.tf` file.
```terraform
# Terraform will prompt for values, unless a tfvars file is present.
variable "username" {
    type = string
}
variable "password" {
    type = string
    sensitive = true
}
variable "validate_certs" {
    type = bool
}

```

### Create A Volume
Now let us create a volume. First, you'll want to have the documentation for [netapp-ontap_volume](https://registry.terraform.io/providers/NetApp/netapp-ontap/latest/docs/resources/storage_volume_resource) open in another tab.
This will show you all the configuration options for the volume resource, including examples.

We are just going to make a volume with the required variables
* cx_profile_name - The name of the connection profile we created earlier
* name - The name of the volume
* svm_name - The name of the SVM to create the volume in
* aggregates - A list of aggregates to create the volume on
* space.size - The size of the volume
* space.size_unit - The unit of the size of the volume

When calling a resource in Terraform, you need to specify the resource type in this case `netapp-ontap_volume`, the name of the resource `volume1`.
The type and name of the resource must be unique in your configuration, and are used to identify the resource in the state file.

```terraform
resource "netapp-ontap_volume" "volume1" {
  cx_profile_name = "cluster4"
  name = "terraformTest5"
  svm_name = "terraformSVM"
  aggregates = [
    {
      name = "aggr2"
    },
  ]
  space = {
    size = 20
    size_unit = "mb"
  }
}
```

With this you have everything need to create a volume. Now run `terraform init` to initialize the provider and download the required plugins. 
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

Terraform will perform the following actions:

  # netapp-ontap_volume.volume1 will be created
  + resource "netapp-ontap_volume" "volume1" {
      + aggregates       = [
          + {
              + name = "aggr2"
            },
        ]
      + analytics        = (known after apply)
      + comment          = (known after apply)
      + cx_profile_name  = "cluster4"
      + efficiency       = (known after apply)
      + encryption       = (known after apply)
      + id               = (known after apply)
      + language         = (known after apply)
      + name             = "terraformTest5"
      + nas              = (known after apply)
      + qos_policy_group = (known after apply)
      + snaplock         = (known after apply)
      + snapshot_policy  = (known after apply)
      + space            = {
          + logical_space          = (known after apply)
          + percent_snapshot_space = (known after apply)
          + size                   = 20
          + size_unit              = "mb"
        }
      + space_guarantee  = (known after apply)
      + state            = (known after apply)
      + svm_name         = "terraformSVM"
      + tiering          = (known after apply)
      + type             = (known after apply)
    }

Plan: 1 to add, 0 to change, 0 to destroy.

──────────────────────────────────────────────────────────────────────────────────

Note: You didn't use the -out option to save this plan, so Terraform can't
guarantee to take exactly these actions if you run "terraform apply" now.

```

You can now run `Terraform apply` to create the volume.

```bash
$ terraform apply
var.password
  Enter a value: 

var.username
  Enter a value: admin

var.validate_certs
  Enter a value: false


Terraform used the selected providers to generate the following execution plan.
Resource actions are indicated with the following symbols:
  + create

Terraform will perform the following actions:

  # netapp-ontap_volume.volume1 will be created
  + resource "netapp-ontap_volume" "volume1" {
      + aggregates       = [
          + {
              + name = "aggr2"
            },
        ]
      + analytics        = (known after apply)
      + comment          = (known after apply)
      + cx_profile_name  = "cluster4"
      + efficiency       = (known after apply)
      + encryption       = (known after apply)
      + id               = (known after apply)
      + language         = (known after apply)
      + name             = "terraformTest5"
      + nas              = (known after apply)
      + qos_policy_group = (known after apply)
      + snaplock         = (known after apply)
      + snapshot_policy  = (known after apply)
      + space            = {
          + logical_space          = (known after apply)
          + percent_snapshot_space = (known after apply)
          + size                   = 20
          + size_unit              = "mb"
        }
      + space_guarantee  = (known after apply)
      + state            = (known after apply)
      + svm_name         = "terraformSVM"
      + tiering          = (known after apply)
      + type             = (known after apply)
    }

Plan: 1 to add, 0 to change, 0 to destroy.

Do you want to perform these actions?
  Terraform will perform the actions described above.
  Only 'yes' will be accepted to approve.

  Enter a value: yes

netapp-ontap_volume.volume1: Creating...
netapp-ontap_volume.volume1: Creation complete after 2s [id=b6742203-7f43-11ee-8c83-005056b34578]

Apply complete! Resources: 1 added, 0 changed, 0 destroyed.
```

This will create a volume on your ONTAP system. You can verify this by logging into your ONTAP system and running `volume show -vserver terraformSVM -volume terraformTest5`
```bash
ontap_cluster_1::> volume show -vserver terraformSVM -volume terraformTest5

                                      Vserver Name: terraformSVM
                                       Volume Name: terraformTest5
                                    Aggregate Name: aggr2
     List of Aggregates for FlexGroup Constituents: aggr2
                                   Encryption Type: none
                  List of Nodes Hosting the Volume: ontap_cluster_1-01
                                       Volume Size: 20MB
                                Volume Data Set ID: 4075
                         Volume Master Data Set ID: 2157109356
...
```

Also in this directory terraform has created a file called `terraform.tfstate` which contains the state of your infrastructure. This is used to track changes to your infrastructure.

You can cat this file to see the state of your infrastructure.

```bash
cat terraform.tfstate
{
  "version": 4,
  "terraform_version": "1.4.6",
  "serial": 3,
  "lineage": "83b85278-3541-a0dd-60b8-68fca5e9d218",
  "outputs": {},
  "resources": [
    {
      "mode": "managed",
      "type": "netapp-ontap_volume",
      "name": "volume1",
      "provider": "provider[\"registry.terraform.io/netapp/netapp-ontap\"]",
      "instances": [
        {
          "schema_version": 0,
          "attributes": {
            "aggregates": [
              {
                "name": "aggr2"
              }
            ],
            "analytics": {
              "state": "off"
            },
            "comment": "",
            "cx_profile_name": "cluster4",
            "efficiency": {
              "compression": "none",
              "policy_name": "-"
            },
            "encryption": false,
            "id": "b6742203-7f43-11ee-8c83-005056b34578",
            "language": "c.utf_8",
            "name": "terraformTest5",
            "nas": {
              "export_policy_name": "default",
              "group_id": 0,
              "junction_path": "",
              "security_style": "unix",
              "unix_permissions": 755,
              "user_id": 0
            },
            "qos_policy_group": "",
            "snaplock": {
              "type": "non_snaplock"
            },
            "snapshot_policy": "default",
            "space": {
              "logical_space": {
                "enforcement": false,
                "reporting": false
              },
              "percent_snapshot_space": 5,
              "size": 20,
              "size_unit": "mb"
            },
            "space_guarantee": "volume",
            "state": "online",
            "svm_name": "terraformSVM",
            "tiering": {
              "minimum_cooling_days": 0,
              "policy_name": "none"
            },
            "type": "rw"
          },
          "sensitive_attributes": []
        }
      ]
    }
  ],
  "check_results": null
}
```

## Destroying Infrastructure
Now that we have a volume managed by Terraform, we can destroy it. To do this, we can run `terraform destroy` and Terraform will destroy the volume.

```bash
terraform destroy
var.password
  Enter a value: 

var.username
  Enter a value: admin

var.validate_certs
  Enter a value: false

netapp-ontap_volume.volume1: Refreshing state... [id=b6742203-7f43-11ee-8c83-005056b34578]

Terraform used the selected providers to generate the following execution plan.
Resource actions are indicated with the following symbols:
  - destroy

Terraform will perform the following actions:

  # netapp-ontap_volume.volume1 will be destroyed
  - resource "netapp-ontap_volume" "volume1" {
      - aggregates      = [
          - {
              - name = "aggr2" -> null
            },
        ] -> null
      - analytics       = {
          - state = "off" -> null
        } -> null
      - cx_profile_name = "cluster4" -> null
      - efficiency      = {
          - compression = "none" -> null
          - policy_name = "-" -> null
        } -> null
      - encryption      = false -> null
      - id              = "b6742203-7f43-11ee-8c83-005056b34578" -> null
      - language        = "c.utf_8" -> null
      - name            = "terraformTest5" -> null
      - nas             = {
          - export_policy_name = "default" -> null
          - group_id           = 0 -> null
          - junction_path      = "" -> null
          - security_style     = "unix" -> null
          - unix_permissions   = 755 -> null
          - user_id            = 0 -> null
        } -> null
      - snaplock        = {
          - type = "non_snaplock" -> null
        } -> null
      - snapshot_policy = "default" -> null
      - space           = {
          - logical_space          = {
              - enforcement = false -> null
              - reporting   = false -> null
            } -> null
          - percent_snapshot_space = 5 -> null
          - size                   = 20 -> null
          - size_unit              = "mb" -> null
        } -> null
      - space_guarantee = "volume" -> null
      - state           = "online" -> null
      - svm_name        = "terraformSVM" -> null
      - tiering         = {
          - minimum_cooling_days = 0 -> null
          - policy_name          = "none" -> null
        } -> null
      - type            = "rw" -> null
    }

Plan: 0 to add, 0 to change, 1 to destroy.

Do you really want to destroy all resources?
  Terraform will destroy all your managed infrastructure, as shown above.
  There is no undo. Only 'yes' will be accepted to confirm.

  Enter a value: yes

netapp-ontap_volume.volume1: Destroying... [id=b6742203-7f43-11ee-8c83-005056b34578]
netapp-ontap_volume.volume1: Destruction complete after 1s

Destroy complete! Resources: 1 destroyed.
```

You can confirm that the volume has been destroyed by running `volume show -vserver terraformSVM -volume terraformTest5` on your ONTAP system.

```bash
ontap_cluster_1::> volume show -vserver terraformSVM -volume terraformTest5
There are no entries matching your query.
```

## Handling resource dependencies
Now that you have created a volume, let's say you wanted to create a volume and create a snapshot for that volume. Something like this will fail.
The reason for this is that the volume resource is not yet created, so the snapshot resource cannot find the volume to create a snapshot for.
Terrafrom will create all resources at the same time, so it will try to create the snapshot before the volume is created.

```terraform
resource "netapp-ontap_volume" "volume1" {
  cx_profile_name = "cluster4"
  name            = "terraformTest5"
  svm_name        = "terraformSVM"
  aggregates      = [
    {
      name = "aggr2"
    },
  ]
  space = {
    size      = 20
    size_unit = "mb"
  }
}

resource "netapp-ontap_volume_snapshot" "snapshot1" {
  cx_profile_name = "cluster4"
  name = "snaptest"
  volume_name = "terraformTest5"
  svm_name = "terraformSVM"
}
```
You can see this by running `terraform graph | dot -Tsvg > graph.svg` and opening the graph.svg file.

<img src="https://github.com/NetApp/terraform-provider-netapp-ontap/tree/integration/main/docs/guides/graph.svg">

In this image, you scan see both volume and snapshot resources are created at the same time.

To tell Terraform that the snapshot resource depends on the volume resource, we can use an expression reference. 
In this case `volume_name = netapp-ontap_volume.volume1.name` which tells Terraform that the snapshot resource depends on the volume resource, and to wait until the volume resource is created before creating the snapshot resource.

```terraform
resource "netapp-ontap_volume" "volume1" {
  cx_profile_name = "cluster4"
  name = "terraformTest2"
  svm_name = "terraformSVM"
  aggregates = [
    {
      name = "aggr2"
    },
  ]
  space = {
    size = 20
    size_unit = "mb"
  }
}

resource "netapp-ontap_volume_snapshot" "snapshot1" {
  cx_profile_name = "cluster4"
  name = "snaptest"
  volume_name = netapp-ontap_volume.volume1.name
  svm_name = "terraformSVM"
}
````

If we run `terraform graph | dot -Tsvg > graph.svg` we can see that the snapshot resource now depends on the volume resource.

<img src="https://github.com/NetApp/terraform-provider-netapp-ontap/tree/integration/main/docs/guides/graph2.svg">

You can now run `terraform apply` to create the volume and snapshot.
