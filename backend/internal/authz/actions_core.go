package authz

// crud monta as ações create/read/update/delete de um recurso de configuração.
func crud(resource, label string, cat Category, verbs ...string) []Action {
	names := map[string]string{"create": "Criar", "read": "Ler", "update": "Atualizar", "delete": "Excluir"}
	var out []Action
	for _, v := range verbs {
		out = append(out, Action{
			ID:       "admin::" + resource + "." + v,
			Domain:   DomainAdmin,
			Category: cat,
			Name:     names[v] + " " + label,
			Resource: resource,
		})
	}
	return out
}

// RegisterCoreActions registra as ações administrativas fixas do sistema
// (configurações, plugins e aplicação). As ações de conteúdo ficam com cada recurso.
func RegisterCoreActions(r *Registry) {
	all := []string{"create", "read", "update", "delete"}
	r.MustRegister(crud("settings.users", "administradores", CategorySettings, all...)...)
	r.MustRegister(crud("settings.roles", "roles de administrador", CategorySettings, all...)...)
	r.MustRegister(crud("settings.api-tokens", "tokens de API", CategorySettings, all...)...)
	r.MustRegister(Action{ID: "admin::settings.api-tokens.regenerate", Domain: DomainAdmin, Category: CategorySettings, Name: "Regenerar tokens de API", Resource: "settings.api-tokens"})
	r.MustRegister(crud("settings.email-templates", "templates de e-mail", CategorySettings, "read", "update")...)
	r.MustRegister(crud("settings.audit-logs", "registros de auditoria", CategorySettings, "read")...)
	r.MustRegister(crud("plugin.enduser-roles", "roles de usuário final", CategoryPlugins, all...)...)
	r.MustRegister(crud("plugin.enduser-users", "usuários finais", CategoryPlugins, all...)...)
	r.MustRegister(crud("plugin.enduser-settings", "configurações avançadas de usuários finais", CategoryPlugins, "read", "update")...)
}
