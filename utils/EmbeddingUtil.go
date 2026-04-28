package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/viper"
)

type RequestContent struct {
	Text string `json:"text,omitempty"`
}

type EmbeddingRequest struct {
	Model string `json:"model"`
	Input struct {
		Contents []RequestContent `json:"contents"`
	} `json:"input"`
	Parameters struct {
		Dimension int `json:"dimension,omitempty"` // 控制输出的维度
	} `json:"parameters,omitempty"`
}

type EmbeddingResponse struct {
	Output struct {
		Embeddings []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"embeddings"`
	} `json:"output"`
}

// 获取文本的向量 (1152 维)
func GetEmbeddings(texts []string) ([][]float32, error) {
	apiKey := viper.GetString("dashscope.api_key")
	url := "https://dashscope.aliyuncs.com/api/v1/services/embeddings/multimodal-embedding/multimodal-embedding"

	model := viper.GetString("dashscope.embedding_model")
	if model == "" {
		model = "tongyi-embedding-vision-plus-2026-03-06"
	}

	// 将字符串切片转换为 RequestContent结构 （把字符串存入contents中，赋值给请求体）
	var contents []RequestContent
	for _, text := range texts {
		contents = append(contents, RequestContent{Text: text})
	}

	reqBody := EmbeddingRequest{
		Model: model,
	}
	reqBody.Input.Contents = contents

	// 在 Milvus 中创建的 Collection 是 1152 维的，这里需要显式指定降维
	reqBody.Parameters.Dimension = 1152

	jsonData, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	if len(result.Output.Embeddings) == 0 {
		return nil, fmt.Errorf("failed to get embeddings")
	}

	var vectors [][]float32
	for _, emb := range result.Output.Embeddings {
		vectors = append(vectors, emb.Embedding)
	}
	return vectors, nil
}

// GetEmbeddingsInBatches 分批获取向量，解决单次请求数量限制
// chunks: 所有文本切片
// batchSize: 每次请求的最大文本数量（阿里云要求的是 20）
func GetEmbeddingsInBatches(chunks []string, batchSize int) ([][]float32, error) {
	var allVectors [][]float32
	totalChunks := len(chunks)

	// 遍历所有切片，步长为 batchSize
	for i := 0; i < totalChunks; i += batchSize {
		// 计算当前批次的结束索引
		end := i + batchSize
		if end > totalChunks {
			end = totalChunks
		}

		// 截取当前批次的切片
		batch := chunks[i:end]

		// 调用底层 API 请求函数获取当前批次切片的向量化结果
		batchVectors, err := GetEmbeddings(batch)
		if err != nil {
			return nil, fmt.Errorf("获取第 %d-%d 个切片向量时失败: %v", i+1, end, err)
		}

		// 将当前批次的向量追加到总结果集中
		allVectors = append(allVectors, batchVectors...)
	}

	return allVectors, nil
}
