package mediatype

import (
	"errors"
	"mime"
	"reflect"
	"strconv"
	"strings"
)

// TypeWildcard defines the character that identifies wildcard types and subtypes.
const TypeWildcard = "*"

// ParameterQuality is the type parameter identifying the quality value of a media type.
const ParameterQuality = "q"

// Nil is a special MediaType that indicates "no media type".
// The nil value for MediaType is Nil.
// This value is usually accompanied by an error giving the reason why a valid media type could not be constructed.
//
// The Nil type has some special semantics:
//   - The type, subtype and suffix are all empty
//   - Nil is not contained in any other type, nor does it contain other types
//   - Nil is incompatible with other types, even "*/*"
//   - Nil is less specific than all non-Nil types
//   - Nil has quality 0
//   - [MediaType.IsNil] returns true
var Nil = MediaType{}

// The MediaType implements a RFC 6838 media type.
// A media type value is immutable.
// Instead of mutating a media type, use the "With..." methods to create a copy with the appropriate values.
type MediaType struct {
	tpe, subtype string
	params       map[string]string
}

// MustParse works like [Parse] but panics if v cannot be parsed.
func MustParse(v string) MediaType {
	t, err := Parse(v)
	if err != nil {
		panic(err)
	}
	return t
}

// Parse parses v into a MediaType.
// If v does not describe a media type in a valid way, err will be non-nil.
// Depending on the type of error t might still be a valid media type.
// Especially if parsing media type parameters cause an error, t will still be set to the media type without parameters.
func Parse(v string) (t MediaType, err error) {
	t.tpe, t.params, err = mime.ParseMediaType(v)
	if err != nil && !errors.Is(err, mime.ErrInvalidMediaParameter) {
		return Nil, err
	}
	if len(t.params) == 0 {
		t.params = nil
	}
	t.tpe, t.subtype, _ = strings.Cut(t.tpe, "/")
	if t.subtype == "" {
		return Nil, errors.New("mediatype: missing subtype")
	}
	if t.tpe == TypeWildcard {
		t.subtype = TypeWildcard
	}
	if t.HasParameter(ParameterQuality) {
		// normalize quality parameter
		t = t.WithQuality(t.Quality())
	}
	// If err != nil this indicates an invalid parameter.
	// We do return t in that case without parameters.
	return t, err
}

// NewMediaType creates a new media type with the given type, subtype and parameters.
// The type, subtype and parameter keys must be valid tokens as defined in RFC 1521 and RFC 2045.
// If this is not the case, Nil is returned.
//
// The parameters are specified as a sequence of key, value pairs.
// If an uneven number of parameters is supplied, the last one will be ignored.
//
// Known parameter values may be normalized.
// Types like "*/example" will be converted to "*/*".
func NewMediaType(tpe, subtype string, params ...string) MediaType {
	if !isToken(tpe) || !isToken(subtype) {
		return Nil
	}
	if tpe == TypeWildcard {
		// use */* as the only valid wildcard
		subtype = TypeWildcard
	}
	t := MediaType{
		tpe:     strings.ToLower(tpe),
		subtype: strings.ToLower(subtype),
		params:  nil,
	}
	if len(params) > 0 {
		t.params = make(map[string]string, len(params)/2)
	}
	for i := 0; i < len(params)-1; i += 2 {
		if !isToken(params[i]) {
			return Nil
		}
		t.params[strings.ToLower(params[i])] = params[i+1]
	}
	if t.HasParameter(ParameterQuality) {
		// normalize quality parameter
		t = t.WithQuality(t.Quality())
	}
	return t
}

// Type returns the major type of t ("application" in "application/json").
func (t MediaType) Type() string {
	return t.tpe
}

// WithType returns a new media type with its type set to tpe.
// If t.IsNil() is true or tpe is not a valid token, Nil is returned.
func (t MediaType) WithType(tpe string) MediaType {
	if t.IsNil() || !isToken(tpe) {
		return Nil
	}
	t.tpe = strings.ToLower(tpe)
	if t.IsWildcardType() {
		// use */* as the only valid wildcard
		t.subtype = TypeWildcard
	}
	return t
}

