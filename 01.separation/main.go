package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
)

// Action Layer

// ErrUserNotFound ...
var ErrUserNotFound = errors.New("User not found")

// User ...
type User struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

// UserStorer ...
type UserStorer interface {
	Get(ctx context.Context, email string) (*User, error)
	Save(ctx context.Context, user *User) error
}

// MemoryUserStorage ...
type MemoryUserStorage struct {
	store map[string]*User
}

// NewMemoUserStorage ...
func NewMemoUserStorage() *MemoryUserStorage {
	return &MemoryUserStorage{
		store: map[string]*User{},
	}
}

func (ms *MemoryUserStorage) Get(ctx context.Context, email string) (*User, error) {
	if u, ok := ms.store[email]; ok {
		return u, nil
	}
	return nil, ErrUserNotFound
}

func (ms *MemoryUserStorage) Save(ctx context.Context, user *User) error {
	ms.store[user.Email] = user
	return nil
}

// Business Logic

// RegisterParams ...
type RegisterParams struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (rp *RegisterParams) Validate() error {
	if rp.Email == "" {
		return errors.New(("Email connot be empty"))
	}

	if !strings.ContainsRune(rp.Email, '@') {
		return errors.New("Email must include an '@' symbol")
	}

	if rp.Name == "" {
		return errors.New("Name cannot be empty")
	}

	return nil
}

// UserService ...
type UserService interface {
	// Register may return an ErrEmailExist error
	Register(context.Context, *RegisterParams) error
	// GetByEmail may retturn an ErrUserNotFound error
	GetByEmail(context.Context, string) (*User, error)
}

// ErrEmailExist ...
var ErrEmailExist = errors.New("Email is already in user")

// UserServiceImpl ...
type UserServiceImpl struct {
	userStorage UserStorer
}

// NewUserServiceImpl ...
func NewUserServiceImpl(us UserStorer) *UserServiceImpl {
	return &UserServiceImpl{
		userStorage: us,
	}
}

// Register ...
func (us *UserServiceImpl) Register(ctx context.Context, params *RegisterParams) error {
	_, err := us.userStorage.Get(ctx, params.Email)

	if err == nil {
		return ErrEmailExist
	} else if err != ErrUserNotFound {
		return err
	}

	return us.userStorage.Save(ctx, &User{
		Email: params.Email,
		Name:  params.Name,
	})
}

// GetByEmail ...
func (us *UserServiceImpl) GetByEmail(ctx context.Context, email string) (*User, error) {
	return us.userStorage.Get(ctx, email)
}

// Access Layer

// JsonOverHTTP ...
type JsonOverHTTP struct {
	router  *http.ServeMux
	usrServ UserService
}

// NewJSONOverHTTP ..
func NewJSONOverHTTP(usrServ UserService) *JsonOverHTTP {
	r := http.NewServeMux()

	joh := &JsonOverHTTP{
		router:  r,
		usrServ: usrServ,
	}

	r.HandleFunc("/register", joh.Register)
	r.HandleFunc("/user", joh.GetUser)

	return joh
}

func (j *JsonOverHTTP) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	j.router.ServeHTTP(w, r)
}

// Register ...
func (j *JsonOverHTTP) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Register requires a post request", http.StatusMethodNotAllowed)
		return
	}

	params := &RegisterParams{}
	err := json.NewDecoder(r.Body).Decode(params)

	if err != nil {
		http.Error(w, "Unable to read your request", http.StatusBadRequest)
		return
	}

	err = params.Validate()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = j.usrServ.Register(r.Context(), params)

	if err == ErrEmailExist {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (j *JsonOverHTTP) validateEmail(email string) error {
	if email == "" {
		return errors.New("Email must not be empty")
	}

	if !strings.ContainsRune(email, '@') {
		return errors.New("Email must include an '@' sympol")
	}

	return nil
}

// GetUser ...
func (j *JsonOverHTTP) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GetUser requires a get request", http.StatusMethodNotAllowed)
		return
	}

	email := r.FormValue("email")
	err := j.validateEmail(email)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	u, err := j.usrServ.GetByEmail(r.Context(), email)

	if err == ErrUserNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// Wire together

func main() {
	println("Separate server register & get user!")

	usrStor := NewMemoUserStorage()
	usrServ := NewUserServiceImpl(usrStor)
	joh := NewJSONOverHTTP(usrServ)

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	err := http.ListenAndServe(":"+port, joh)
	if err != nil {
		panic(err)
	}

}

/*
TEST
	Register
	~ curl -XPOST -d '{"email":"thanhdungfb@gmail.com", "Name":"Alex Lee"}' localhost:8080/register

	Get Detail User
	~ curl localhost:8080/user\?email=thanhdungfb@gmail.com

Test with Insomidia
	1.
	POST : localhost:8080/register
	BODY JSON:
	{
		"email":"thanhdungfb@gmail.com",
		"Name":"Alex Lee"
	}

	2.
	GET: localhost:8080/user\?email=thanhdungfb@gmail.com

	(Input the param in the Query Params)
*/
