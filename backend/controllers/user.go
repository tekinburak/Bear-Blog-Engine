package controllers

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/alanqchen/MGBlog/backend/app"
	"github.com/alanqchen/MGBlog/backend/models"
	"github.com/alanqchen/MGBlog/backend/repositories"
	"github.com/alanqchen/MGBlog/backend/services"
	"github.com/alanqchen/MGBlog/backend/util"
	"github.com/gorilla/mux"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"
)

type UserController struct {
	*app.App
	repositories.UserRepository
	repositories.PostRepository
}

/*
type PasswordResetController struct { // Same as Auth
	App *app.App
	repositories.UserRepository
	jwtService services.JWTAuthService
}
*/

func NewUserController(a *app.App, ur repositories.UserRepository, pr repositories.PostRepository) *UserController {
	return &UserController{a, ur, pr}
}

func (uc *UserController) HelloWorld(w http.ResponseWriter, r *http.Request) {
	log.Println(r.Context().Value("userId"))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "Hello gopher!")
}

func (uc *UserController) Profile(w http.ResponseWriter, r *http.Request) {
	uid, err := services.UserIdFromContext(r.Context())
	if err != nil {
		NewAPIError(&APIError{false, "Something went wrong", http.StatusInternalServerError}, w)
		return
	}

	NewAPIResponse(&APIResponse{Data: uid}, w, http.StatusOK)
}

func (uc *UserController) Create(w http.ResponseWriter, r *http.Request) {
	// Validate the length of the body since some users could send a big payload
	/*required := []string{"name", "email", "password"}
	if len(params) != len(required) {
		err := NewAPIError(false, "Invalid request")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(err)
		return
	}*/

	j, err := GetJSON(r.Body)
	if err != nil {
		NewAPIError(&APIError{false, "Invalid request", http.StatusBadRequest}, w)
		return
	}

	name, err := j.GetString("name")
	if err != nil {
		NewAPIError(&APIError{false, "Name is required", http.StatusBadRequest}, w)
		return
	}
	// TODO: Implement something like this and embed in a basecontroller https://stackoverflow.com/a/23960293/2554631
	if len(name) < 2 || len(name) > 32 {
		NewAPIError(&APIError{false, "Name must be between 2 and 32 characters", http.StatusBadRequest}, w)
		return
	}

	email, err := j.GetString("email")
	if err != nil {
		NewAPIError(&APIError{false, "Email is required", http.StatusBadRequest}, w)
		return
	}
	if ok := util.IsEmail(email); !ok {
		NewAPIError(&APIError{false, "You must provide a valid email address", http.StatusBadRequest}, w)
		return
	}
	exists := uc.UserRepository.Exists(email)
	if exists {
		NewAPIError(&APIError{false, "The email address is already in use", http.StatusBadRequest}, w)
		return
	}
	pw, err := j.GetString("password")
	if err != nil {
		NewAPIError(&APIError{false, "Password is required", http.StatusBadRequest}, w)
		return
	}
	if len(pw) < 6 {
		NewAPIError(&APIError{false, "Password must not be less than 6 characters", http.StatusBadRequest}, w)
		return
	}

	var newID int
	for collision := true; collision; collision = !(err == pgx.ErrNoRows) {
		newID = int(rand.Int31())
		_, err = uc.Database.Conn.Prepare(context.Background(), "id-exists-query", "SELECT user_schema.\"user\".id FROM user_schema.\"user\" WHERE id = $1;")
		err = uc.Database.Conn.QueryRow(context.Background(), "id-exists-query", newID).Scan(exists)
	}
	u := &models.User{
		ID:        newID,
		Name:      name,
		Email:     email,
		Admin:     false,
		CreatedAt: time.Now(),
	}
	u.SetPassword(pw)

	err = uc.UserRepository.Create(u)
	if err != nil {
		NewAPIError(&APIError{false, "Could not create user", http.StatusBadRequest}, w)
		return
	}

	defer r.Body.Close()
	NewAPIResponse(&APIResponse{Success: true, Message: "User created"}, w, http.StatusOK)
}

func (uc *UserController) GetAll(w http.ResponseWriter, r *http.Request) {
	users, err := uc.UserRepository.GetAll()
	if err != nil {
		// something went wrong
		NewAPIError(&APIError{false, "Could not fetch users", http.StatusBadRequest}, w)
		return
	}

	NewAPIResponse(&APIResponse{Success: true, Data: users}, w, http.StatusOK)
}

