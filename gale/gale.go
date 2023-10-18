package main

// Gale is a Dagger module for running Github Actions workflows.
type Gale struct {
	Config *AuthConfig
}

// SetToken sets the Github token used for authentication.
func (g *Gale) SetToken(token string) *Gale {
	return &Gale{
		Config: &AuthConfig{
			Token: dag.SetSecret("GITHUB_TOKEN", token),
		},
	}
}
