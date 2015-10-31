package registrar

type AccountsService struct {
	client *Client
}

type Account struct {
	Id    int    `json:"id"`
	Email string `json:"email"`
}

type AccountParams struct {
}

func (s *AccountsService) Current(params *AccountParams) (*Account, error) {
	var a Account
	err := s.client.Call("GET", "userinfo", params, &a)
	if err != nil {
		return nil, err
	}

	return &a, nil
}
