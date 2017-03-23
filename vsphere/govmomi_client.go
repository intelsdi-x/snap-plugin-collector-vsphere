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

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/methods"
	"github.com/vmware/govmomi/vim25/mo"
	"github.com/vmware/govmomi/vim25/types"
)

type govmomiClient struct {
	client  *govmomi.Client
	finder  *find.Finder
	pc      *property.Collector
	cluster *mo.ClusterComputeResource
	metrics []types.PerfCounterInfo
}

const (
	// Default IntervalId for QueryPerf(). Has to be 20 to retrieve most recent data,
	// and ignore aggregated, historical data. See README.md for more details.
	defaultIntervalId = 20

	// Default metric instance for QueryPerf(). Some counters have more than one instance
	// (for example, for each CPU core). Asterisk specifies all the instances. For memory
	// counters, there's only one instance
	defaultMetricInstance = "*"

	// vSphere counters strings
	vcMemConsumedKb = "mem.consumed.average"
)

// Init initializes all necessary objects to send API calls to vSphere
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

	if c.client == nil {
		c.client, err = initializeClient(ctx, url, username, password, insecure)
		if err != nil {
			return fmt.Errorf("unable to initialize vSphere client: %v", err)
		}
	}

	if c.finder == nil {
		c.finder, err = initializeFinder(ctx, c.client)
		if err != nil {
			return fmt.Errorf("unable to initialize vSphere finder: %v", err)
		}
	}

	if c.pc == nil {
		c.pc = property.DefaultCollector(c.client.Client)
	}

	if c.cluster == nil {
		c.cluster, err = findCluster(ctx, c.finder, c.pc, clusterName)
		if err != nil {
			return fmt.Errorf("unable to find cluster: %v", err)
		}
	}

	return nil
}

// RetrieveMetrics retrieves vSphere cluster metric list that are available for user
func (c *govmomiClient) RetrieveMetrics(ctx context.Context) error {
	var perfManager mo.PerformanceManager

	err := c.client.RetrieveOne(ctx, *c.client.ServiceContent.PerfManager, nil, &perfManager)
	if err != nil {
		return err
	}

	c.metrics = perfManager.PerfCounter
	return nil
}

// GetHosts finds all hosts on given vSphere cluster
func (c *govmomiClient) GetHosts(ctx context.Context) ([]mo.HostSystem, error) {
	var hosts []mo.HostSystem
	err := c.pc.Retrieve(ctx, c.cluster.Host, []string{"name", "vm", "systemResources", "hardware"}, &hosts)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve hosts: %v", err)
	}

	return hosts, nil
}

// CallQueryPerf calls queryPerf() govmomi method to retrieve single host metric (given by metricName)
func (c *govmomiClient) CallQueryPerf(ctx context.Context, host mo.HostSystem, metricName string) (int64, error) {
	var metric types.PerfCounterInfo

	// Find metric
	found := false
	for _, pc := range c.metrics {
		groupInfo := pc.GroupInfo.GetElementDescription()
		nameInfo := pc.NameInfo.GetElementDescription()
		fullName := groupInfo.Key + "." + nameInfo.Key + "." + fmt.Sprint(pc.RollupType)

		if fullName == metricName {
			metric = pc
			found = true
		}

	}

	if !found {
		return 0, fmt.Errorf("metric %v not found in vSphere", metricName)
	}

	// Prepare PerfMetricId, PerfQuerySpec and QueryPerf types
	metricId := types.PerfMetricId{
		CounterId: metric.Key,
		Instance:  defaultMetricInstance, // Get data for all instances (i.e. all cores).
	}

	querySpec := types.PerfQuerySpec{
		Entity:     host.Reference(),
		IntervalId: defaultIntervalId,
		MaxSample:  1,
		Format:     "normal",
		MetricId:   []types.PerfMetricId{metricId},
	}

	query := types.QueryPerf{
		QuerySpec: []types.PerfQuerySpec{querySpec},
		This:      *c.client.ServiceContent.PerfManager,
	}

	// Call
	results, err := methods.QueryPerf(ctx, c.client.RoundTripper, &query)
	if err != nil {
		return 0, fmt.Errorf("unable to call QueryPerf: %v", err)
	}

	/*
		QueryPerf results:
		> results.Returnval - array of results for all entities (host/vm) specified in
		in QuerySpec field in query (types.QueryPerf).
		> Each element of results.Returnval contains (Value) an array (perfEntityResult) of
		results for each metric specified in MetricId field in querySpec (types.PerfQuerySpec)
		and has to be type-asserted for proper type.
		> Each element of of perfEntityResults contains (Value) an array of results for
		requested instances.

		Right now, we specify single metric, single entity, and every metric has a single
		instance (since it's a memory metric), so expected lenght of each array is 1.
	*/

	// Parse result
	if len(results.Returnval) != 1 {
		return 0, fmt.Errorf("Incorrect number of QueryPerf results (for metrics)")
	}

	perfEntityResults := results.Returnval[0].(*types.PerfEntityMetric)
	if len(perfEntityResults.Value) != 1 {
		return 0, fmt.Errorf("Incorrect number (%d) of QueryPerf results (for entities)", len(perfEntityResults.Value))
	}

	instanceIntResults := perfEntityResults.Value[0].(*types.PerfMetricIntSeries)
	if len(instanceIntResults.Value) != 1 {
		return 0, fmt.Errorf("Incorrect number (%d) of QueryPerf results (for instances)", len(instanceIntResults.Value))
	}

	return instanceIntResults.Value[0], nil
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
		return nil, fmt.Errorf("Unable to find default datacenter: %v", err)
	}
	f.SetDatacenter(dc)

	return f, nil
}

// findCluster finds cluster (mo.ClusterComputeResource type) with specified name (clusterName),
// using previously initialized property collector and finder
func findCluster(ctx context.Context, f *find.Finder, pc *property.Collector, clusterName string) (*mo.ClusterComputeResource, error) {

	clusterList, err := f.ClusterComputeResourceList(ctx, clusterName)
	if err != nil {
		return nil, fmt.Errorf("Unable to find cluster compute resource list: %v", err)
	}

	// There's only one cluster in cluster list if found
	clusterRef := clusterList[0].Reference()

	cluster := new(mo.ClusterComputeResource)
	err = pc.Retrieve(ctx, []types.ManagedObjectReference{clusterRef}, []string{"name", "host"}, cluster)
	if err != nil {
		return nil, fmt.Errorf("Unable to retrieve cluster from reference: %v", err)
	}

	return cluster, nil

}
