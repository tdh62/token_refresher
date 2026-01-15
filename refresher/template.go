package refresher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"jwt_refresher/models"
	"text/template"
)

func RenderTemplate(tmpl string, project *models.Project) (string, error) {
	// 解析自定义变量
	var customVars map[string]interface{}
	if project.CustomVariables != "" {
		if err := json.Unmarshal([]byte(project.CustomVariables), &customVars); err != nil {
			return "", fmt.Errorf("failed to parse custom variables: %w", err)
		}
	} else {
		customVars = make(map[string]interface{})
	}

	// 添加RefreshToken到变量中
	customVars["RefreshToken"] = project.CurrentRefreshToken

	t, err := template.New("body").Parse(tmpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, customVars); err != nil {
		return "", err
	}

	return buf.String(), nil
}
