package main

import (
	"fmt"
	"net/http"

	"github.com/justinas/nosurf"
)
/*WriteToConsole:

This middleware function simply logs a message to the console when an HTTP request is received.
It takes an http.Handler called next as an argument, which represents the next middleware or the final request handler.
After logging the message, it calls the ServeHTTP method of the next handler to continue processing the request.
*/
func WriteToConsole(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Hit the page")
		next.ServeHTTP(w, r)
	})
}

/*NoSurf:

This middleware function is responsible for preventing Cross-Site Request Forgery (CSRF) attacks using the "nosurf" package.
It wraps the next handler with CSRF protection by creating a new nosurf.CSRFHandler with the next handler as its target.
It configures the base CSRF cookie with settings like HttpOnly, Path, and whether it should be set as Secure (likely depending on whether the application is in production).
The CSRF handler adds protection against CSRF attacks by including and validating CSRF tokens in requests.
*/
func NoSurf(next http.Handler) http.Handler {
	csrfHandler := nosurf.New(next)

	csrfHandler.SetBaseCookie(http.Cookie{
		HttpOnly: true,
		Path:     "/",
		Secure:   app.InProduction,
		SameSite: http.SameSiteLaxMode,
	})
	return csrfHandler
}
/*SessionLoad:

This middleware function is responsible for managing sessions.
It takes the next handler and wraps it to load and save sessions.
It's assumed that the session variable is a global object responsible for session management.
This middleware ensures that session data is loaded and saved for each request.
*/
func SessionLoad(next http.Handler) http.Handler {
	return session.LoadAndSave(next)
}

/*Auth:

This is an authentication middleware that checks if a user is authenticated before allowing access to certain routes.
It takes the next handler and wraps it to perform the authentication check.
Inside the middleware function, it uses the helpers.IsAuthenticated function to determine if the user is authenticated. If not, it sets an error message in the session and redirects the user to the login page.
If the user is authenticated, it allows the request to continue to the next handler.
*/
func Auth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !helpers.IsAuthenticated(r) {
			session.Put(r.Context(), "error", "Log in first!")
			http.Redirect(w, r, "/user/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
