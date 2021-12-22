package cmd

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/treeverse/lakefs/pkg/api"
)

const annotateTemplate = `{{  $val := .Commit}}
{{ $.Object|ljust 15}} {{ $val.Committer|ljust 20 }} {{ $val.Id | printf "%.16s"|ljust 20 }} {{ $val.CreationDate|date }}  {{ $.CommitMessage |ljust 30 }}
`

var annotateCmd = &cobra.Command{
	Use:   "annotate <path uri>",
	Short: "List entries under a given path, annotating each with the latest modifying commit ",
	// Long:    "Show who created the latest commit to an object or objects in a give path",
	Aliases: []string{"blame"},
	Args:    cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		pathURI := MustParsePathURI("path", args[0])
		client := getClient()
		amount := 1
		pfx := api.PaginationPrefix(*pathURI.Path)
		res, err := client.ListObjectsWithResponse(cmd.Context(), pathURI.Repository, pathURI.Ref, &api.ListObjectsParams{Prefix: &pfx})
		DieOnResponseError(res, err)
		var from string
		for {
			pfx := api.PaginationPrefix(*pathURI.Path)
			params := &api.ListObjectsParams{
				Prefix: &pfx,
				After:  api.PaginationAfterPtr(from),
			}
			resp, err := client.ListObjectsWithResponse(cmd.Context(), pathURI.Repository, pathURI.Ref, params)
			DieOnResponseError(resp, err)
			for _, obj := range resp.JSON200.Results {
				prfx := []string{obj.Path}
				logCommitsParams := &api.LogCommitsParams{
					Amount:  api.PaginationAmountPtr(amount),
					Objects: &prfx,
				}
				res, err := client.LogCommitsWithResponse(cmd.Context(), pathURI.Repository, pathURI.Ref, logCommitsParams)
				DieOnResponseError(res, err)
				data := struct {
					Commit        api.Commit
					Object        string
					CommitMessage string
				}{
					Commit:        res.JSON200.Results[0],
					Object:        obj.Path,
					CommitMessage: setMessageSize(100, (res.JSON200.Results[0].Message)),
				}
				Write(annotateTemplate, data)
			}
			pagination := resp.JSON200.Pagination
			if !pagination.HasMore {
				break
			}
			from = pagination.NextOffset
		}
	},
}

func setMessageSize(size int, str string) string {
	if len(str) > size {
		str = str[:size] + "..."
	}
	str = strings.Split(str, "\\n")[0]
	return str
}

//nolint:gochecknoinits
func init() {
	rootCmd.AddCommand(annotateCmd)

}