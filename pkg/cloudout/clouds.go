package cloudout

import (
	"crypto/tls"
	"fmt"
	"io"
	"os"

	"github.com/gophercloud/gophercloud/v2"
	"github.com/gophercloud/gophercloud/v2/openstack/config/clouds"
	"gopkg.in/yaml.v2"
)

func Write(authOptions gophercloud.AuthOptions, endpointOptions gophercloud.EndpointOpts, tlsConfig *tls.Config, identityEndpoint, tlsCertPath string, w io.Writer) error {
	var authType clouds.AuthType
	switch {
	case authOptions.Username != "":
		authType = clouds.AuthV3Password
	case authOptions.ApplicationCredentialID != "":
		authType = clouds.AuthV3ApplicationCredential
	default:
		return fmt.Errorf("unknown authentication type")
	}

	c := clouds.Clouds{
		Clouds: map[string]clouds.Cloud{
			os.Getenv("OS_CLOUD"): {
				AuthInfo: &clouds.AuthInfo{
					AuthURL:                     identityEndpoint,
					Username:                    authOptions.Username,
					UserID:                      authOptions.UserID,
					Password:                    authOptions.Password,
					ApplicationCredentialID:     authOptions.ApplicationCredentialID,
					ApplicationCredentialName:   authOptions.ApplicationCredentialName,
					ApplicationCredentialSecret: authOptions.ApplicationCredentialSecret,
					UserDomainName:              authOptions.DomainName,
					UserDomainID:                authOptions.DomainID,
				},
				AuthType:     authType,
				RegionName:   endpointOptions.Region,
				EndpointType: endpointOptions.Type,
				CACertFile:   tlsCertPath,
			},
		},
	}

	if err := yaml.NewEncoder(w).Encode(c); err != nil {
		return fmt.Errorf("failed to encode clouds.yaml to YAML: %w", err)
	}

	return nil
}
