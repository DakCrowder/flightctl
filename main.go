package main

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/flightctl/flightctl/internal/consts"
	"github.com/flightctl/flightctl/pkg/aap_client"
)

func main() {
	gatewayUrl := "https://192.168.122.98"

	// Dev tokens from a local auth provider
	// token := "9rVtDCufPf4Q1GEMVUqq2coeCgptXw"
	token := "S3E5Afs36v39tecnsmhFJBUAPjU44n"
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
