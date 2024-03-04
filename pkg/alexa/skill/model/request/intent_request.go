package request

type Intent struct {
	Name  string `json:"name"`
	Slots map[string]struct {
		Name               string `json:"name"`
		Value              string `json:"value"`
		ConfirmationStatus string `json:"confirmationStatus"`
		Resolutions        struct {
			ResolutionsPerAuthority []struct {
				Authority string `json:"authority"`
				Status    struct {
					Code string `json:"code"`
				} `json:"status"`
				Values []struct {
					Value struct {
						Name string `json:"name"`
						ID   string `json:"id"`
					} `json:"value"`
				} `json:"values"`
			} `json:"resolutionsPerAuthority"`
		} `json:"resolutions"`
	} `json:"slots"`
	ConfirmationStatus string `json:"confirmationStatus"`
}

type IntentRequest struct {
	BaseRequest
	DialogState string `json:"dialogState"`
	Intent      Intent `json:"intent"`
}
