//
// Copyright 2021 METRO Digital GmbH
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
//
//go:build unitTests
// +build unitTests

package main_test

import (
	"context"

	. "github.com/metro-digital/kustomize-google-secret-manager/main"
)

var ctx = context.Background()

func createEncryptedGCPSecret(name string, secretKey string) KGCPSecret {
	return KGCPSecret{
		TypeMeta: TypeMeta{
			APIVersion: "metro.digital/v1",
			Kind:       "KGCPSecret",
		},
		GCPObjectMeta: GCPObjectMeta{
			Dc:          "",
			Stage:       "",
			Name:        name,
			Namespace:   "",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		GCPProjectID:          "cf-2tier-uhd-test-d7",
		DisableNameSuffixHash: true,
		Type:                  "",
		Behavior:              "",
		DataType:              "",
		Keys: []string{
			secretKey,
		},
	}
}

func createExpectedK8SSecret(name string, secretKey string, secretValue string) K8SSecret {
	return K8SSecret{
		TypeMeta: TypeMeta{
			APIVersion: "v1",
			Kind:       "Secret",
		},
		ObjectMeta: ObjectMeta{
			Name:        name,
			Namespace:   "",
			Labels:      map[string]string{},
			Annotations: map[string]string{},
		},
		Data: map[string]string{
			secretKey: secretValue,
		},
		Type: "",
	}
}
