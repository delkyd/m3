// Copyright (c) 2018 Uber Technologies, Inc.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

package client

import (
	"fmt"
	"io"
	"math/rand"
	"os"
	"sort"
	"testing"
	"time"

	"github.com/m3db/m3db/encoding"
	"github.com/m3db/m3db/encoding/m3tsz"
	"github.com/m3db/m3db/generated/thrift/rpc"
	"github.com/m3db/m3db/serialize"
	"github.com/m3db/m3db/x/xpool"
	"github.com/m3db/m3x/ident"
	"github.com/m3db/m3x/pool"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/stretchr/testify/require"
)

func TestFetchTaggedResultsAccumulatorClearResetsState(t *testing.T) {
	pools := newTestFetchTaggedPools()
	accum := newFetchTaggedResultAccumulator()
	iter, exhaustive, err := accum.AsEncodingSeriesIterators(100, pools)
	require.NoError(t, err)
	require.True(t, exhaustive)
	require.Equal(t, 0, iter.Len())
	iter.Close()

	results, err := accum.AsIndexQueryResults(100, pools)
	require.NoError(t, err)
	require.True(t, results.Exhaustive)
	require.False(t, results.Iterator.Next())
	require.NoError(t, results.Iterator.Err())
}

func TestFetchTaggedShardConsistencyResultsInitializeLength(t *testing.T) {
	var results fetchTaggedShardConsistencyResults
	require.Len(t, results, 0)
	results = results.initialize(10)
	require.Len(t, results, 10)
	results = results.initialize(100)
	require.Len(t, results, 100)
}

func TestFetchTaggedShardConsistencyResultsInitializeLengthContract(t *testing.T) {
	var results fetchTaggedShardConsistencyResults
	require.Len(t, results, 0)
	results = results.initialize(100)
	require.Len(t, results, 100)
	results = results[:0]
	results = results.initialize(1)
	require.Len(t, results, 1)
}

func TestFetchTaggedShardConsistencyResultsInitializeResetsValues(t *testing.T) {
	var (
		empty   fetchTaggedShardConsistencyResult
		results fetchTaggedShardConsistencyResults
	)
	require.Len(t, results, 0)
	results = results.initialize(10)
	require.Len(t, results, 10)
	for _, elem := range results {
		require.Equal(t, empty, elem)
	}
}

func TestFetchTaggedForEachIDFn(t *testing.T) {
	input := fetchTaggedIDResults{
		&rpc.FetchTaggedIDResult_{
			ID: []byte("abc"),
		},
		&rpc.FetchTaggedIDResult_{
			ID: []byte("def"),
		},
		&rpc.FetchTaggedIDResult_{
			ID: []byte("abc"),
		},
		&rpc.FetchTaggedIDResult_{
			ID: []byte("xyz"),
		},
	}
	sort.Sort(fetchTaggedIDResultsSortedByID(input))
	numElements := 0
	input.forEachID(func(_ fetchTaggedIDResults, hasMore bool) bool {
		numElements++
		switch numElements {
		case 1:
			require.True(t, hasMore)
		case 2:
			require.True(t, hasMore)
		case 3:
			require.False(t, hasMore)
		default:
			require.Fail(t, "should never reach here")
		}
		return true
	})
	require.Equal(t, 3, numElements)
}

func TestFetchTaggedForEachIDFnEarlyTerminate(t *testing.T) {
	input := fetchTaggedIDResults{
		&rpc.FetchTaggedIDResult_{
			ID: []byte("xyz"),
		},
		&rpc.FetchTaggedIDResult_{
			ID: []byte("abc"),
		},
		&rpc.FetchTaggedIDResult_{
			ID: []byte("def"),
		},
		&rpc.FetchTaggedIDResult_{
			ID: []byte("abc"),
		},
	}
	sort.Sort(fetchTaggedIDResultsSortedByID(input))
	numElements := 0
	input.forEachID(func(elems fetchTaggedIDResults, hasMore bool) bool {
		numElements++
		switch numElements {
		case 1:
			require.Equal(t, "abc", string(elems[0].ID))
			require.True(t, hasMore)
			return true
		case 2:
			require.Equal(t, "def", string(elems[0].ID))
			require.True(t, hasMore)
			return false
		}
		require.Fail(t, fmt.Sprintf("illegal state: %v %+v", string(elems[0].ID), elems))
		return true
	})
	require.Equal(t, 2, numElements)
}