// Subtype returns the minor type of t ("json" in "application/json").
func (t MediaType) Subtype() string {
	return t.subtype
}

// WithSubtype returns a new media type with its subtype set to subtype.
// If t.IsNil() is true or subtype is not a valid token, Nil is returned.
func (t MediaType) WithSubtype(subtype string) MediaType {
	if t.IsNil() || !isToken(subtype) {
		return Nil
	}
	if t.IsWildcardType() {
		// We cannot make a subtype of */*
		return t
	}
	t.subtype = strings.ToLower(subtype)
	return t
}

// FullType returns the full type of t in the format type/subtype.
func (t MediaType) FullType() string {
	if t.IsNil() {
		return ""
	}
	return t.tpe + "/" + t.subtype
}

// HasParameter checks if a parameter with the specified key is defined in t.
func (t MediaType) HasParameter(key string) bool {
	_, ok := t.params[strings.ToLower(key)]
	return ok
}

// Parameter returns the value for the parameter key from t.
// If no such parameter exists, an empty string is returned.
func (t MediaType) Parameter(key string) string {
	return t.params[strings.ToLower(key)]
}

// Parameters returns a non-nil map of all parameters of t.
func (t MediaType) Parameters() map[string]string {
	params := make(map[string]string, len(t.params))
	for key, value := range t.params {
		params[key] = value
	}
	return params
}

// WithParameters returns a new media type with all parameters from t and the specified params.
// If a parameter is present in t and params, the value from params takes precedence.
//
// In contrast to [MediaType.WithExactParameters] this method retains the parameters of t that are not set in params.
func (t MediaType) WithParameters(params map[string]string) MediaType {
	if t.IsNil() {
		return t
	}
	newParams := make(map[string]string, len(t.params)+len(params)/2)
	for k, v := range t.params {
		newParams[k] = v
	}
	for k, v := range params {
		if !isToken(k) {
			return Nil
		}
		newParams[strings.ToLower(k)] = v
	}
	t.params = newParams
	return t
}

// WithExactParameters returns a new media type with the specified set of parameters.
//
// In contrast to [MediaType.WithParameters] any existing parameters of t are discarded.
func (t MediaType) WithExactParameters(params map[string]string) MediaType {
	if t.IsNil() {
		return Nil
	}
	t.params = make(map[string]string, len(params)/2)
	for k, v := range params {
		if !isToken(k) {
			return Nil
		}
		t.params[strings.ToLower(k)] = v
	}
	return t
}

// WithoutParameters returns a new media type with some or all parameters of t removed.
// If you don't specify any params, the returned media type will not have any parameters set.
// If you do specify at least one param, only the specified parameters will be removed.
func (t MediaType) WithoutParameters(params ...string) MediaType {
	if len(t.params) == 0 || len(params) == 0 {
		// this is also the case for Nil
		t.params = nil
	} else {
		t.params = make(map[string]string, len(t.params))
		for k, v := range t.params {
			t.params[k] = v
		}
		for _, k := range params {
			delete(t.params, strings.ToLower(k))
		}
		if len(t.params) == 0 {
			t.params = nil
		}
	}
	return t
}

// WithQuality returns a new media type that is equivalent to t but has the quality parameter "q" set to q.
// The value q is truncated to 3 decimal places (as per the spec).
// q is not normalized any further, values outside the [0, 1] interval are taken as-is.
// You can use [MediaType.WithNormalizedQuality] to make sure the quality value is within bounds.
func (t MediaType) WithQuality(q float64) MediaType {
	if t.IsNil() {
		return t
	}
	return t.WithParameters(map[string]string{ParameterQuality: strconv.FormatFloat(q, 'f', 3, 64)})
}

