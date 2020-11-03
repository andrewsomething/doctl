/*
Copyright 2018 The Doctl Authors All rights reserved.
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

package do

import (
	"context"
	"fmt"

	"github.com/digitalocean/godo"
)

// RegistryHostname is the hostname for the DO registry
const RegistryHostname = "registry.digitalocean.com"

// Registry wraps a godo Registry.
type Registry struct {
	*godo.Registry
}

// Repository wraps a godo Repository
type Repository struct {
	*godo.Repository
}

// RepositoryTag wraps a godo RepositoryTag
type RepositoryTag struct {
	*godo.RepositoryTag
}

// GarbageCollection wraps a godo GarbageCollection
type GarbageCollection struct {
	*godo.GarbageCollection
}

// RegistrySubscriptionTier wraps a godo RegistrySubscriptionTier
type RegistrySubscriptionTier struct {
	*godo.RegistrySubscriptionTier
}

// Endpoint returns the registry endpoint for image tagging
func (r *Registry) Endpoint() string {
	return fmt.Sprintf("%s/%s", RegistryHostname, r.Registry.Name)
}

// RegistryService is the godo RegistryService interface.
type RegistryService interface {
	Get() (*Registry, error)
	Create(*godo.RegistryCreateRequest) (*Registry, error)
	Delete() error
	DockerCredentials(*godo.RegistryDockerCredentialsRequest) (*godo.DockerCredentials, error)
	ListRepositoryTags(string, string) ([]RepositoryTag, error)
	ListRepositories(string) ([]Repository, error)
	DeleteTag(string, string, string) error
	DeleteManifest(string, string, string) error
	Endpoint() string
	StartGarbageCollection(string) (*GarbageCollection, error)
	GetGarbageCollection(string) (*GarbageCollection, error)
	ListGarbageCollections(string) ([]GarbageCollection, error)
	CancelGarbageCollection(string, string) (*GarbageCollection, error)
	GetSubscriptionTiers() ([]RegistrySubscriptionTier, error)
}

type registryService struct {
	client *godo.Client
	ctx    context.Context
}

var _ RegistryService = &registryService{}

// NewRegistryService builds an instance of RegistryService.
func NewRegistryService(client *godo.Client) RegistryService {
	return &registryService{
		client: client,
		ctx:    context.Background(),
	}
}

func (rs *registryService) Get() (*Registry, error) {
	r, _, err := rs.client.Registry.Get(rs.ctx)
	if err != nil {
		return nil, err
	}

	return &Registry{Registry: r}, nil
}

func (rs *registryService) Create(cr *godo.RegistryCreateRequest) (*Registry, error) {
	r, _, err := rs.client.Registry.Create(rs.ctx, cr)
	if err != nil {
		return nil, err
	}

	return &Registry{Registry: r}, nil
}

func (rs *registryService) Delete() error {
	_, err := rs.client.Registry.Delete(rs.ctx)
	return err
}

func (rs *registryService) DockerCredentials(request *godo.RegistryDockerCredentialsRequest) (*godo.DockerCredentials, error) {
	dockerConfig, _, err := rs.client.Registry.DockerCredentials(rs.ctx, request)
	if err != nil {
		return nil, err
	}

	return dockerConfig, nil
}

func (rs *registryService) ListRepositories(registry string) ([]Repository, error) {
	f := func(opt *godo.ListOptions) ([]interface{}, *godo.Response, error) {
		list, resp, err := rs.client.Registry.ListRepositories(rs.ctx, registry, opt)
		if err != nil {
			return nil, nil, err
		}

		si := make([]interface{}, len(list))
		for i := range list {
			si[i] = list[i]
		}

		return si, resp, err
	}

	si, err := PaginateResp(f)
	if err != nil {
		return nil, err
	}

	list := make([]Repository, len(si))
	for i := range si {
		a := si[i].(*godo.Repository)
		list[i] = Repository{Repository: a}
	}

	return list, nil
}

func (rs *registryService) ListRepositoryTags(registry, repository string) ([]RepositoryTag, error) {
	f := func(opt *godo.ListOptions) ([]interface{}, *godo.Response, error) {
		list, resp, err := rs.client.Registry.ListRepositoryTags(rs.ctx, registry, repository, opt)
		if err != nil {
			return nil, nil, err
		}

		si := make([]interface{}, len(list))
		for i := range list {
			si[i] = list[i]
		}

		return si, resp, err
	}

	si, err := PaginateResp(f)
	if err != nil {
		return nil, err
	}

	list := make([]RepositoryTag, len(si))
	for i := range si {
		a := si[i].(*godo.RepositoryTag)
		list[i] = RepositoryTag{RepositoryTag: a}
	}

	return list, nil
}

func (rs *registryService) DeleteTag(registry, repository, tag string) error {
	_, err := rs.client.Registry.DeleteTag(rs.ctx, registry, repository, tag)
	return err
}

func (rs *registryService) DeleteManifest(registry, repository, digest string) error {
	_, err := rs.client.Registry.DeleteManifest(rs.ctx, registry, repository, digest)
	return err
}

func (rs *registryService) StartGarbageCollection(registry string) (*GarbageCollection, error) {
	gc, _, err := rs.client.Registry.StartGarbageCollection(rs.ctx, registry)
	if err != nil {
		return nil, err
	}

	return &GarbageCollection{gc}, nil
}

func (rs *registryService) GetGarbageCollection(registry string) (*GarbageCollection, error) {
	gc, _, err := rs.client.Registry.GetGarbageCollection(rs.ctx, registry)
	if err != nil {
		return nil, err
	}

	return &GarbageCollection{gc}, nil
}

func (rs *registryService) ListGarbageCollections(registry string) ([]GarbageCollection, error) {
	f := func(opt *godo.ListOptions) ([]interface{}, *godo.Response, error) {
		list, resp, err := rs.client.Registry.ListGarbageCollections(rs.ctx, registry, opt)
		if err != nil {
			return nil, nil, err
		}

		si := make([]interface{}, len(list))
		for i := range list {
			si[i] = list[i]
		}

		return si, resp, err
	}

	si, err := PaginateResp(f)
	if err != nil {
		return nil, err
	}

	list := make([]GarbageCollection, len(si))
	for i := range si {
		a := si[i].(*godo.GarbageCollection)
		list[i] = GarbageCollection{a}
	}

	return list, err
}

func (rs *registryService) CancelGarbageCollection(registry, garbageCollection string) (*GarbageCollection, error) {
	gc, _, err := rs.client.Registry.UpdateGarbageCollection(rs.ctx, registry,
		garbageCollection, &godo.UpdateGarbageCollectionRequest{
			Cancel: true,
		})
	if err != nil {
		return nil, err
	}

	return &GarbageCollection{gc}, nil
}

func (rs *registryService) Endpoint() string {
	return RegistryHostname
}

func (rs *registryService) GetSubscriptionTiers() ([]RegistrySubscriptionTier, error) {
	opts, _, err := rs.client.Registry.GetOptions(rs.ctx)
	if err != nil {
		return nil, err
	}

	ret := make([]RegistrySubscriptionTier, len(opts.SubscriptionTiers))
	for i, tier := range opts.SubscriptionTiers {
		ret[i] = RegistrySubscriptionTier{tier}
	}

	return ret, nil
}
