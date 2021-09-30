package sqlite3

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

////////////////////////////////////////////////////////////////////////////////
// TYPES

// A single sample
type sample struct {
	t time.Time
	d time.Duration
}

// An array of samples
type samplearray struct {
	key     string
	samples []sample
	cap     int
	n       int
}

type profilearray struct {
	sync.RWMutex
	sync.WaitGroup
	m   map[string]*samplearray
	cap int
	n   int
	age time.Duration
}

type profilesample struct {
	key            string
	count          int
	delta          time.Duration
	min, mean, max time.Duration
}

type samplearr []*profilesample

////////////////////////////////////////////////////////////////////////////////
// GLOBALS

const (
	defaultSampleSize  = 100            // Maximum number of samples per profile
	defaultProfileSize = 100            // Maximum number of profiles
	defaultAge         = time.Hour * 24 // Default age of profiles to cull
)

////////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// NewSampleArray returns a new set of samples, up to "capacity" samples.
func NewSampleArray(cap int) *samplearray {
	a := new(samplearray)
	a.cap = intMin(intMax(1, cap), defaultSampleSize)
	a.n = 0
	a.samples = make([]sample, 0, a.cap)
	return a
}

// NewProfileArray returns a new set of profiles, up to "capacity" profiles,
// removing the oldest profiles based on age
func NewProfileArray(cap, samples int, age time.Duration) *profilearray {
	p := new(profilearray)
	p.cap = intMin(intMax(1, cap), defaultProfileSize)
	p.n = samples
	p.m = make(map[string]*samplearray, p.cap)
	if age <= 0 {
		p.age = defaultAge
	} else {
		p.age = age
	}
	return p
}

func (p *profilearray) Close() {
	// Wait for garbage collection to complete
	p.Wait()
	// Release resources
	p.m = nil
}

////////////////////////////////////////////////////////////////////////////////
// STRINGIFY

func (a *samplearray) String() string {
	str := "<samples"
	str += fmt.Sprint(" n=", len(a.samples))
	for i := 0; i < len(a.samples); i++ {
		str += fmt.Sprintf(" %v", a.samples[a.i(i)].d)
	}
	return str + ">"
}

func (s *profilesample) String() string {
	str := "<sample"
	str += fmt.Sprintf(" key=%q", s.key)
	str += fmt.Sprint(" min=", s.min.Truncate(time.Millisecond))
	str += fmt.Sprint(" mean=", s.mean.Truncate(time.Millisecond))
	str += fmt.Sprint(" max=", s.max.Truncate(time.Millisecond))
	str += fmt.Sprint(" count=", s.count)
	str += fmt.Sprint(" delta=", s.delta)
	str += fmt.Sprintf(" ops/s=%.1f", float64(s.count)/s.delta.Seconds())
	return str + ">"
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PROFILE ARRAY

// Return the query text
func (s *profilesample) SQL() string {
	return s.key
}

// Return the minimum query time
func (s *profilesample) Min() time.Duration {
	return s.min
}

// Return the maximum query time
func (s *profilesample) Max() time.Duration {
	return s.max
}

// Return the mean average query time
func (s *profilesample) Mean() time.Duration {
	return s.mean
}

// Return the number of samples
func (s *profilesample) Count() int {
	return s.count
}

// Return the period over which the samples were taken
func (s *profilesample) Delta() time.Duration {
	return s.delta
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - PROFILE ARRAY

// Add a new profile into the array of profiles, bump out the oldest profile
func (p *profilearray) Add(key string, d time.Duration) {
	p.RWMutex.Lock()
	defer p.RWMutex.Unlock()
	if _, exists := p.m[key]; !exists {
		p.m[key] = NewSampleArray(p.n)
	}
	p.m[key].Add(d)

	// Cull the least used profiles by date and by oldest once we
	// reach capacity
	if len(p.m) > p.cap {
		p.WaitGroup.Add(1)
		go func() {
			defer p.WaitGroup.Done()
			p.garbagecollect(p.cap >> 1)
		}()
	}
}

// Return n profiles - key, delta, count and mean - sorted by mean so the
// slowest queries are first. Stats are then ops per second and
// mean time taken per query.
func (p *profilearray) SlowQueries(n int) []*profilesample {
	// Check arguments
	if n < 1 {
		return nil
	}
	// Populate results
	results := make(samplearr, 0, len(p.m))
	p.RWMutex.RLock()
	for key, samples := range p.m {
		results = append(results, samples.NewSlowQuery(key))
	}
	p.RWMutex.RUnlock()

	// Sort results with the slowest queries first
	sort.Sort(results)

	// Return top n results
	n = intMin(n, len(p.m))
	return results[:n]
}

////////////////////////////////////////////////////////////////////////////////
// PUBLIC METHODS - SAMPLE ARRAY

func (a *samplearray) NewSlowQuery(key string) *profilesample {
	s := new(profilesample)
	s.key = key
	s.count = len(a.samples)
	s.delta = a.Delta()

	var sum float64
	for i := 0; i < len(a.samples); i++ {
		v := a.samples[a.i(i)].d
		if i == 0 {
			s.min, s.max = durationMin(v, s.min), durationMax(v, s.max)
		}
	}
	s.mean = time.Duration(sum / float64(s.count))
	return s
}

// Count returns the number of samples in the array
func (a *samplearray) Count() int {
	return len(a.samples)
}

// Last returns the time since the last sample time
func (a *samplearray) Last() time.Duration {
	if len(a.samples) == 0 {
		return 0
	} else {
		return time.Since(a.samples[a.i(-1)].t)
	}
}

// Add a new sample into the array of samples, bump out the oldest sample
func (a *samplearray) Add(d time.Duration) {
	if a.n >= len(a.samples) {
		a.samples = a.samples[:a.n+1]
	}
	a.samples[a.n] = sample{t: time.Now(), d: d}
	a.n = a.n + 1
	if a.n >= a.cap {
		a.n = 0
	}
}

// Delta returns the duration between the minumum and maximum samples
func (a *samplearray) Delta() time.Duration {
	if len(a.samples) == 0 {
		return 0
	}
	return a.samples[a.i(-1)].t.Sub(a.samples[a.i(0)].t)
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - PROFILE ARRAY

// Remove the profiles down to "cap" profiles, removing oldest profiles first
func (p *profilearray) garbagecollect(cap int) {
	// Remove by age
	keys := make([]string, 0, cap)
	p.RWMutex.RLock()
	for key, sample := range p.m {
		if len(keys) < cap {
			if sample.Last() > p.age {
				keys = append(keys, key)
			}
		}
	}
	p.RWMutex.RUnlock()

	// Remove keys
	if len(keys) > 0 {
		p.RWMutex.Lock()
		for _, key := range keys {
			delete(p.m, key)
		}
		p.RWMutex.Unlock()
	}

	// TODO: Cull by query time - sort samples by mean and remove oldest N
	if len(p.m) > cap {
		fmt.Println("TODO: required cap ", cap, " actual cap", len(p.m))
	}
}

func (a samplearr) Len() int {
	return len(a)
}

func (a samplearr) Less(i, j int) bool {
	return a[i].mean > a[j].mean
}

func (a samplearr) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

////////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS - SAMPLE ARRAY

// Return the index of the i'th sample
func (a *samplearray) i(i int) int {
	j := (a.n + i) % len(a.samples)
	if j < 0 {
		j = len(a.samples) + j
	}
	return j
}
