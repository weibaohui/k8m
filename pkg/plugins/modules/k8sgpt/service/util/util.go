/*
Copyright 2023 The K8sGPT Authors.
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

package util

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/weibaohui/kom/kom"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	v2 "k8s.io/api/events/v1"
	networkingv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/labels"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var anonymizePattern = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*()-_=+[]{}|;':\",./<>?")

func GetParent(ctx context.Context, clusterID string, meta metav1.ObjectMeta) (string, bool) {
	if meta.OwnerReferences != nil {
		for _, owner := range meta.OwnerReferences {
			switch owner.Kind {
			case "ReplicaSet":
				var rs *appsv1.ReplicaSet

				err := kom.Cluster(clusterID).WithContext(ctx).Resource(&appsv1.ReplicaSet{}).Namespace(meta.Namespace).Name(owner.Name).Get(&rs).Error
				if err != nil {
					return "", false
				}
				if rs.OwnerReferences != nil {
					return GetParent(ctx, clusterID, rs.ObjectMeta)
				}
				return "ReplicaSet/" + rs.Name, true

			case "Deployment":
				var dep *appsv1.Deployment
				err := kom.Cluster(clusterID).WithContext(ctx).Resource(&appsv1.Deployment{}).Namespace(meta.Namespace).Name(owner.Name).Get(&dep).Error
				if err != nil {
					return "", false
				}
				if dep.OwnerReferences != nil {
					return GetParent(ctx, clusterID, dep.ObjectMeta)
				}
				return "Deployment/" + dep.Name, true

			case "StatefulSet":
				var sts *appsv1.StatefulSet
				err := kom.Cluster(clusterID).WithContext(ctx).Resource(&appsv1.StatefulSet{}).Namespace(meta.Namespace).Name(owner.Name).Get(&sts).Error
				if err != nil {
					return "", false
				}
				if sts.OwnerReferences != nil {
					return GetParent(ctx, clusterID, sts.ObjectMeta)
				}
				return "StatefulSet/" + sts.Name, true

			case "DaemonSet":
				var ds *appsv1.DaemonSet
				err := kom.Cluster(clusterID).WithContext(ctx).Resource(&appsv1.DaemonSet{}).Namespace(meta.Namespace).Name(owner.Name).Get(&ds).Error

				if err != nil {
					return "", false
				}
				if ds.OwnerReferences != nil {
					return GetParent(ctx, clusterID, ds.ObjectMeta)
				}
				return "DaemonSet/" + ds.Name, true

			case "Ingress":
				var ing *networkingv1.Ingress
				err := kom.Cluster(clusterID).WithContext(ctx).Resource(&networkingv1.Ingress{}).Namespace(meta.Namespace).Name(owner.Name).Get(&ing).Error

				if err != nil {
					return "", false
				}
				if ing.OwnerReferences != nil {
					return GetParent(ctx, clusterID, ing.ObjectMeta)
				}
				return "Ingress/" + ing.Name, true

			case "MutatingWebhookConfiguration":
				var mw *admissionregistrationv1.MutatingWebhookConfiguration
				err := kom.Cluster(clusterID).WithContext(ctx).Resource(&admissionregistrationv1.MutatingWebhookConfiguration{}).Namespace(meta.Namespace).Name(owner.Name).Get(&mw).Error
				if err != nil {
					return "", false
				}
				if mw.OwnerReferences != nil {
					return GetParent(ctx, clusterID, mw.ObjectMeta)
				}
				return "MutatingWebhook/" + mw.Name, true

			case "ValidatingWebhookConfiguration":
				var vw *admissionregistrationv1.ValidatingWebhookConfiguration
				err := kom.Cluster(clusterID).WithContext(ctx).Resource(&admissionregistrationv1.ValidatingWebhookConfiguration{}).Namespace(meta.Namespace).Name(owner.Name).Get(&vw).Error
				if err != nil {
					return "", false
				}
				if vw.OwnerReferences != nil {
					return GetParent(ctx, clusterID, vw.ObjectMeta)
				}
				return "ValidatingWebhook/" + vw.Name, true
			}
		}
	}
	return "", false
}

func RemoveDuplicates(slice []string) ([]string, []string) {
	set := make(map[string]bool)
	duplicates := []string{}
	for _, val := range slice {
		if _, ok := set[val]; !ok {
			set[val] = true
		} else {
			duplicates = append(duplicates, val)
		}
	}
	uniqueSlice := make([]string, 0, len(set))
	for val := range set {
		uniqueSlice = append(uniqueSlice, val)
	}
	return uniqueSlice, duplicates
}

func SliceDiff(source, dest []string) []string {
	mb := make(map[string]struct{}, len(dest))
	for _, x := range dest {
		mb[x] = struct{}{}
	}
	var diff []string
	for _, x := range source {
		if _, found := mb[x]; !found {
			diff = append(diff, x)
		}
	}
	return diff
}

func MaskString(input string) string {
	key := make([]byte, len(input))
	result := make([]rune, len(input))
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	for i := range result {
		result[i] = anonymizePattern[int(key[i])%len(anonymizePattern)]
	}
	return base64.StdEncoding.EncodeToString([]byte(string(result)))
}

func ReplaceIfMatch(text string, pattern string, replacement string) string {
	re := regexp.MustCompile(fmt.Sprintf(`%s(\b)`, pattern))
	if re.MatchString(text) {
		text = re.ReplaceAllString(text, replacement)
	}
	return text
}

func MapToString(m map[string]string) string {
	// Handle empty map case
	if len(m) == 0 {
		return ""
	}

	var pairs []string
	for k, v := range m {
		pairs = append(pairs, fmt.Sprintf("%s=%s", k, v))
	}

	// Efficient string joining
	return strings.Join(pairs, ",")
}

func LabelsIncludeAny(predefinedSelector, Labels map[string]string) bool {
	// Check if any label in the predefinedSelector exists in Labels
	for key := range predefinedSelector {
		if _, exists := Labels[key]; exists {
			return true
		}
	}

	return false
}

func FetchLatestEvent(ctx context.Context, clusterID string, namespace string, name string) (*v2.Event, error) {

	// get the list of events
	var events []*v2.Event
	err := kom.Cluster(clusterID).WithContext(ctx).Resource(&v2.Event{}).WithLabelSelector("involvedObject.name=" + name).Namespace(namespace).List(&events).Error

	if err != nil {
		return nil, err
	}
	// find most recent event
	var latestEvent *v2.Event
	for _, event := range events {
		if latestEvent == nil {
			// this is required, as a pointer to a loop variable would always yield the latest value in the range
			latestEvent = event
		}
		if event.EventTime.After(latestEvent.EventTime.Time) {
			// this is required, as a pointer to a loop variable would always yield the latest value in the range
			latestEvent = event
		}
	}
	return latestEvent, nil
}

// NewHeaders parses a slice of strings in the format "key:value" into []http.Header
// It handles headers with the same key by appending values
func NewHeaders(customHeaders []string) []http.Header {
	headers := make(map[string][]string)

	for _, header := range customHeaders {
		vals := strings.SplitN(header, ":", 2)
		if len(vals) != 2 {
			// TODO: Handle error instead of ignoring it
			continue
		}
		key := strings.TrimSpace(vals[0])
		value := strings.TrimSpace(vals[1])

		if _, ok := headers[key]; !ok {
			headers[key] = []string{}
		}
		headers[key] = append(headers[key], value)
	}

	// Convert map to []http.Header format
	var result []http.Header
	for key, values := range headers {
		header := make(http.Header)
		for _, value := range values {
			header.Add(key, value)
		}
		result = append(result, header)
	}

	return result
}

func LabelStrToSelector(labelStr string) labels.Selector {
	if labelStr == "" {
		return nil
	}
	labelSelectorMap := make(map[string]string)
	for _, s := range strings.Split(labelStr, ",") {
		parts := strings.SplitN(s, "=", 2)
		if len(parts) == 2 {
			labelSelectorMap[parts[0]] = parts[1]
		}
	}
	return labels.SelectorFromSet(labels.Set(labelSelectorMap))
}
