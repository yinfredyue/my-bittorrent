package encode

import (
	"fmt"
	"sort"
	"strings"
)

func encodeInt(sb *strings.Builder, i int) error {
	_, err := sb.WriteString(fmt.Sprintf("i%de", i))
	return err
}

func encodeString(sb *strings.Builder, s string) error {
	_, err := sb.WriteString(fmt.Sprintf("%d:%s", len(s), s))
	return err
}

func encode(sb *strings.Builder, v interface{}) error {
	switch v := v.(type) {
	case int:
		return encodeInt(sb, v)
	case string:
		return encodeString(sb, v)
	case []interface{}:
		sb.WriteString("l")
		for _, e := range v {
			err := encode(sb, e)
			if err != nil {
				return err
			}
		}
		sb.WriteString("e")
	case map[string]interface{}:
		sb.WriteString("d")

		// sort keys
		keys := make([]string, 0)
		for k := range v {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			err := encode(sb, k)
			if err != nil {
				return err
			}
			err = encode(sb, v[k])
			if err != nil {
				return err
			}
		}
		sb.WriteString("e")
	default:
		return fmt.Errorf("unhandled case")
	}

	return nil
}

func Encode(v interface{}) (string, error) {
	var sb strings.Builder

	err := encode(&sb, v)

	if err != nil {
		return "", err
	}

	return sb.String(), nil
}
