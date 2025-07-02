package upload

import "encoding/json"

func InitUpload(jsonConfig string) error {
	config := Config{}
	err := json.Unmarshal([]byte(jsonConfig), &config)
	if err != nil {
		return err
	}

	CloseAllStrategy()

	switch config.Type {
	case GITHUB:
		g, err := NewGithubBed(config.Github)
		if err != nil {
			return err
		}
		RegisterStrategy(&g)
	}

	return nil
}
