/*
Copyright The Kubernetes Authors.

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

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/mrogers950/csr-approver-operator/pkg/apis/csrapprover.config.openshift.io/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// CSRApproverOperatorConfigLister helps list CSRApproverOperatorConfigs.
type CSRApproverOperatorConfigLister interface {
	// List lists all CSRApproverOperatorConfigs in the indexer.
	List(selector labels.Selector) (ret []*v1alpha1.CSRApproverOperatorConfig, err error)
	// Get retrieves the CSRApproverOperatorConfig from the index for a given name.
	Get(name string) (*v1alpha1.CSRApproverOperatorConfig, error)
	CSRApproverOperatorConfigListerExpansion
}

// cSRApproverOperatorConfigLister implements the CSRApproverOperatorConfigLister interface.
type cSRApproverOperatorConfigLister struct {
	indexer cache.Indexer
}

// NewCSRApproverOperatorConfigLister returns a new CSRApproverOperatorConfigLister.
func NewCSRApproverOperatorConfigLister(indexer cache.Indexer) CSRApproverOperatorConfigLister {
	return &cSRApproverOperatorConfigLister{indexer: indexer}
}

// List lists all CSRApproverOperatorConfigs in the indexer.
func (s *cSRApproverOperatorConfigLister) List(selector labels.Selector) (ret []*v1alpha1.CSRApproverOperatorConfig, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.CSRApproverOperatorConfig))
	})
	return ret, err
}

// Get retrieves the CSRApproverOperatorConfig from the index for a given name.
func (s *cSRApproverOperatorConfigLister) Get(name string) (*v1alpha1.CSRApproverOperatorConfig, error) {
	obj, exists, err := s.indexer.GetByKey(name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("csrapproveroperatorconfig"), name)
	}
	return obj.(*v1alpha1.CSRApproverOperatorConfig), nil
}
