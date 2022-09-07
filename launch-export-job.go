// This app will start a New Relic bulk export job.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/machinebox/graphql"
)

//Define the API call response structure.
//Don't forget that the interactive GraphQL endpoint adds the data part of the struct, while the
//programmatic endpoint doesn't.
type nrResponseStruct struct {
	//Data struct {
	HistoricalDataExportCreateExport struct {
		ID              string      `json:"id"`
		Message         interface{} `json:"message"`
		Nrql            string      `json:"nrql"`
		PercentComplete float32     `json:"percentComplete"`
		Status          string      `json:"status"`
	} `json:"historicalDataExportCreateExport"`
	//} `json:"data"`
}

func main() {

	//Define command line flags.
	nrAPI := flag.String("apikey", "", "New Relic admin user API Key")
	logVerbose := flag.Bool("verbose", false, "Writes verbose logs for debugging")
	accountId := flag.Int("account", 0, "New Relic account ID")
	exportQuery := flag.String("nrql", "", "NRQL query to execute")
	flag.Parse()

	if *logVerbose {
		fmt.Println("Bulk export util v2.0")
		fmt.Println("Verbose logging enabled")
	}

	//Define API endpoint
	graphqlClient := graphql.NewClient("https://api.newrelic.com/graphql")

	//Define request structure.
	graphqlRequest := graphql.NewRequest(`
	mutation ($accountId: Int!, $query: Nrql!) {
		historicalDataExportCreateExport(accountId: $accountId, nrql: $query) {
		id
		message
		nrql
		percentComplete
		status
		}
	}
  `)

	var graphqlResponse nrResponseStruct

	//Fill in the request variables.
	graphqlRequest.Var("query", *exportQuery)
	graphqlRequest.Var("accountId", *accountId)
	graphqlRequest.Header.Set("API-Key", *nrAPI)

	if *logVerbose {
		fmt.Println(*exportQuery)
		fmt.Println(graphqlRequest)
	}

	//Execute the request.  If the request is malformed (i.e. bad NRQL), we panic and post the error message.
	if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlResponse); err != nil {
		panic(err)
		fmt.Println(graphqlResponse)
	}

	//We successfully launched the job, so now we write the job ID to the screen.
	fmt.Println("Export job ID: " + graphqlResponse.HistoricalDataExportCreateExport.ID)

	if *logVerbose {
		fmt.Println(graphqlResponse)
		//fmt.Println(result)
	}

	// Almost there, let's create an info file so we remember what this export job was about.
	fmt.Println("Creating info file")
	infoFile, err := os.Create(graphqlResponse.HistoricalDataExportCreateExport.ID + ".launch")
	if err != nil {
		fmt.Println(err)
	}
	defer infoFile.Close()

	infoFile.WriteString("Export Job: " + graphqlResponse.HistoricalDataExportCreateExport.ID + "\n")
	infoFile.WriteString("Account: " + strconv.Itoa(*accountId) + "\n")
	infoFile.WriteString("Query status: " + graphqlResponse.HistoricalDataExportCreateExport.Status + "\n")
	infoFile.WriteString("Query: " + graphqlResponse.HistoricalDataExportCreateExport.Nrql + "\n")
}
