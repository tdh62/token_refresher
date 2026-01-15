package refresher

import (
	"fmt"

	"github.com/tidwall/gjson"
)

func ExtractToken(jsonBody, path string) (string, error) {
	result := gjson.Get(jsonBody, path)
	if !result.Exists() {
		return "", fmt.Errorf("path %s not found in response", path)
	}
	return result.String(), nil
}

func ExtractExpiresIn(jsonBody, path string) (int64, error) {
	result := gjson.Get(jsonBody, path)
	if !result.Exists() {
		return 0, fmt.Errorf("path %s not found in response", path)
	}
	return result.Int(), nil
}
