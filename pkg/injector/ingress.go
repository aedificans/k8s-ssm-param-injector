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

	networkingV1 "k8s.io/api/networking/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func (s *SSMParameterInjector) handleIngress(ctx context.Context, req admission.Request) admission.Response {
	ingress := &networkingV1.Ingress{}

	log.Log.V(1).Info("Decoding Ingress from request")
	err := s.Decoder.Decode(req, ingress)
	if err != nil {
		log.Log.Error(err, "unable to decode Ingress")
		return admission.Errored(http.StatusBadRequest, err)
	}
	log.Log.WithValues("name", ingress.Name, "namespace", ingress.Namespace).
		V(1).Info("Ingress successfully decoded")

	hasUpdatedAnnotations := false
	if ingress.Annotations != nil {
		hasUpdatedAnnotations, err = s.processAnnotations(ctx, ingress.Annotations)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	hasUpdatedRuleHosts := false
	if ingress.Spec.Rules != nil {
		hasUpdatedRuleHosts, err = s.processIngressRules(ctx, ingress)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	hasUpdatedTLSHosts := false
	if ingress.Spec.TLS != nil {
		hasUpdatedTLSHosts, err = s.processIngressTLS(ctx, ingress)
		if err != nil {
			return admission.Errored(http.StatusInternalServerError, err)
		}
	}

	if !hasUpdatedAnnotations && !hasUpdatedRuleHosts && !hasUpdatedTLSHosts {
		log.Log.Info("No SSM parameters found")
		return admission.Allowed("No modifications required")
	}

	ingressJson, err := json.Marshal(ingress)
	if err != nil {
		log.Log.Error(err, "unable to marshal modified Ingress to JSON")
		return admission.Errored(http.StatusInternalServerError, err)
	}

	log.Log.Info("Returning JSON patch for value injection(s)")
	return admission.PatchResponseFromRaw(req.Object.Raw, ingressJson)
}

func (s *SSMParameterInjector) processIngressRules(ctx context.Context, ingress *networkingV1.Ingress) (bool, error) {
	wasModified := false

	for i, rule := range ingress.Spec.Rules {
		if strings.HasPrefix(rule.Host, "ssm:/") {
			log.Log.Info("SSM Parameter detected in Ingress rule")
			log.Log.WithValues("paramKey", rule.Host).
				V(1).Info("SSM Parameter detected")
			paramName := strings.TrimPrefix(rule.Host, "ssm:/")
			paramValue, err := s.getSSMParameter(ctx, paramName)
			if err != nil {
				return false, err
			}
			log.Log.V(1).Info("Updating Ingress rule hostname with SSM Parameter value")
			ingress.Spec.Rules[i].Host = paramValue
			wasModified = true
		}
	}

	return wasModified, nil
}

func (s *SSMParameterInjector) processIngressTLS(ctx context.Context, ingress *networkingV1.Ingress) (bool, error) {
	wasModified := false

	for i, tls := range ingress.Spec.TLS {
		if tls.Hosts != nil {
			for j, host := range tls.Hosts {
				if strings.HasPrefix(host, "ssm:/") {
					log.Log.Info("SSM Parameter detected in Ingress TLS hosts")
					log.Log.WithValues("paramKey", host).
						V(1).Info("SSM Parameter detected")
					paramName := strings.TrimPrefix(host, "ssm:/")
					paramValue, err := s.getSSMParameter(ctx, paramName)
					if err != nil {
						return false, err
					}
					log.Log.V(1).Info("Updating Ingress TLS host with SSM Parameter value")
					ingress.Spec.TLS[i].Hosts[j] = paramValue
					wasModified = true
				}
			}
		}
	}

	return wasModified, nil
}
