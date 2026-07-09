package search

import (
	"context"
	"math"
	"slices"
	"strings"
	"sync"
)

type Index struct {
	mu sync.RWMutex
	docs map[string]*Document
	postings map[string][]Posting
}

type Document struct {
	Path string
	TermCount int
}

type Posting struct {
	Path string
	Positions []int
	TF float64
}

type Result struct {
	Path string
	Score float64
}

func NewIndex() *Index {
	return &Index{
		docs: make(map[string]*Document),
		postings: make(map[string][]Posting),
	}
}

func (idx *Index) Add(path string, content []byte) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	text := string(content)
	terms := tokenize(text)
	termPos := make(map[string][]int)
	termCount := len(terms)
	d := Document{
		Path: path,
		TermCount: termCount,
	}

	idx.docs[path] = &d
	for i, t := range terms {
		pos, ok := termPos[t]
		if ok {
			termPos[t] = append(pos, i)
		} else {
			termPos[t] = []int{i}
		}
	}

	for term, pos := range termPos {
		c := len(pos)
		tf := float64(c) / float64(termCount)
		posting := Posting{
			Path: path,
			Positions: pos,
			TF: tf,
		}
		postings, ok := idx.postings[term]
		if ok {
			idx.postings[term] = append(postings, posting)
		} else {
			idx.postings[term] = []Posting{posting}
		}
	}
}

func (idx *Index) Remove(path string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	delete(idx.docs, path)
	for t, postings := range idx.postings {
		remainingPostings := make([]Posting, 0)
		for _, p := range postings {
			if p.Path != path {
				remainingPostings = append(remainingPostings, p)
			}
		}
		if len(remainingPostings) == 0 {
			// delete term
			delete(idx.postings, t)
		} else if len(remainingPostings) != len(postings) {
			idx.postings[t] = remainingPostings
		}
	}
}


func (idx *Index) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	idx.mu.Lock()
	defer idx.mu.Unlock()
	terms := tokenize(query)
	results, err := idx.bm25(ctx, terms, limit)
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (idx *Index) bm25(ctx context.Context, terms []string, limit int) ([]Result, error) {
	const k1 = 1.5
	const b = 0.75

	N := float64(len(idx.docs))
	if N == 0 {
		return []Result{}, nil
	}

	// Check for ctx cancel
	if err := ctx.Err(); err != nil {
		return nil, ctx.Err()
	}
	// Average document length
	var totalTerms int
	for _, doc := range idx.docs {
		totalTerms += doc.TermCount
	}

	avgdl := float64(totalTerms) / N

	// Check for ctx cancel
	if err := ctx.Err(); err != nil {
		return nil, ctx.Err()
	}
	// Accumulate scores per document
	scores := make(map[string]float64)
	for _, term := range terms {
		postings, ok := idx.postings[term]
		if !ok {
			continue
		}

		// IDF: log((N - n + 0.5) / (n + 0.5) + 1)
		n := float64(len(postings))
		idf := math.Log((N - n+0.5)/(n+0.5) + 1)

		// Check for ctx cancel
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		for _, p := range postings {
			// Check for ctx cancel
			if err := ctx.Err(); err != nil {
				return nil, ctx.Err()
			}
			doc := idx.docs[p.Path]
			if doc == nil {
				continue
			}

			f := float64(len(p.Positions)) // raw term frequency
			dl := float64(doc.TermCount) // doc length

			// BM35 score contribution
			num := f * (k1 + 1)
			denom := f + k1*(1-b+b*dl/avgdl)
			scores[p.Path] += idf * num / denom
		}
	}

	// Convert to sorted results
	results := make([]Result, 0, len(scores))
	for path, score := range scores {
		results = append(results, Result{Path: path, Score: score})
	}

	// Check for ctx cancel
	if err := ctx.Err(); err != nil {
		return nil, ctx.Err()
	}
	slices.SortFunc(results, func(a, b Result) int {
		if a.Score > b.Score {
			return -1
		}
		if a.Score < b.Score {
			return 1
		}
		return strings.Compare(a.Path, b.Path) // stable tie-breaker
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}
	return results, nil
}

func squeeze(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func tokenize(text string) []string {
	// lowercase, split on non-alphanumeric
	// return unique terms with positions
	text = strings.ToLower(text)
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\r", " ")
	text = squeeze(text)
	terms := strings.Split(text, " ")
	return terms
}
