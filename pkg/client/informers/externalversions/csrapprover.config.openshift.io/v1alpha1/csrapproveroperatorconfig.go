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

// Code generated by informer-gen. DO NOT EDIT.

package v1alpha1

import (
	time "time"

	csrapproverconfigopenshiftiov1alpha1 "github.com/mrogers950/csr-approver-operator/pkg/apis/csrapprover.config.openshift.io/v1alpha1"
	versioned "github.com/mrogers950/csr-approver-operator/pkg/client/clientset/versioned"
	internalinterfaces "github.com/mrogers950/csr-approver-operator/pkg/client/informers/externalversions/internalinterfaces"
	v1alpha1 "github.com/mrogers950/csr-approver-operator/pkg/client/listers/csrapprover.config.openshift.io/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
	watch "k8s.io/apimachinery/pkg/watch"
	cache "k8s.io/client-go/tools/cache"
)

// CSRApproverOperatorConfigInformer provides access to a shared informer and lister for
// CSRApproverOperatorConfigs.
type CSRApproverOperatorConfigInformer interface {
	Informer() cache.SharedIndexInformer
	Lister() v1alpha1.CSRApproverOperatorConfigLister
}

type cSRApproverOperatorConfigInformer struct {
	factory          internalinterfaces.SharedInformerFactory
	tweakListOptions internalinterfaces.TweakListOptionsFunc
}

// NewCSRApproverOperatorConfigInformer constructs a new informer for CSRApproverOperatorConfig type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewCSRApproverOperatorConfigInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers) cache.SharedIndexInformer {
	return NewFilteredCSRApproverOperatorConfigInformer(client, resyncPeriod, indexers, nil)
}

// NewFilteredCSRApproverOperatorConfigInformer constructs a new informer for CSRApproverOperatorConfig type.
// Always prefer using an informer factory to get a shared informer instead of getting an independent
// one. This reduces memory footprint and number of connections to the server.
func NewFilteredCSRApproverOperatorConfigInformer(client versioned.Interface, resyncPeriod time.Duration, indexers cache.Indexers, tweakListOptions internalinterfaces.TweakListOptionsFunc) cache.SharedIndexInformer {
	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(options v1.ListOptions) (runtime.Object, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CsrapproverV1alpha1().CSRApproverOperatorConfigs().List(options)
			},
			WatchFunc: func(options v1.ListOptions) (watch.Interface, error) {
				if tweakListOptions != nil {
					tweakListOptions(&options)
				}
				return client.CsrapproverV1alpha1().CSRApproverOperatorConfigs().Watch(options)
			},
		},
		&csrapproverconfigopenshiftiov1alpha1.CSRApproverOperatorConfig{},
		resyncPeriod,
		indexers,
	)
}

func (f *cSRApproverOperatorConfigInformer) defaultInformer(client versioned.Interface, resyncPeriod time.Duration) cache.SharedIndexInformer {
	return NewFilteredCSRApproverOperatorConfigInformer(client, resyncPeriod, cache.Indexers{cache.NamespaceIndex: cache.MetaNamespaceIndexFunc}, f.tweakListOptions)
}

func (f *cSRApproverOperatorConfigInformer) Informer() cache.SharedIndexInformer {
	return f.factory.InformerFor(&csrapproverconfigopenshiftiov1alpha1.CSRApproverOperatorConfig{}, f.defaultInformer)
}

func (f *cSRApproverOperatorConfigInformer) Lister() v1alpha1.CSRApproverOperatorConfigLister {
	return v1alpha1.NewCSRApproverOperatorConfigLister(f.Informer().GetIndexer())
}
