package main

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/flightctl/flightctl/internal/auth/common"
	"github.com/flightctl/flightctl/internal/consts"
	"github.com/flightctl/flightctl/internal/org/providers"
	"github.com/flightctl/flightctl/pkg/aap_client"
)

// For testing: org 6 is "private" and has no members

func main() {
	gatewayUrl := "https://192.168.122.98"

	// Dev tokens from a local auth provider
	// admin token
	// token := "S3E5Afs36v39tecnsmhFJBUAPjU44n"
	// testuser token
	token := "hu1MuJwNyg2WzklWpLhPjWmdrPc5b6"

	useClient(gatewayUrl, token)
	useProvider(gatewayUrl, token)
}

func useProvider(gatewayUrl string, token string) {
	ctx := context.WithValue(context.Background(), consts.TokenCtxKey, token)

	provider, err := providers.NewAAPOrganizationProvider(gatewayUrl, &tls.Config{
		InsecureSkipVerify: true,
	})
	if err != nil {
		fmt.Printf("err: %s", err)
		return
	}

	identity := common.NewBaseIdentity("test", "3", []string{})

	organizations, err := provider.GetUserOrganizations(ctx, identity)
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	fmt.Println(organizations)

	isMember, err := provider.IsMemberOf(ctx, identity, "5")
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	fmt.Printf("is member of 5: %t\n", isMember)

	isMember, err = provider.IsMemberOf(ctx, identity, "6")
	if err != nil {
		fmt.Printf("err: %s", err)
	}
	fmt.Printf("is member of 6: %t\n", isMember)
}

func useClient(gatewayUrl string, token string) {
	ctx := context.WithValue(context.Background(), consts.TokenCtxKey, token)

	aapClient, err := aap_client.NewAAPGatewayClient(aap_client.AAPGatewayClientOptions{
		GatewayUrl: gatewayUrl,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
	if err != nil {
		fmt.Printf("err: %s", err)
		return
	}

	organizations, err := aapClient.GetUserOrganizations(ctx, "3")
	if err != nil {
		fmt.Printf("err: %s", err)
		return
	}
	fmt.Println(organizations)

	teams, err := aapClient.GetUserTeams(ctx, "3")
	if err != nil {
		fmt.Printf("err: %s", err)
		return
	}
	fmt.Println(teams)

	fmt.Println("--------------------------------")

	organizations, err = aapClient.GetOrganizations(ctx)
	if err != nil {
		fmt.Printf("err: %s", err)
		return
	}
	fmt.Println(organizations)

	organization, err := aapClient.GetOrganization(ctx, "5")
	if err != nil {
		fmt.Printf("err: %s", err)
		return
	}
	fmt.Println(organization)
}