// WithNormalizedQuality returns a new media type making sure that its quality value q is within 0 <= q <= 1.
func (t MediaType) WithNormalizedQuality() MediaType {
	if t.IsNil() {
		return t
	}
	if !t.HasParameter(ParameterQuality) {
		return t
	}
	q := t.Quality()
	if q < 0 {
		q = 0
	} else if q > 1 {
		q = 1
	} else {
		return t
	}
	return t.WithQuality(q)
}

// IsWildcardType indicates whether t describes the "*/*" type.
// Nil is the only type that is neither concrete nor a wildcard.
func (t MediaType) IsWildcardType() bool {
	return t.tpe == TypeWildcard
}

// IsWildcardSubtype indicates whether t describes a wildcard subtype or a suffix wildcard.
// Both "example/*" and "example/*+json" are considered wildcard subtypes.
// Nil is the only type that is neither concrete nor a wildcard.
func (t MediaType) IsWildcardSubtype() bool {
	return t.subtype == TypeWildcard || strings.HasPrefix(t.subtype, "*+")
}

// IsConcrete indicates whether t is a concrete media type (that is not a wildcard).
// Nil is the only type that is neither concrete nor a wildcard.
func (t MediaType) IsConcrete() bool {
	return !t.IsNil() && !t.IsWildcardType() && !t.IsWildcardSubtype()
}

// SubtypeSuffix returns the subtype suffix (if any).
// For example "application/problem+json" has a subtype suffix of "json".
// If t does not have a subtype suffix, an empty string is returned.
func (t MediaType) SubtypeSuffix() string {
	idx := strings.LastIndexByte(t.subtype, '+')
	if idx < 0 {
		return ""
	}
	return t.subtype[idx+1:]
}

// Quality returns a parsed value for the "q" parameter of t.
// If t does not have a "q" parameter, 1 is returned.
// If the q-value cannot be parsed, 0 is returned.
// Nil has quality 0.
func (t MediaType) Quality() float64 {
	if t.IsNil() {
		return 0
	}
	if v, ok := t.params[ParameterQuality]; ok {
		f, _ := strconv.ParseFloat(v, 64)
		return f
	}
	return 1
}

// IsNil checks whether t identifies the Nil type (and should thus be considered invalid).
func (t MediaType) IsNil() bool {
	return t.tpe == ""
}

// Equals compares t to other and returns true if both media types should be considered equal.
// This method checks parameter equality as well.
// If you only want to check for equal types, use [MediaType.EqualsType].
func (t MediaType) Equals(other MediaType) bool {
	return reflect.DeepEqual(t, other)
}

// EqualsType checks if t and other describe the same fundamental type.
// In contrast to [MediaType.Equals] this method ignores any type parameters.
// Wildcards and their subtypes are NOT considered equal.
func (t MediaType) EqualsType(other MediaType) bool {
	return t.tpe == other.tpe && t.subtype == other.subtype
}

// Includes checks whether t includes other.
// Inclusion rules are as follows:
//   - The wildcard type "*/*" includes every other type.
//   - A subtype wildcard like "example/*" includes every type in its subtree.
//   - A suffix wildcard like "example/*+json" includes types in the same tree ending with the same suffix.
//     For example both "example/json" and "example/foo+json" are included in "example/*+json".
//   - A concrete type like "application/json" only includes itself.
//
// Type parameters are not considered for inclusion.
//
// Note that this method is not symmetric.
// "text/*" includes "text/plain" but not the other way round.
// If you need a symmetric variant, use [MediaType.IsCompatibleWith] instead.
func (t MediaType) Includes(other MediaType) bool {
	if t.IsNil() || other.IsNil() {
		return false
	} else if t.IsWildcardType() {
		// */* includes anything
		return true
	} else if t.tpe != other.tpe {
		return false
	} else if t.subtype == other.subtype {
		// t is exactly equal to other
		return true
	} else if !t.IsWildcardSubtype() {
		return false
	}
	idx := strings.LastIndexByte(t.subtype, '+')
	if idx < 0 {
		// example/* includes example/<anything>
		return true
	}
	suffix := t.subtype[idx+1:]
	if suffix == other.subtype || suffix == other.SubtypeSuffix() {
		// example/*+xml includes example/xml
		// example/*+xml includes example/foo+xml
		return true
	}
	return false
}

