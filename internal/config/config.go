package config

type Config struct {
    InputPath  string `json:"input_path"`
    OutputPath string `json:"output_path"`
    Limit      int    `json:"limit"`
    Format     string `json:"format"` // "json", "sql", "text"
}

func InitConfig()*Config{
	return &Config{
		InputPath: "./dabkrs/dabkrs_1.dsl",
		OutputPath: "./output/",
		Limit: 100,
		Format: "json",
	}
}