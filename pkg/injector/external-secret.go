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

	externalsecrets "github.com/external-secrets/external-secrets/apis/externalsecrets/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (s *SSMParameterInjector) handleExternalSecret(ctx context.Context, req admission.Request) admission.Response {
	externalSecret := &externalsecrets.ExternalSecret{}

	log.Log.V(1).Info("Decoding ExternalSecret from request")
	err := s.Decoder.Decode(req, externalSecret)
	if err != nil {
		log.Log.Error(err, "unable to decode ExternalSecret")
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.Log.WithValues("name", externalSecret.Name, "namespace", externalSecret.Namespace).
		V(1).Info("ExternalSecret successfully decoded")

	wasModified := false
	if externalSecret.Spec.Data != nil {
		for i, data := range externalSecret.Spec.Data {
			if strings.HasPrefix(data.RemoteRef.Key, "ssm:/") {
				log.Log.Info("SSM Parameter detected in ExternalSecret remoteRef.key")
				log.Log.WithValues("paramKey", data.RemoteRef.Key).
					V(1).Info("SSM Parameter detected")
				paramName := strings.TrimPrefix(data.RemoteRef.Key, "ssm:/")
				paramValue, err := s.getSSMParameter(ctx, paramName)
				if err != nil {
					return admission.Errored(http.StatusInternalServerError, err)
				}
				log.Log.Info("Updating ExternalSecret remoteRef.key with SSM Parameter value")
				externalSecret.Spec.Data[i].RemoteRef.Key = paramValue
				wasModified = true
			}
		}
	}

	if !wasModified {
		log.Log.Info("No SSM parameters found")
		return admission.Allowed("No modifications required")
	}

	externalSecretJson, err := json.Marshal(externalSecret)
	if err != nil {
		log.Log.Error(err, "unable to marshal modified ExternalSecret to JSON")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	log.Log.Info("Returning JSON patch for value injection(s)")
	return admission.PatchResponseFromRaw(req.Object.Raw, externalSecretJson)
}
