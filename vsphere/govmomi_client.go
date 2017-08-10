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

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

const (
	// Default IntervalId for QueryPerf(). Has to be 20 to retrieve most recent data,
	// and ignore historical data. See README.md for more details.
	defaultIntervalID = 20

	// Default metric instance for QueryPerf(). Some counters have more than one instance
	// (for example, for each CPU core). Asterisk specifies all the instances. Empty field specifies aggregated instances.
	defaultMetricInstance = "*"
)

// API - vSphere API interface for testing purposes
type API interface {
	// Initialize all necessary objects to send API calls to vSphere
	Init(ctx context.Context, url, username, password, clusterName string, datacenterName string, insecure bool) error

	// RetrieveCounters retrieves vSphere cluster metric list that are available for user
	RetrieveCounters(ctx context.Context) ([]types.PerfCounterInfo, error)

	// Get all datastores for cluster
	RetrieveDatastores(ctx context.Context) ([]mo.Datastore, error)

	// Get all hosts for cluster
	RetrieveHosts(ctx context.Context) ([]mo.HostSystem, error)

	// Find all VMs for given host
	RetrieveVMs(ctx context.Context, host mo.HostSystem) ([]mo.VirtualMachine, error)

	// Call performance query to retrieve perf data
	PerfQuery(ctx context.Context, querySpecs []types.PerfQuerySpec) (*types.QueryPerfResponse, error)

	// Clear API retrieve cache
	ClearCache()
}

// govmomiClient is proxy for API calls, providing more functionality and allowing to mock API calls separately for testing
type govmomiClient struct {
	api API
}

func (c *govmomiClient) Init(ctx context.Context, cfg plugin.Config) error {
	url, err := cfg.GetString("url")
	if err != nil {
		return err
	}
	username, err := cfg.GetString("username")
	if err != nil {
		return err
	}
	password, err := cfg.GetString("password")
	if err != nil {
		return err
	}

	insecure, err := cfg.GetBool("insecure")
	if err != nil {
		return err
	}
	clusterName, err := cfg.GetString("clusterName")
	if err != nil {
		return err
	}

	datacenterName, err := cfg.GetString("datacenterName")
	if err != nil {
		return err
	}

	return c.api.Init(ctx, url, username, password, clusterName, datacenterName, insecure)
}

func (c *govmomiClient) PerfQuery(ctx context.Context, querySpecs []types.PerfQuerySpec) (*types.QueryPerfResponse, error) {
	return c.api.PerfQuery(ctx, querySpecs)
}

func (c *govmomiClient) RetrieveCounters(ctx context.Context) ([]types.PerfCounterInfo, error) {
	return c.api.RetrieveCounters(ctx)
}

// ClearCache clears cache for all API retrieve operations
func (c *govmomiClient) ClearCache() {
	c.api.ClearCache()
}

// FindHosts returns all hosts for configured cluster
func (c *govmomiClient) FindHosts(ctx context.Context, hostName string) ([]mo.HostSystem, error) {
	hosts, err := c.api.RetrieveHosts(ctx)
	if err != nil {
		return nil, err
	}
	results := []mo.HostSystem{}
	for _, host := range hosts {
		if host.Name == hostName || hostName == "*" {
			results = append(results, host)
		}
	}
	return results, nil
}

// FindVMs retuns all virtual machines for given host
func (c *govmomiClient) FindVMs(ctx context.Context, host mo.HostSystem, vmName string) ([]mo.VirtualMachine, error) {
	vms, err := c.api.RetrieveVMs(ctx, host)
	if err != nil {
		return nil, err
	}
	results := []mo.VirtualMachine{}
	for _, vm := range vms {
		if vm.Name == vmName || vmName == "*" {
			results = append(results, vm)
		}
	}
	return results, nil
}

// FindCounter returns vSphere counter info by counter name, for example cpu.idle.summation
func (c *govmomiClient) FindCounter(ctx context.Context, counterFullName string) (*types.PerfCounterInfo, error) {
	// Retrieve all vCenter metrics
	vMetrics, err := c.api.RetrieveCounters(ctx)
	if err != nil {
		return nil, err
	}
	for _, pc := range vMetrics {
		nameInfo := pc.GroupInfo.GetElementDescription().Key + "." + pc.NameInfo.GetElementDescription().Key + "." + fmt.Sprint(pc.RollupType)

		if nameInfo == counterFullName {
			return &pc, nil
		}
	}
	return nil, fmt.Errorf("no vsphere perf counters found for %s", counterFullName)
}

// FindCounterByKey returns vSphere counter info by counter key (ID)
func (c *govmomiClient) FindCounterByKey(ctx context.Context, key int32) (*types.PerfCounterInfo, error) {
	// Retrieve all vCenter metrics
	vMetrics, err := c.api.RetrieveCounters(ctx)
	if err != nil {
		return nil, err
	}
	for _, pc := range vMetrics {
		if key == pc.Key {
			return &pc, nil
		}

	}
	return nil, fmt.Errorf("no vsphere perf counters found for key %d", key)
}

func (c *govmomiClient) FindDatastoreByRef(ctx context.Context, ref types.ManagedObjectReference) (*mo.Datastore, error) {
	datastores, err := c.api.RetrieveDatastores(ctx)
	if err != nil {
		return nil, err
	}
	for _, ds := range datastores {
		if ds.Reference().Value == ref.Value {
			return &ds, nil
		}
	}
	return nil, fmt.Errorf("cannot find dataqstore by reference %s", ref.Value)
}

// FindHostByRef returns mo.HostSystem for given reference
func (c *govmomiClient) FindHostByRef(ctx context.Context, ref types.ManagedObjectReference) (*mo.HostSystem, error) {
	hosts, err := c.api.RetrieveHosts(ctx)
	if err != nil {
		return nil, err
	}
	for _, host := range hosts {
		if host.Reference().Value == ref.Value {
			return &host, nil
		}
	}
	return nil, fmt.Errorf("cannot find host by reference %s", ref.Value)
}

// FindVMByRef returns mo.VirtualMachine for given reference
func (c *govmomiClient) FindVMByRef(ctx context.Context, ref types.ManagedObjectReference) (*mo.VirtualMachine, error) {
	hosts, err := c.api.RetrieveHosts(ctx)
	if err != nil {
		return nil, err
	}
	for _, host := range hosts {
		vms, err := c.api.RetrieveVMs(ctx, host)
		if err != nil {
			return nil, err
		}
		for _, vm := range vms {
			if vm.Reference().Value == ref.Value {
				return &vm, nil
			}
		}
	}

	return nil, fmt.Errorf("cannot find virtual machine by reference %s", ref.Value)
}

// GetInstances extracts instance list from provided metric
// Each metric contains multiple instances, for example disk-related metrics returns 1 instance for each disk
func (c *govmomiClient) GetInstances(metric types.BasePerfEntityMetricBase) ([]types.BasePerfMetricSeries, error) {
	result := metric.(*types.PerfEntityMetric).Value
	if len(result) == 0 {
		return nil, fmt.Errorf("No instances found for specified metric")
	}
	return result, nil
}

// GetInstanceSeries retrieves metric series for given instance
// types.PerfMetricIntSeries contains metric info (such as counter key) and slice of int64 values.
func (c *govmomiClient) GetInstanceSeries(metric types.BasePerfMetricSeries) (*types.PerfMetricIntSeries, error) {
	return metric.(*types.PerfMetricIntSeries), nil
}
