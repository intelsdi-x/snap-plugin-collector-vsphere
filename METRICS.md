# Metrics collected by vSphere plugin

Plugin allows user to collect metrics retrieved directly from vCenter server, using `perfCounters` - internal vSphere cluster monitoring mechanism.  
Current implementation collects data for two types of entities - Host and Virtual Machine, with 4 metric groups (`cpu`, `mem`, `net` for Host and `virtualDisk` for VM). 
Metric list relies on `VMware vCenter Server 6.5.0`, but most of them are available for previous versions (included in table below).

## Namespace

Namespaces destriptions contain following dynamic elements:
* `<hostname>` - name of the hosts (IP by default)
* `<vm>` - name of the VM
* `<metric_group>` - group of vSphere metrics (`cpu`, `mem`, etc. - described in paragraphs)
* `<instance>` - instance of the metric. Most of the metrics has only one instance, but some of them are multiple-instanced and allow you to collect more detailed results - for example CPU idle time can be measured per-core (core = instance) or overall (aggregated metric).
For multiple-instanced metrics, you can set `<instance>` as name of the instance (i.e. core number), or use `aggr` to get only aggregated metric or `*` to retrieve both per-instance data and aggregated metric. 
Instance informations are included in metric tables below. 

Tables contain per-metric instance information, internal vSphere `perfCounter` name and vCenter API versions from which the specific counter is available (if mentioned in vSphere documentation).
Some metrics does not require `perfCounters`, since they are static and pre-defined (i.e. host memory).

## Host metrics
Namespaces for host metrics are built in the following way:
`/intel/vmware/vsphere/host/<hostname>/<metric_group>/<instance>/metric_name`

### CPU metric group
Namespace metric group prefix: `cpu`

| Metric name |Unit| Instances          | perfCounter |API version| Description | 
|-------------|-|--------------------|------|------------|-|
| idle |%| `aggr`, per core| cpu.usage.average (100 - usage)  |`>5.0`| Total & per-core time CPU spent in an idle as a percentage during the last 20s || 
| wait |%| `aggr` | cpu.latency.average  |`>5.0`| Total time CPU spent in a wait state as a percentage during the last 20s||
| load |number| `aggr` | rescpu.actav1.latest | `>5.0` | CPU load average over last 1 minute ||

Namespace examples:
* All per-instance and aggregated CPU metrics for all hosts:

  `/intel/vmware/vsphere/host/*/cpu/*`

* CPU aggregated metrics for host `1.1.1.1`:

  `/inte/vmware/vsphere/host/1.1.1.1/cpu/aggr/*`

### Memory metric group
Namespace metric group prefix: `mem`

| Metric name |Unit| Instances          | perfCounter |API version| Description | 
|-------------|-|--------------------|------|------------|-|
| available |MB| `0` | *static*  || Host memory size ||
| usage |MB| `0` | mem.consumed.average  |`>5.5`| Host memory usage || 
| free |MB| `0` | mem.consumed.average (available - used)   |`>5.5`| Host memory free|| 


Namespace examples:
* All available memory metrics for all hosts:

  `/intel/vmware/vsphere/host/*/mem/*`

* Memory available metric for all hosts:

  `/inte/vmware/vsphere/host/*/mem/*/available` 

### Network metric group
Namespace metric group prefix: `net`

| Metric name |Unit| Instances          | perfCounter |API version| Description | 
|-------------|-|--------------------|------|------------|-|
| packets_tx |num| `aggr`, per VM NIC | net.packetsTx.summation  |`>5.0`| Number of packets transmitted during the last 20s ||
| packets_rx |num| `aggr`, per VM NIC | net.packetsRx.summation |`>5.0`|Number of packets received during the last 20s || 
| kbrate_tx |kBps| `aggr`, per VM NIC | net.bytesTx.summation    |`>5.0`| Average amount of data transmitted per second during the last 20s|| 
| kbrate_rx |kBps| `aggr`, per VM NIC | net.bytesRx.summation  |`>5.0`| Average amount of data received per second during the last 20s


Namespace examples:
* All available network metrics for all hosts:

  `/intel/vmware/vsphere/host/*/net/*`

* packets_tx for `vmnic0` on host `1.1.1.1`

  `/intel/vmware/vsphere/host/1.1.1.1/net/vmnic0/packets_tx` 


## VMDK metrics
Namespaces for virtual machine disks metrics are built in the following way:
`/intel/vmware/vsphere/host/<hostname>/vm/<vm>/<metric_group>/<instance>/metric_name`

### Virtual Disk metric group
Namespace metric group prefix: `virtualDisk`

| Metric name |Unit| Instances          | perfCounter |API version| Description | 
|-------------|-|--------------------|------|------------|-|
| read_throughput |kBps| `aggr`, per SCSI | virtualDisk.read.average  |`>5.0`| Rate of reading data from the virtual disk ||
| write_throughput |kBps| `aggr`, per SCSI | virtualDisk.write.average |`>5.0`| Rate of reading data from the virtual disk || 
| read_iops |num| per SCSI | virtualDisk.numberReadAveraged.average    |`>5.0`| Average number of read commands issued per second to the virtual disk during the last 20s|| 
| write_iops |num| per SCSI | virtualDisk.numberWriteAveraged.average  |`>5.0`| Average number of write commands issued per second to the virtual disk during the last 20s
| read_latency |ms| per SCSI | virtualDisk.totalReadLatency.average  |`>5.0`| Average latency for writing to virtual disk
| write_latency |ms| per SCSI | virtualDisk.totalWriteLatency.average  |`>5.0`| Average latency for reading from virtual disk

Namespace examples:
* All available VMDK metrics for all hosts:

  `/intel/vmware/vsphere/host/*/vm/*`

* Read throughput for VM `vm1` on host `1.1.1.1` for `scsi0:0` 

  `/intel/vmware/vsphere/host/1.1.1.1/vm/vm1/virtualDisk/scsi0:0/read_throughput`
