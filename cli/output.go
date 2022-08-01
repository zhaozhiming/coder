package cli

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/xerrors"

	"github.com/coder/coder/cli/cliui"
)

var outputFormatJSON = "json"

type OutputFormatter[T any] struct {
	// Name is the name of the formatter as it should be supplied via the
	// --output flag.
	Name string
	// AttachFlags is an optional function that can be used to attach any
	// additional flags to the command to be used when formatting the output.
	// The `name` argument is the name of the formatter.
	AttachFlags func(cmd *cobra.Command, name string)
	// Fn is the render function. It should return a string and optionally an
	// error with no side effects. The `cmd` object is supplied only for reading
	// flags and should not be used for writing to stdout. Writing to stderr is
	// permitted if necessary, however.
	Fn func(cmd *cobra.Command, out T) (string, error)
}

func Formatter[T any](name string, fn func(*cobra.Command, T) (string, error)) OutputFormatter[T] {
	return FormatterWithFlags(name, nil, fn)
}

func FormatterWithFlags[T any](name string, attachFlags func(*cobra.Command, string), fn func(*cobra.Command, T) (string, error)) OutputFormatter[T] {
	return OutputFormatter[T]{
		Name:        name,
		AttachFlags: attachFlags,
		Fn:          fn,
	}
}

func SetupDisplay[T any](defaultFormatter string, formatters ...OutputFormatter[T]) (display func(cmd *cobra.Command, out T) error, attachFlags func(cmd *cobra.Command)) {
	// We don't insert our default formatters here to avoid extra generics spam
	// blowing up the binary size. They do get added to the names array though.
	formatterMap := map[string]func(*cobra.Command, T) (string, error){}
	formatterNames := []string{}
	for _, formatter := range formatters {
		if _, ok := formatterMap[formatter.Name]; ok {
			panic("duplicate output formatter name: " + formatter.Name)
		}
		formatterMap[formatter.Name] = formatter.Fn
		formatterNames = append(formatterNames, formatter.Name)
	}

	// Add our default formatter names to the names array if they aren't already
	// in there. They don't exist in the map to avoid generics.
	if _, ok := formatterMap[outputFormatJSON]; !ok {
		formatterNames = append(formatterNames, outputFormatJSON)
	}
	sort.Strings(formatterNames)

	// Verify that the "default" formatter exists.
	if _, ok := formatterMap[defaultFormatter]; !ok {
		panic("default output formatter not found: " + defaultFormatter)
	}

	displayFn := func(cmd *cobra.Command, out T) error {
		format, err := cmd.Flags().GetString(varOutputFormat)
		if err != nil {
			return xerrors.Errorf("determine output format type: %w", err)
		}
		if format == "" {
			format = defaultFormatter
		}

		var outputString string
		if fn, ok := formatterMap[format]; ok {
			outputString, err = fn(cmd, out)
		} else {
			// Default formatters.
			switch format {
			case outputFormatJSON:
				outputString, err = displayJSON(out)
			default:
				return xerrors.Errorf("unknown output format %q, acceptable formats: %q", format, strings.Join(formatterNames, `", "`))
			}
		}

		if err != nil {
			return xerrors.Errorf("format output with %q formatter: %w", format, err)
		}

		_, err = fmt.Fprint(cmd.OutOrStdout(), outputString)
		if err != nil {
			return xerrors.Errorf("write formatted output: %w", err)
		}

		return nil
	}

	attachFlagsFn := func(cmd *cobra.Command) {
		// Add the --output flag.
		usage := "Output format. Available formats are: " + strings.Join(formatterNames, ", ")
		value := &enumValue{allowed: formatterNames, value: defaultFormatter}
		cmd.Flags().VarP(value, varOutputFormat, "o", usage)

		// Add any additional flags for the formatters.
		for _, formatter := range formatters {
			if formatter.AttachFlags != nil {
				formatter.AttachFlags(cmd, formatter.Name)
			}
		}
	}

	return displayFn, attachFlagsFn
}

