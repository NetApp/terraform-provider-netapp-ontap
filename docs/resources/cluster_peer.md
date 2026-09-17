---
page_title: "ONTAP: Cluster Peer"
subcategory: "Cluster"
description: |-
  Cluster peer resource
---

# Resource Cluster Peer

Create/Modify/Delete a cluster peer.

## Related ONTAP commands

```commandline
* cluster peer create
* cluster peer modify
* cluster peer delete
```

## Supported Platforms

* On-prem ONTAP system 9.6 or higher
* Amazon FSx for NetApp ONTAP
* Google Cloud NetApp Volumes (GCNV) ONTAP-mode

## Example Usage

```terraform
resource "netapp-ontap_cluster_peer" "cluster_peer" {
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

### GCNV to ONTAP cluster peering

When peering a Google Cloud NetApp Volumes (GCNV) ONTAP-mode cluster with a standard ONTAP cluster,
GCNV places intercluster LIFs in the `Gcnv` IPspace while ONTAP uses `Default`. Use `ipspace` for
the local (GCNV) cluster and `peer_ipspace` for the remote (ONTAP) cluster:

```terraform
resource "netapp-ontap_cluster_peer" "gcnv_to_ontap" {
  cx_profile_name      = "gcnv_profile"
  peer_cx_profile_name = "ontap_profile"
  name = "gcnv-ontap-peer"
  remote = {
    ip_addresses = ["10.5.0.5"]
  }
  source_details = {
    ip_addresses = ["10.5.1.211", "10.5.1.210"]
  }
  passphrase        = "my-passphrase"
  peer_applications = ["snapmirror"]
  ipspace      = { name = "Gcnv" }
  peer_ipspace = { name = "Default" }
}
```

## Argument Reference

### Required

- `cx_profile_name` (String) Connection profile name
- `remote` (Attributes) (see [below for nested schema](#nestedatt--remote))
- `source_details` (Attributes) (see [below for nested schema](#nestedatt--source_details))

### Optional

- `passphrase` (String) User generated passphrase for use in authentication
- `generate_passphrase` (String) When true, ONTAP automatically generates a passphrase to authenticate cluster peer
- `name` (String) Name of the peering relationship or name of the remote peer
- `peer_applications` (String) SVM peering applications
- `peer_cx_profile_name` (String) Peer connection profile name, to be accepted from peer side to make the status OK
- `ipspace` (Attributes) IPspace for the **local** cluster peer LIFs. Required for GCNV ONTAP-mode clusters (use `"Gcnv"`). Cannot be updated after creation. (see [below for nested schema](#nestedatt--ipspace))
- `peer_ipspace` (Attributes) IPspace for the **peer** cluster peer LIFs. Use when the peer cluster uses a different IPspace than the local cluster (e.g. `"Default"` for a standard ONTAP cluster when local is GCNV). Cannot be updated after creation. (see [below for nested schema](#nestedatt--peer_ipspace))

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

### Nested Schema form `ipspace`

<a id="nestedatt--ipspace"></a>

Required:

- `name` (String) Name of the IPspace for the local cluster peer LIFs (e.g. `"Gcnv"` for GCNV ONTAP-mode, `"Default"` for standard ONTAP).

<a id="nestedatt--peer_ipspace"></a>

### Nested Schema for `peer_ipspace`

Required:

- `name` (String) Name of the IPspace for the peer cluster peer LIFs (e.g. `"Default"` for standard ONTAP when local is GCNV).


## Import

This Resource supports import, which allows you to import existing cluster peer relation into the state of this resource.
Import require a unique ID composed of the cluster name and cx_profile_name, separated by a comma.

 id = `name`,`cx_profile_name`

### Terraform Import

 For example

 ```shell
  terraform import netapp-ontap_cluster_peer.example clutername-1,cluster4
 ```

!> The terraform import CLI command can only import resources into the state. Importing via the CLI does not generate configuration. If you want to generate the accompanying configuration for imported resources, use the import block instead.

### Terraform Import Block

This requires Terraform 1.5 or higher, and will auto create the configuration for you

First create the block

```terraform
import {
  to = netapp-ontap_cluster_peer.example.cluster_import
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
resource "netapp-ontap_cluster_peer.example" "cluster_peer_import" {
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
