package supabase

import (
	"fmt"
	"net/url"
)

type SignUpParams struct {
	Email    string         `json:"email"`
	Password string         `json:"password"`
	Data     map[string]any `json:"data,omitempty"`
}

type SignUpResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Aud       string `json:"aud"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type Session struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type User struct {
	ID               string         `json:"id"`
	Aud              string         `json:"aud"`
	Role             string         `json:"role"`
	Email            string         `json:"email"`
	EmailConfirmedAt string         `json:"email_confirmed_at,omitempty"`
	Phone            string         `json:"phone,omitempty"`
	ConfirmedAt      string         `json:"confirmed_at,omitempty"`
	LastSignInAt     string         `json:"last_sign_in_at,omitempty"`
	CreatedAt        string         `json:"created_at"`
	UpdatedAt        string         `json:"updated_at"`
	UserMetadata     map[string]any `json:"user_metadata"`
	AppMetadata      map[string]any `json:"app_metadata"`
}

type AdminUserResponse struct {
	Users []User `json:"users"`
}

type GoTrueError struct {
	Message string `json:"msg"`
}

func (c *Client) SignUp(params SignUpParams) (*SignUpResponse, error) {
	data, err := c.doAuth("POST", "/auth/v1/signup", params)
	if err != nil {
		return nil, err
	}
	return jsonUnmarshal[SignUpResponse](data)
}

func (c *Client) SignInWithEmail(email, password string) (*Session, error) {
	params := map[string]string{
		"email":    email,
		"password": password,
	}
	path := "/auth/v1/token?grant_type=password"
	data, err := c.doAuth("POST", path, params)
	if err != nil {
		return nil, err
	}
	session, err := jsonUnmarshal[Session](data)
	if err != nil {
		return nil, err
	}
	c.SetAuthToken(session.AccessToken)
	return session, nil
}

func (c *Client) SignOut() error {
	_, err := c.doAuth("POST", "/auth/v1/logout", nil)
	if err != nil {
		return fmt.Errorf("sign out: %w", err)
	}
	c.SetAuthToken("")
	return nil
}

func (c *Client) GetUser() (*User, error) {
	data, err := c.doAuth("GET", "/auth/v1/user", nil)
	if err != nil {
		return nil, err
	}
	return jsonUnmarshal[User](data)
}

func (c *Client) RefreshToken(refreshToken string) (*Session, error) {
	params := map[string]string{
		"refresh_token": refreshToken,
	}
	path := "/auth/v1/token?grant_type=refresh_token"
	data, err := c.doAuth("POST", path, params)
	if err != nil {
		return nil, err
	}
	session, err := jsonUnmarshal[Session](data)
	if err != nil {
		return nil, err
	}
	c.SetAuthToken(session.AccessToken)
	return session, nil
}

func (c *Client) SignInWithOAuth(provider string) (string, error) {
	u, err := url.Parse(c.config.URL + "/auth/v1/authorize")
	if err != nil {
		return "", err
	}
	q := u.Query()
	q.Set("provider", provider)
	q.Set("redirect_to", "http://localhost:9999/auth/callback")
	u.RawQuery = q.Encode()
	return u.String(), nil
}

func (c *Client) AdminListUsers() ([]User, error) {
	data, err := c.doAdmin("GET", "/auth/v1/admin/users", nil)
	if err != nil {
		return nil, err
	}
	result, err := jsonUnmarshal[AdminUserResponse](data)
	if err != nil {
		return nil, err
	}
	return result.Users, nil
}

func (c *Client) AdminCreateUser(email, password string, data map[string]any) (*User, error) {
	params := map[string]any{
		"email":    email,
		"password": password,
		"data":     data,
	}
	body, err := c.doAdmin("POST", "/auth/v1/admin/users", params)
	if err != nil {
		return nil, err
	}
	return jsonUnmarshal[User](body)
}

func (c *Client) AdminDeleteUser(userID string) error {
	_, err := c.doAdmin("DELETE", "/auth/v1/admin/users/"+userID, nil)
	if err != nil {
		return err
	}
	return nil
}
