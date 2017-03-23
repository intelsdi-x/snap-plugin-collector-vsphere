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

	"github.com/intelsdi-x/snap-plugin-lib-go/v1/plugin"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/vmware/govmomi/vim25/mo"
)

// Test config
func TestInit(t *testing.T) {
	ctx := context.Background()

	full_cfg := plugin.Config{
		"url":         "test",
		"username":    "test",
		"password":    "test",
		"insecure":    true,
		"clusterName": "test",
	}

	Convey("test parameters", t, func() {
		c := New(true)

		cfg := full_cfg
		So(c.GovmomiResources.Init(ctx, cfg), ShouldBeNil)

		delete(cfg, "url")
		So(c.GovmomiResources.Init(ctx, cfg), ShouldNotBeNil)

		cfg = full_cfg
		delete(cfg, "username")
		So(c.GovmomiResources.Init(ctx, cfg), ShouldNotBeNil)

		cfg = full_cfg
		delete(cfg, "password")
		So(c.GovmomiResources.Init(ctx, cfg), ShouldNotBeNil)

		cfg = full_cfg
		delete(cfg, "clusterName")
		So(c.GovmomiResources.Init(ctx, cfg), ShouldNotBeNil)
	})

}

func TestCallQueryPerf(t *testing.T) {
	ctx := context.Background()

	Convey("test empty metric", t, func() {
		c := New(true)

		_, err := c.GovmomiResources.CallQueryPerf(ctx, mo.HostSystem{}, "")
		So(err, ShouldNotBeNil)
	})

}

func TestCollectMetrics(t *testing.T) {

	var metric plugin.Metric
	cfg := plugin.Config{
		"url":         "test",
		"username":    "test",
		"password":    "test",
		"insecure":    true,
		"clusterName": "test",
	}

	Convey("test no metric", t, func() {
		c := New(true)

		// Get correct metric for next tests
		mts, err := c.GetMetricTypes(plugin.Config{})
		metric = mts[0]
		metric.Config = cfg

		mts, err = c.CollectMetrics([]plugin.Metric{})
		So(mts, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("test client failure", t, func() {
		c := New(true)
		c.GovmomiResources.(*mockClient).ClientFailure = true

		mts, err := c.CollectMetrics([]plugin.Metric{metric})
		So(mts, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("test retrieve metrics failure", t, func() {
		c := New(true)
		c.GovmomiResources.(*mockClient).RetrieveMetricsFailure = true

		mts, err := c.CollectMetrics([]plugin.Metric{metric})
		So(mts, ShouldBeNil)
		So(err, ShouldNotBeNil)

	})

	Convey("test get hosts failure", t, func() {
		c := New(true)
		c.GovmomiResources.(*mockClient).GetHostsFailure = true

		mts, err := c.CollectMetrics([]plugin.Metric{metric})
		So(mts, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("test queryPerf call failure", t, func() {
		c := New(true)
		c.GovmomiResources.(*mockClient).CallQueryPerfFailure = true

		mts, err := c.CollectMetrics([]plugin.Metric{metric})
		So(mts, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("test no vSphere hosts", t, func() {
		c := New(true)
		c.GovmomiResources.(*mockClient).GetHostsEmpty = true

		// If no hosts are found
		mts, err := c.CollectMetrics([]plugin.Metric{metric})
		So(mts, ShouldBeEmpty)
		So(err, ShouldBeNil)
	})

	Convey("test positive configuration", t, func() {
		c := New(true)
		mts, err := c.CollectMetrics([]plugin.Metric{metric})
		So(mts, ShouldNotBeEmpty)
		So(err, ShouldBeNil)
	})

}
