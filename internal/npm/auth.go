package npm

import "context"

type loginRequest struct {
	Identity string `json:"identity"`
	Secret   string `json:"secret"`
}

type loginResponse struct {
	Token string `json:"token"`
}

func (c *Client) Login(ctx context.Context) error {
	var response loginResponse

	err := c.doJSONNoAuth(ctx, "POST", "/tokens", loginRequest{
		Identity: c.email,
		Secret:   c.password,
	}, &response)

	if err != nil {
		return err
	}

	c.token = response.Token

	c.logger.Info(
		"npm login success",
		"event", "npm_login_success",
		"base_url", c.baseURL,
		"email", c.email,
	)

	return nil
}
