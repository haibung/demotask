package enum

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/demotask/backend/utilities"
)

type OAuthProvider int

const (
	OAuthProviderLocal    OAuthProvider = 0
	OAuthProviderPayPal   OAuthProvider = 1
	OAuthProviderGoogle   OAuthProvider = 2
	OAuthProviderFacebook OAuthProvider = 3
)

func (t OAuthProvider) String() string {
	switch t {
	case OAuthProviderLocal:
		return "LOCAL"
	case OAuthProviderPayPal:
		return "PAYPAL"
	case OAuthProviderGoogle:
		return "GOOGLE"
	case OAuthProviderFacebook:
		return "FACEBOOK"
	default:
		return "unknown"
	}
}

func (t OAuthProvider) IsValid() error {
	switch t {
	case
		OAuthProviderLocal,
		OAuthProviderPayPal,
		OAuthProviderGoogle,
		OAuthProviderFacebook:
		return nil
	}
	return fmt.Errorf(utilities.DataNotFound, "OAuth Provider")
}

func OAuthProviderFromString(value string) (OAuthProvider, error) {
	switch strings.ToLower(value) {
	case "local":
		return OAuthProviderLocal, nil
	case "paypal":
		return OAuthProviderPayPal, nil
	case "google":
		return OAuthProviderGoogle, nil
	case "facebook":
		return OAuthProviderFacebook, nil
	default:
		return 0, fmt.Errorf(utilities.DataNotFound, "OAuth Provider")
	}
}

func OAuthProviderToArray(value string) ([]int, error) {
	var result []int
	providers := strings.Split(value, ",")
	for i := 0; i < len(providers); i++ {
		providerID, err := strconv.Atoi(providers[i])
		if err != nil {
			return nil, errors.New(utilities.NumberNotValid)
		}
		if err := OAuthProvider(providerID).IsValid(); err != nil {
			return nil, err
		}
		result = append(result, providerID)
	}
	return result, nil
}
