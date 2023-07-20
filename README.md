# Sync

## Description

Sync is a CLI that synchronizes two directories: a source directory and a destination directory.
The program minimizes the overall copy operation. The source folder is considered source of truth during the synchronization.
If the destination folder doesn't exist it will be created unless the source folder is empty.
## Build

If you have make installed:
```shell
make build
```
this will build the binary in the root directory.

If you don't have make:
```shell
cd cmd/sync
go build
```
this will build the binary in cmd/sync

## Run

### usage
```shell
sync -s path_to_source_dir -d path_to_destination_dir
```