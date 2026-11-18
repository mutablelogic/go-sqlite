package sqlite

import (
	"net/url"
	"testing"

	assert "github.com/stretchr/testify/assert"
)

func Test_OffsetLimit_Clamp_001(t *testing.T) {
	assert := assert.New(t)

	limit := uint64(50)
	offsetLimit := OffsetLimit{Offset: 0, Limit: &limit}
	offsetLimit.Clamp(100)

	assert.Equal(uint64(50), limit)
}

func Test_OffsetLimit_Clamp_002(t *testing.T) {
	assert := assert.New(t)

	limit := uint64(50)
	offsetLimit := OffsetLimit{Offset: 90, Limit: &limit}
	offsetLimit.Clamp(100)

	assert.Equal(uint64(10), limit)
}

func Test_OffsetLimit_Clamp_003(t *testing.T) {
	assert := assert.New(t)

	limit := uint64(50)
	offsetLimit := OffsetLimit{Offset: 100, Limit: &limit}
	offsetLimit.Clamp(100)

	assert.Equal(uint64(0), limit)
}

func Test_OffsetLimit_Clamp_004(t *testing.T) {
	assert := assert.New(t)

	limit := uint64(50)
	offsetLimit := OffsetLimit{Offset: 120, Limit: &limit}
	offsetLimit.Clamp(100)

	assert.Equal(uint64(0), limit)
}

func Test_OffsetLimit_Clamp_005(t *testing.T) {
	offsetLimit := OffsetLimit{Offset: 10, Limit: nil}
	offsetLimit.Clamp(100)
}

func Test_OffsetLimit_Query_001(t *testing.T) {
	assert := assert.New(t)

	params := url.Values{}
	offsetLimit := OffsetLimit{}
	offsetLimit.Query(params)

	assert.Empty(params)
}

func Test_OffsetLimit_Query_002(t *testing.T) {
	assert := assert.New(t)

	limit := uint64(50)
	params := url.Values{}
	offsetLimit := OffsetLimit{Limit: &limit}
	offsetLimit.Query(params)

	assert.Equal("50", params.Get("limit"))
	assert.False(params.Has("offset"))
}

func Test_OffsetLimit_Query_003(t *testing.T) {
	assert := assert.New(t)

	params := url.Values{}
	offsetLimit := OffsetLimit{Offset: 10}
	offsetLimit.Query(params)

	assert.Equal("10", params.Get("offset"))
	assert.False(params.Has("limit"))
}

func Test_OffsetLimit_Query_004(t *testing.T) {
	assert := assert.New(t)

	limit := uint64(50)
	params := url.Values{}
	offsetLimit := OffsetLimit{Offset: 10, Limit: &limit}
	offsetLimit.Query(params)

	assert.Equal("limit=50&offset=10", params.Encode())
}

func Test_OffsetLimit_Query_005(t *testing.T) {
	assert := assert.New(t)

	// A zero limit is explicitly set, so it should be included
	limit := uint64(0)
	params := url.Values{}
	offsetLimit := OffsetLimit{Limit: &limit}
	offsetLimit.Query(params)

	assert.Equal("0", params.Get("limit"))
}

func Test_OffsetLimit_Query_006(t *testing.T) {
	assert := assert.New(t)

	// Existing values are replaced, and other parameters are preserved
	limit := uint64(50)
	params := url.Values{"limit": {"1", "2"}, "offset": {"3"}, "q": {"test"}}
	offsetLimit := OffsetLimit{Offset: 10, Limit: &limit}
	offsetLimit.Query(params)

	assert.Equal([]string{"50"}, params["limit"])
	assert.Equal([]string{"10"}, params["offset"])
	assert.Equal("test", params.Get("q"))
}

func Test_OffsetLimit_Query_007(t *testing.T) {
	assert := assert.New(t)

	// Zero offset does not remove an existing offset parameter
	params := url.Values{"offset": {"3"}}
	offsetLimit := OffsetLimit{}
	offsetLimit.Query(params)

	assert.Equal("3", params.Get("offset"))
}
