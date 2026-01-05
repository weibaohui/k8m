package sso

import (
	"github.com/go-chi/chi/v5"
	"github.com/weibaohui/k8m/pkg/controller/admin/config"
	"github.com/weibaohui/k8m/pkg/response"
)

type AuthController struct{}

func RegisterAuthRoutes(auth chi.Router) {
	ctrl := &AuthController{}
	ldap := &config.LdapConfigController{}
	auth.Get("/sso/config", response.Adapter(ctrl.GetSSOConfig))
	auth.Get("/oidc/{name}/sso", response.Adapter(ctrl.GetAuthCodeURL))
	auth.Get("/oidc/{name}/callback", response.Adapter(ctrl.HandleCallback))
	auth.Get("/ldap/config", response.Adapter(ldap.GetLdapConfig))
}
