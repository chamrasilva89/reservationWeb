package config

import (
	"html/template"
	"log"

	"github.com/alexedwards/scs/v2"
)

const FileUploadPath = "E:\\GO\\FF\\"

// Appconfig hold the app config
type AppConfig struct {
	UseCache      bool
	TemplateCache map[string]*template.Template
	InfoLog       *log.Logger
	ErrorLog      *log.Logger
	InProduction  bool
	Session       *scs.SessionManager
}
