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

package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	booksv1 "www-github.cisco.com/bota/maglev-bootcamp/track-controlplane/week-3/bookstore-operator/api/v1"
)

// Product has all the information about book product.
type Product struct {
	ID          string `db:"product_id" json:"id"`
	Name        string `db:"name" json:"name"`
	Author      string `db:"author" json:"author"`
	ISBN        string `db:"isbn" json:"isbn"`
	Genre       string `db:"genre" json:"genre"`
	DateCreated string `db:"datecreated" json:"date_created"`
	DateUpdated string `db:"dateupdated" json:"date_updated"`
}

// NewProduct get new product from user.
type NewProduct struct {
	Name   string `db:"name" json:"name"`
	Author string `db:"author" json:"author"`
	ISBN   string `db:"isbn" json:"isbn"`
	Genre  string `db:"genre" json:"genre"`
}

// UpdateProduct defines what information may be provided to modify an
// existing Product.
type UpdateProduct struct {
	Name   *string `json:"name"`
	Author *string `json:"author"`
	ISBN   *string `json:"isbn"`
	Genre  *string `json:"genre"`
}

// BookReconciler reconciles a Book object
type BookReconciler struct {
	client.Client
	//Log    logr.Logger
	Log    *log.Logger
	Scheme *runtime.Scheme
	SVC    string
}

// BooksFinalizerLabel defines the finalizer.
var BooksFinalizerLabel = "books.finalizer.bookstore.com"

// Helper functions to check and remove string from a slice of strings.
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// +kubebuilder:rbac:groups=books.bookstore.com,resources=books,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=books.bookstore.com,resources=books/status,verbs=get;update;patch

// Reconcile will get the request from the queue and reconcile the spec and status.
func (r *BookReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	ctx := context.Background()
	log := r.Log
	//log := r.Log.WithValues("book", req.NamespacedName)
	//log := r.Log.WithName("bookstore")

	// your logic here

	book := &booksv1.Book{}
	err := r.Get(ctx, req.NamespacedName, book)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Println("book resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Println(err, "Failed to get book")
		return ctrl.Result{}, err
	}

	//log.Printf("book %#v", book)

	if book.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object. This is equivalent
		// registering our finalizer.
		if !containsString(book.ObjectMeta.Finalizers, BooksFinalizerLabel) {
			book.ObjectMeta.Finalizers = append(book.ObjectMeta.Finalizers, BooksFinalizerLabel)
			if err := r.Update(ctx, book); err != nil {
				return ctrl.Result{}, err
			}
		}
	} else {
		// The object is being deleted
		if containsString(book.ObjectMeta.Finalizers, BooksFinalizerLabel) {
			// our finalizer is present, so lets handle any external dependency
			// in this case to delete the book.
			log.Printf("deleting book %v", book)
			err = r.deleteBook(book.Status.ID)
			if err != nil {
				log.Printf("delete book %s: %s failed %s", req.Name, book.Status.ID, err)
			}
		}

		// remove our finalizer from the list and update it.
		book.ObjectMeta.Finalizers = removeString(book.ObjectMeta.Finalizers, BooksFinalizerLabel)
		if err := r.Update(ctx, book); err != nil {
			return ctrl.Result{}, err
		}
		return ctrl.Result{}, nil
	}

	// New CRD, call create book.
	if book.Status.ID == "" {
		log.Println("creating book: ", book.Spec.Name)
		createStatus, err := r.createBook(book)
		if err != nil {
			log.Printf("create %#v failed: %v", book.Spec, err)
			return ctrl.Result{}, err
		}
		book.Status = *createStatus
		err = r.Status().Update(ctx, book)
		if err != nil {
			log.Println("update failed: ", err)
			return ctrl.Result{}, err
		}
	} else { //Existing CRD, call get book or update book.
		// If any information changed, we need to update it.
		if isUpdated(book) {
			log.Printf("Need to update book %#v", book)
			updatestatus, err := r.updateBook(book)
			if err != nil {
				log.Printf("update %#v failed: %v", book.Spec, err)
				return ctrl.Result{}, err
			}
			book.Status = *updatestatus
			err = r.Status().Update(ctx, book)
			if err != nil {
				log.Println("update failed: ", err)
				return ctrl.Result{}, err
			}
		} else {
			gBook, err := r.getBook(book)
			//log.Printf("get book %#v", gBook)
			if err != nil || gBook.ID == "" {
				log.Println("Need to handler get failed, maybe recreate the book.")
				createStatus, err := r.createBook(book)
				if err != nil {
					log.Printf("create %#v failed: %v", book.Spec, err)
					return ctrl.Result{}, err
				}
				book.Status = *createStatus
				err = r.Status().Update(ctx, book)
				if err != nil {
					log.Println("update failed: ", err)
					return ctrl.Result{}, err
				}
			}
		}
	}
	//return ctrl.Result{}, nil
	// requeue the object.
	return ctrl.Result{Requeue: true}, nil
}

