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

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	"github.com/vmware/govmomi/vim25/mo"
)

// API - vSphere API interface for testing purposes.
type API interface {
	// Initialize all necessary objects to send API calls to vSphere
	Init(ctx context.Context, cfg plugin.Config) error

	// Retrieve all metrics available on vSphere cluster.
	RetrieveMetrics(ctx context.Context) error

	// Get all hosts for cluster
	GetHosts(ctx context.Context) ([]mo.HostSystem, error)

	// Call queryPerf() govmomi method to retrieve single host metric (given by metricName)
	CallQueryPerf(ctx context.Context, host mo.HostSystem, metricName string) (int64, error)
}
