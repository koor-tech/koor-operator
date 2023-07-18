/*
Copyright 2023 Koor Technologies, Inc. All rights reserved.

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

package utils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bufbuild/connect-go"
	koapi "github.com/koor-tech/koor-operator/api/v1alpha1"
	vsapi "github.com/koor-tech/version-service/api/v1"
	"github.com/koor-tech/version-service/api/v1/apiv1connect"
)

type VersionService interface {
	LatestVersions(ctx context.Context, endpoint string,
		versions *koapi.ProductVersions) (*koapi.DetailedProductVersions, error)
}

func NewVersionServiceClient() VersionService {
	return &versionServiceClient{}
}

type versionServiceClient struct{}

func (vc *versionServiceClient) LatestVersions(ctx context.Context, endpoint string,
	versions *koapi.ProductVersions) (*koapi.DetailedProductVersions, error) {
	if versions == nil {
		return nil, fmt.Errorf("current versions is empty")
	}
	client := apiv1connect.NewVersionServiceClient(
		http.DefaultClient,
		endpoint,
	)
	resp, err := client.Operator(ctx, connect.NewRequest(&vsapi.OperatorRequest{
		Versions: &vsapi.ProductVersions{
			KoorOperator: versions.KoorOperator,
			Ksd:          versions.Ksd,
			Ceph:         versions.Ceph,
		},
	}))
	if err != nil {
		return nil, fmt.Errorf("connecting to endpoint %s failed: %w", endpoint, err)
	}
	latestVersions := &koapi.DetailedProductVersions{
		KoorOperator: convertDetailedVersion(resp.Msg.Versions.KoorOperator),
		Ksd:          convertDetailedVersion(resp.Msg.Versions.Ksd),
		Ceph:         convertDetailedVersion(resp.Msg.Versions.Ceph),
	}
	return latestVersions, nil
}

func convertDetailedVersion(dv *vsapi.DetailedVersion) *koapi.DetailedVersion {
	return &koapi.DetailedVersion{
		Version:        dv.Version,
		ImageUri:       dv.ImageUri,
		ImageHash:      dv.ImageHash,
		HelmRepository: dv.HelmRepository,
		HelmChart:      dv.HelmChart,
	}
}
