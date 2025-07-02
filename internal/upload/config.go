package upload

type Config struct {
	Type   BedType      `json:"type"`
	Github GithubConfig `json:"github"`
}

var (
	GITHUB BedType = "github"
)
