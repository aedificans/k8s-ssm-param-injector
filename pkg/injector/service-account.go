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

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (s *SSMParameterInjector) handleServiceAccount(ctx context.Context, req admission.Request) admission.Response {
	serviceAccount := &corev1.ServiceAccount{}

	log.Log.V(1).Info("Decoding ServiceAccount from request")
	err := s.Decoder.Decode(req, serviceAccount)
	if err != nil {
		log.Log.Error(err, "unable to decode ServiceAccount")
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.Log.WithValues("name", serviceAccount.Name, "namespace", serviceAccount.Namespace).
		V(1).Info("ServiceAccount successfully decoded")

	if serviceAccount.Annotations == nil {
		log.Log.Info("No annotations present on the ServiceAccount")
		return admission.Allowed("No modifications required")
	}

	wasModified, err := s.processAnnotations(ctx, serviceAccount.Annotations)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	if !wasModified {
		log.Log.Info("No SSM parameters found")
		return admission.Allowed("No modifications required")
	}

	serviceAccountJson, err := json.Marshal(serviceAccount)
	if err != nil {
		log.Log.Error(err, "unable to marshal modified ServiceAccount to JSON")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	log.Log.Info("Returning JSON patch for value injection(s)")
	return admission.PatchResponseFromRaw(req.Object.Raw, serviceAccountJson)
}
