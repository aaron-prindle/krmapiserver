package openstack

import (
	"os"

	"github.com/aaron-prindle/krmapiserver/included/github.com/gophercloud/gophercloud"
)

var nilOptions = gophercloud.AuthOptions{}

/*
AuthOptionsFromEnv fills out an identity.AuthOptions structure with the
settings found on the various OpenStack OS_* environment variables.

The following variables provide sources of truth: OS_AUTH_URL, OS_USERNAME,
OS_PASSWORD, OS_TENANT_ID, and OS_TENANT_NAME.

Of these, OS_USERNAME, OS_PASSWORD, and OS_AUTH_URL must have settings,
or an error will result.  OS_TENANT_ID, OS_TENANT_NAME, OS_PROJECT_ID, and
OS_PROJECT_NAME are optional.

OS_TENANT_ID and OS_TENANT_NAME are mutually exclusive to OS_PROJECT_ID and
OS_PROJECT_NAME. If OS_PROJECT_ID and OS_PROJECT_NAME are set, they will
still be referred as "tenant" in Gophercloud.

To use this function, first set the OS_* environment variables (for example,
by sourcing an `openrc` file), then:

	opts, err := openstack.AuthOptionsFromEnv()
	provider, err := openstack.AuthenticatedClient(opts)
*/
func AuthOptionsFromEnv() (gophercloud.AuthOptions, error) {
	authURL := os.Getenv("OS_AUTH_URL")
	username := os.Getenv("OS_USERNAME")
	userID := os.Getenv("OS_USERID")
	password := os.Getenv("OS_PASSWORD")
	tenantID := os.Getenv("OS_TENANT_ID")
	tenantName := os.Getenv("OS_TENANT_NAME")
	domainID := os.Getenv("OS_DOMAIN_ID")
	domainName := os.Getenv("OS_DOMAIN_NAME")
	applicationCredentialID := os.Getenv("OS_APPLICATION_CREDENTIAL_ID")
	applicationCredentialName := os.Getenv("OS_APPLICATION_CREDENTIAL_NAME")
	applicationCredentialSecret := os.Getenv("OS_APPLICATION_CREDENTIAL_SECRET")

	// If OS_PROJECT_ID is set, overwrite tenantID with the value.
	if v := os.Getenv("OS_PROJECT_ID"); v != "" {
		tenantID = v
	}

	// If OS_PROJECT_NAME is set, overwrite tenantName with the value.
	if v := os.Getenv("OS_PROJECT_NAME"); v != "" {
		tenantName = v
	}

	if authURL == "" {
		err := gophercloud.ErrMissingEnvironmentVariable{
			EnvironmentVariable: "OS_AUTH_URL",
		}
		return nilOptions, err
	}

	if userID == "" && username == "" {
		// Empty username and userID could be ignored, when applicationCredentialID and applicationCredentialSecret are set
		if applicationCredentialID == "" && applicationCredentialSecret == "" {
			err := gophercloud.ErrMissingAnyoneOfEnvironmentVariables{
				EnvironmentVariables: []string{"OS_USERID", "OS_USERNAME"},
			}
			return nilOptions, err
		}
	}

	if password == "" && applicationCredentialID == "" && applicationCredentialName == "" {
		err := gophercloud.ErrMissingEnvironmentVariable{
			EnvironmentVariable: "OS_PASSWORD",
		}
		return nilOptions, err
	}

	if (applicationCredentialID != "" || applicationCredentialName != "") && applicationCredentialSecret == "" {
		err := gophercloud.ErrMissingEnvironmentVariable{
			EnvironmentVariable: "OS_APPLICATION_CREDENTIAL_SECRET",
		}
		return nilOptions, err
	}

	if applicationCredentialID == "" && applicationCredentialName != "" && applicationCredentialSecret != "" {
		if userID == "" && username == "" {
			return nilOptions, gophercloud.ErrMissingAnyoneOfEnvironmentVariables{
				EnvironmentVariables: []string{"OS_USERID", "OS_USERNAME"},
			}
		}
		if username != "" && domainID == "" && domainName == "" {
			return nilOptions, gophercloud.ErrMissingAnyoneOfEnvironmentVariables{
				EnvironmentVariables: []string{"OS_DOMAIN_ID", "OS_DOMAIN_NAME"},
			}
		}
	}

	ao := gophercloud.AuthOptions{
		IdentityEndpoint:            authURL,
		UserID:                      userID,
		Username:                    username,
		Password:                    password,
		TenantID:                    tenantID,
		TenantName:                  tenantName,
		DomainID:                    domainID,
		DomainName:                  domainName,
		ApplicationCredentialID:     applicationCredentialID,
		ApplicationCredentialName:   applicationCredentialName,
		ApplicationCredentialSecret: applicationCredentialSecret,
	}

	return ao, nil
}
