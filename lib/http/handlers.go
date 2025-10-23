package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"libraryes/library"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type HTTPHandlers struct {
	books *library.Library
}

func NewHTTPHandlers(library *library.Library) *HTTPHandlers{
	return &HTTPHandlers{
		books: library,
	}
}

func (h *HTTPHandlers) HandleCreateBook(w http.ResponseWriter, r *http.Request) {
	var BookDTO = BookDTO{}
	if err := json.NewDecoder(r.Body).Decode(&BookDTO); err != nil{
		errDTO := CreateErrDTO(err.Error(), time.Now())

		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
		return
	}

	if err := BookDTO.ValidateToCreate(); err != nil {
 		errDTO := CreateErrDTO(err.Error(), time.Now())

		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
	}

	books := library.NewBook(BookDTO.Title, BookDTO.Author, BookDTO.Pages)
	if err := h.books.AddBook(books); err != nil{
		errDTO := CreateErrDTO(err.Error(), time.Now())
		
		if errors.Is(err, library.ErrBookAlreadyExists){
			http.Error(w, errDTO.ToString(), http.StatusConflict)
		} else {
			http.Error(w, errDTO.ToString(), http.StatusConflict) 
		}
	return
	}

	b, err := json.MarshalIndent(books, "", "	")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusCreated)

	if _, err := w.Write(b); err != nil {
		fmt.Print(err)
	}
}


func (h *HTTPHandlers) HandleGetBook(w http.ResponseWriter, r *http.Request) {
	title := mux.Vars(r)["title"] // http://localhost:8081/books/title

	book, err := h.books.GetBook(title) 
	if err != nil {
		errDTO := CreateErrDTO(err.Error(), time.Now())
		if errors.Is(err, library.ErrBookNotFound){
			http.Error(w, errDTO.ToString(), http.StatusNotFound)
		} else{
			http.Error(w, errDTO.ToString(), http.StatusInternalServerError)
		}
		return
	}

	b, err := json.MarshalIndent(book, "", "	")
	if err != nil{
		panic(err)
	}

	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("failed to write http response", err)
		return
	}


}

func (h *HTTPHandlers) HandleReadBook(w http.ResponseWriter, r *http.Request) {
	var completeDTO = CompleteBookDTO{}
	if err := json.NewDecoder(r.Body).Decode(&completeDTO); err != nil {
		errDTO := CreateErrDTO(err.Error(), time.Now())
		b := []byte("error in request body")
		w.Write(b)
		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
	}

	title := mux.Vars(r)["title"]

	if completeDTO.Complete {
		if err := h.books.ReadBook(title); err != nil {
			errDTO := CreateErrDTO(err.Error(), time.Now())
			if errors.Is(err, library.ErrBookNotFound) {
				http.Error(w, errDTO.ToString(), http.StatusNotFound)
			} else {
				http.Error(w, errDTO.ToString(), http.StatusInternalServerError)
			}

			return

		}
	} else {
		if err := h.books.UnReadBook(title); err != nil {
			errDTO := CreateErrDTO(err.Error(), time.Now())
			if errors.Is(err, library.ErrBookNotFound) {
				http.Error(w, errDTO.ToString(), http.StatusNotFound)
			} else {
				http.Error(w, errDTO.ToString(), http.StatusInternalServerError)
			}

			return

		}
	}
}

func (h *HTTPHandlers) HandleGetAllBook(w http.ResponseWriter, r *http.Request) {
	books := h.books.ListBooks()

	b, err := json.MarshalIndent(books, "", "	")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("failed to write http response")
		return
	}
}

func (h *HTTPHandlers) HandleDeleteBook(w http.ResponseWriter, r *http.Request) {
	title := mux.Vars(r)["title"]
	if err := h.books.DeleteBook(title); err != nil {
		errDTO := CreateErrDTO(err.Error(), time.Now())

			if errors.Is(err, library.ErrBookNotFound){
				http.Error(w, errDTO.ToString(), http.StatusNotFound)
			} else {
				http.Error(w, errDTO.ToString(), http.StatusInternalServerError)
			}
	}
}

func (h *HTTPHandlers) HandleGetUnReadedBook(w http.ResponseWriter, r *http.Request) {
	books := h.books.ListUnReadedBooks()
	b, err := json.MarshalIndent(books, "", "    ")
		if err != nil {
		panic(err)
	}
	if _, err := w.Write(b); err != nil {
		http.Error(w, "failed to write http response", http.StatusBadGateway)
	}	
	
}

func (h *HTTPHandlers) HandleGetReadedBook(w http.ResponseWriter, r *http.Request) {
	books := h.books.ListReadedBooks()
	b, err := json.MarshalIndent(books, "", "    ")
		if err != nil {
		panic(err)
	}
	if _, err := w.Write(b); err != nil {
		http.Error(w, "failed to write http response", http.StatusBadGateway)
	}
	
}

func (h *HTTPHandlers) HandleListBookAuthor(w http.ResponseWriter, r *http.Request){
	author := r.URL.Query().Get("author")
	if author == "" {
		w.Write([]byte("Query param `author` is null"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	AuthorListBooks := h.books.ListBooksAuthor(author)
	b, err := json.MarshalIndent(AuthorListBooks, "", "    ")
	if err != nil {
		panic(err)
	}
	if _, err := w.Write(b); err != nil {
		http.Error(w, "failed to write http response", http.StatusBadGateway)
	}

}

func (h *HTTPHandlers) HandleCreateAuthor(w http.ResponseWriter, r *http.Request) {
	AuthorDTO := AuthorDTO{}
	if err := json.NewDecoder(r.Body).Decode(&AuthorDTO); err != nil{
		errDTO := CreateErrDTO(err.Error(), time.Now())

		http.Error(w, errDTO.ToString(), http.StatusBadRequest)
		return
	}

	author := library.NewAuthor(AuthorDTO.Author)
	if err := h.books.AddAuthor(author); err != nil {
		errDTO := CreateErrDTO(err.Error(), time.Now())
		
		if errors.Is(err, library.ErrAuthorAlreadyExists){
			http.Error(w, errDTO.ToString(), http.StatusConflict)
		} else {
			http.Error(w, errDTO.ToString(), http.StatusConflict) 
		}
	return
	}
	b, err := json.MarshalIndent(author, "", "	")
	if err != nil {
		panic(err)
	}
	

	if _, err := w.Write(b); err != nil {
		fmt.Print(err)
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *HTTPHandlers) HandleDeleteAuthor(w http.ResponseWriter, r *http.Request) {
	author := mux.Vars(r)["author"]
	if err := h.books.DeleteAuthor(author); err != nil {
		errDTO := CreateErrDTO(err.Error(), time.Now())
			if errors.Is(err, library.ErrAuthorNotFound){
				http.Error(w, errDTO.ToString(), http.StatusNotFound)
			} else {
				http.Error(w, errDTO.ToString(), http.StatusInternalServerError)
			}
	}
}

func (h *HTTPHandlers) HandleListAuthors(w http.ResponseWriter, r *http.Request) {
	author := h.books.ListAuthors()

	b, err := json.MarshalIndent(author, "", "	")
	if err != nil {
		panic(err)
	}
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(b); err != nil {
		fmt.Println("failed to write http response")
	}
}

func (h *HTTPHandlers) HandleLogin(w http.ResponseWriter, r *http.Request) {
	loginFunc(h.HandleLogin)
}

func (h *HTTPHandlers) HandleRegister(w http.ResponseWriter, r *http.Request) {
	regFunc(h.HandleRegister)
}
