/*


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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BookSpec defines the desired state of Book
type BookSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Book. Edit Book_types.go to remove/update
	Foo    string `json:"foo,omitempty"`
	Name   string `db:"name" json:"name" validate:"required"`
	Author string `db:"author" json:"author"`
	ISBN   string `db:"isbn" json:"isbn"`
	Genre  string `db:"genre" json:"genre"`
}

// BookStatus defines the observed state of Book
type BookStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ID          string `db:"product_id" json:"id"`
	Name        string `db:"name" json:"name"`
	Author      string `db:"author" json:"author"`
	ISBN        string `db:"isbn" json:"isbn"`
	Genre       string `db:"genre" json:"genre"`
	DateCreated string `db:"datecreated" json:"date_created"`
	DateUpdated string `db:"dateupdated" json:"date_updated"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Book is the Schema for the books API
type Book struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BookSpec   `json:"spec,omitempty"`
	Status BookStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// BookList contains a list of Book
type BookList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Book `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Book{}, &BookList{})
}
