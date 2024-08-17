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

func (s *SSMParameterInjector) handlePod(ctx context.Context, req admission.Request) admission.Response {
	pod := &corev1.Pod{}

	log.Log.V(1).Info("Decoding Pod from request")
	err := s.Decoder.Decode(req, pod)
	if err != nil {
		log.Log.Error(err, "unable to decode Pod")
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.Log.WithValues("name", pod.Name, "namespace", pod.Namespace).
		V(1).Info("Pod successfully decoded")

	hasUpdatedContainers := false
	if pod.Spec.Containers != nil {
		hasUpdatedContainers, err = s.processContainers(ctx, pod.Spec.Containers)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	hasUpdatedInitContainers := false
	if pod.Spec.InitContainers != nil {
		hasUpdatedInitContainers, err = s.processContainers(ctx, pod.Spec.InitContainers)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	if !hasUpdatedContainers && !hasUpdatedInitContainers {
		log.Log.Info("No SSM parameters found")
		return admission.Allowed("No modifications required")
	}

	podJson, err := json.Marshal(pod)
	if err != nil {
		log.Log.Error(err, "unable to marshal modified Pod to JSON")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	log.Log.Info("Returning JSON patch for value injection(s)")
	return admission.PatchResponseFromRaw(req.Object.Raw, podJson)
}
