package dmmcutil

import (
	"errors"
	"fmt"

	"github.com/go-nlp/dmmclust"
	"github.com/grokify/mogo/type/maputil"
)

type DMMClusterer struct {
	Config           dmmclust.Config
	Docs             []dmmclust.Document
	DocClusters      []dmmclust.Cluster // same length as `Docs`
	DocClusterScores [][]float64        // same length as `Docs`
	CenterDocs       map[int]int        // k=clusterID, v=docID
	Clusters         []dmmclust.Cluster
	ClustersMap      map[int]dmmclust.Cluster
	ClustersMetaMap  map[int]ClusterMeta
	//CenterDocAlg4 int // supports duplicate words in doc
}

func NewDMMClusterer(numClusters, clusterSize uint, randSeed int64, docs []dmmclust.Document) (*DMMClusterer, error) {
	cfg, err := ConfigDefault(numClusters, randSeed)
	if err != nil {
		return &DMMClusterer{}, err
	}
	if numClusters == 0 && clusterSize > 0 {
		cfg.K = int(float64(len(docs)) / float64(clusterSize))
	}
	docs2 := Documents(docs)
	cfg.Vocabulary = len(docs2.IDs())
	dmmc := &DMMClusterer{
		Config: cfg,
		Docs:   docs,
	}

	var clustered []dmmclust.Cluster // len(clustered) == len(docs)
	if clustered, err = dmmclust.FindClusters(docs, cfg); err != nil {
		return dmmc, err
	}
	if len(docs) != len(clustered) {
		panic("mismatch")
	}
	dmmc.DocClusters = clustered
	for i, clust := range clustered {
		//doc := docs[i]
		fmt.Printf("\t%d: %q\n", clust.ID(), i)
	}
	err = dmmc.Inflate()
	return dmmc, err
}

func (dmmc *DMMClusterer) Inflate() error {
	dmmc.BuildClustersCanonical()
	err := dmmc.BuildScores()
	if err != nil {
		return err
	}
	dmmc.BuildCenterDocs() // Must run canonical clusters first
	return dmmc.BuildClusterMetas()
}

var ErrNoClusters = errors.New("no clusters error")

func (dmmc *DMMClusterer) BuildScores() error {
	if len(dmmc.Clusters) == 0 {
		return ErrNoClusters
	}
	docsScores := [][]float64{}
	for _, doc := range dmmc.Docs {
		docScores := dmmc.Config.Score(doc, dmmc.Docs, dmmc.Clusters, dmmc.Config)
		docsScores = append(docsScores, docScores)
	}
	dmmc.DocClusterScores = docsScores
	return nil
}

func (dmmc *DMMClusterer) BuildCenterDocs() {
	centerDocs := map[int]int{}              // k=clusterID, v=docID
	centerDocScoresHigh := map[int]float64{} // k=clusterID, v=highest score seen
	for docID, docScores := range dmmc.DocClusterScores {
		for cluID, docCluScore := range docScores {
			if _, ok := centerDocScoresHigh[cluID]; !ok {
				centerDocScoresHigh[cluID] = 0
			}
			if docCluScore > centerDocScoresHigh[cluID] {
				centerDocs[cluID] = docID
				centerDocScoresHigh[cluID] = docCluScore
			}
		}
	}
	dmmc.CenterDocs = centerDocs
}

func (dmmc *DMMClusterer) BuildClustersCanonical() {
	clusMap := map[int]dmmclust.Cluster{}
	for _, clu := range dmmc.DocClusters {
		if _, ok := clusMap[clu.ID()]; !ok {
			clusMap[clu.ID()] = clu
		}
	}
	clus := []dmmclust.Cluster{}
	ids := maputil.IntKeys(clusMap)
	for _, id := range ids {
		clus = append(clus, clusMap[id])
	}
	dmmc.Clusters = clus
	dmmc.ClustersMap = clusMap
}

func (dmmc *DMMClusterer) ClusterDocCounts() (map[int]int, float64, error) {
	counts := map[int]int{}
	if len(dmmc.Clusters) == 0 {
		return counts, -1, ErrNoClusters
	}
	for _, clu := range dmmc.Clusters {
		counts[clu.ID()] = clu.Docs()
	}
	return counts, maputil.NumberValuesAverage(counts), nil
}

func (dmmc *DMMClusterer) BuildClusterMetas() error {
	_, c2d, err := dmmc.BuildDocToCluserMap()
	if err != nil {
		return err
	}
	cMetas := map[int]ClusterMeta{}
	for _, clu := range dmmc.ClustersMap {
		docIDs := map[int]int{}
		if docIDsTry, ok := c2d[clu.ID()]; ok {
			docIDs = docIDsTry
		}
		cm := ClusterMeta{
			ID:      clu.ID(),
			DocIDs:  maputil.IntKeys(docIDs),
			Cluster: clu}
		if centerDocID, ok := dmmc.CenterDocs[clu.ID()]; ok {
			cm.CenterDocID = centerDocID
		}
		cMetas[clu.ID()] = cm
	}
	dmmc.ClustersMetaMap = cMetas
	return nil
}

// BuildDocToCluserMap returns a `map[int]int` where the kesy are the input
// document index and the value is the clusterID.
func (dmmc *DMMClusterer) BuildDocToCluserMap() (map[int]int, map[int]map[int]int, error) {
	c2d := map[int]map[int]int{}
	d2c := map[int]int{}
	for i := range dmmc.Docs {
		if i >= len(dmmc.DocClusters) {
			return d2c, c2d, errors.New("doc cluster not found")
		}
		clu := dmmc.DocClusters[i]
		d2c[i] = clu.ID()
	}
	for docID, cluID := range d2c {
		if _, ok := c2d[cluID]; !ok {
			c2d[cluID] = map[int]int{}
		}
		c2d[cluID][docID]++
	}
	return d2c, c2d, nil
}
