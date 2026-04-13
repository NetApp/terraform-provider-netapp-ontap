---
page_title: "Setup to use ONTAP-mode in Google Cloud NetApp Volumes"
subcategory: ""
description: |-
---

# Setup to use ONTAP-mode in Google Cloud NetApp Volumes

## What is ONTAP-mode?

Google Cloud NetApp Volumes is a managed NFS/SMB/iSCSI/NVMe storage service in Google Cloud. It is built on ONTAP technology, but is consumed as a service in Google Cloud. It comes in multiple [service levels](https://docs.cloud.google.com/netapp/volumes/docs/discover/service-levels) with different capabilities. The Flex Unified service level comes with two management models, see [Default-mode versus ONTAP-mode](https://docs.cloud.google.com/netapp/volumes/docs/discover/features#flex_unified_default-mode_versus_ontap-mode).

All service levels except Flex Unified ONTAP-mode are managed through Google Cloud Console, gcloud, Google APIs and the [google](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/netapp_storage_pool) Terraform provider (see all resources starting with "netapp_"). If you wish to manage any service level other than Flex Unified **ONTAP-mode**, stop reading here and head over to:

* [google](https://registry.terraform.io/providers/hashicorp/google/latest/docs/resources/netapp_storage_pool) Terraform provider documentation
* [Google Cloud NetApp Volumes - Terraform integration blog](https://community.netapp.com/t5/Tech-ONTAP-Blogs/Google-Cloud-NetApp-Volumes-Terraform-integration/ba-p/450777)

If you want to manage resources in Flex Unified ONTAP-mode using Terraform, this page is for you.

## ONTAP-mode management concept

In ONTAP-mode, the storage pool is deployed through Google APIs (Cloud Console, gcloud, google provider). This deploys an ONTAP cluster for you in Google Cloud. All pool management happens through Google means, but all the resources inside the pool like volumes, shares, export policies, snapshots and SnapMirror replications are managed through ONTAP APIs. The `netapp-ontap` provider can do this for you.

All ONTAP API calls need to go through an Google API proxy specific to the individual ONTAP-mode pool. These API calls need to be authenticated using Google auth. This provides a highly secure, strongly authenticated and transport encrypted API access to the ONTAP cluster without the requirement of direct network access. As long as you can reach the public Google API endpoint, you can manage your pool. This `netapp-ontap` Terraform provider knows how to use the Google API proxy to talk to your ONTAP-mode pool.

The pools comes with a pre-deployed SVM, data aggregate, LIFs and an NFS server. Other resources need to be deployed by you.

See [About ONTAP-mode](https://docs.cloud.google.com/netapp/volumes/docs/ontap/overview#about_ontap-mode) for more details.

## Authentication

To authenticate the API calls to the Google API proxy, this provider uses Application Default Credentials (ADC). ADCs are commonly used by most Google tools to provide flexible authentication options through Google IDs or service account credentials provided through Workload Identity or key files.

To learn about ADC, see [How Application Default Credentials works](https://docs.cloud.google.com/docs/authentication/application-default-credentials). The Google identity you use needs roles/netapp.admin IAM permissions ([details](https://docs.cloud.google.com/netapp/volumes/docs/plan-and-prepare/iam)).

High-level guidance: If the `google` provider works in your runtime environment, the `netapp-ontap` provider will do too.

## Requirements

To create ONTAP-mode Flex Unified pools, you need google-beta provider >= v7.23.0.

Support in google provider requires v7.27.0 or later.

## How to use the netapp-ontap provider with an ONTAP-mode pool

For every ONTAP-mode storage pool you want to manage, you need to create an individual connection profile.  Connection profiles for ONTAP-mode pools require a `google_netapp_unified_pool` parameter block to specify pool detail.

The following example shows HCL code which creates an ONTAP-mode storage pool using the `google` provider, configures a `netapp-ontap` provider connection profile and queries the SVM to fetch the SVM name, the aggregate name and all interface IPs.

```terraform
locals {
  # Set your environment details here
  project      = "your-google-project"
  region       = "us-east1-b"
  zone         = "us-east1-b"
  network      = "your-vpc"
  # Leave host_project empty for standalone VPCs or specify host project for shared-VPCs
  host_project = ""
  pool_name    = "ontap-mode-pool"

  # Pool related variables
  cx_name        = google_netapp_storage_pool.unified-ontap-mode.id
  svm_name       = data.netapp-ontap_svms.unified-ontap-mode-svms.svms[0].name
  aggregate_name = data.netapp-ontap_svms.unified-ontap-mode-svms.svms[0].aggregates[0]
  nas_lifs       = toset([for lif in data.netapp-ontap_network_ip_interfaces.unified-ontap-mode-lifs.ip_interfaces : lif.ip.address if lif.service_policy == "gcnv-data-files-policy"])
  san_lifs       = toset([for lif in data.netapp-ontap_network_ip_interfaces.unified-ontap-mode-lifs.ip_interfaces : lif.ip.address if lif.service_policy == "gcnv-data-block-policy"])
  ic_lifs        = toset([for lif in data.netapp-ontap_network_ip_interfaces.unified-ontap-mode-lifs.ip_interfaces : lif.ip.address if lif.service_policy == "default-intercluster"])
}

provider "google" {
  project = local.project
  region  = local.region
}

# Get details of exiting VPC
data "google_compute_network" "vpc" {
  name    = local.network
  project = local.host_project == "" ? local.project : local.host_project
}

# Creating an ONTAP-mode GCNV Storage Pool
resource "google_netapp_storage_pool" "unified-ontap-mode" {
  # provider                   = google-beta
  name                       = local.pool_name
  location                   = local.zone
  service_level              = "FLEX"
  type                       = "UNIFIED"
  mode                       = "ONTAP"
  capacity_gib               = "1024"
  custom_performance_enabled = true
  total_throughput_mibps     = 64
  total_iops                 = 1024
  network                    = data.google_compute_network.vpc.id
}

# Creates GCNV connection profile for NetApp ONTAP provider
provider "netapp-ontap" {
  connection_profiles = [
    {
      # We use the Google URN for the connection profile name
      # The URN is only available after the pool is created
      # This ensures that the connection profile is initialized after the pool is created
      # If you initialize the profile before the pool is created, you will get an error
      name = local.cx_name
      google_netapp_unified_pool = {
        project_id      = google_netapp_storage_pool.unified-ontap-mode.project
        location        = google_netapp_storage_pool.unified-ontap-mode.location
        storage_pool    = google_netapp_storage_pool.unified-ontap-mode.name
        # custom_base_url = "https://netapp.googleapis.com/v1beta1"
      }
    }
  ]
}

# Get SVMs from the pool. Many resources require you to specify the SVM and aggregate name.
# This datasource is fetching them to store them in local variables for easy reference.
data "netapp-ontap_svms" "unified-ontap-mode-svms" {
  cx_profile_name = local.cx_name
}

# Get LIF IP addresses for the pool. For client access you will need these IPs.
# This datasource is fetching the IPs to store them in local variables for easy reference.
data "netapp-ontap_network_ip_interfaces" "unified-ontap-mode-lifs" {
  cx_profile_name = local.cx_name
}

# Optional: Output the SVM, aggregate and LIF IP addresses.
output "svm_name" {
  description = "SVM name"
  value       = local.svm_name
}

output "aggregate_name" {
  description = "Aggregate name"
  value       = local.aggregate_name
}

output "nas_lifs" {
  description = "List of NAS IP addresses"
  value       = local.nas_lifs
}

output "san_lifs" {
  description = "List of SAN IP addresses"
  value       = local.san_lifs
}

output "ic_lifs" {
  description = "List of intercluster IP addresses"
  value       = local.ic_lifs
}
```
