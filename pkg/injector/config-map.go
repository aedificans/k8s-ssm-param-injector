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
	"encoding/json"
	"net/http"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (s *SSMParameterInjector) handleConfigMap(ctx context.Context, req admission.Request) admission.Response {
	configMap := &corev1.ConfigMap{}

	log.Log.V(1).Info("Decoding ConfigMap from request")
	err := s.Decoder.Decode(req, configMap)
	if err != nil {
		log.Log.Error(err, "unable to decode ConfigMap")
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.Log.WithValues("name", configMap.Name, "namespace", configMap.Namespace).
		V(1).Info("ConfigMap successfully decoded")

	wasModified := false
	if configMap.Data != nil {
		for key, value := range configMap.Data {
			if strings.HasPrefix(value, "ssm:/") {
				log.Log.Info("SSM Parameter detected in ConfigMap data")
				log.Log.WithValues("paramKey", value).
					V(1).Info("SSM Parameter detected")
				paramName := strings.TrimPrefix(value, "ssm:/")
				paramValue, err := s.getSSMParameter(ctx, paramName)
				if err != nil {
					return admission.Errored(http.StatusInternalServerError, err)
				}
				log.Log.V(1).Info("Updating ConfigMap data with SSM Parameter value")
				configMap.Data[key] = paramValue
				wasModified = true
			}
		}
	}

	if !wasModified {
		log.Log.Info("No SSM parameters found")
		return admission.Allowed("No modifications required")
	}

	configMapJson, err := json.Marshal(configMap)
	if err != nil {
		log.Log.Error(err, "unable to marshal modified ConfigMap to JSON")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	log.Log.Info("Returning JSON patch for value injection(s)")
	return admission.PatchResponseFromRaw(req.Object.Raw, configMapJson)
}
