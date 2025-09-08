package i18n

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"
)

type Translations map[string]string

type Localizer struct {
	locales         map[string]Translations
	defaultLanguage string
}

type Config struct {
	DefaultLanguage string
	Languages       []string
	TranslationDir  string
}

func NewLocalizer(config *Config) (*Localizer, error) {
	loc := &Localizer{locales: map[string]Translations{}, defaultLanguage: config.DefaultLanguage}
	files, err := os.ReadDir(config.TranslationDir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		lang := f.Name()[0:2]
		data, err := os.ReadFile(filepath.Join(config.TranslationDir, f.Name()))
		if err != nil {
			return nil, err
		}
		var t Translations
		if err := yaml.Unmarshal(data, &t); err != nil {
			return nil, err
		}
		loc.locales[lang] = t
	}

	return loc, nil
}

func (l *Localizer) Locales() map[string]Translations {
	return l.locales
}

func (l *Localizer) IsLanguageSupported(lang string) bool {
	_, ok := l.locales[lang]
	return ok
}

func (l *Localizer) DefaultLanguage() string {
	return l.defaultLanguage
}

func (l *Localizer) ParseAcceptLanguage(acceptLanguage string) string {
	lang := strings.Split(acceptLanguage, ",")[0]
	return strings.Split(lang, ";")[0]
}

// LocalizeError returns localized message using template data
func (l *Localizer) LocalizeError(lang, key string, data map[string]interface{}) string {
	trans, ok := l.locales[lang][key]
	if !ok {
		trans = key
	}
	tmpl, err := template.New("").Parse(trans)
	if err != nil {
		return trans
	}
	var out string
	err = tmpl.Execute(&stringWriter{&out}, data)
	if err != nil {
		return trans
	}
	return out
}

func (l *Localizer) SetLanguageInContext(ctx context.Context, lang string) context.Context {
	return context.WithValue(ctx, "lang", lang)
}

// Get language from context
func (l *Localizer) GetLanguageFromContext(ctx context.Context) string {
	if ctx == nil {
		return "en"
	}
	if lang, ok := ctx.Value("lang").(string); ok && lang != "" {
		return lang
	}
	return "en"
}

type stringWriter struct{ s *string }

func (w *stringWriter) Write(p []byte) (int, error) {
	*w.s += string(p)
	return len(p), nil
}

// IsDev returns true if APP_ENV != production
func IsDev() bool {
	return os.Getenv("APP_ENV") != "production"
}
