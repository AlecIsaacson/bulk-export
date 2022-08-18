// If you feed this app a New Relic export job ID, it'll go to AWS and download the results for you.
package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/machinebox/graphql"
)

//Don't forget that the interactive GraphQL endpoint adds the data part of the struct, while the
//programmatic endpoint doesn't.
type nrResponseStruct struct {
	//Data struct {
	Actor struct {
		Account struct {
			HistoricalDataExport struct {
				Export struct {
					Account struct {
						ID int `json:"id"`
					} `json:"account"`
					AvailableUntil  int64       `json:"availableUntil"`
					BeginTime       int64       `json:"beginTime"`
					EndTime         int64       `json:"endTime"`
					ID              string      `json:"id"`
					InternalStatus  string      `json:"internalStatus"`
					Message         interface{} `json:"message"`
					Nrql            string      `json:"nrql"`
					PercentComplete float32     `json:"percentComplete"`
					Results         []string    `json:"results"`
					Status          string      `json:"status"`
					SubmittedAt     int64       `json:"submittedAt"`
					UpdatedAt       int64       `json:"updatedAt"`
				} `json:"export"`
			} `json:"historicalDataExport"`
		} `json:"account"`
	} `json:"actor"`
	//} `json:"data"`
}

func main() {

	nrAPI := flag.String("apikey", "", "New Relic admin user API Key")
	logVerbose := flag.Bool("verbose", false, "Writes verbose logs for debugging")
	accountId := flag.Int("account", 0, "New Relic account ID")
	exportId := flag.String("exportId", "", "GUID of export job")
	flag.Parse()

	if *logVerbose {
		fmt.Println("Export retriever util v1.0")
		fmt.Println("Verbose logging enabled")
		fmt.Println("Getting Export", *exportId)
	}

	graphqlClient := graphql.NewClient("https://api.newrelic.com/graphql")

	// Note that the query is of type String! here.  That may not be respecting the standard rules for our API calls.
	// In most of the cases I've seen, it should be of type Nrql!.
	// Also note that the API won't let me query for internalStatus or updatedAt programmatically.  I can do this via the GraphQL UI, though.
	graphqlRequest := graphql.NewRequest(`
	query($accountId: Int!, $exportId: ID!){
		actor {
		  account(id: $accountId) {
			historicalDataExport {
			  export(id: $exportId) {
				account {
				  id
				}
				availableUntil
				beginTime
				endTime
				id
				message
				nrql
				percentComplete
				results
				status
				submittedAt
			  }
			}
		  }
		}
	  }
  `)

	var graphqlResponse nrResponseStruct

	graphqlRequest.Var("exportId", *exportId)
	graphqlRequest.Var("accountId", *accountId)
	graphqlRequest.Header.Set("API-Key", *nrAPI)

	if *logVerbose {
		fmt.Println(graphqlRequest)
	}

	// Get the job info from the GraphQL API.
	if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlResponse); err != nil {
		panic(err)
		fmt.Println(graphqlResponse)
	}

	if *logVerbose {
		fmt.Println(graphqlResponse)
	}

	// Don't try to download incomplete export jobs.
	pctComplete := graphqlResponse.Actor.Account.HistoricalDataExport.Export.PercentComplete
	if pctComplete != 100.0 {
		fmt.Println("Bulk export job is incomplete.  Please try again later.")
		fmt.Printf("Job is currently %v%% complete\n", pctComplete)
		os.Exit(1)
	}

	// The export job is complete, parse out the results for the URLs of each file to be downloaded and then download them.
	for i, result := range graphqlResponse.Actor.Account.HistoricalDataExport.Export.Results {
		filename := strings.Split(path.Base(result), "?")

		fmt.Println("Getting file: ", i, filename[0])

		resp, err := http.Get(result)
		if err != nil {
			fmt.Println(err)
		}
		defer resp.Body.Close()

		outFile, err := os.Create(filename[0])
		if err != nil {
			fmt.Println(err)
		}

		_, err = io.Copy(outFile, resp.Body)
		outFile.Close()
	}

	// Almost there, let's create an info file so we remember what this export job was about.
	fmt.Println("Creating info file")
	infoFile, err := os.Create(*exportId + ".info")
	if err != nil {
		fmt.Println(err)
	}
	defer infoFile.Close()

	infoFile.WriteString("Export Job: " + *exportId + "\n")
	infoFile.WriteString("Account: " + strconv.Itoa(*accountId) + "\n")
	infoFile.WriteString("Query status: " + graphqlResponse.Actor.Account.HistoricalDataExport.Export.Status + "\n")
	infoFile.WriteString("Query: " + graphqlResponse.Actor.Account.HistoricalDataExport.Export.Nrql + "\n")
	infoFile.WriteString("Query since: " + strconv.FormatInt(graphqlResponse.Actor.Account.HistoricalDataExport.Export.BeginTime, 10) + "\n")
	infoFile.WriteString("Query until: " + strconv.FormatInt(graphqlResponse.Actor.Account.HistoricalDataExport.Export.EndTime, 10) + "\n")
	infoFile.WriteString("Submitted at: " + strconv.FormatInt(graphqlResponse.Actor.Account.HistoricalDataExport.Export.SubmittedAt, 10) + "\n")
	infoFile.WriteString("Updated at: " + strconv.FormatInt(graphqlResponse.Actor.Account.HistoricalDataExport.Export.UpdatedAt, 10) + "\n")
}
