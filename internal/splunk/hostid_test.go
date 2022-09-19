// Copyright The OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package splunk

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/collector/pdata/pcommon"
	conventions "go.opentelemetry.io/collector/semconv/v1.6.1"
)

var (
	ec2Resource = func() pcommon.Resource {
		res := pcommon.NewResource()
		attr := res.Attributes()
		attr.PutString(conventions.AttributeCloudProvider, conventions.AttributeCloudProviderAWS)
		attr.PutString(conventions.AttributeCloudAccountID, "1234")
		attr.PutString(conventions.AttributeCloudRegion, "us-west-2")
		attr.PutString(conventions.AttributeHostID, "i-abcd")
		return res
	}()
	ec2WithHost = func() pcommon.Resource {
		res := pcommon.NewResource()
		attr := res.Attributes()
		attr.PutString(conventions.AttributeCloudProvider, conventions.AttributeCloudProviderAWS)
		attr.PutString(conventions.AttributeCloudAccountID, "1234")
		attr.PutString(conventions.AttributeCloudRegion, "us-west-2")
		attr.PutString(conventions.AttributeHostID, "i-abcd")
		attr.PutString(conventions.AttributeHostName, "localhost")
		return res
	}()
	ec2PartialResource = func() pcommon.Resource {
		res := pcommon.NewResource()
		attr := res.Attributes()
		attr.PutString(conventions.AttributeCloudProvider, conventions.AttributeCloudProviderAWS)
		attr.PutString(conventions.AttributeHostID, "i-abcd")
		return res
	}()
	gcpResource = func() pcommon.Resource {
		res := pcommon.NewResource()
		attr := res.Attributes()
		attr.PutString(conventions.AttributeCloudProvider, conventions.AttributeCloudProviderGCP)
		attr.PutString(conventions.AttributeCloudAccountID, "1234")
		attr.PutString(conventions.AttributeHostID, "i-abcd")
		return res
	}()
	gcpPartialResource = func() pcommon.Resource {
		res := pcommon.NewResource()
		attr := res.Attributes()
		attr.PutString(conventions.AttributeCloudProvider, conventions.AttributeCloudProviderGCP)
		attr.PutString(conventions.AttributeCloudAccountID, "1234")
		return res
	}()
	azureResource = func() pcommon.Resource {
		res := pcommon.NewResource()
		attrs := res.Attributes()
		attrs.PutString(conventions.AttributeCloudProvider, conventions.AttributeCloudProviderAzure)
		attrs.PutString(conventions.AttributeCloudPlatform, conventions.AttributeCloudPlatformAzureVM)
		attrs.PutString(conventions.AttributeHostName, "myHostName")
		attrs.PutString(conventions.AttributeCloudRegion, "myCloudRegion")
		attrs.PutString(conventions.AttributeHostID, "myHostID")
		attrs.PutString(conventions.AttributeCloudAccountID, "myCloudAccount")
		attrs.PutString("azure.vm.name", "myVMName")
		attrs.PutString("azure.vm.size", "42")
		attrs.PutString("azure.resourcegroup.name", "myResourcegroupName")
		return res
	}()
	azureScalesetResource = func() pcommon.Resource {
		res := pcommon.NewResource()
		attrs := res.Attributes()
		attrs.PutString(conventions.AttributeCloudProvider, conventions.AttributeCloudProviderAzure)
		attrs.PutString(conventions.AttributeCloudPlatform, conventions.AttributeCloudPlatformAzureVM)
		attrs.PutString(conventions.AttributeHostName, "my.fq.host.name")
		attrs.PutString(conventions.AttributeCloudRegion, "myCloudRegion")
		attrs.PutString(conventions.AttributeHostID, "myHostID")
		attrs.PutString(conventions.AttributeCloudAccountID, "myCloudAccount")
		attrs.PutString("azure.vm.name", "myVMScalesetName_1")
		attrs.PutString("azure.vm.size", "42")
		attrs.PutString("azure.vm.scaleset.name", "myVMScalesetName")
		attrs.PutString("azure.resourcegroup.name", "myResourcegroupName")
		return res
	}()
	azureMissingCloudAcct = func() pcommon.Resource {
		res := pcommon.NewResource()
		attrs := res.Attributes()
		attrs.PutString(conventions.AttributeCloudProvider, conventions.AttributeCloudProviderAzure)
		attrs.PutString(conventions.AttributeCloudPlatform, conventions.AttributeCloudPlatformAzureVM)
		attrs.PutString(conventions.AttributeCloudRegion, "myCloudRegion")
		attrs.PutString(conventions.AttributeHostID, "myHostID")
		attrs.PutString("azure.vm.size", "42")
		attrs.PutString("azure.resourcegroup.name", "myResourcegroupName")
		return res
	}()
	azureMissingResourceGroup = func() pcommon.Resource {
		res := pcommon.NewResource()
		attrs := res.Attributes()
		attrs.PutString(conventions.AttributeCloudProvider, conventions.AttributeCloudProviderAzure)
		attrs.PutString(conventions.AttributeCloudPlatform, conventions.AttributeCloudPlatformAzureVM)
		attrs.PutString(conventions.AttributeCloudRegion, "myCloudRegion")
		attrs.PutString(conventions.AttributeHostID, "myHostID")
		attrs.PutString(conventions.AttributeCloudAccountID, "myCloudAccount")
		attrs.PutString("azure.vm.size", "42")
		return res
	}()
	azureMissingHostName = func() pcommon.Resource {
		res := pcommon.NewResource()
		attrs := res.Attributes()
		attrs.PutString(conventions.AttributeCloudProvider, conventions.AttributeCloudProviderAzure)
		attrs.PutString(conventions.AttributeCloudPlatform, conventions.AttributeCloudPlatformAzureVM)
		attrs.PutString(conventions.AttributeCloudRegion, "myCloudRegion")
		attrs.PutString(conventions.AttributeHostID, "myHostID")
		attrs.PutString(conventions.AttributeCloudAccountID, "myCloudAccount")
		attrs.PutString("azure.resourcegroup.name", "myResourcegroupName")
		attrs.PutString("azure.vm.size", "42")
		return res
	}()
	hostResource = func() pcommon.Resource {
		res := pcommon.NewResource()
		attr := res.Attributes()
		attr.PutString(conventions.AttributeHostName, "localhost")
		return res
	}()
	unknownResource = func() pcommon.Resource {
		res := pcommon.NewResource()
		attr := res.Attributes()
		attr.PutString(conventions.AttributeCloudProvider, "unknown")
		attr.PutString(conventions.AttributeCloudAccountID, "1234")
		attr.PutString(conventions.AttributeHostID, "i-abcd")
		return res
	}()
)