func TestFetchTaggedForEachIDFnNumberCalls(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	seed := time.Now().UnixNano()
	parameters.MinSuccessfulTests = 1000
	parameters.MaxSize = 40
	parameters.Rng = rand.New(rand.NewSource(seed))
	properties := gopter.NewProperties(parameters)

	properties.Property("ForEach is called once per ID", prop.ForAll(
		func(results fetchTaggedIDResults) bool {
			sort.Sort(fetchTaggedIDResultsSortedByID(results))
			ids := make(map[string]struct{})
			results.forEachID(func(elems fetchTaggedIDResults, hasMore bool) bool {
				id := elems[0].ID
				for _, elem := range elems {
					require.Equal(t, id, elem.ID)
				}
				_, ok := ids[string(id)]
				require.False(t, ok)
				ids[string(id)] = struct{}{}
				return true
			})
			return true
		},
		gen.SliceOf(genFetchTaggedIDResult()),
	))

	properties.Property("ForEach correctly indicates it has more elements", prop.ForAll(
		func(results fetchTaggedIDResults) bool {
			sort.Sort(fetchTaggedIDResultsSortedByID(results))
			returnedElems := make(fetchTaggedIDResults, 0, len(results))
			results.forEachID(func(elems fetchTaggedIDResults, hasMore bool) bool {
				returnedElems = append(returnedElems, elems...)
				return hasMore
			})
			return len(results) == len(returnedElems)
		},
		gen.SliceOf(genFetchTaggedIDResult()),
	))

	reporter := gopter.NewFormatedReporter(true, 160, os.Stdout)
	if !properties.Run(reporter) {
		t.Errorf("failed with initial seed: %d", seed)
	}
}

func genFetchTaggedIDResult() gopter.Gen {
	return gen.Identifier().Map(func(s string) *rpc.FetchTaggedIDResult_ {
		return &rpc.FetchTaggedIDResult_{
			ID: []byte(s),
		}
	})
}

func newTestFetchTaggedPools() testFetchTaggedPools {
	pools := testFetchTaggedPools{}
	opts := pool.NewObjectPoolOptions().SetSize(1)

	pools.readerSlices = newReaderSliceOfSlicesIteratorPool(opts)
	pools.readerSlices.Init()

	pools.multiReader = encoding.NewMultiReaderIteratorPool(opts)
	pools.multiReader.Init(func(r io.Reader) encoding.ReaderIterator {
		return m3tsz.NewReaderIterator(r, m3tsz.DefaultIntOptimizationEnabled, encoding.NewOptions())
	})

	pools.seriesIter = encoding.NewSeriesIteratorPool(opts)
	pools.seriesIter.Init()

	pools.mutableSeriesIter = encoding.NewMutableSeriesIteratorsPool(nil)
	pools.mutableSeriesIter.Init()

	pools.iteratorArray = encoding.NewIteratorArrayPool(nil)
	pools.iteratorArray.Init()

	pools.id = ident.NewPool(nil, opts)

	pools.checkedBytesWrapper = xpool.NewCheckedBytesWrapperPool(opts)
	pools.checkedBytesWrapper.Init()

	pools.tagDecoder = serialize.NewTagDecoderPool(serialize.NewTagDecoderOptions(), opts)
	pools.tagDecoder.Init()

	return pools
}

// ensure testFetchTaggedPools satisfies the fetchTaggedPools interface.
var _ fetchTaggedPools = testFetchTaggedPools{}

type testFetchTaggedPools struct {
	readerSlices        *readerSliceOfSlicesIteratorPool
	multiReader         encoding.MultiReaderIteratorPool
	seriesIter          encoding.SeriesIteratorPool
	mutableSeriesIter   encoding.MutableSeriesIteratorsPool
	iteratorArray       encoding.IteratorArrayPool
	id                  ident.Pool
	checkedBytesWrapper xpool.CheckedBytesWrapperPool
	tagDecoder          serialize.TagDecoderPool
}

func (p testFetchTaggedPools) ReaderSliceOfSlicesIterator() *readerSliceOfSlicesIteratorPool {
	return p.readerSlices
}

func (p testFetchTaggedPools) MultiReaderIterator() encoding.MultiReaderIteratorPool {
	return p.multiReader
}

func (p testFetchTaggedPools) SeriesIterator() encoding.SeriesIteratorPool {
	return p.seriesIter
}

func (p testFetchTaggedPools) MutableSeriesIterators() encoding.MutableSeriesIteratorsPool {
	return p.mutableSeriesIter
}

func (p testFetchTaggedPools) IteratorArray() encoding.IteratorArrayPool {
	return p.iteratorArray
}

func (p testFetchTaggedPools) ID() ident.Pool {
	return p.id
}

func (p testFetchTaggedPools) CheckedBytesWrapper() xpool.CheckedBytesWrapperPool {
	return p.checkedBytesWrapper
}

func (p testFetchTaggedPools) TagDecoder() serialize.TagDecoderPool {
	return p.tagDecoder
}