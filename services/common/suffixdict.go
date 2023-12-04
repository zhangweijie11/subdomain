package common

var SubdomainSuffixDict *[]string

func init() {
	var subdomainSuffixDict = []string{"111", "222"}
	SubdomainSuffixDict = &subdomainSuffixDict
}
