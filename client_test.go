package restful_test

import (
	"fmt"
	"log"

	"github.com/jncornett/restful"
)

// Example usage of the default JSON client
func ExampleClient() {
	client := restful.NewJSONClient(
		"http://www.example.com/api",
		// Default constructor for an anonymous struct
		func() interface{} { return &struct{ First, Last string }{} },
		// Default constructor for an anonymous slice list
		func() interface{} { return &[]struct{ First, Last string }{} },
	)
	// lookup all records
	list, err := client.GetAll()
	if err != nil {
		log.Fatal(err)
	}
	for _, item := range list.([]struct{ First, Last string }) {
		fmt.Println(item)
	}
}
