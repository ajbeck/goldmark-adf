//go:build goexperiment.jsonv2

package adf_test

import (
	"bytes"
	"fmt"

	"github.com/ajbeck/goldmark-adf"
)

// This example demonstrates basic Markdown to ADF conversion using the
// convenience function [adf.Convert].
func Example() {
	output, err := adf.Convert([]byte("Hello **world**"))
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(string(output))
	// Output:
	// {
	//   "version": 1,
	//   "type": "doc",
	//   "content": [
	//     {
	//       "type": "paragraph",
	//       "content": [
	//         {
	//           "type": "text",
	//           "text": "Hello "
	//         },
	//         {
	//           "type": "text",
	//           "marks": [
	//             {
	//               "type": "strong"
	//             }
	//           ],
	//           "text": "world"
	//         }
	//       ]
	//     }
	//   ]
	// }
}

// This example demonstrates creating a reusable goldmark instance with [adf.New].
// This approach is more efficient when converting multiple documents.
func Example_reusableInstance() {
	md := adf.New()
	var buf bytes.Buffer

	if err := md.Convert([]byte("# Title"), &buf); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(buf.String())
	// Output:
	// {
	//   "version": 1,
	//   "type": "doc",
	//   "content": [
	//     {
	//       "type": "heading",
	//       "attrs": {
	//         "level": 1
	//       },
	//       "content": [
	//         {
	//           "type": "text",
	//           "text": "Title"
	//         }
	//       ]
	//     }
	//   ]
	// }
}

// This example demonstrates GFM (GitHub Flavored Markdown) support with
// [adf.NewWithGFM], which enables tables, strikethrough, autolinks, and task lists.
func Example_withGFM() {
	md := adf.NewWithGFM()
	var buf bytes.Buffer

	if err := md.Convert([]byte("Hello ~~world~~"), &buf); err != nil {
		fmt.Println("Error:", err)
		return
	}
	fmt.Println(buf.String())
	// Output:
	// {
	//   "version": 1,
	//   "type": "doc",
	//   "content": [
	//     {
	//       "type": "paragraph",
	//       "content": [
	//         {
	//           "type": "text",
	//           "text": "Hello "
	//         },
	//         {
	//           "type": "text",
	//           "marks": [
	//             {
	//               "type": "strike"
	//             }
	//           ],
	//           "text": "world"
	//         }
	//       ]
	//     }
	//   ]
	// }
}
