package ssm

import (
	"fmt"

	"github.com/disneystreaming/ssm-helpers/util"
)

// ListSlice is for standard Var flag usage
type ListSlice []string

// CommaSlice enables comma-delimited input for a Var flag
type CommaSlice []string

// SemiSlice enables semicolon-delimited input for a Var flag
type SemiSlice []string

func (l *ListSlice) String() string {
	return fmt.Sprintf("%s", *l)
}

// Set appends the value provided to the current ListSlice object
func (l *ListSlice) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func (l *ListSlice) Type() string {
	return "stringSlice"
}

func (s *SemiSlice) String() string {
	return fmt.Sprintf("%s", *s)
}

// Set takes a list of semicolon-delimited values, splits them, and appends each item to the current SemiSlice object
func (s *SemiSlice) Set(value string) error {
	list := util.SemiSplit(value)
	for _, v := range list {
		*s = append(*s, v)
	}
	return nil
}

func (s *SemiSlice) Type() string {
	return "stringSlice"
}

func (c *CommaSlice) String() string {
	return fmt.Sprintf("%s", *c)
}

// Set takes a list of comma-delimited values, splits them, and appends each item to the current CommaSlice object
func (c *CommaSlice) Set(value string) error {
	list := util.CommaSplit(value)
	for _, v := range list {
		*c = append(*c, v)
	}
	return nil
}

func (c *CommaSlice) Type() string {
	return "stringSlice"
}
