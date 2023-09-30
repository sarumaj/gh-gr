package main

import (
	"context"
	"fmt"

	client "github.com/sarumaj/gh-pr/pkg/restclient"
)

func main() {
	client, err := client.NewRESTClient(client.ClientOptions{})
	if err != nil {
		fmt.Println(err)
		return
	}

	ctx := context.Background()

	rate, err := client.GetRateLimit(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	if rate.IsExhausted() {
		fmt.Printf("Rate limit exceed, wait until: %s\n", rate.GetResetTime())
		return
	}

	user, err := client.GetUser(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Hello %s\n", user.Login)

	orgs, err := client.GetUserOrgs(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	orgs.Print(nil)

	repos, err := client.GetUserRepos(ctx)
	if err != nil {
		fmt.Println(err)
		return
	}

	repos.Print(nil)

	org, err := client.GetOrg(ctx, "github")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Got org %s\n", org.Login)

	repos, err = client.GetOrgRepos(ctx, "github")
	if err != nil {
		fmt.Println(err)
		return
	}

	repos.Print(nil)
}
