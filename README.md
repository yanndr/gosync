# Sync

## Description
Sync is a CLI to synchronize to directories, source and destination.
If the destination folder doesn't exist it will be created unless the source folder is empty.
## Build

If you have make installed:
```shell
make build
```
that will build the binary in the root directory.

If you don't have make:
```shell
cd cmd/sync
go build
```
that will build the binary in cmd/sync

## Run

### usage
```shell
sync -s path_to_source_dir -d path_to_destination_dir
```