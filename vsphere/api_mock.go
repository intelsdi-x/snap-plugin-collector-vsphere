/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2017 Intel Corporation

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vsphere

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type counterInfo struct {
	key    int32
	name   string
	group  string
	rollup string
}

type counterData struct {
	key      int32
	instance string
	data     int64
}

type mockAPI struct {
	ClientFailure       bool
	RetrieveHostsErr    bool
	RetrieveVMsErr      bool
	RetrieveCountersErr bool
	PerfQueryErr        bool
}

var (
	testHosts []mo.HostSystem
	testVMs   map[string][]mo.VirtualMachine // map[Host Reference Name]Virtual Machines

	// Fixtures with server responses for counter queries with instances and example data
	testCountersInstances []counterData

	// Fixtures for available counters
	testCountersInfo []counterInfo
)

func initFixtures() {
	// Fixtures with all hosts and VMs available on cluster
	testHosts = []mo.HostSystem{mo.HostSystem{}, mo.HostSystem{}}
	testHosts[0].Name = "1.1.1.1"
	testHosts[0].Self.Type = "HostSystem"
	testHosts[0].Self.Value = "host-1"
	testHosts[0].Hardware = &types.HostHardwareInfo{
		MemorySize: 1234567890,
	}
	testHosts[1].Name = "2.2.2.2"
	testHosts[1].Self.Type = "HostSystem"
	testHosts[1].Self.Value = "host-2"
	testHosts[1].Hardware = &types.HostHardwareInfo{
		MemorySize: 4567890123,
	}

	testVMs = make(map[string][]mo.VirtualMachine)
	testVMs["host-1"] = []mo.VirtualMachine{mo.VirtualMachine{}, mo.VirtualMachine{}}
	testVMs["host-1"][0].Name = "VM1"
	testVMs["host-1"][0].Self.Type = "VirtualMachine"
	testVMs["host-1"][0].Self.Value = "vm-1"
	testVMs["host-1"][0].Summary = types.VirtualMachineSummary{
		Runtime: types.VirtualMachineRuntimeInfo{
			Host: &testHosts[0].Self,
		},
	}
	testVMs["host-1"][1].Name = "VM2"
	testVMs["host-1"][1].Self.Type = "VirtualMachine"
	testVMs["host-1"][1].Self.Value = "vm-2"
	testVMs["host-1"][1].Summary = types.VirtualMachineSummary{
		Runtime: types.VirtualMachineRuntimeInfo{
			Host: &testHosts[0].Self,
		},
	}

	// Fixtures with all counter instances and data on server
	testCountersInstances = []counterData{
		// cpu
		counterData{key: 1, instance: "0", data: 100},
		counterData{key: 1, instance: "1", data: 110},
		counterData{key: 1, instance: "2", data: 120},
		counterData{key: 2, instance: "", data: 200},

		// rescpu
		counterData{key: 3, instance: "", data: 300},

		// mem
		counterData{key: 4, instance: "0", data: 123456},

		// net
		counterData{key: 5, instance: "eth0", data: 400},
		counterData{key: 6, instance: "eth0", data: 500},
		counterData{key: 7, instance: "eth0", data: 600},
		counterData{key: 8, instance: "eth0", data: 700},

		// virtualDisk
		counterData{key: 9, instance: "0", data: 800},
		counterData{key: 10, instance: "0", data: 900},
		counterData{key: 11, instance: "0", data: 1000},
		counterData{key: 12, instance: "0", data: 1100},
		counterData{key: 13, instance: "0", data: 1200},
		counterData{key: 14, instance: "0", data: 1300},
	}

	// Fixtures with all available counters on server
	testCountersInfo = []counterInfo{
		counterInfo{key: 1, group: "cpu", name: "usage", rollup: "average"},
		counterInfo{key: 2, group: "cpu", name: "latency", rollup: "average"},
		counterInfo{key: 3, group: "rescpu", name: "actav1", rollup: "latest"},
		counterInfo{key: 4, group: "mem", name: "consumed", rollup: "average"},
		counterInfo{key: 5, group: "net", name: "bytesTx", rollup: "average"},
		counterInfo{key: 6, group: "net", name: "bytesRx", rollup: "average"},
		counterInfo{key: 7, group: "net", name: "packetsTx", rollup: "summation"},
		counterInfo{key: 8, group: "net", name: "packetsRx", rollup: "summation"},
		counterInfo{key: 9, group: "virtualDisk", name: "numberReadAveraged", rollup: "average"},
		counterInfo{key: 10, group: "virtualDisk", name: "numberWriteAveraged", rollup: "average"},
		counterInfo{key: 11, group: "virtualDisk", name: "read", rollup: "average"},
		counterInfo{key: 12, group: "virtualDisk", name: "write", rollup: "average"},
		counterInfo{key: 13, group: "virtualDisk", name: "totalReadLatency", rollup: "average"},
		counterInfo{key: 14, group: "virtualDisk", name: "totalWriteLatency", rollup: "average"},
	}
}

