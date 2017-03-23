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
)

type mockClient struct {
	ClientFailure          bool
	RetrieveMetricsFailure bool
	GetHostsFailure        bool
	GetHostsEmpty          bool
	CallQueryPerfFailure   bool
}

// Init - mock initialization of govmomi API objects
func (c *mockClient) Init(ctx context.Context, cfg plugin.Config) error {

	if _, err := cfg.GetString("url"); err != nil {
		return err
	}

	if _, err := cfg.GetString("username"); err != nil {
		return err
	}

	if _, err := cfg.GetString("password"); err != nil {
		return err
	}

	if _, err := cfg.GetBool("insecure"); err != nil {
		return err
	}

	if _, err := cfg.GetString("clusterName"); err != nil {
		return err
	}

	if c.ClientFailure {
		return fmt.Errorf("unable to initialize client")
	}
	return nil
}

// RetrieveMetrics - mock retrieving vSphere metrics
func (c *mockClient) RetrieveMetrics(ctx context.Context) error {
	if c.RetrieveMetricsFailure {
		return fmt.Errorf("unable to retrieve metrics")
	}
	return nil
}

// RetrieveMetrics - mock retrieving vSphere metrics
func (c *mockClient) GetHosts(ctx context.Context) ([]mo.HostSystem, error) {
	hosts := []mo.HostSystem{}
	if !c.GetHostsEmpty {
		hosts = []mo.HostSystem{mo.HostSystem{}, mo.HostSystem{}}
	}

	if c.GetHostsFailure {
		return hosts, fmt.Errorf("Unable to find hosts")
	}

	return hosts, nil
}

// RetrieveMetrics - mock queryPerf() govmomi call
func (c *mockClient) CallQueryPerf(ctx context.Context, host mo.HostSystem, metricName string) (int64, error) {

	// No metric found
	if metricName == "" {
		return 0, fmt.Errorf("no metric specified")
	}

	if c.CallQueryPerfFailure {
		return 0, fmt.Errorf("unable to call query perf")
	}
	return 0, nil
}
