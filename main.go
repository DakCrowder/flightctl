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

	token := "9rVtDCufPf4Q1GEMVUqq2coeCgptXw"
	ctx := context.WithValue(context.Background(), consts.TokenCtxKey, token)

	aapClient := aap_client.NewAAPGatewayClient(aap_client.AAPGatewayClientOptions{
		GatewayUrl: gatewayUrl,
		ClientTlsConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	})
	organizations, err := aapClient.GetUserOrganizations(ctx, "3")
	if err != nil {
		fmt.Printf("err: %s", err)
		return
	}
	fmt.Println(organizations)
}
