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
package controllers

import (
	"fmt"
)

type VersionService interface {
	LatestCephVersion(endpoint string) (string, error)
	LatestRookVersion(endpoint string) (string, error)
}

type VersionServiceClient struct {
}

func (vc *VersionServiceClient) LatestCephVersion(_ string) (string, error) {
	return "", fmt.Errorf("Not implemented")
}

func (vc *VersionServiceClient) LatestRookVersion(_ string) (string, error) {
	return "", fmt.Errorf("Not implemented")
}
