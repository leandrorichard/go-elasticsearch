// Licensed to Elasticsearch B.V under one or more agreements.
// Elasticsearch B.V. licenses this file to you under the Apache 2.0 License.
// See the LICENSE file in the project root for more information.
//
// Code generated, DO NOT EDIT

package elasticsearch_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
)

var (
	_ = fmt.Printf
	_ = os.Stdout
	_ = elasticsearch.NewDefaultClient
)

// <https://github.com/elastic/elasticsearch/blob/master/docs/reference/docs/get.asciidoc#L53>
//
// --------------------------------------------------------------------------------
// GET twitter/_doc/0?_source=false
// --------------------------------------------------------------------------------

func Test_docs_get_138ccd89f72aa7502dd9578403dcc589(t *testing.T) {
	es, _ := elasticsearch.NewDefaultClient()

	// tag:138ccd89f72aa7502dd9578403dcc589[]
	res, err := es.Get(
		"twitter",
		"0",
		es.Get.WithSource("false"),
		es.Get.WithPretty(),
	)
	// end:138ccd89f72aa7502dd9578403dcc589[]
	if err != nil {
		fmt.Println("Error getting the response:", err)
		os.Exit(1)
	}
	defer res.Body.Close()
	fmt.Println(res)
}