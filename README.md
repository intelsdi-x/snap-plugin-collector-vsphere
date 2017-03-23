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
This plugin has the ability to gather memory metrics for each host that appearing in your vSphere infractructure

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
NAMESPACE                                        VERSION         UNIT            DESCRIPTION
/intel/vmware/vsphere/host/[host]/memAvailable    1               megabytes       Host memory available
/intel/vmware/vsphere/host/[host]/memFree         1               megabytes       Host free memory
/intel/vmware/vsphere/host/[host]/memUsage        1               megabytes       Host memory usage
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
        /intel/vmware/vsphere/host/*/memUsage: {}
        /intel/vmware/vsphere/host/*/memFree: {}
      config:
          "/intel/vmware/vsphere":
            "url": "localhost"
            "username": "admin@domain"
            "password": "pass"
            "insecure": true
            "clusterName": "cluster"
      publish:
        - plugin_name: "file"
          config:
            file: "/tmp/memory.log"
```


Create a task
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
  {
      "data": 48201,
      "last_advertised_time": "2017-03-16T09:48:13.468157696+01:00",
      "namespace": "/intel/vmware/vsphere/host/100.1.1.1/memUsage",
      "tags": {
          "plugin_running_on": "dev"
      },
      "timestamp": "2017-03-16T09:48:11.917464846+01:00",
      "unit": "megabytes",
      "version": 1
  },
  {
      "data": 26809,
      "last_advertised_time": "2017-03-16T09:48:13.468159549+01:00",
      "namespace": "/intel/vmware/vsphere/host/100.1.1.2/memUsage",
      "tags": {
          "plugin_running_on": "dev"
      },
      "timestamp": "2017-03-16T09:48:12.169159107+01:00",
      "unit": "megabytes",
      "version": 1
  },
  {
      "data": 213849,
      "last_advertised_time": "2017-03-16T09:48:13.468161495+01:00",
      "namespace": "/intel/vmware/vsphere/host/100.1.1.1/memFree",
      "tags": {
          "plugin_running_on": "dev"
      },
      "timestamp": "2017-03-16T09:48:12.96940815+01:00",
      "unit": "megabytes",
      "version": 1
  },
  {
      "data": 235241,
      "last_advertised_time": "2017-03-16T09:48:13.468163415+01:00",
      "namespace": "/intel/vmware/vsphere/host/100.1.1.2/memFree",
      "tags": {
          "plugin_running_on": "dev"
      },
      "timestamp": "2017-03-16T09:48:13.218007203+01:00",
      "unit": "megabytes",
      "version": 1
  }

```
Plugin found 2 hosts in vSphere's cluster "**cluster**" (as given in task manifest)

### Roadmap
This plugin is still in active development and its metric catalog has to be extended. Items that must be covered as first are:
- [ ] Add more host memory metrics (swap usage, buffer, caches)
- [ ] Add CPU metrics (idle, iowait, load, steal, usertime, etc.) 
- [ ] Add dynamic VM metrics - Get memory and CPU metrics for VMs
- [ ] Extend plugin with IO metric (for vSphere datastore)
- [ ] Refactor - Reduce **QueryPerf()** calls to make plugin more scalable 

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

* Author: [@jjlakis](https://github.com/jjlakis/)

