# Snap collector plugin - VMWare vSphere

This plugin collects performance counters data for VMWare vSphere cluster, using [govmomi](https://github.com/vmware/govmomi) - library communicating with  [vCenter server](http://www.vmware.com/products/vcenter-server.html). 

1. [Getting Started](#getting-started)
  * [System Requirements](#system-requirements)
  * [Operating systems](#openrating-systems)
  * [Installation](#installation)
  * [Configuration and Usage](#configuration-and-usage)
2. [Documentation](#documentation)
  * [Collected Metrics](#collected-metrics)
  * [Examples](#examples)
  * [Roadmap](#roadmap)
3. [Community Support](#community-support)
4. [Contributing](#contributing)
5. [License](#license)
6. [Acknowledgements](#acknowledgements)

## Getting Started

In order to use this plugin you need cretendials (URL, username, password) to vCenter server - VM in vSphere cluster used for managing and retrieving metrics from entire cluster.
Plugin is designed for vSphere clusters with *performance counter* feature (from version `2.5`).

### System Requirements

* [Snap](http://github.com/intelsdi-x/snap)
* [golang 1.6+](https://golang.org/dl/) (for building)
* [snap-plugin-utilities](http://github.com/intelsdi-x/snap-plugin-utilities) (for building)

### Operating systems
All OSs currently supported by Snap:
* Linux/amd64
* Darwin/amd64

### Installation

#### To build the plugin binary:
Get the source by running a `go get` to fetch the code:
```
$ go get -d github.com/intelsdi-x/snap-plugin-collector-vsphere
```

Build the plugin by running make within the cloned repo:
```
$ cd $GOPATH/src/github.com/intelsdi-x/snap-plugin-collector-vsphere && make
```
This builds the plugin in `/build/`.


### Configuration and Usage

* Set up the [Snap framework](https://github.com/intelsdi-x/snap/blob/master/README.md#getting-started).

* Add vSphere plugin configuration information to the Task Manifest ([example](./examples/file.yaml))

* Load the plugin and create the task

## Documentation 

### Collected Metrics
This plugin has the ability to gather memory, CPU and network metrics for each host, and detailed IO metrics for each VM-mounted virtual disk from your vSphere infractructure.

**vSphere Metric Catalog**

See [METRICS.md](METRICS.md)

**Counter interval**

vCenter and ESXi servers collect data for each performance counter every 20 seconds and maintain this data for an hour. So, recommended interval is the multiple of 20 seconds to prevent ineffective API calls.
For more details, see VMWare vSphere documentation - [Retrieving performance data](http://pubs.vmware.com/vsphere-65/index.jsp?topic=%2Fcom.vmware.wssdk.pg.doc%2FPG_Performance.19.4.html) and [Performance intervals](http://pubs.vmware.com/vsphere-65/index.jsp#com.vmware.wssdk.pg.doc/PG_Performance.19.6.html) .


### Examples
Example retrieving memory metrics from hosts in snap-plugin-collector-vsphere and writing data to a file. Assuming you have snap-plugin-collector-vsphere and [snap-publisher-file](https://github.com/intelsdi-x/snap-plugin-publisher-file) binaries.  
 
Ensure [Snap daemon is running](https://github.com/intelsdi-x/snap#running-snap):
```
$ snapteld -l 1 -t 0 &
```

Load Snap plugins:
```
$ snaptel plugin load snap-plugin-collector-vsphere
$ snaptel plugin load snap-plugin-publisher-file
```

See available metrics for your system:
```
$ snaptel metric list --verbose  
NAMESPACE                                                                                        VERSION        UNIT                     DESCRIPTION
/intel/vmware/vsphere/host/[hostname]/cpu/[instance]/idle                                        1              percent                  Total & per-core time CPU spent in an idle as a percentage during the last 20s
/intel/vmware/vsphere/host/[hostname]/cpu/[instance]/wait                                        1              percent                  Total time CPU spent in a wait state as a percentage during the last 20s
/intel/vmware/vsphere/host/[hostname]/cpu/[instance]/load                                        1              percent                  CPU load average over last 1 minute
/intel/vmware/vsphere/host/[hostname]/mem/[instance]/available                                   1              megabyte                 Available memory in megabytes
/intel/vmware/vsphere/host/[hostname]/mem/[instance]/free                                        1              megabyte                 Free memory in megabytes
/intel/vmware/vsphere/host/[hostname]/mem/[instance]/usage                                       1              megabyte                 Memory usage in megabytes
/intel/vmware/vsphere/host/[hostname]/net/[instance]/kbrate_rx                                   1              kiloBytesPerSecond       Average amount of data received per second during the last 20s
/intel/vmware/vsphere/host/[hostname]/net/[instance]/kbrate_tx                                   1              kiloBytesPerSecond       Average amount of data transmitted per second during the last 20s
/intel/vmware/vsphere/host/[hostname]/net/[instance]/packets_rx                                  1              number           Number of packets received during the last 20s
/intel/vmware/vsphere/host/[hostname]/net/[instance]/packets_tx                                  1              number           Number of packets transmitted during the last 20s
/intel/vmware/vsphere/host/[hostname]/vm/[vmname]/virtualDisk/[instance]/read_iops               1              number           Read I/O operations per second
/intel/vmware/vsphere/host/[hostname]/vm/[vmname]/virtualDisk/[instance]/read_latency            1              millisecond              Read latency
/intel/vmware/vsphere/host/[hostname]/vm/[vmname]/virtualDisk/[instance]/read_throughput         1              kiloBytesPerSecond       Read throughput
/intel/vmware/vsphere/host/[hostname]/vm/[vmname]/virtualDisk/[instance]/write_iops              1              number           Write I/O operations per second
/intel/vmware/vsphere/host/[hostname]/vm/[vmname]/virtualDisk/[instance]/write_latency           1              millisecond              Write latency
/intel/vmware/vsphere/host/[hostname]/vm/[vmname]/virtualDisk/[instance]/write_throughput        1              kiloBytesPerSecond       Write throughput
```

Add URL, username, password, and vSphere cluster name to task configuration, you can use an example:
```
$ cat examples/task/file.yaml 
---
  version: 1
  schedule:
    type: "simple"
    interval: "20s"
  max-failures: 10
  workflow:
    collect:
      metrics:
        /intel/vmware/vsphere/host/*/mem/*/free: {}
        /intel/vmware/vsphere/host/*/cpu/*/idle: {}
        /intel/vmware/vsphere/host/*/vm/*/virtualDisk/*/read_iops: {}
      config:
          "/intel/vmware/vsphere":
            "url": "https://localhost/sdk"
            "username": "admin@domain"
            "password": "pass"
            "insecure": true
            "clusterName": "cluster"
      publish:
        - plugin_name: "file"
          config:
            file: "/tmp/memory.log"
```


Provide correct credentials to vCenter endpoint and create a task
```
$ snaptel task create -t examples/task/file.yaml
Using task manifest to create task
Task created
ID: 37cd9903-daf6-4e53-b15d-b9082666a830
Name: Task-37cd9903-daf6-4e53-b15d-b9082666a830
State: Running
```

See the file output (part of the output):
```
$ tail -f /tmp/memory.log
```

Or watch task values:
```
$ snaptel task watch 37cd9903-daf6-4e53-b15d-b9082666a830
```

### Roadmap
This plugin is still in active development and its metric catalog has to be extended. Items that must be covered as first are:
- [ ] Add SWAP metrics (figure out how to calculate full SWAP space) 

If you have a feature request, please add it as an [issue](https://github.com/intelsdi-x/snap-plugin-collector-vsphere/issues) 
and feel free to submit a [pull request](https://github.com/intelsdi-x/snap-plugin-collector-vsphere/pulls) that resolves it.

## Community Support
This repository is one of **many** plugins in **Snap**, the open telemetry framework. See the full project at http://github.com/intelsdi-x/snap. To reach out to other users, head to the [main framework](https://github.com/intelsdi-x/snap#community-support).

## Contributing
There's more than one way to give back, from examples to blogs to code updates. See our recommended process in [CONTRIBUTING.md](CONTRIBUTING.md).

And **thank you!** Your contribution, through code and participation, is incredibly important to us.

## License
[Snap](http://github.com:intelsdi-x/snap), along with this plugin, is an Open Source software released under the Apache 2.0 [License](LICENSE).


## Acknowledgements

* Authors: [@jjlakis](https://github.com/jjlakis/), [@mkleina](https://github.com/mkleina/)