func convertBodyToProduct(resp *http.Response, target interface{}) error {
	return json.NewDecoder(resp.Body).Decode(target)
}

func isUpdated(b *booksv1.Book) bool {
	if b.Spec.Name != b.Status.Name ||
		b.Spec.Author != b.Status.Author ||
		b.Spec.ISBN != b.Status.ISBN ||
		b.Spec.Genre != b.Status.Genre {
		return true
	}
	return false
}

// getJson will get the response as json format.
func getJSON(url string, target interface{}) error {
	r, err := http.Get(url)
	if err != nil {
		return err
	}
	defer r.Body.Close()

	//return json.NewDecoder(r.Body).Decode(target)
	return convertBodyToProduct(r, target)
}

func (r *BookReconciler) getBook(book *booksv1.Book) (*booksv1.BookStatus, error) {
	getURL := fmt.Sprintf("%s/%s/%s", r.SVC, "books", book.Status.ID)
	prod := Product{}

	err := getJSON(getURL, &prod)
	if err != nil {
		log.Printf("get %v failed: %v", getURL, err)
		return nil, err
	}

	//log.Printf("in getBook book %#v", prod)

	return &booksv1.BookStatus{
		ID:          prod.ID,
		Name:        prod.Name,
		Author:      prod.Author,
		ISBN:        prod.ISBN,
		Genre:       prod.Genre,
		DateCreated: prod.DateCreated,
		DateUpdated: prod.DateUpdated,
	}, nil
}

func (r *BookReconciler) deleteBook(bookID string) error {
	getURL := fmt.Sprintf("%s/%s/%s", r.SVC, "books", bookID)
	req, err := http.NewRequest(http.MethodDelete, getURL, nil)
	if err != nil {
		log.Printf("get delete request %v failed: %v", getURL, err)
		return err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("send request failed: ", err)
		return err
	}
	defer resp.Body.Close()

	log.Println("delete reponse: ", resp.Status)

	return nil
}

func (r *BookReconciler) createBook(book *booksv1.Book) (*booksv1.BookStatus, error) {
	getURL := fmt.Sprintf("%s/%s", r.SVC, "books")

	newProd := NewProduct{
		Name:   book.Spec.Name,
		Author: book.Spec.Author,
		ISBN:   book.Spec.ISBN,
		Genre:  book.Spec.Genre,
	}

	newProdBytes, err := json.Marshal(&newProd)
	if err != nil {
		log.Printf("marchal newProd %#v failed: %v", newProd, err)
		return nil, err
	}
	req, err := http.NewRequest("POST", getURL, bytes.NewBuffer(newProdBytes))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	//body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("response Body:", string(body))

	prod := Product{}
	err = convertBodyToProduct(resp, &prod)
	if err != nil {
		return nil, err
	}
	//log.Printf("create prod %#v", prod)

	return &booksv1.BookStatus{
		ID:          prod.ID,
		Name:        prod.Name,
		Author:      prod.Author,
		ISBN:        prod.ISBN,
		Genre:       prod.Genre,
		DateCreated: prod.DateCreated,
		DateUpdated: prod.DateUpdated,
	}, nil
}

func (r *BookReconciler) updateBook(book *booksv1.Book) (*booksv1.BookStatus, error) {
	getURL := fmt.Sprintf("%s/%s/%s", r.SVC, "books", book.Status.ID)
	prod := NewProduct{
		Name:   book.Spec.Name,
		Author: book.Spec.Author,
		ISBN:   book.Spec.ISBN,
		Genre:  book.Spec.Genre,
	}
	prodJSON, err := json.Marshal(prod)

	if err != nil {
		log.Printf("json prod %v failed: %v", prod, err)
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPut, getURL, bytes.NewBuffer(prodJSON))
	if err != nil {
		log.Printf("get delete request %v failed: %v", getURL, err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Println("send request failed: ", err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Println("response Status:", resp.Status)
	//body, _ := ioutil.ReadAll(resp.Body)
	//fmt.Println("response Body:", string(body))

	return &booksv1.BookStatus{
		ID:          book.Status.ID,
		Name:        book.Spec.Name,
		Author:      book.Spec.Author,
		ISBN:        book.Spec.ISBN,
		Genre:       book.Spec.Genre,
		DateCreated: book.Status.DateCreated,
		DateUpdated: time.Now().String(),
	}, nil
}

// SetupWithManager will setup new controller manager.
func (r *BookReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&booksv1.Book{}).
		Complete(r)
}
