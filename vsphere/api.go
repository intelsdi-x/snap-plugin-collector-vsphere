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
	"net/url"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type govmomiAPI struct {
	// Govmomi API objects
	client  *govmomi.Client
	finder  *find.Finder
	pc      *property.Collector
	cluster *mo.ClusterComputeResource

	// Cache variables
	datastores []mo.Datastore
	hosts      []mo.HostSystem
	vms        map[string][]mo.VirtualMachine
	metrics    []types.PerfCounterInfo
}

// Init initializes all necessary objects to send API calls to vSphere
// TODO: Mock Init's inside functions instead of making 2 versions of Init()
func (a *govmomiAPI) Init(ctx context.Context, url, username, password, clusterName string, insecure bool) error {
	var err error
	if a.client == nil {
		a.client, err = initializeClient(ctx, url, username, password, insecure)
		if err != nil {
			return fmt.Errorf("unable to initialize vSphere client: %v", err)
		}
	}

	if a.finder == nil {
		a.finder, err = initializeFinder(ctx, a.client)
		if err != nil {
			return fmt.Errorf("unable to initialize vSphere finder: %v", err)
		}
	}

	if a.pc == nil {
		a.pc = property.DefaultCollector(a.client.Client)
	}

	if a.cluster == nil {
		a.cluster, err = findCluster(ctx, a.finder, a.pc, clusterName)
		if err != nil {
			return fmt.Errorf("unable to find cluster: %v", err)
		}
	}

	return nil
}

// Clear API retrieve cache
func (a *govmomiAPI) ClearCache() {
	a.datastores = nil
	a.hosts = nil
	a.vms = nil
	a.metrics = nil
}

// RetrieveCounters retrieves vSphere cluster metric list that are available for user
func (a *govmomiAPI) RetrieveCounters(ctx context.Context) ([]types.PerfCounterInfo, error) {
	if a.metrics == nil {
		var perfManager mo.PerformanceManager

		err := a.client.RetrieveOne(ctx, *a.client.ServiceContent.PerfManager, nil, &perfManager)
		if err != nil {
			return nil, err
		}

		a.metrics = perfManager.PerfCounter
	}
	return a.metrics, nil
}

// RetrieveDatastores retrieves all datastores for cluster
// NOTE: For future development, for now datastore metrics are not available due to API limitations
func (a *govmomiAPI) RetrieveDatastores(ctx context.Context) ([]mo.Datastore, error) {
	if a.datastores == nil {
		if len(a.cluster.Datastore) != 0 {
			err := a.pc.Retrieve(ctx, a.cluster.Datastore, nil, &a.datastores)
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve datastores: %v", err)
			}
		}
	}

	return a.datastores, nil
}

// RetrieveHosts finds all hosts on given vSphere cluster
func (a *govmomiAPI) RetrieveHosts(ctx context.Context) ([]mo.HostSystem, error) {
	if a.hosts == nil {
		if len(a.cluster.Host) != 0 {
			err := a.pc.Retrieve(ctx, a.cluster.Host, []string{"name", "vm", "systemResources", "hardware"}, &a.hosts)
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve hosts: %v", err)
			}
		}
	}

	return a.hosts, nil
}

// RetrieveVMs finds all VMs for given host
func (a *govmomiAPI) RetrieveVMs(ctx context.Context, host mo.HostSystem) ([]mo.VirtualMachine, error) {
	if a.vms == nil {
		a.vms = make(map[string][]mo.VirtualMachine)
	}
	if a.vms[host.Reference().Value] == nil {
		if len(host.Vm) != 0 {
			vmsData := []mo.VirtualMachine{}
			err := a.pc.Retrieve(ctx, host.Vm, []string{"name", "summary"}, &vmsData)
			a.vms[host.Reference().Value] = vmsData
			if err != nil {
				return nil, fmt.Errorf("unable to retrieve virtual machines: %v", err)
			}
		}
	}

	return a.vms[host.Reference().Value], nil
}

// PerfQuery builds query object containing given query specs and sends it through govmomi API
// Response from PerfQuery() is built in following way:
// response.Returnval is a slice with results for all given entities (hosts, VMs, disks). Each element of this slice contains a slice with results for all given instances (CPU cores, VM NIC, etc.). Each element from instances slice contains a slice with integer values for given period. In our case, there's only one value in the last slice (we are retrieveing real-time data).
func (a *govmomiAPI) PerfQuery(ctx context.Context, querySpecs []types.PerfQuerySpec) (*types.QueryPerfResponse, error) {
	query := types.QueryPerf{
		QuerySpec: querySpecs,
		This:      *a.client.ServiceContent.PerfManager,
	}
	return methods.QueryPerf(ctx, a.client.RoundTripper, &query)
}

// initializeClient initializes vSphere API client
func initializeClient(ctx context.Context, hosturl, username, password string, insecure bool) (*govmomi.Client, error) {

	// TODO: Check whether user added "https://" prefix and "/sdk" suffix in URL
	vURL, err := url.Parse(hosturl)
	if err != nil {
		return nil, err
	}
	vURL.User = url.UserPassword(username, password)

	// Initialize client
	c, err := govmomi.NewClient(ctx, vURL, insecure)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// initializeFinder initializes and prepares vSphere API finder
func initializeFinder(ctx context.Context, client *govmomi.Client) (*find.Finder, error) {
	f := find.NewFinder(client.Client, true)
	dc, err := f.DefaultDatacenter(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to find default datacenter: %v", err)
	}
	f.SetDatacenter(dc)

	return f, nil
}

// findCluster finds cluster (mo.ClusterComputeResource type) with specified name (clusterName),
// using previously initialized property collector and finder
func findCluster(ctx context.Context, f *find.Finder, pc *property.Collector, clusterName string) (*mo.ClusterComputeResource, error) {

	clusterList, err := f.ClusterComputeResourceList(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("unable to find cluster compute resource list: %v", err)
	}
	if len(clusterList) == 0 {
		return nil, fmt.Errorf("cluster compute resource list is empty")
	}

	// There's only one cluster in cluster list if found
	clusterRef := clusterList[0].Reference()

	cluster := mo.ClusterComputeResource{}
	err = pc.Retrieve(ctx, []types.ManagedObjectReference{clusterRef}, []string{"name", "host", "datastore"}, &cluster)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve cluster from reference: %v", err)
	}

	return &cluster, nil

}
