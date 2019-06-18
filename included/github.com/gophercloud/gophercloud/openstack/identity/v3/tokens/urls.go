package tokens

import "github.com/aaron-prindle/krmapiserver/included/github.com/gophercloud/gophercloud"

func tokenURL(c *gophercloud.ServiceClient) string {
	return c.ServiceURL("auth", "tokens")
}