// IsCompatibleWith checks whether t and other are compatible with each other.
// This is a symmetric version of [MediaType.Includes].
// See the documentation on [MediaType.Includes] for a definition of the inclusion rules.
func (t MediaType) IsCompatibleWith(other MediaType) bool {
	if t.IsNil() || other.IsNil() {
		return false
	} else if t.IsWildcardType() || other.IsWildcardType() {
		// one is */*
		return true
	} else if t.tpe != other.tpe {
		return false
	} else if t.subtype == other.subtype {
		// t and other are exactly equal
		return true
	} else if !t.IsWildcardSubtype() && !other.IsWildcardSubtype() {
		return false
	} else if t.subtype == TypeWildcard || other.subtype == TypeWildcard {
		// One is example/*, the other is example/<anything>
		return true
	} else if t.IsWildcardSubtype() {
		// t is example/*+xml, check if other is example/xml or example/<anything>+xml
		suffix := t.SubtypeSuffix()
		return suffix == other.subtype || suffix == other.SubtypeSuffix()
	} else if other.IsWildcardSubtype() {
		// other is example/*+xml, check if t is example/xml or example/<anything>+xml
		suffix := other.SubtypeSuffix()
		return suffix == t.subtype || suffix == t.SubtypeSuffix()
	}
	return false
}

// IsMoreSpecific indicates whether t is more specific than other.
// The specificity of a type can fall in one of 4 classes (ordered from more specific to less specific):
//  1. Concrete types like "application/json"
//  2. Suffix wildcards like "application/*+json"
//  3. Wildcard subtypes like "example/*"
//  4. General wildcards like "*/*"
//  5. The Nil type
//
// If t and other fall into the same specificity class, the type with more parameters is more specific.
func (t MediaType) IsMoreSpecific(other MediaType) bool {
	if t.IsNil() {
		return false
	} else if other.IsNil() {
		return true
	} else if t.IsWildcardType() != other.IsWildcardType() {
		// exactly one wildcard
		return other.IsWildcardType()
	} else if t.IsWildcardSubtype() != other.IsWildcardSubtype() {
		// exactly one wildcard subtype
		return other.IsWildcardSubtype()
	} else if t.tpe == other.tpe && t.subtype == other.subtype {
		// exactly equal types (might be */*)
		return len(t.params) > len(other.params)
	} else if t.subtype == TypeWildcard || other.subtype == TypeWildcard {
		// t is example/*+xml, other is example/*
		return other.subtype == TypeWildcard
	}
	return false
}

// IsLessSpecific indicates whether t is less specific than other.
// This is the inverse to [MediaType.IsMoreSpecific].
func (t MediaType) IsLessSpecific(other MediaType) bool {
	return other.IsMoreSpecific(t)
}

// HasHigherPriority indicates whether t has a higher priority than other.
// The priority is mainly determined by the [MediaType.Quality] of a media type, where a higher quality indicates a higher priority.
// If t and other have the same quality, the more specific type has higher priority (as indicated by [MediaType.IsMoreSpecific]).
func (t MediaType) HasHigherPriority(other MediaType) bool {
	qt := t.Quality()
	qo := other.Quality()
	if qt != qo {
		return qt > qo
	}
	return t.IsMoreSpecific(other)
}

// HasLowerPriority indicates whether t has a lower priority than other.
// This is the inverse to [MediaType.HasHigherPriority].
func (t MediaType) HasLowerPriority(other MediaType) bool {
	return other.HasHigherPriority(t)
}

// String returns a string representation of the media type.
// The resulting string conforms to the syntax required by the Content-Type HTTP header.
func (t MediaType) String() string {
	return mime.FormatMediaType(t.tpe+"/"+t.subtype, t.params)
}
