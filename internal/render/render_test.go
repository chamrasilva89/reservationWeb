package render

import (
	"net/http"
	"testing"

	"github.com/chamrasilva89/reservationWeb/internal/models"
)

func TestAddDefaultData(t *testing.T) {
	var td models.TemplateData

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}
	session.Put(r.Context(), "flash", "1111")
	result := AddDefaultData(&td, r)

	if result.Flash != "1111" {
		t.Error("flash value of 1111 not found in session")
	}
}

func TestRenderTemplate(t *testing.T) {
	pathToTemplates = "./../../templates"
	tc, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}
	app.TemplateCache = tc

	r, err := getSession()
	if err != nil {
		t.Error(err)
	}

	var ww myWriter

	err = RenderTemplates(&ww, r, "home.page.tmpl", &models.TemplateData{})
	if err != nil {
		t.Error("Error Writing template to browser")
	}
	err = RenderTemplates(&ww, r, "non-exsitent.page.tmpl", &models.TemplateData{})
	if err == nil {
		t.Error("rendered template does not exsit")
	}

}
func getSession() (*http.Request, error) {
	r, err := http.NewRequest("GET", "/some-url", nil)
	if err != nil {
		return nil, err
	}

	ctx := r.Context()
	ctx, _ = session.Load(ctx, r.Header.Get("X-Session"))
	r = r.WithContext(ctx)

	return r, nil
}

func TestNewTemplates(t *testing.T) {
	NewTemplates(app)
}

func TestCreateTemplateCache(t *testing.T) {
	pathToTemplates = "./../../templates"
	_, err := CreateTemplateCache()
	if err != nil {
		t.Error(err)
	}

}