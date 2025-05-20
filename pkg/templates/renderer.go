package templates

import (
	"log"

	"github.com/gin-contrib/multitemplate"
)

func SetupTemplates() multitemplate.Renderer {
	r := multitemplate.NewRenderer()

	// Daftar template dengan layout yang benar
	templates := map[string][]string{
		"index": {
			"templates/layouts/guest.html",
			"templates/partials/header.html",
			"templates/partials/footer.html",
			"templates/pages/index.html",
		},
		"auth/login": {
			"templates/layouts/guest.html",
			"templates/partials/header.html",
			"templates/partials/footer.html",
			"templates/pages/auth/login.html",
		},
		"auth/register": {
			"templates/layouts/guest.html",
			"templates/partials/header.html",
			"templates/partials/footer.html",
			"templates/pages/auth/register.html",
		},
		"dashboard": {
			"templates/layouts/app.html",
			"templates/partials/top-navbar.html",
			"templates/partials/bottom-navbar.html",
			"templates/pages/dashboard.html",
		},
		"demo/dashboard": {
			"templates/layouts/guest.html",
			"templates/partials/header.html",
			"templates/partials/footer.html",
			"templates/pages/demo/dashboard.html",
		},
		"demo/cashier": {
			"templates/layouts/guest.html",
			"templates/partials/header.html",
			"templates/partials/footer.html",
			"templates/pages/demo/cashier.html",
		},
		"demo/order": {
			"templates/layouts/guest.html",
			"templates/partials/header.html",
			"templates/partials/footer.html",
			"templates/pages/demo/order.html",
		},
	}

	for name, files := range templates {
		r.AddFromFiles(name, files...)
		log.Printf("Registered template %q with files: %v\n", name, files)
	}

	return r
}
