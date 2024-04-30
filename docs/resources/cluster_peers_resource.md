---
page_title: "ONTAP: Cluster Peers"
subcategory: "Cluster"
description: |-
  Cluster peers resource
---

# Resource Cluster Peers

Create/Modify/Delete a cluster peer.

### Related ONTAP commands
```commandline
* cluster peer create
* cluster peer modify
* cluster peer delete
```

## Supported Platforms
* On-perm ONTAP system 9.6 or higher
* Amazon FSx for NetApp ONTAP

## Example Usage

```
resource "netapp-ontap_cluster_peers_resource" "cluster_peers" {
  # required to know which system to interface with
  cx_profile_name = "cluster3"
  name = "testme"
  remote = {
    ip_addresses = ["10.10.10.10", "10.10.10.11"]
  }
  source_details = {
    ip_addresses = ["10.10.10.12"]
  }
  peer_cx_profile_name = "cluster2"
  passphrase = "12345678"
  peer_applications = ["snapmirror"]
}
```

## Argument Reference

### Required

- `cx_profile_name` (String) Connection profile name
- `remote` (Attributes) (see [below for nested schema](#nestedatt--remote))
- `source_details` (Attributes) (see [below for nested schema](#nestedatt--source_details))

### Optional

- `passphrase` (String) User generated passphrase for use in authentication
- `generate_passphrase` (String) When true, ONTAP automatically generates a passphrase to authenticate cluster peers
- `name` (String) Name of the peering relationship or name of the remote peer
- `peer_applications` (String) SVM peering applications
- `peer_cx_profile_name` (String) Peer connection profile name, to be accepted from peer side to make the status OK

### Read-Only

- `id` (String) Cluster peer relation source identifier
- `peer_id` (String) Cluster peer relation destination identifier
- `state` (String) Cluster peering state

<a id="nestedatt--remote"></a>
### Nested Schema for `remote`

Required:

- `ip_addresses` (Set of String) list of the remote ip addresses

<a id="nestedatt--source_details"></a>
### Nested Schema for `source_details`

Required:

- `ip_addresses` (Set of String) list of the remote ip addresses


## Import
This Resource supports import, which allows you to import existing cluster peer relation into the state of this resoruce.
Import require a unique ID composed of the cluster name and cx_profile_name, separated by a comma.

 id = `name`,`cx_profile_name`

 ### Terraform Import

 For example
 ```shell
  terraform import netapp-ontap_cluster_peers_resource.example clutername-1,cluster4
 ```

!> The terraform import CLI command can only import resources into the state. Importing via the CLI does not generate configuration. If you want to generate the accompanying configuration for imported resources, use the import block instead.

### Terrafomr Import Block
This requires Terraform 1.5 or higher, and will auto create the configuration for you

First create the block
```terraform
import {
  to = netapp-ontap_cluster_peers_resource.example.cluster_import
  id = "clutername-1,cluster4"
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
# __generated__ by Terraform from "clutername-1,cluster4"
resource "netapp-ontap_cluster_peers_resource.example" "cluster_peers_import" {
  cx_profile_name = "cluster3"
  name       = "test"
  generate_passphrase = false
  passphrase = "12345678"
  peer_applications = ["snapmirror"]
  peer_cx_profile_name = "cluster2"
  remote = {
    ip_addresses = [
    "10.10.10.10"
    ]
  }
  source_details = {
    ip_addresses = [
    "10.10.10.11"
    ]
  }
  state = "pending"
}
```
