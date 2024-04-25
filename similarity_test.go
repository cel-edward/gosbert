package gosbert

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSbert(t *testing.T) {
	assert := assert.New(t)

	sbert, err := NewSbert()
	assert.NoError(err)
	assert.NotNil(sbert.module)
}

// func Test() {
// 	sbert, err := NewSbert()
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	defer sbert.Finalize()

// 	similarity, err := sbert.Similarity(
// 		"this is a test phrase",
// 		[]string{
// 			"this is a similar testing phrase",
// 			"i am totally unrelated",
// 			"i also have no relevance",
// 		},
// 	)
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	log.Print(similarity)
// }
