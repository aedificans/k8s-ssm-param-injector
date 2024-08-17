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
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

type SSMParameterInjector struct {
	SsmClient *ssm.Client
	Decoder   admission.Decoder
}

func (s *SSMParameterInjector) Handle(ctx context.Context, req admission.Request) admission.Response {
	switch req.Kind.Kind {
	case "ConfigMap":
		log.Log.WithValues("action", req.Operation).Info("ConfigMap request received")
		return s.handleConfigMap(ctx, req)
	case "CronJob":
		log.Log.WithValues("action", req.Operation).Info("CronJob request received")
		return s.handleCronJob(ctx, req)
	case "ExternalSecret":
		log.Log.WithValues("action", req.Operation).Info("ExternalSecret request received")
		return s.handleExternalSecret(ctx, req)
	case "Ingress":
		log.Log.WithValues("action", req.Operation).Info("Ingress request received")
		return s.handleIngress(ctx, req)
	case "Job":
		log.Log.WithValues("action", req.Operation).Info("Job request received")
		return s.handleJob(ctx, req)
	case "Pod":
		log.Log.WithValues("action", req.Operation).Info("Pod request received")
		return s.handlePod(ctx, req)
	case "ServiceAccount":
		log.Log.WithValues("action", req.Operation).Info("ServiceAccount request received")
		return s.handleServiceAccount(ctx, req)
	default:
		log.Log.WithValues("action", req.Operation).Error(nil, "unsupported Kind")
		return admission.Errored(http.StatusBadRequest, fmt.Errorf("unsupported Kind: %s", req.Kind.Kind))
	}
}