func TestResourceToHostID(t *testing.T) {
	type args struct {
		res pcommon.Resource
	}
	tests := []struct {
		name string
		args args
		want HostID
		ok   bool
	}{
		{
			name: "nil resource",
			args: args{pcommon.NewResource()},
			want: HostID{},
			ok:   false,
		},
		{
			name: "ec2",
			args: args{ec2Resource},
			want: HostID{
				Key: "AWSUniqueId",
				ID:  "i-abcd_us-west-2_1234",
			},
			ok: true,
		},
		{
			name: "ec2 with hostname prefers ec2",
			args: args{ec2WithHost},
			want: HostID{
				Key: "AWSUniqueId",
				ID:  "i-abcd_us-west-2_1234",
			},
			ok: true,
		},
		{
			name: "gcp",
			args: args{gcpResource},
			want: HostID{
				Key: "gcp_id",
				ID:  "1234_i-abcd",
			},
			ok: true,
		},
		{
			name: "azure",
			args: args{azureResource},
			want: HostID{
				Key: "azure_resource_id",
				ID:  "mycloudaccount/myresourcegroupname/microsoft.compute/virtualmachines/myvmname",
			},
			ok: true,
		},
		{
			name: "azure scaleset",
			args: args{azureScalesetResource},
			want: HostID{
				Key: "azure_resource_id",
				ID:  "mycloudaccount/myresourcegroupname/microsoft.compute/virtualmachinescalesets/myvmscalesetname/virtualmachines/1",
			},
			ok: true,
		},
		{
			name: "azure cloud account missing",
			args: args{azureMissingCloudAcct},
			want: HostID{},
			ok:   false,
		},
		{
			name: "azure resource group missing",
			args: args{azureMissingResourceGroup},
			want: HostID{},
			ok:   false,
		},
		{
			name: "azure hostname missing",
			args: args{azureMissingHostName},
			want: HostID{},
			ok:   false,
		},
		{
			name: "ec2 attributes missing",
			args: args{ec2PartialResource},
			want: HostID{},
			ok:   false,
		},
		{
			name: "gcp attributes missing",
			args: args{gcpPartialResource},
			want: HostID{},
			ok:   false,
		},
		{
			name: "unknown provider",
			args: args{unknownResource},
			want: HostID{},
			ok:   false,
		},
		{
			name: "host provider",
			args: args{hostResource},
			want: HostID{
				Key: "host.name",
				ID:  "localhost",
			},
			ok: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hostID, ok := ResourceToHostID(tt.args.res)
			assert.Equal(t, tt.ok, ok)
			assert.Equal(t, tt.want, hostID)
		})
	}
}
