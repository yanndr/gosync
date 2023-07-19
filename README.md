# Sync

## Description
Sync is a CLI to synchronize to directories, source and destination.

## Build

if you have make install :
```shell
make build
```
that will build the binary in the root directory

if you don't have make:
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