// This app will start a New Relic bulk export job.  I've hardcoded my specific query here.  You'll probably want to change that.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/machinebox/graphql"
)

//Don't forget that the interactive GraphQL endpoint adds the data part of the struct, while the
// programmatic endpoint doesn't.
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
	var periodStart time.Time
	var periodEnd time.Time

	nrAPI := flag.String("apikey", "", "New Relic admin user API Key")
	logVerbose := flag.Bool("verbose", false, "Writes verbose logs for debugging")
	accountId := flag.Int("account", 0, "New Relic account ID")
	startDate := flag.String("start", "", "Starting date you want to query YYYY-MM-DD format")
	endDate := flag.String("end", "", "Ending date you want to query YYYY-MM-DD format")
	flag.Parse()

	dateLayout := "2006-01-02"
	periodStart, _ = time.Parse(dateLayout, *startDate)
	periodEnd, _ = time.Parse(dateLayout, *endDate)

	if *logVerbose {
		fmt.Println("Bulk export util v1.0")
		fmt.Println("Verbose logging enabled")
		fmt.Println(periodStart, periodEnd)
	}

	graphqlClient := graphql.NewClient("https://api.newrelic.com/graphql")

	// Note that the query is of type String! here.  That may not be respecting the standard rules for our API calls.
	// In most of the cases I've seen, it should be of type Nrql!.
	graphqlRequest := graphql.NewRequest(`
	mutation ($accountId: Int!, $query: String!) {
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

	exportQuery := fmt.Sprintf("FROM Query SELECT user, account, timestamp WHERE targetEventType = 'nrAQMIncident' since '%v' until '%v'", periodStart.Format(dateLayout), periodEnd.Format(dateLayout))
	graphqlRequest.Var("query", exportQuery)
	graphqlRequest.Var("accountId", *accountId)
	graphqlRequest.Header.Set("API-Key", *nrAPI)

	if *logVerbose {
		fmt.Println(exportQuery)
		fmt.Println(graphqlRequest)
	}

	if err := graphqlClient.Run(context.Background(), graphqlRequest, &graphqlResponse); err != nil {
		panic(err)
		fmt.Println(graphqlResponse)
	}

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
