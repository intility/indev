package outputformat

import "errors"

var ErrInvalidOutputFormat = errors.New(`must be one of "wide", "json", "yaml"`)

type Format string

func (o *Format) String() string {
	return string(*o)
}

func (o *Format) Set(value string) error {
	switch value {
	case "wide", "json", "yaml":
		*o = Format(value)
		return nil
	default:
		return ErrInvalidOutputFormat
	}
}

func (o *Format) Type() string {
	return "outputFormat"
}
