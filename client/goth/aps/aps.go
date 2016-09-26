package aps

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/markbates/goth"
	"golang.org/x/oauth2"
)

const (
	authURL         string = "http://localhost:9096/authorize"
	tokenURL        string = "http://localhost:9096/token"
	endpointProfile string = "http://localhost:9096/userinfo"
)

// Provider is the implementation of `goth.Provider` for accessing APS.
type Provider struct {
	ClientKey   string
	Secret      string
	CallbackURL string
	config      *oauth2.Config
	prompt      oauth2.AuthCodeOption
}

// New - Please fill the code
func New(clientKey, secret, callbackURL string, scopes ...string) *Provider {
	///return nil
	p := &Provider{
		ClientKey:   clientKey,
		Secret:      secret,
		CallbackURL: callbackURL,
	}
	p.config = newConfig(p, scopes)
	return p
}

// FetchUser - Please fill the code
func (p *Provider) FetchUser(session goth.Session) (goth.User, error) {
	//return goth.User{}, nil
	sess := session.(*Session)
	user := goth.User{
		AccessToken:  sess.AccessToken,
		Provider:     p.Name(),
		RefreshToken: sess.RefreshToken,
		ExpiresAt:    sess.ExpiresAt,
	}

	response, err := http.Get(endpointProfile + "?access_token=" + url.QueryEscape(sess.AccessToken))
	if err != nil {
		if response != nil {
			response.Body.Close()
		}
		return user, err
	}
	defer response.Body.Close()

	bits, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return user, err
	}

	err = json.NewDecoder(bytes.NewReader(bits)).Decode(&user.RawData)
	if err != nil {
		return user, err
	}

	err = userFromReader(bytes.NewReader(bits), &user)
	return user, err
}

func userFromReader(reader io.Reader, user *goth.User) error {
	u := struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		FirstName string `json:"given_name"`
		LastName  string `json:"family_name"`
		Link      string `json:"link"`
		Picture   string `json:"picture"`
	}{}

	err := json.NewDecoder(reader).Decode(&u)
	if err != nil {
		return err
	}

	user.Name = u.Name
	user.FirstName = u.FirstName
	user.LastName = u.LastName
	user.NickName = u.Name
	user.Email = u.Email
	//user.Description = u.Bio
	user.AvatarURL = u.Picture
	user.UserID = u.ID
	//user.Location = u.Location.Name

	return err
}

// RefreshToken - Please fill the code
func (p *Provider) RefreshToken(refreshToken string) (*oauth2.Token, error) {
	///return nil, nil
	token := &oauth2.Token{RefreshToken: refreshToken}
	ts := p.config.TokenSource(oauth2.NoContext, token)
	newToken, err := ts.Token()
	if err != nil {
		return nil, err
	}
	return newToken, err	
}

// RefreshTokenAvailable - Please fill the code
func (p *Provider) RefreshTokenAvailable() bool {
	return true
}

// Name is the name used to retrieve this provider later.
func (p *Provider) Name() string {
	return "aps"
}

// Debug is a no-op for the APS package.
func (p *Provider) Debug(debug bool) {}

// BeginAuth - Please fill the code
func (p *Provider) BeginAuth(state string) (goth.Session, error) {
	///return nil, nil
	var opts []oauth2.AuthCodeOption
	if p.prompt != nil {
		opts = append(opts, p.prompt)
	}
	url := p.config.AuthCodeURL(state, opts...)
	session := &Session{
		AuthURL: url,
	}
	return session, nil	
}

// SetPrompt - Please fill the code
func (p *Provider) SetPrompt(prompt ...string) {
	if len(prompt) == 0 {
		return
	}
	p.prompt = oauth2.SetAuthURLParam("prompt", strings.Join(prompt, " "))
}

func newConfig(provider *Provider, scopes []string) *oauth2.Config {
	c := &oauth2.Config{
		ClientID:     provider.ClientKey,
		ClientSecret: provider.Secret,
		RedirectURL:  provider.CallbackURL,
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURL,
			TokenURL: tokenURL,
		},
		Scopes: []string{},
	}

	if len(scopes) > 0 {
		for _, scope := range scopes {
			c.Scopes = append(c.Scopes, scope)
		}
	} else {
		c.Scopes = []string{"profile", "email", "openid"}
	}
	return c
}

