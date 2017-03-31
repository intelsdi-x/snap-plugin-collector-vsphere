// +build small

/*
http://www.apache.org/licenses/LICENSE-2.0.txt


Copyright 2016 Intel Corporation

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
	"testing"

	"strings"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vmware/govmomi/vim25/types"
)

var testCtx = context.Background()

func Start() {
	initFixtures()
}

// Test config
func TestInit(t *testing.T) {
	testCfg := plugin.Config{
		"url":         "test",
		"username":    "test",
		"password":    "test",
		"insecure":    true,
		"clusterName": "test",
	}

	Convey("test parameters", t, func() {
		c := New(true)

		cfg := testCfg
		So(c.GovmomiResources.Init(testCtx, cfg), ShouldBeNil)

		delete(cfg, "url")
		So(c.GovmomiResources.Init(testCtx, cfg), ShouldNotBeNil)

		cfg = testCfg
		delete(cfg, "username")
		So(c.GovmomiResources.Init(testCtx, cfg), ShouldNotBeNil)

		cfg = testCfg
		delete(cfg, "password")
		So(c.GovmomiResources.Init(testCtx, cfg), ShouldNotBeNil)

		cfg = testCfg
		delete(cfg, "clusterName")
		So(c.GovmomiResources.Init(testCtx, cfg), ShouldNotBeNil)
	})

}

func TestFindHosts(t *testing.T) {
	initFixtures()

	Convey("test FindHosts error", t, func() {
		c := New(true)
		c.GovmomiResources.api.(*mockAPI).RetrieveHostsErr = true

		hosts, err := c.GovmomiResources.FindHosts(testCtx, "*")
		So(hosts, ShouldBeEmpty)
		So(err, ShouldNotBeNil)
	})

	Convey("test FindHosts success for *", t, func() {
		c := New(true)

		hosts, err := c.GovmomiResources.FindHosts(testCtx, "*")
		So(hosts, ShouldNotBeEmpty)
		So(len(hosts), ShouldEqual, 2)
		So(err, ShouldBeNil)
	})

	Convey("test FindHosts success for specific host", t, func() {
		c := New(true)

		hosts, err := c.GovmomiResources.FindHosts(testCtx, "1.1.1.1")
		So(hosts, ShouldNotBeEmpty)
		So(len(hosts), ShouldEqual, 1)
		So(hosts[0].Name, ShouldEqual, "1.1.1.1")
		So(err, ShouldBeNil)
	})

	Convey("test FindHosts success for specific not existing host", t, func() {
		c := New(true)

		hosts, err := c.GovmomiResources.FindHosts(testCtx, "1.1.1.2")
		So(hosts, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestFindHostByRef(t *testing.T) {
	initFixtures()

	Convey("test FindHostByRef error", t, func() {
		c := New(true)
		c.GovmomiResources.api.(*mockAPI).RetrieveHostsErr = true

		host, err := c.GovmomiResources.FindHostByRef(testCtx, testHosts[0].Reference())
		So(host, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("test FindHostByRef success for specified host reference", t, func() {
		c := New(true)

		host, err := c.GovmomiResources.FindHostByRef(testCtx, testHosts[0].Reference())
		So(host, ShouldNotBeNil)
		So(host.Entity().Name, ShouldEqual, "1.1.1.1")
		So(err, ShouldBeNil)
	})

	Convey("test FindHostByRef success for bad reference", t, func() {
		c := New(true)

		host, err := c.GovmomiResources.FindHostByRef(testCtx, testVMs["host-1"][0].Reference())
		So(host, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestFindVMs(t *testing.T) {
	initFixtures()

	Convey("test FindVMs error", t, func() {
		c := New(true)
		c.GovmomiResources.api.(*mockAPI).RetrieveVMsErr = true

		vms, err := c.GovmomiResources.FindVMs(testCtx, testHosts[0], "*")
		So(vms, ShouldBeEmpty)
		So(err, ShouldNotBeNil)
	})

	Convey("test FindVMs success for *", t, func() {
		c := New(true)

		vms, err := c.GovmomiResources.FindVMs(testCtx, testHosts[0], "*")
		So(vms, ShouldNotBeEmpty)
		So(len(vms), ShouldEqual, 2)
		So(err, ShouldBeNil)
	})

	Convey("test FindVMs success for specific vm", t, func() {
		c := New(true)

		vms, err := c.GovmomiResources.FindVMs(testCtx, testHosts[0], "VM1")
		So(vms, ShouldNotBeEmpty)
		So(len(vms), ShouldEqual, 1)
		So(err, ShouldBeNil)
	})

	Convey("test FindVMs success for specific vm (with host which does not contain any vm)", t, func() {
		c := New(true)

		vms, err := c.GovmomiResources.FindVMs(testCtx, testHosts[1], "VM1")
		So(vms, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})
}

func TestFindVMByRef(t *testing.T) {
	initFixtures()

	Convey("test FindVMByRef error", t, func() {
		c := New(true)
		c.GovmomiResources.api.(*mockAPI).RetrieveVMsErr = true

		vm, err := c.GovmomiResources.FindVMByRef(testCtx, testVMs["host-1"][0].Reference())
		So(vm, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("test FindVMByRef success for specified vm reference", t, func() {
		c := New(true)

		vm, err := c.GovmomiResources.FindVMByRef(testCtx, testVMs["host-1"][0].Reference())
		So(vm, ShouldNotBeNil)
		So(vm.Entity().Name, ShouldEqual, "VM1")
		So(vm.Summary.Runtime.Host.Reference().Value, ShouldEqual, "host-1")
		So(err, ShouldBeNil)
	})

	Convey("test FindVMByRef success for bad reference", t, func() {
		c := New(true)

		vm, err := c.GovmomiResources.FindVMByRef(testCtx, testHosts[0].Reference())
		So(vm, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestFindCounter(t *testing.T) {
	initFixtures()

	Convey("test FindCounter error", t, func() {
		c := New(true)
		c.GovmomiResources.api.(*mockAPI).RetrieveCountersErr = true

		counter, err := c.GovmomiResources.FindCounter(testCtx, "cpu.usage.average")
		So(counter, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("test FindCounter success", t, func() {
		c := New(true)
		counter, err := c.GovmomiResources.FindCounter(testCtx, "cpu.usage.average")
		So(counter, ShouldNotBeNil)
		So(counter.GroupInfo.GetElementDescription().Key, ShouldEqual, "cpu")
		So(counter.NameInfo.GetElementDescription().Key, ShouldEqual, "usage")
		So(counter.RollupType, ShouldEqual, "average")
		So(err, ShouldBeNil)
	})

	Convey("test FindCounter for non existing counter", t, func() {
		c := New(true)
		counter, err := c.GovmomiResources.FindCounter(testCtx, "cpu.test.summation")
		So(counter, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestFindCounterByID(t *testing.T) {
	initFixtures()

	Convey("test FindCounterByID error", t, func() {
		c := New(true)
		c.GovmomiResources.api.(*mockAPI).RetrieveCountersErr = true

		counter, err := c.GovmomiResources.FindCounterByKey(testCtx, 1)
		So(counter, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("test FindCounterByID success (1)", t, func() {
		c := New(true)
		counter, err := c.GovmomiResources.FindCounterByKey(testCtx, 1)
		So(counter, ShouldNotBeNil)
		So(counter.GroupInfo.GetElementDescription().Key, ShouldEqual, "cpu")
		So(counter.NameInfo.GetElementDescription().Key, ShouldEqual, "usage")
		So(counter.RollupType, ShouldEqual, "average")
		So(err, ShouldBeNil)
	})

	Convey("test FindCounterByID for non existing counter", t, func() {
		c := New(true)
		counter, err := c.GovmomiResources.FindCounterByKey(testCtx, 999999999)
		So(counter, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})
}

func TestPerfQuery(t *testing.T) {
	initFixtures()

	testQuerySpecs := []types.PerfQuerySpec{

		// cpu.idle.summation
		types.PerfQuerySpec{
			Entity:     testHosts[0].Reference(),
			IntervalId: 20,
			MaxSample:  1,
			Format:     "normal",
			MetricId: []types.PerfMetricId{
				types.PerfMetricId{
					Instance:  "*",
					CounterId: 1,
				},
			},
		},

		// cpu.wait.summation
		types.PerfQuerySpec{
			Entity:     testHosts[0].Reference(),
			IntervalId: 20,
			MaxSample:  1,
			Format:     "normal",
			MetricId: []types.PerfMetricId{
				types.PerfMetricId{
					Instance:  "*",
					CounterId: 2,
				},
			},
		},
	}

	Convey("test PerfQuery success", t, func() {
		c := New(true)

		response, err := c.GovmomiResources.PerfQuery(testCtx, testQuerySpecs)
		So(err, ShouldBeNil)
		So(len(response.Returnval), ShouldEqual, 2)
		So(len(response.Returnval[0].(*types.PerfEntityMetric).Value), ShouldEqual, 3)
		So(len(response.Returnval[1].(*types.PerfEntityMetric).Value), ShouldEqual, 1)
	})

	Convey("test PerfQuery error", t, func() {
		c := New(true)
		c.GovmomiResources.api.(*mockAPI).PerfQueryErr = true

		response, err := c.GovmomiResources.PerfQuery(testCtx, testQuerySpecs)
		So(err, ShouldNotBeNil)
		So(response, ShouldBeNil)
	})
}

func TestCollectMetrics(t *testing.T) {
	testCfg := plugin.Config{
		"url":         "test",
		"username":    "test",
		"password":    "test",
		"insecure":    true,
		"clusterName": "test",
	}

	testMetrics := []plugin.Metric{
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "cpu", "*", "idle"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "cpu", "*", "wait"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "cpu", "*", "load"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "mem", "*", "usage"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "mem", "*", "free"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "2.2.2.2", "mem", "*", "free"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "net", "*", "kbrate_tx"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "net", "*", "kbrate_rx"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "net", "*", "packets_tx"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "net", "*", "packets_rx"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "vm", "*", "virtualDisk", "*", "read_iops"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "vm", "*", "virtualDisk", "*", "write_iops"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "vm", "*", "virtualDisk", "*", "read_throughput"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "vm", "*", "virtualDisk", "*", "write_throughput"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "vm", "*", "virtualDisk", "*", "read_latency"),
			Config:    testCfg,
		},
		plugin.Metric{
			Namespace: plugin.NewNamespace("intel", "vmware", "vsphere", "host", "1.1.1.1", "vm", "*", "virtualDisk", "*", "write_latency"),
			Config:    testCfg,
		},
	}

	Convey("test CollectMetrics success", t, func() {
		c := New(true)

		result, err := c.CollectMetrics(testMetrics)

		So(err, ShouldBeNil)
		So(len(result), ShouldEqual, 27)

		// Checking CollectMetrics output based on data in mock fixtures
		for _, r := range result {
			ns := strings.Join(r.Namespace.Strings(), "/")
			switch ns {
			case "intel/vmware/vsphere/host/1.1.1.1/cpu/aggr/wait":
				So(r.Data, ShouldEqual, 2)
			case "intel/vmware/vsphere/host/1.1.1.1/cpu/aggr/load":
				So(r.Data, ShouldEqual, 0.003)
			case "intel/vmware/vsphere/host/1.1.1.1/cpu/0/idle":
				So(r.Data, ShouldEqual, 99)
			case "intel/vmware/vsphere/host/1.1.1.1/cpu/1/idle":
				So(r.Data, ShouldEqual, 98.9)
			case "intel/vmware/vsphere/host/1.1.1.1/cpu/2/idle":
				So(r.Data, ShouldEqual, 98.8)
			case "intel/vmware/vsphere/host/1.1.1.1/mem/0/usage":
				So(r.Data, ShouldEqual, 120)
			case "intel/vmware/vsphere/host/1.1.1.1/mem/0/free":
				So(r.Data, ShouldEqual, 1057)
			case "intel/vmware/vsphere/host/2.2.2.2/mem/0/free":
				So(r.Data, ShouldEqual, 4236)
			case "intel/vmware/vsphere/host/1.1.1.1/net/eth0/kbrate_tx":
				So(r.Data, ShouldEqual, 400)
			case "intel/vmware/vsphere/host/1.1.1.1/net/eth0/kbrate_rx":
				So(r.Data, ShouldEqual, 500)
			case "intel/vmware/vsphere/host/1.1.1.1/net/eth0/packets_tx":
				So(r.Data, ShouldEqual, 600)
			case "intel/vmware/vsphere/host/1.1.1.1/net/eth0/packets_rx":
				So(r.Data, ShouldEqual, 700)
			case "intel/vmware/vsphere/host/1.1.1.1/vm/VM1/virtualDisk/0/read_iops":
				So(r.Data, ShouldEqual, 800)
			case "intel/vmware/vsphere/host/1.1.1.1/vm/VM1/virtualDisk/0/write_iops":
				So(r.Data, ShouldEqual, 900)
			case "intel/vmware/vsphere/host/1.1.1.1/vm/VM1/virtualDisk/0/read_throughput":
				So(r.Data, ShouldEqual, 1000)
			case "intel/vmware/vsphere/host/1.1.1.1/vm/VM1/virtualDisk/0/write_throughput":
				So(r.Data, ShouldEqual, 1100)
			case "intel/vmware/vsphere/host/1.1.1.1/vm/VM1/virtualDisk/0/read_latency":
				So(r.Data, ShouldEqual, 1200)
			case "intel/vmware/vsphere/host/1.1.1.1/vm/VM1/virtualDisk/0/write_latency":
				So(r.Data, ShouldEqual, 1300)
			}
		}
	})

	Convey("test CollectMetrics (RetrieveCounters fail)", t, func() {
		c := New(true)
		c.GovmomiResources.api.(*mockAPI).RetrieveCountersErr = true

		result, err := c.CollectMetrics(testMetrics)

		So(err, ShouldNotBeNil)
		So(result, ShouldBeEmpty)
	})

	Convey("test CollectMetrics (RetrieveHosts fail)", t, func() {
		c := New(true)
		c.GovmomiResources.api.(*mockAPI).RetrieveHostsErr = true

		result, err := c.CollectMetrics(testMetrics)

		So(err, ShouldNotBeNil)
		So(result, ShouldBeEmpty)
	})

	Convey("test CollectMetrics (RetrieveVMs fail)", t, func() {
		c := New(true)
		c.GovmomiResources.api.(*mockAPI).RetrieveVMsErr = true

		result, err := c.CollectMetrics(testMetrics)

		So(err, ShouldNotBeNil)
		So(result, ShouldBeEmpty)
	})
}

func TestGetMetricTypes(t *testing.T) {
	testCfg := plugin.Config{
		"url":         "test",
		"username":    "test",
		"password":    "test",
		"insecure":    true,
		"clusterName": "test",
	}

	Convey("test TestGetMetricTypes output not empty", t, func() {
		c := New(true)

		result, err := c.GetMetricTypes(testCfg)
		So(err, ShouldBeNil)
		So(result, ShouldNotBeEmpty)
	})
}

func TestGetConfigPolicy(t *testing.T) {
	Convey("test GetConfigPolicy returns non-nil value", t, func() {
		c := New(true)

		result, err := c.GetConfigPolicy()
		So(err, ShouldBeNil)
		So(result, ShouldNotBeNil)
	})
}