// Init initializes all necessary objects to send API calls to vSphere
func (a *mockAPI) Init(ctx context.Context, url, username, password, clusterName string, insecure bool) error {
	if a.ClientFailure {
		return fmt.Errorf("unable to initialize client")
	}
	return nil
}

func (a *mockAPI) ClearCache() {
}

// RetrieveCounters retrieves vSphere cluster metric list that are available for user
func (a *mockAPI) RetrieveCounters(ctx context.Context) ([]types.PerfCounterInfo, error) {
	if a.RetrieveCountersErr {
		return nil, fmt.Errorf("test error")
	}

	testCounters := []types.PerfCounterInfo{}
	for _, c := range testCountersInfo {
		testCounters = append(testCounters, types.PerfCounterInfo{
			Key: c.key,
			NameInfo: &types.ElementDescription{
				Key: c.name,
			},
			GroupInfo: &types.ElementDescription{
				Key: c.group,
			},
			RollupType: types.PerfSummaryType(c.rollup),
		})
	}

	return testCounters, nil
}

// RetrieveCounters retrieves vSphere cluster datastore list that are available for user
func (a *mockAPI) RetrieveDatastores(ctx context.Context) ([]mo.Datastore, error) {
	return nil, nil
}

// RetrieveHosts finds all hosts on given vSphere cluster
func (a *mockAPI) RetrieveHosts(ctx context.Context) ([]mo.HostSystem, error) {
	if a.RetrieveHostsErr {
		return nil, fmt.Errorf("test error")
	}
	return testHosts, nil
}

// RetrieveVMs finds all VMs for given host
func (a *mockAPI) RetrieveVMs(ctx context.Context, host mo.HostSystem) ([]mo.VirtualMachine, error) {
	if a.RetrieveVMsErr {
		return nil, fmt.Errorf("test error")
	}
	return testVMs[host.Reference().Value], nil
}

// PerfQuery retrieves all metric data for provided query specs
// This method builds query perf response from provided query specs using
// testCountersInstances fixtures
func (a *mockAPI) PerfQuery(ctx context.Context, querySpecs []types.PerfQuerySpec) (*types.QueryPerfResponse, error) {
	if a.PerfQueryErr {
		return nil, fmt.Errorf("test error")
	}

	result := &types.QueryPerfResponse{
		Returnval: []types.BasePerfEntityMetricBase{},
	}

	for _, querySpec := range querySpecs {
		for _, metricID := range querySpec.MetricId {
			// Build metric entity info
			var entity types.BasePerfEntityMetricBase
			entity = &types.PerfEntityMetric{}
			entity.(*types.PerfEntityMetric).Entity = querySpec.Entity
			entity.(*types.PerfEntityMetric).Value = []types.BasePerfMetricSeries{}

			// Loop through testCountersInstances and match fixtures
			for _, data := range testCountersInstances {
				if data.key == metricID.CounterId {
					if data.instance == metricID.Instance || metricID.Instance == "*" {
						// Build metric instance and data
						var instance types.BasePerfMetricSeries
						instance = &types.PerfMetricIntSeries{}
						instance.(*types.PerfMetricIntSeries).Id = types.PerfMetricId{Instance: data.instance, CounterId: data.key}
						instance.(*types.PerfMetricIntSeries).Value = []int64{data.data}

						// Append complete metric instance to entity
						entity.(*types.PerfEntityMetric).Value = append(entity.(*types.PerfEntityMetric).Value, instance)
					}
				}
			}

			// Append entity to query perf response result
			result.Returnval = append(result.Returnval, entity)
		}
	}

	return result, nil
}
