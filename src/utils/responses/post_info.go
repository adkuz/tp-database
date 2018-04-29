package responses

import "github.com/Alex-Kuz/tp-database/src/models"

type PostInfo struct {
	Post   *models.Post   `json:"post"`

	Thread *models.Thread `json:"thread,omitempty"`

	Forum  *models.Forum  `json:"forum,omitempty"`

	Author *models.User   `json:"author,omitempty"`
}
