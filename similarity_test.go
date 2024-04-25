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

func TestGetSimilarity(t *testing.T) {
	assert := assert.New(t)

	sbert, err := NewSbert()
	assert.NoError(err)

	t.Run("identical", func(t *testing.T) {
		got, err := sbert.GetSimilarity(
			"I am a factual statement",
			[]string{
				"I am a factual statement",
			},
		)
		assert.NoError(err)
		assert.Len(got, 1)
		assert.InDelta(1, got[0], 0.001)
	})

	t.Run("similarity", func(t *testing.T) {
		got, err := sbert.GetSimilarity(
			"I am a factual statement",
			[]string{
				"I am also a factual phrase",
				"There is no relevance here",
				"Nothing contained within has a similarity to above",
			},
		)
		assert.NoError(err)
		assert.Len(got, 3)
		assert.Equal(true, got[0] > got[1])
		assert.Equal(true, got[0] > got[2])
	})

}