func (uc *UserController) GetById(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		NewAPIError(&APIError{false, "Invalid request", http.StatusBadRequest}, w)
		return
	}
	user, err := uc.UserRepository.FindById(id)
	if err != nil {
		// user was not found
		NewAPIError(&APIError{false, "Could not find user", http.StatusNotFound}, w)
		return
	}

	NewAPIResponse(&APIResponse{Success: true, Data: user}, w, http.StatusOK)
}

func (uc *UserController) Update(w http.ResponseWriter, r *http.Request) {
	uid, err := services.UserIdFromContext(r.Context())
	if err != nil {
		NewAPIError(&APIError{false, "Something went wrong", http.StatusInternalServerError}, w)
		return
	}

	user, err := uc.UserRepository.FindById(uid)
	if err != nil {
		NewAPIError(&APIError{false, "Could not find user", http.StatusBadRequest}, w)
		return
	}

	j, err := GetJSON(r.Body)
	if err != nil {
		NewAPIError(&APIError{false, "Invalid request", http.StatusBadRequest}, w)
		return
	}

	name, err := j.GetString("name")
	if name != "" && err == nil {
		user.Name = name
	}

	newpw, err := j.GetString("newpassword")
	if newpw != "" && err == nil {
		// confirm password
		oldpw, err := j.GetString("oldpassword")
		if err != nil {
			NewAPIError(&APIError{false, "Old password is required", http.StatusBadRequest}, w)
			return
		}
		ok := user.CheckPassword(oldpw)
		if !ok {
			NewAPIError(&APIError{false, "Old password do not match", http.StatusBadRequest}, w)
			return
		}
		if len(newpw) < 6 {
			NewAPIError(&APIError{false, "Password must not be less than 6 characters", http.StatusBadRequest}, w)
			return
		}
		user.SetPassword(newpw)
	}

	user.UpdatedAt = pgtype.Timestamptz{Time: time.Now()}

	err = uc.UserRepository.Update(user)
	if err != nil {
		NewAPIError(&APIError{false, "Could not update user", http.StatusBadRequest}, w)
		return
	}

	authUser := &models.AuthUser{
		User:  user,
		Admin: user.Admin,
	}

	NewAPIResponse(&APIResponse{Success: true, Data: authUser}, w, http.StatusOK)
}

// Password reset functionality - Might look into this more later
/*
func (prc *PasswordResetController) ResetPasswordRequest(w http.ResponseWriter, r *http.Request) {
	j, err := GetJSON(r.Body)
	if err != nil {
		NewAPIError(&APIError{false, "Invalid request", http.StatusBadRequest}, w)
		return
	}
	if err != nil {
		NewAPIError(&APIError{false, "Invalid request", http.StatusBadRequest}, w)
		return
	}
	email, err := j.GetString("email")
	if err != nil {
		NewAPIError(&APIError{false, "Email is required", http.StatusBadRequest}, w)
		return
	}
	if ok := util.IsEmail(email); !ok {
		NewAPIError(&APIError{false, "You must provide a valid email address", http.StatusBadRequest}, w)
		return
	}

	user, err := prc.UserRepository.FindByEmail(email)

	data := struct {
		Tokens *services.Tokens `json:"tokens"`
		User   *models.AuthUser `json:"user"`
	}{
		nil,
		nil,
	}

	if user != nil {
		tokens, err := prc.jwtService.GenerateResetToken(user)
		if err != nil {
			NewAPIError(&APIError{false, "Something went wrong", http.StatusBadRequest}, w)
			return
		}

		err := smtp.SendMail(
			email := gmail.Compose("Email subject", tokens.AccessToken)
			email.From = "empty@gmail.com"
			email.Password = "empty"

			// Defaults to "text/plain; charset=utf-8" if unset.
			email.ContentType = "text/html; charset=utf-8"

			// Normally you'll only need one of these, but I thought I'd show both.
			email.AddRecipient(email)

			err := email.Send()
			if err != nil {
				log.err("Failed to send email")
				return
			}
		)
		if err != nil {
			log.Fatal(err)
		}

	}
	NewAPIResponse(&APIResponse{Success: true, Message: "Login successful", Data: data}, w, http.StatusOK)
	// TODO return jwt tokens
	//NewAPIResponse(&APIResponse{Success: true, Data: j}, w, http.StatusOK)

}
*/