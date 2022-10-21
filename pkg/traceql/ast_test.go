package traceql

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatic_Equals(t *testing.T) {
	areEqual := []struct {
		lhs, rhs Static
	}{
		{NewStaticInt(1), NewStaticInt(1)},
		{NewStaticFloat(1.5), NewStaticFloat(1.5)},
		{NewStaticString("foo"), NewStaticString("foo")},
		{NewStaticBool(true), NewStaticBool(true)},
		{NewStaticDuration(1 * time.Second), NewStaticDuration(1000 * time.Millisecond)},
		{NewStaticStatus(StatusOk), NewStaticStatus(StatusOk)},
		// Status and int comparison
		{NewStaticStatus(StatusError), NewStaticInt(0)},
		{NewStaticStatus(StatusOk), NewStaticInt(1)},
		{NewStaticStatus(StatusUnset), NewStaticInt(2)},
	}
	areNotEqual := []struct {
		lhs, rhs Static
	}{
		{NewStaticInt(1), NewStaticInt(2)},
		{NewStaticInt(1), NewStaticFloat(1)},
		{NewStaticBool(true), NewStaticInt(1)},
		{NewStaticString("foo"), NewStaticString("bar")},
		{NewStaticDuration(0), NewStaticInt(0)},
		{NewStaticStatus(StatusError), NewStaticStatus(StatusOk)},
		{NewStaticStatus(StatusOk), NewStaticInt(0)},
		{NewStaticStatus(StatusError), NewStaticFloat(0)},
	}
	for _, tt := range areEqual {
		t.Run(fmt.Sprintf("%v == %v", tt.lhs, tt.rhs), func(t *testing.T) {
			assert.True(t, tt.lhs.Equals(tt.rhs))
			assert.True(t, tt.rhs.Equals(tt.lhs))
		})
	}
	for _, tt := range areNotEqual {
		t.Run(fmt.Sprintf("%v != %v", tt.lhs, tt.rhs), func(t *testing.T) {
			assert.False(t, tt.lhs.Equals(tt.rhs))
			assert.False(t, tt.rhs.Equals(tt.lhs))
		})
	}
}

func TestSpansetFilterEvaluate(t *testing.T) {
	testCases := []struct {
		query  string
		input  []Spanset
		output []Spanset
	}{
		{
			"{ true }",
			[]Spanset{
				// Empty spanset is dropped
				{Spans: []Span{}},
				{Spans: []Span{{}}},
			},
			[]Spanset{
				{Spans: []Span{{}}},
			},
		},
		{
			"{ .foo = `a` }",
			[]Spanset{
				{Spans: []Span{
					// Second span should be dropped here
					{ID: []byte{1}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticString("a")}},
					{ID: []byte{2}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticString("b")}},
				}},
				{Spans: []Span{
					// This entire spanset will be dropped
					{ID: []byte{3}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticString("b")}},
				}},
			},
			[]Spanset{
				{Spans: []Span{
					{ID: []byte{1}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticString("a")}}}},
			},
		},
		{
			"{ .foo = 1 || (.foo >= 4 && .foo < 6) }",
			[]Spanset{
				{Spans: []Span{
					// Second span should be dropped here
					{ID: []byte{1}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticInt(1)}},
					{ID: []byte{2}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticInt(2)}},
				}},
				{Spans: []Span{
					// First span should be dropped here
					{ID: []byte{3}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticInt(3)}},
					{ID: []byte{4}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticInt(4)}},
					{ID: []byte{5}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticInt(5)}},
				}},
				{Spans: []Span{
					// Entire spanset should be dropped
					{ID: []byte{3}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticInt(6)}},
					{ID: []byte{4}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticInt(7)}},
				}},
			},
			[]Spanset{
				{Spans: []Span{
					{ID: []byte{1}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticInt(1)}}}},
				{Spans: []Span{
					{ID: []byte{4}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticInt(4)}},
					{ID: []byte{5}, Attributes: map[Attribute]Static{NewAttribute("foo"): NewStaticInt(5)}},
				}},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.query, func(t *testing.T) {
			ast, err := Parse(tc.query)
			require.NoError(t, err)

			filt := ast.Pipeline.Elements[0].(SpansetFilter)

			actual, err := filt.evaluate(tc.input)
			require.NoError(t, err)
			require.Equal(t, tc.output, actual)
		})
	}
}