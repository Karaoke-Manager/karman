package mediatype

import (
	"strings"
)

// MediaTypes is a list of MediaType objects.
// The MediaTypes type implements an algorithm for content type negotiation.
//
// A MediaTypes can be used to represent an Accept header of an HTTP request.
//
// MediaTypes implements [sort.Interface].
// A sorted lists orders media types in descending order by [MediaType.HasHigherPriority].
type MediaTypes []MediaType

// ParseList parses one or more comma separated lists of media types.
// This function can be used to parse an Accept header into a MediaTypes list.
//
// In contrast to [ParseListStrict] this function discards invalid media types silently.
func ParseList(l ...string) MediaTypes {
	var types MediaTypes
	for _, raw := range l {
		if strings.TrimSpace(raw) == "" {
			// An empty string indicates an empty list
			continue
		}
		elems := strings.Split(raw, ",")
		for _, v := range elems {
			if t, err := Parse(v); err == nil && t.Quality() > 0 {
				types = append(types, t)
			}
		}
	}
	return types
}

// ParseListStrict parses one or more comma separated lists of media types.
// If an invalid media type is encountered, an empty list and an error is returned.
// This function can be used to parse an Accept header into a MediaTypes list.
//
// In contrast to [ParseList] this function aborts with an error if an invalid media type is found.
func ParseListStrict(l ...string) (MediaTypes, error) {
	var types MediaTypes
	for _, raw := range l {
		if strings.TrimSpace(raw) == "" {
			// An empty string indicates an empty list
			continue
		}
		elems := strings.Split(raw, ",")
		for _, v := range elems {
			t, err := Parse(v)
			if err != nil {
				return nil, err
			}
			types = append(types, t)
		}
	}
	return types, nil
}

// FindMatches returns a list of matching content types between l and the list of available content types.
// A match is a content type that fits both a media type (or wildcard) in l and a media type (or wildcard) in the available types.
// The match will always be the more concrete of the two types.
// The quality of a match will be set to the product of the qualities of the matching types from l and the available types.
//
// If l is an empty list, every available type is considered a match.
// If the available list is empty, there will be no matches.
// An empty return value indicates that no matches were found.
//
// The returned list is not sorted.
// Sorting the list would give you all matching media types in decreasing order of priority.
// It is strongly recommended to use a stable sorting algorithm to preserve the order of types with equal priorities.
func (l MediaTypes) FindMatches(available ...MediaType) MediaTypes {
	matches := make(MediaTypes, 0)
	for _, c := range available {
		q := c.Quality()
		if len(l) == 0 {
			matches = append(matches, c)
		}
		for _, t := range l {
			if c.IsCompatibleWith(t) {
				tq := t.Quality() * q
				if tq <= 0 {
					continue
				}
				if c.IsMoreSpecific(t) {
					t = c
				}
				t = t.WithQuality(tq)
				matches = append(matches, t)
			}
		}
	}
	return matches
}

// BestMatch compares l with the available media types and selects the best match.
// If no matches are available, Nil is returned.
func (l MediaTypes) BestMatch(available ...MediaType) MediaType {
	return l.FindMatches(available...).Best().WithoutParameters(ParameterQuality)
}

// Best returns the media type from l that has the highest priority.
// If l is empty, Nil is returned.
func (l MediaTypes) Best() MediaType {
	if len(l) == 0 {
		return Nil
	}
	best := l[0]
	for _, t := range l {
		if t.HasHigherPriority(best) {
			best = t
		}
	}
	return best
}

// GetType returns the first element of l whose type equals t, or Nil if no such element exists.
func (l MediaTypes) GetType(t MediaType) MediaType {
	for _, c := range l {
		if c.EqualsType(t) {
			return c
		}
	}
	return Nil
}

// GetCompatible returns the first element of l that is compatible with t, or Nil if no such element exists.
func (l MediaTypes) GetCompatible(t MediaType) MediaType {
	for _, c := range l {
		if c.IsCompatibleWith(t) {
			return c
		}
	}
	return Nil
}

// Includes determines if any of the types in l include t.
func (l MediaTypes) Includes(t MediaType) bool {
	for _, c := range l {
		if c.Includes(t) {
			return true
		}
	}
	return false
}

// String returns a string representation of l that can be used as the value for an Accept header.
func (l MediaTypes) String() string {
	comps := make([]string, len(l))
	for i := range l {
		comps[i] = l[i].String()
	}
	return strings.Join(comps, ", ")
}

// Len implements [sort.Interface].
func (l MediaTypes) Len() int {
	return len(l)
}

// Less implements [sort.Interface].
func (l MediaTypes) Less(i, j int) bool {
	return l[i].HasHigherPriority(l[j])
}

// Swap implements [sort.Interface].
func (l MediaTypes) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
