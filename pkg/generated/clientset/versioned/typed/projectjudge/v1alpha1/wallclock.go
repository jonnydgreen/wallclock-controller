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

// Code generated by client-gen. DO NOT EDIT.

package v1alpha1

import (
	"context"
	v1alpha1 "projectjudge/projects/wallclock-controller/pkg/apis/projectjudge/v1alpha1"
	scheme "projectjudge/projects/wallclock-controller/pkg/generated/clientset/versioned/scheme"
	"time"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// WallClocksGetter has a method to return a WallClockInterface.
// A group's client should implement this interface.
type WallClocksGetter interface {
	WallClocks() WallClockInterface
}

// WallClockInterface has methods to work with WallClock resources.
type WallClockInterface interface {
	Create(ctx context.Context, wallClock *v1alpha1.WallClock, opts v1.CreateOptions) (*v1alpha1.WallClock, error)
	Update(ctx context.Context, wallClock *v1alpha1.WallClock, opts v1.UpdateOptions) (*v1alpha1.WallClock, error)
	UpdateStatus(ctx context.Context, wallClock *v1alpha1.WallClock, opts v1.UpdateOptions) (*v1alpha1.WallClock, error)
	Delete(ctx context.Context, name string, opts v1.DeleteOptions) error
	DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error
	Get(ctx context.Context, name string, opts v1.GetOptions) (*v1alpha1.WallClock, error)
	List(ctx context.Context, opts v1.ListOptions) (*v1alpha1.WallClockList, error)
	Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error)
	Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.WallClock, err error)
	WallClockExpansion
}

// wallClocks implements WallClockInterface
type wallClocks struct {
	client rest.Interface
}

// newWallClocks returns a WallClocks
func newWallClocks(c *ProjectjudgeV1alpha1Client) *wallClocks {
	return &wallClocks{
		client: c.RESTClient(),
	}
}

// Get takes name of the wallClock, and returns the corresponding wallClock object, and an error if there is any.
func (c *wallClocks) Get(ctx context.Context, name string, options v1.GetOptions) (result *v1alpha1.WallClock, err error) {
	result = &v1alpha1.WallClock{}
	err = c.client.Get().
		Resource("wallclocks").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do(ctx).
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of WallClocks that match those selectors.
func (c *wallClocks) List(ctx context.Context, opts v1.ListOptions) (result *v1alpha1.WallClockList, err error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	result = &v1alpha1.WallClockList{}
	err = c.client.Get().
		Resource("wallclocks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Do(ctx).
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested wallClocks.
func (c *wallClocks) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	var timeout time.Duration
	if opts.TimeoutSeconds != nil {
		timeout = time.Duration(*opts.TimeoutSeconds) * time.Second
	}
	opts.Watch = true
	return c.client.Get().
		Resource("wallclocks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Timeout(timeout).
		Watch(ctx)
}

// Create takes the representation of a wallClock and creates it.  Returns the server's representation of the wallClock, and an error, if there is any.
func (c *wallClocks) Create(ctx context.Context, wallClock *v1alpha1.WallClock, opts v1.CreateOptions) (result *v1alpha1.WallClock, err error) {
	result = &v1alpha1.WallClock{}
	err = c.client.Post().
		Resource("wallclocks").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(wallClock).
		Do(ctx).
		Into(result)
	return
}

// Update takes the representation of a wallClock and updates it. Returns the server's representation of the wallClock, and an error, if there is any.
func (c *wallClocks) Update(ctx context.Context, wallClock *v1alpha1.WallClock, opts v1.UpdateOptions) (result *v1alpha1.WallClock, err error) {
	result = &v1alpha1.WallClock{}
	err = c.client.Put().
		Resource("wallclocks").
		Name(wallClock.Name).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(wallClock).
		Do(ctx).
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *wallClocks) UpdateStatus(ctx context.Context, wallClock *v1alpha1.WallClock, opts v1.UpdateOptions) (result *v1alpha1.WallClock, err error) {
	result = &v1alpha1.WallClock{}
	err = c.client.Put().
		Resource("wallclocks").
		Name(wallClock.Name).
		SubResource("status").
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(wallClock).
		Do(ctx).
		Into(result)
	return
}

// Delete takes name of the wallClock and deletes it. Returns an error if one occurs.
func (c *wallClocks) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("wallclocks").
		Name(name).
		Body(&opts).
		Do(ctx).
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *wallClocks) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	var timeout time.Duration
	if listOpts.TimeoutSeconds != nil {
		timeout = time.Duration(*listOpts.TimeoutSeconds) * time.Second
	}
	return c.client.Delete().
		Resource("wallclocks").
		VersionedParams(&listOpts, scheme.ParameterCodec).
		Timeout(timeout).
		Body(&opts).
		Do(ctx).
		Error()
}

// Patch applies the patch and returns the patched wallClock.
func (c *wallClocks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *v1alpha1.WallClock, err error) {
	result = &v1alpha1.WallClock{}
	err = c.client.Patch(pt).
		Resource("wallclocks").
		Name(name).
		SubResource(subresources...).
		VersionedParams(&opts, scheme.ParameterCodec).
		Body(data).
		Do(ctx).
		Into(result)
	return
}
