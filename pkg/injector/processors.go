/*
Copyright 2024.

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

package injector

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

func (s *SSMParameterInjector) processAnnotations(ctx context.Context, annotations map[string]string) (bool, error) {
	wasModified := false

	for key, value := range annotations {
		if strings.HasPrefix(value, "ssm:/") {
			log.Log.Info("SSM Parameter detected in annotation")
			log.Log.WithValues("paramKey", value).
				V(1).Info("SSM Parameter detected")
			paramName := strings.TrimPrefix(value, "ssm:/")
			paramValue, err := s.getSSMParameter(ctx, paramName)
			if err != nil {
				return false, err
			}
			log.Log.V(1).Info("Updating annotation with SSM Parameter value")
			annotations[key] = paramValue
			wasModified = true
		}
	}

	return wasModified, nil
}

func (s *SSMParameterInjector) processContainers(ctx context.Context, containers []corev1.Container) (bool, error) {
	wasModified := false

	for i, container := range containers {
		for j, envVar := range container.Env {
			if strings.HasPrefix(envVar.Value, "ssm:/") {
				log.Log.Info("SSM Parameter detected in container environment variable value")
				log.Log.WithValues("paramKey", envVar.Value).
					V(1).Info("SSM Parameter detected")
				paramName := strings.TrimPrefix(envVar.Value, "ssm:/")
				paramValue, err := s.getSSMParameter(ctx, paramName)
				if err != nil {
					return false, err
				}
				log.Log.V(1).Info("Updating container environment variable value with SSM Parameter value")
				containers[i].Env[j].Value = paramValue
				wasModified = true
			}
		}
	}

	return wasModified, nil
}

func (s *SSMParameterInjector) getSSMParameter(ctx context.Context, paramName string) (string, error) {
	WithDecryption := true
	ssmRequestInput := &ssm.GetParameterInput{
		Name:           &paramName,
		WithDecryption: &WithDecryption,
	}

	log.Log.WithValues("paramName", paramName).V(1).Info("Retrieving SSM Parameter value")
	ssmResponse, err := s.SsmClient.GetParameter(ctx, ssmRequestInput)
	if err != nil {
		log.Log.WithValues("paramName", paramName).Error(err, "failed to retrieve SSM parameter")
		return "", fmt.Errorf("failed to retrieve SSM parameter: %s", err)
	}

	log.Log.WithValues("paramValue", *ssmResponse.Parameter.Value).
		V(2).Info("SSM Parameter retrieved value")
	log.Log.V(1).Info("Returning retrieved SSM Parameter value")
	return *ssmResponse.Parameter.Value, nil
}
