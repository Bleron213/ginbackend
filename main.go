package main

import (
	"fmt"
	"golang-angular/handlers"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"

	"github.com/auth0-community/go-auth0"
	"github.com/gin-gonic/gin"
	jose "gopkg.in/square/go-jose.v2"
)

/*

set env variables in the cmd (not powershell) using these two:
	set AUTH0_API_IDENTIFIER=https://my-golang-api
	set AUTH0_DOMAIN=dev-fly1l0bd.auth0.com

and then run "go run main.go"

special characters should be preceded with a ^. (special charaters: <,>,|,&,^)

*/

var (
	audience string
	domain   string
)

func main() {
	setAuth0Variables()

	r := gin.Default()
	fmt.Printf("%T", r)
	fmt.Println()
	r.NoRoute(func(c *gin.Context) {
		dir, file := path.Split(c.Request.RequestURI)
		ext := filepath.Ext(file)
		if file == "" || ext == "" {
			c.File("./ui/dist/ui/index.html")
		} else {
			c.File("./ui/dist/ui/" + path.Join(dir, file))
		}
	})

	authorized := r.Group("/")
	authorized.Use(authRequired())
	authorized.GET("/todo", handlers.GetTodoListHandler)
	authorized.POST("/todo", handlers.AddTodoHandler)
	authorized.DELETE("/todo/:id", handlers.DeleteTodoHandler)
	authorized.PUT("/todo", handlers.CompleteTodoHandler)

	fmt.Println("AUTH0_API_IDENTIFIER:", os.Getenv("AUTH0_API_IDENTIFIER"))
	fmt.Println("AUTH0_DOMAIN:", os.Getenv("AUTH0_DOMAIN"))
	err := r.Run(":3000")
	if err != nil {
		panic(err)
	}
}

func setAuth0Variables() {
	audience = os.Getenv("AUTH0_API_IDENTIFIER")
	domain = os.Getenv("AUTH0_DOMAIN")
}

func authRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		var auth0Domain = "https://" + domain + "/"

		client := auth0.NewJWKClient(auth0.JWKClientOptions{URI: auth0Domain + ".well-known/jwks.json"}, nil)
		configuration := auth0.NewConfiguration(client, []string{audience}, auth0Domain, jose.RS256)
		validator := auth0.NewValidator(configuration, nil)

		_, err := validator.ValidateRequest(c.Request)

		if err != nil {
			log.Println(err)
			terminateWithError(http.StatusUnauthorized, "token is not valid", c)
			return
		}
		c.Next()

	}
}

func terminateWithError(statusCode int, message string, c *gin.Context) {
	c.JSON(statusCode, gin.H{"error": message})
	c.Abort()
}
