# Metrics collected by vSphere plugin
The prefix of metric's namespace for hosts is `/intel/vmware/vsphere/host/<hostname>/`

Namespace    | Data Type | Unit | Description
-------------|-----------|------|----------------
memAvailable | uint64    |  MB  | Memory available on the host
memFree      | uint64    |  MB  | Host free memory
memUsage     | uint64    |  MB  | Host memory usage

The dynamic metric queries are supported. You may view the [exemplary task](./examples/file.yaml) with dynamic host.
