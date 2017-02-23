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
	"time"

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"

	"github.com/vmware/govmomi/vim25/mo"
)

const (
	vendor       = "intel"
	class        = "vmware"
	name         = "vsphere"
	hostPosition = 4 // Position of host element
)

/*
VsphereCollector contains govmomi abstract types such as client, finder and properties collector
and implements plugin interface
*/
type VsphereCollector struct {
	GovmomiResources API
}

// New returns instance of VsphereCollector
func New(isTest bool) *VsphereCollector {
	collector := &VsphereCollector{}
	if isTest {
		collector.GovmomiResources = &mockClient{}
	} else {
		collector.GovmomiResources = &govmomiClient{}
	}
	return collector
}

// CollectMetrics collects requested metrics
func (c *VsphereCollector) CollectMetrics(mts []plugin.Metric) ([]plugin.Metric, error) {

	if len(mts) < 1 {
		return nil, fmt.Errorf("No metrics specified")
	}

	metrics := []plugin.Metric{}
	ctx := context.Background()

	if err := c.GovmomiResources.Init(ctx, mts[0].Config); err != nil {
		return nil, fmt.Errorf("Unable to initialize: %v", err)
	}

	// Retrieve vSphere available metrics. Has to be retrieved
	// on each CollectMetrics() call - available metrics can be changed in runtime
	err := c.GovmomiResources.RetrieveMetrics(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to retrieve vSphere perf counters: %v", err)
	}

	for _, mt := range mts {

		if isDynamic, dynInd := mt.Namespace.IsDynamic(); isDynamic {
			if len(dynInd) != 1 || dynInd[0] != hostPosition {
				return nil, fmt.Errorf("incorrect metric")
			}
		} else {
			// Not dynamic metric provided
			return nil, fmt.Errorf("incorrect metric")
		}

		var hosts []mo.HostSystem
		hosts, err := c.GovmomiResources.GetHosts(ctx)
		if err != nil {
			return nil, fmt.Errorf("Unable to get hosts: %v", err)
		}

		// Single host
		if requestedHost := mt.Namespace[hostPosition].Value; requestedHost != "*" {
			found := false
			for _, h := range hosts {
				if h.Name == requestedHost {
					found = true
					hosts = []mo.HostSystem{h}
					break
				}
			}
			if !found {
				return nil, fmt.Errorf("Unable to find host %s", requestedHost)
			}
		}

		// Prepare metric for request
		for _, host := range hosts {
			var data int64

			switch mt.Namespace[len(mt.Namespace)-1].Value {
			case "memUsage":
				memUsage, err := c.GovmomiResources.CallQueryPerf(ctx, host, vcMemConsumedKb) // in kilobytes
				if err != nil {
					return nil, fmt.Errorf("unable retrieve memUsage: %v", err)
				}

				data = memUsage / 1024
				break

			case "memFree":
				memUsage, err := c.GovmomiResources.CallQueryPerf(ctx, host, vcMemConsumedKb) // in kilobytes
				if err != nil {
					return nil, fmt.Errorf("unable retrieve memUsage: %v", err)
				}
				memAvailable := host.Hardware.MemorySize // in bytes

				data = memAvailable/(1024*1024) - memUsage/1024
				break

			case "memAvailable":
				memAvailable := host.Hardware.MemorySize // in bytes
				data = memAvailable / (1024 * 1024)
				break

			}

			ns := make([]plugin.NamespaceElement, len(mt.Namespace))
			copy(ns, mt.Namespace)

			// 4 is an offset for host name in metric
			ns[hostPosition].Value = host.Name
			hostMetric := plugin.Metric{
				Data:      data,
				Namespace: ns,
				Timestamp: time.Now(),
				Unit:      mt.Unit,
				Version:   mt.Version,
			}

			metrics = append(metrics, hostMetric)

		}

	}

	return metrics, nil
}

// GetMetricTypes returns available metrics
func (c *VsphereCollector) GetMetricTypes(cfg plugin.Config) ([]plugin.Metric, error) {
	metrics := []plugin.Metric{}

	/*
		Dynamic hostname metric
	*/
	// Host memory usage
	metrics = append(metrics, plugin.Metric{
		Namespace:   plugin.NewNamespace(vendor, class, name, "host").AddDynamicElement("host", "name of the vSphere host").AddStaticElement("memUsage"),
		Description: "Host memory usage",
		Unit:        "megabytes",
	})

	// Host memory available
	metrics = append(metrics, plugin.Metric{
		Namespace:   plugin.NewNamespace(vendor, class, name, "host").AddDynamicElement("host", "name of the vSphere host").AddStaticElement("memAvailable"),
		Description: "Host memory available",
		Unit:        "megabytes",
	})

	// Host memory free
	metrics = append(metrics, plugin.Metric{
		Namespace:   plugin.NewNamespace(vendor, class, name, "host").AddDynamicElement("host", "name of the vSphere host").AddStaticElement("memFree"),
		Description: "Host free memory",
		Unit:        "megabytes",
	})

	return metrics, nil
}

// GetConfigPolicy retrieves config for the plugin
func (c *VsphereCollector) GetConfigPolicy() (plugin.ConfigPolicy, error) {
	policy := plugin.NewConfigPolicy()

	policy.AddNewStringRule([]string{vendor, class, name}, "url", true, plugin.SetDefaultString(""))
	policy.AddNewStringRule([]string{vendor, class, name}, "username", true, plugin.SetDefaultString(""))
	policy.AddNewStringRule([]string{vendor, class, name}, "password", true, plugin.SetDefaultString(""))
	policy.AddNewBoolRule([]string{vendor, class, name}, "insecure", false, plugin.SetDefaultBool(false))
	policy.AddNewStringRule([]string{vendor, class, name}, "clusterName", true, plugin.SetDefaultString(""))

	return *policy, nil
}
