terraform {
  required_providers {
    netapp-ontap = {
      source = "NetApp/netapp-ontap"
      version = "0.0.1"
    }
  }
}


provider "netapp-ontap" {
  # A connection profile defines how to interface with an ONTAP cluster or vserver.
  # At least one is required.
  connection_profiles = [
    {
      name = "cluster1"
      hostname = "10.193.78.219"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    },
    {
      name = "cluster2"
      hostname = "10.193.78.222"
      username = var.username
      password = var.password
      validate_certs = var.validate_certs
    }
  ]
}
