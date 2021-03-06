/*
Copyright 2017 The Kubernetes Authors.

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

package handlers

import (
	"context"
	"fmt"
	"net/http"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apiserver/pkg/endpoints/handlers"
	"k8s.io/apiserver/pkg/endpoints/handlers/responsewriters"
	"k8s.io/apiserver/pkg/endpoints/request"
	"k8s.io/klog"
)

func writeError(scope *handlers.RequestScope, err error, w http.ResponseWriter, req *http.Request) {
	responsewriters.ErrorNegotiated(err, scope.Serializer, scope.Kind.GroupVersion(), w, req)
}

// setSelfLink sets the self link of an object (or the child items in a list) to the base URL of the request
// plus the path and query generated by the provided linkFunc
func setSelfLink(obj runtime.Object, requestInfo *request.RequestInfo, namer handlers.ScopeNamer) error {
	// TODO: SelfLink generation should return a full URL?
	uri, err := namer.GenerateLink(requestInfo, obj)
	if err != nil {
		return nil
	}

	return namer.SetSelfLink(obj, uri)
}

// setListSelfLink sets the self link of a list to the base URL, then sets the self links
// on all child objects returned. Returns the number of items in the list.
func setListSelfLink(obj runtime.Object, ctx context.Context, req *http.Request, namer handlers.ScopeNamer) (int, error) {
	if !meta.IsListType(obj) {
		return 0, nil
	}

	uri, err := namer.GenerateListLink(req)
	if err != nil {
		return 0, err
	}
	if err := namer.SetSelfLink(obj, uri); err != nil {
		klog.V(4).Infof("Unable to set self link on object: %v", err)
	}
	requestInfo, ok := request.RequestInfoFrom(ctx)
	if !ok {
		return 0, fmt.Errorf("missing requestInfo")
	}

	count := 0
	err = meta.EachListItem(obj, func(obj runtime.Object) error {
		count++
		return setSelfLink(obj, requestInfo, namer)
	})
	return count, err
}
