package config

type Config struct {
	Server ServerConfig `yaml:"server"`
	Mysql  MysqlConfig  `yaml:"mysql"`
	Rag    RagConfig    `yaml:"rag"`
	LLM    LLMConfig    `yaml:"llm"`
}

type ServerConfig struct {
	Port string `yaml:"port"`
}

type MysqlConfig struct {
	Dsn string `yaml:"dsn"`
}

type RagConfig struct {
	PineconeHost        string  `yaml:"pinecone_host"`
	PineconeAPIKey      string  `yaml:"pinecone_api_key"`
	PineconeIndex       string  `yaml:"pinecone_index"`
	EmbeddingURL        string  `yaml:"embedding_url"`
	EmbedDim            int     `yaml:"embed_dim"`
	TopK                int     `yaml:"topk"`
	SimilarityThreshold float64 `yaml:"similarity_threshold"`
}

type LLMConfig struct {
	URL       string `yaml:"url"`
	Model     string `yaml:"model"`
	MaxTokens int    `yaml:"max_tokens"`
}
