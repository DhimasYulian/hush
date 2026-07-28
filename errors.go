package hush

import (
	"github.com/DhimasYulian/hush/internal/build"
	"github.com/DhimasYulian/hush/internal/parse"
	"github.com/DhimasYulian/hush/internal/query"
	"github.com/DhimasYulian/hush/internal/schema"
	"github.com/DhimasYulian/hush/internal/validate"
)

// Parse phase errors.
var (
	ErrEmptyKey            = parse.ErrEmptyKey
	ErrEmptySegment        = parse.ErrEmptySegment
	ErrInvalidSyntax       = parse.ErrInvalidSyntax
	ErrUnexpectedCharacter = parse.ErrUnexpectedCharacter
)

// Build phase errors.
var (
	ErrInvalidFields  = build.ErrInvalidFields
	ErrInvalidSort    = build.ErrInvalidSort
	ErrInvalidFilters = build.ErrInvalidFilters
	ErrInvalidGroupBy = build.ErrInvalidGroupBy
)

// Schema builder errors.
var (
	ErrEmptyName         = schema.ErrEmptyName
	ErrNoOperators       = schema.ErrNoOperators
	ErrUnknownOperator   = schema.ErrUnknownOperator
	ErrDuplicateField    = schema.ErrDuplicateField
	ErrDuplicateRelation = schema.ErrDuplicateRelation
	ErrNilTarget         = schema.ErrNilTarget
	ErrInvalidMaxDepth   = schema.ErrInvalidMaxDepth
	ErrInvalidMaxLimit   = schema.ErrInvalidMaxLimit
	ErrUnknownFieldType  = schema.ErrUnknownFieldType
)

// Validation phase errors.
var (
	ErrInvalidPath        = validate.ErrInvalidPath
	ErrUnknownField       = validate.ErrUnknownField
	ErrOperatorNotAllowed = validate.ErrOperatorNotAllowed
	ErrNestingTooDeep     = validate.ErrNestingTooDeep
	ErrInvalidValue       = validate.ErrInvalidValue
	ErrUnknownFilterNode  = validate.ErrUnknownFilterNode
	ErrMissingSchema      = validate.ErrMissingSchema
	ErrUnknownGroupBy     = validate.ErrUnknownGroupBy
)

// Shared cross-domain errors.
var (
	ErrInvalidPopulate    = query.ErrInvalidPopulate
	ErrInvalidPagination  = query.ErrInvalidPagination
	ErrInvalidAggregation = query.ErrInvalidAggregation
)

// Error is a structured validation error with context about what failed.
type Error = query.Error