// DisplayTable renders a table as a string. The input argument must be a slice
// of structs. At least one field in the struct must have a `table:""` tag
// containing the name of the column in the outputted table.
//
// Nested structs are processed if the field has the `table:"$NAME,recursive"`
// tag and their fields will be named as `$PARENT_NAME $NAME`.
//
// If sort is empty, the input order will be used. If filterColumns is empty or
// nil, all available columns are included.
func DisplayTable[T any](out []T, sort string, filterColumns []string) (string, error) {
	v := reflect.Indirect(reflect.ValueOf(out))

	tw := cliui.Table()
	headersRaw := typeToTableHeaders(v.Type().Elem())
	headers := make(table.Row, len(headersRaw))
	for i, header := range headersRaw {
		headers[i] = header
	}
	tw.AppendHeader(headers)
	tw.SetColumnConfigs(cliui.FilterTableColumns(headers, filterColumns))
	tw.SortBy([]table.SortBy{{
		Name: sort,
	}})

	// Write each struct to the table.
	for i := 0; i < v.Len(); i++ {
		// Format the row as a slice.
		rowMap := valueToTableMap(v.Index(i))
		rowSlice := make([]interface{}, len(headers))
		for i, h := range headersRaw {
			v, ok := rowMap[h]
			if !ok {
				v = nil
			}

			rowSlice[i] = v
		}

		tw.AppendRow(table.Row(rowSlice))
	}

	return tw.Render(), nil
}

func parseTableStructTag(tag string) (name string, recurse bool) {
	tagSplit := strings.Split(tag, ",")
	if len(tagSplit) == 0 || tagSplit[0] == "" || (len(tagSplit) == 2 && strings.TrimSpace(tagSplit[1]) != "recursive") || len(tagSplit) > 2 {
		panic(fmt.Sprintf(`invalid table tag %q, must be a non-empty string, optionally followed by ",recursive"`, tag))
	}

	return tagSplit[0], len(tagSplit) == 2
}

// typeToTableHeaders converts a type to a slice of column names. If the given
// type is not a struct, this function will panic.
func typeToTableHeaders(t reflect.Type) []string {
	if t.Kind() != reflect.Struct {
		panic("typeToTableHeaders called with non-struct type")
	}

	headers := []string{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("table")
		if tag == "" {
			continue
		}

		name, recurse := parseTableStructTag(tag)
		if recurse {
			// If it's not a struct or a pointer to a struct.
			if field.Type.Kind() != reflect.Struct || (field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() != reflect.Struct) {
				panic(fmt.Sprintf("invalid table tag %q, field %q is not a struct or a pointer to a struct so we cannot recurse", tag, field.Name))
			}
			fieldType := field.Type
			if field.Type.Kind() == reflect.Pointer {
				fieldType = field.Type.Elem()
			}

			childNames := typeToTableHeaders(fieldType)
			for _, childName := range childNames {
				headers = append(headers, fmt.Sprintf("%s %s", name, childName))
			}
			continue
		}

		headers = append(headers, name)
	}

	return headers
}

// valueToTableMap converts a struct to a map of column name to value. If the
// given value is not a struct, this function will panic.
func valueToTableMap(val reflect.Value) map[string]interface{} {
	val = reflect.Indirect(val)
	if val.Type().Kind() != reflect.Struct {
		panic("valueToTableMap called with non-struct value")
	}

	if val.IsNil() {
		return nil
	}

	row := map[string]interface{}{}
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		tag := field.Tag.Get("table")
		if tag == "" {
			continue
		}

		// If the field is a struct, recursively process it.
		name, recurse := parseTableStructTag(tag)
		if recurse {
			// If it's not a struct or a pointer to a struct.
			if field.Type.Kind() != reflect.Struct || (field.Type.Kind() == reflect.Pointer && field.Type.Elem().Kind() != reflect.Struct) {
				panic(fmt.Sprintf("invalid table tag %q, field %q is not a struct or a pointer to a struct so we cannot recurse", tag, field.Name))
			}
			if field.Type.Kind() == reflect.Pointer {
				val = val.Elem()
			}

			childMap := valueToTableMap(val.Field(i))
			for childName, childValue := range childMap {
				row[fmt.Sprintf("%s %s", name, childName)] = childValue
			}
			continue
		}

		// Otherwise, we just use the field value.
		row[name] = val.Field(i).Interface()
	}

	return row
}

func displayJSON(out interface{}) (string, error) {
	b, err := json.Marshal(out)
	if err != nil {
		return "", xerrors.Errorf("marshal JSON: %w", err)
	}
	return string(b), nil
}

// enumValue provides a flag value that validates that the given value is one of
// the allowed values.
type enumValue struct {
	allowed []string
	value   string
}

var _ pflag.Value = &enumValue{}

// Type implements pflag.Value.
func (*enumValue) Type() string {
	return "string"
}

// String implements pflag.Value.
func (f *enumValue) String() string {
	return f.value
}

// Set implements pflag.Value.
func (f *enumValue) Set(val string) error {
	for _, allowed := range f.allowed {
		if val == allowed {
			f.value = val
			return nil
		}
	}

	return xerrors.Errorf("the flag value %q is not accepted here, acceptable values: %q", val, strings.Join(f.allowed, `", "`))
}
