#!/bin/bash

NODE="nodes/nyarlathotep/"
STORAGE_LOCAL="storage/Nyarlathotep-thin/"
STORAGE_IMAGES="storage/NAS-NFS/"
CONTENT="content"

# Check current images, to avoid conflicting names
URL_BUILDER_IMAGES=`$PROXMOX_API_URL + $NODE + $STORAGE_IMAGES + $CONTENT`
URL_BUILDER_VMDISKS=`$PROXMOX_API_URL + $NODE + STORAGE_LOCAL + $CONTENT`
CURL_COMMAND=`curl -H 'Authorization: PVEAPIToken=$TOKEN_ID=$SECRET' $URL_BUILDER_IMAGES`


# Build package
GO11MODULE=off GOOS=linux go build

# Create image and instance
ops image create frontend -t proxmox
ops instance create frontend -t proxmox