# bulk-export

These are two utilities that help with New Relic's bulk export feature.

## launch-export-job
This command line app can be used to launch a bulk export job.  

**Please note that the export query is hard coded in the app.  You'll probably want to change it to reflect your specific query**

The app expects the following arguments:

- account - The ID of the New Relic account you want to run the query against.  
- apikey - A user key that corresponds to the account you're querying.  
- start - The beginning of the date range for your query in YYYY-MM-DD form.  
- end - The end of the date range for your query in YYYY-MM-DD form.  

You can optionally set -verbose=true for more info when you run the app.

As an example:

./launch-export-job -apikey=(MyNRUserKey) -account=(MyAccountId) -start=2022-05-01 -end=2022-05-07

On success, the utility returns the ID of your export job.  It also creates a file called (exportJobID).launch with details on the export job.


## get-export-results
This command line app can be used to retrive the results of your export job when it's done.

The app expects the following arguments:

- account - The ID of the New Relic account you want to run the query against.  
- apikey - A user key that corresponds to the account you're querying.  
- exportId - The ID of the export job you launched.

You can optionally set -verbose=true for more info when you run the app.

As an example:

get-export-results -apikey=(MyNRUserKey) -account=(MyAccountId) -exportId=(MyExportID)

When run, the app first checks to see if the job is 100% complete.  If it is not, the app posts a message and terminates.

If the job is complete, it downloads the results files to the local directory and then creates a file called (exportJobID).info with details on the export job.
