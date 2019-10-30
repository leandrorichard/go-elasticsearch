// Licensed to Elasticsearch B.V under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.
//
// Code generated, DO NOT EDIT

package elasticsearch_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
)

var (
	_ = fmt.Printf
	_ = os.Stdout
	_ = elasticsearch.NewDefaultClient
)

// <https://github.com/elastic/elasticsearch/blob/master/docs/reference/getting-started.asciidoc#L214>
//
// --------------------------------------------------------------------------------
// PUT /customer/_doc/1
// {
//   "name": "John Doe"
// }
// --------------------------------------------------------------------------------

func Test_getting_started_311c4b632a29b9ead63b02d01f10096b(t *testing.T) {
	es, _ := elasticsearch.NewDefaultClient()

	// tag:311c4b632a29b9ead63b02d01f10096b[]
	res, err := es.Index(
		"customer",
		strings.NewReader(`{
		  "name": "John Doe"
		}`),
		es.Index.WithDocumentID("1"),
		es.Index.WithPretty(),
	)
	// end:311c4b632a29b9ead63b02d01f10096b[]
	if err != nil {
		fmt.Println("Error getting the response:", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	fmt.Println(res)
}