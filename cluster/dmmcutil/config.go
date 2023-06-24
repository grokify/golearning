package dmmcutil

import (
	"math/rand"

	"github.com/go-nlp/dmmclust"
	"github.com/grokify/mogo/crypto/randutil"
)

func ConfigDefault(numClusters uint) (dmmclust.Config, error) {
	var err error
	if numClusters <= 0 {
		numClusters = 10 // default number of clusters
	}
	//if randSeed == 0 {
	//	randSeed, err = randutil.NewSeedInt64Crypto()
	//}
	// Vocabulary is set with Docs are known. Corpus size = number of words
	return dmmclust.Config{
		K:          int(numClusters),                                            // maximum 10 clusters expected
		Vocabulary: 0,                                                           // simple example: the vocab is the same as the corpus size
		Iter:       100,                                                         // iterate 100 times
		Alpha:      0.0001,                                                      // smaller probability of joining an empty group
		Beta:       0.1,                                                         // higher probability of joining groups like me
		Score:      dmmclust.Algorithm4,                                         // use Algorithm3 to score
		Sampler:    dmmclust.NewGibbs(rand.New(randutil.NewCryptoRandSource())), // #nosec G404 - `NewCryptoRandSource()` uses `crypto/rand`; use Gibbs to sample
		// Sampler:    dmmclust.NewGibbs(rand.New(rand.NewSource(randSeed))), // use Gibbs to sample
	}, err
}
