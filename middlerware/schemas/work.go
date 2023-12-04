package schemas

type DomainParams struct {
	Domains []string `json:"domains" binding:"required"`
}
