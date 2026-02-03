---
page_title: "SSL Certificate Authentication with NetApp ONTAP Provider"
subcategory: ""
description: |-
---

# SSL Certificate Authentication with NetApp ONTAP Provider

The NetApp ONTAP Provider supports SSL certificate-based authentication as an alternative to username/password authentication. This guide will walk you through setting up SSL certificate authentication for secure communication with your ONTAP cluster.

## Overview

Certificate-based authentication provides enhanced security by using digital certificates instead of traditional username/password credentials. 

## Prerequisites

Before setting up SSL certificate authentication, you will need:

- ONTAP 9.6 or later
- Terraform 1.4 or later
- A client certificate and private key
- The Certificate Authority (CA) certificate that signed your client certificate
- Administrative access to configure SSL certificate authentication on ONTAP

## ONTAP Configuration

Before configuring the Terraform provider, you need to set up ONTAP to accept certificate-based authentication. This involves creating a Certificate Authority (CA), generating client and server certificates, and configuring ONTAP appropriately.

### Prerequisites for ONTAP Setup

You will need:
- OpenSSL installed on your system
- Administrative access to the ONTAP cluster
- SSH or HTTPS access to ONTAP management interface

### Certificate Infrastructure Setup

The certificate setup process involves several components:

1. **Root Certificate Authority (CA)**: Creates and signs both client and server certificates
2. **Client Certificate**: Used by Terraform to authenticate with ONTAP
3. **Server Certificate**: Used by ONTAP for HTTPS connections with proper Subject Alternative Names (SAN)

### 1. Certificate Authority and Certificate Generation

You'll need to create a root CA and generate the necessary certificates. The certificates should include:

- **Root CA certificate and private key** (4096-bit for enhanced security)
- **Client certificate and private key** for Terraform authentication (with CN=tfuser)
- **Server certificate and private key** for ONTAP with proper SAN entries including the cluster IP address

### 2. Install Certificates in ONTAP

**Install the CA Certificate:**
Install the CA certificate as both server-ca and client-ca type in ONTAP to enable certificate validation.

**Create Certificate-Based User:**
Create a user account (e.g., "tfuser") that uses certificate-based authentication for both HTTP and ONTAPI applications with admin role privileges.

**Install Server Certificate:**
Install both the server certificate and its private key in ONTAP to enable secure HTTPS connections.

### 3. Configure ONTAP Web Services

**Enable SSL and Certificate Authentication:**
Configure ONTAP to use the server certificate for web services and enable both server and client SSL authentication.

**Web Service Security:**
- Enable HTTPS and optionally disable HTTP for enhanced security
- Configure the web service to use the installed server certificate
- Verify the configuration using the REST API

**User Authentication:**
Ensure certificate authentication is properly configured for the user account that Terraform will use.

## Certificate Files

You will need three certificate files for the Terraform provider configuration:

- **Client Certificate** (`cert_filepath`): Your client certificate in PEM format with CN matching the ONTAP user (e.g., CN=tfuser)
- **Private Key** (`key_filepath`): The private key corresponding to your client certificate in PEM format  
- **CA Certificate** (`ca_cert_file`): The Certificate Authority certificate used to sign the client certificate in PEM format

### Example Certificate File Structure

```
/path/to/certificates/
├── client-cert.pem      # Client certificate (terraform_client.crt)
├── client-key.pem       # Private key (terraform_client.key)
└── ca-cert.pem         # CA certificate (rootCA.crt)
```

## Provider Configuration

Configure the NetApp ONTAP Provider to use SSL certificate authentication by specifying the certificate file paths in your connection profile:

### Basic Certificate Authentication

```terraform
terraform {
  required_providers {
    netapp-ontap = {
      source = "NetApp/netapp-ontap"
      version = "~> 2.1"
    }
  }
}

provider "netapp-ontap" {
  connection_profiles = [
    {
      name = "cluster1"
      hostname = "ontap-cluster.example.com"
      # Certificate-based authentication
      cert_filepath = var.cert_filepath
      key_filepath = var.key_filepath
      ca_cert_file = var.ca_cert_file
      validate_certs = var.validate_certs
    }
  ]
}
```

### Using Variables for Certificate Paths

For better security and flexibility, use Terraform variables for certificate file paths:

```terraform
# variables.tf
# Terraform will prompt for values, unless a tfvars file is present.
variable "username" {
    type = string
    default = null
}
variable "password" {
    type = string
    sensitive = true
    default = null
}
variable "validate_certs" {
    type = bool
    default = false
}
variable "cert_filepath" {
    type = string
    default = null
    description = "Path to the client certificate file for certificate-based authentication"
}
variable "key_filepath" {
    type = string
    default = null
    description = "Path to the private key file for certificate-based authentication"
}
variable "ca_cert_file" {
    type = string
    default = null
    description = "Path to the CA certificate file for validating server certificates"
}
```

### Terraform Variables File (terraform.tfvars)

```
cert_filepath = "/root/cert/terraform_client.crt"
key_filepath = "/root/cert/terraform_client.key"
ca_cert_file = "/root/cert/rootCA.crt"
```

### Security Best Practices

### File Permissions
Ensure certificate and key files have appropriate permissions for security:

- **Private keys**: Should have 600 permissions (readable by owner only)
- **Certificates**: Can have 644 permissions (readable by all)
- **Configuration files**: Should have 600 permissions for security

Ensure proper ownership of certificate files by the user running Terraform.

## Troubleshooting

### Common Issues and Solutions

**1. Certificate Validation Failed**
```
Error: x509: certificate signed by unknown authority
```
- Verify the CA certificate path is correct and matches the CA that signed the client certificate
- Ensure `validate_certs` is set to `true` and the CA certificate is properly installed in ONTAP
- Check that the CA certificate file is readable by the Terraform process

**2. Authentication Failed**
```
Error: authentication failed
```
- Verify the certificate-based user exists in ONTAP with proper permissions
- Ensure certificate authentication is enabled for both HTTP and ONTAPI applications
- Check that the client certificate CN matches the username configured in ONTAP
- Confirm the client certificate hasn't expired

**3. Private Key/Certificate Mismatch**
```
Error: private key does not match certificate
```
- Verify that the private key corresponds to the client certificate
- Check that both files are in PEM format and properly formatted
- Ensure the files haven't been corrupted during transfer

**4. Subject Alternative Name Mismatch**
```
Error: x509: certificate is valid for [IP], not [different IP]
```
- This typically affects the server certificate installed in ONTAP
- Ensure the ONTAP server certificate includes the correct IP address or hostname in the Subject Alternative Names
- The server certificate must include all possible connection endpoints

**5. TLS Handshake Failure**
```
Error: tls: handshake failure
```
- Verify ONTAP is using the correct server certificate for web services
- Check that the server certificate hasn't expired
- Ensure the certificate chain is complete and valid

**6. File Not Found Errors**
```
Error: no such file or directory
```
- Verify that certificate file paths in the Terraform configuration are correct and absolute
- Ensure Terraform has read permissions on the certificate files and directories
- Check that the files exist at the specified locations

## Note

SSL certificate authentication is not applicable when using AWS Lambda for FSx ONTAP connections. For FSx ONTAP, continue to use the standard username/password authentication with AWS Lambda configuration.
