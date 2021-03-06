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

package fake

import (
	"context"

	bookcontrollerv1 "github.com/boknowswiki/boknows_services/bookstore-controller/pkg/apis/bookcontroller/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeBooks implements BookInterface
type FakeBooks struct {
	Fake *FakeBookcontrollerV1
	ns   string
}

var booksResource = schema.GroupVersionResource{Group: "bookcontroller.com", Version: "v1", Resource: "books"}

var booksKind = schema.GroupVersionKind{Group: "bookcontroller.com", Version: "v1", Kind: "Book"}

// Get takes name of the book, and returns the corresponding book object, and an error if there is any.
func (c *FakeBooks) Get(ctx context.Context, name string, options v1.GetOptions) (result *bookcontrollerv1.Book, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(booksResource, c.ns, name), &bookcontrollerv1.Book{})

	if obj == nil {
		return nil, err
	}
	return obj.(*bookcontrollerv1.Book), err
}

// List takes label and field selectors, and returns the list of Books that match those selectors.
func (c *FakeBooks) List(ctx context.Context, opts v1.ListOptions) (result *bookcontrollerv1.BookList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(booksResource, booksKind, c.ns, opts), &bookcontrollerv1.BookList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &bookcontrollerv1.BookList{ListMeta: obj.(*bookcontrollerv1.BookList).ListMeta}
	for _, item := range obj.(*bookcontrollerv1.BookList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested books.
func (c *FakeBooks) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(booksResource, c.ns, opts))

}

// Create takes the representation of a book and creates it.  Returns the server's representation of the book, and an error, if there is any.
func (c *FakeBooks) Create(ctx context.Context, book *bookcontrollerv1.Book, opts v1.CreateOptions) (result *bookcontrollerv1.Book, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(booksResource, c.ns, book), &bookcontrollerv1.Book{})

	if obj == nil {
		return nil, err
	}
	return obj.(*bookcontrollerv1.Book), err
}

// Update takes the representation of a book and updates it. Returns the server's representation of the book, and an error, if there is any.
func (c *FakeBooks) Update(ctx context.Context, book *bookcontrollerv1.Book, opts v1.UpdateOptions) (result *bookcontrollerv1.Book, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(booksResource, c.ns, book), &bookcontrollerv1.Book{})

	if obj == nil {
		return nil, err
	}
	return obj.(*bookcontrollerv1.Book), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeBooks) UpdateStatus(ctx context.Context, book *bookcontrollerv1.Book, opts v1.UpdateOptions) (*bookcontrollerv1.Book, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(booksResource, "status", c.ns, book), &bookcontrollerv1.Book{})

	if obj == nil {
		return nil, err
	}
	return obj.(*bookcontrollerv1.Book), err
}

// Delete takes name of the book and deletes it. Returns an error if one occurs.
func (c *FakeBooks) Delete(ctx context.Context, name string, opts v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(booksResource, c.ns, name), &bookcontrollerv1.Book{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeBooks) DeleteCollection(ctx context.Context, opts v1.DeleteOptions, listOpts v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(booksResource, c.ns, listOpts)

	_, err := c.Fake.Invokes(action, &bookcontrollerv1.BookList{})
	return err
}

// Patch applies the patch and returns the patched book.
func (c *FakeBooks) Patch(ctx context.Context, name string, pt types.PatchType, data []byte, opts v1.PatchOptions, subresources ...string) (result *bookcontrollerv1.Book, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(booksResource, c.ns, name, pt, data, subresources...), &bookcontrollerv1.Book{})

	if obj == nil {
		return nil, err
	}
	return obj.(*bookcontrollerv1.Book), err
}
