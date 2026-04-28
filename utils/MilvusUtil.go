package utils

import (
	"context"
	"fmt"
	"log"

	"github.com/antidote-kt/SSE_Library-back/constant"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
	"github.com/milvus-io/milvus-sdk-go/v2/entity"
	"github.com/spf13/viper"
)

var MilvusClient client.Client

// 初始化 Milvus 并创建集合
func InitMilvus() {
	addr := viper.GetString("milvus.address")
	var err error
	MilvusClient, err = client.NewClient(context.Background(), client.Config{Address: addr})
	if err != nil {
		log.Fatalf("Failed to connect to Milvus: %v", err)
	}

	ctx := context.Background()
	has, _ := MilvusClient.HasCollection(ctx, constant.CollectionName)
	if !has {
		// 定义表结构 Schema
		schema := entity.NewSchema().WithName(constant.CollectionName).WithDescription("RAG Collection")
		schema.WithField(entity.NewField().WithName("id").WithDataType(entity.FieldTypeInt64).WithIsAutoID(true).WithIsPrimaryKey(true))
		schema.WithField(entity.NewField().WithName("file_id").WithDataType(entity.FieldTypeInt64)) // 关联 MySQL 中的文档 ID
		schema.WithField(entity.NewField().WithName("content").WithDataType(entity.FieldTypeVarChar).WithMaxLength(65535))
		schema.WithField(entity.NewField().WithName("vector").WithDataType(entity.FieldTypeFloatVector).WithDim(1152))

		err = MilvusClient.CreateCollection(ctx, schema, entity.DefaultShardNumber)
		if err != nil {
			log.Fatalf("Failed to create collection: %v", err)
		}

		// 创建索引加速检索 (HNSW 算法)
		idx, _ := entity.NewIndexHNSW(entity.L2, 8, 96)
		MilvusClient.CreateIndex(ctx, constant.CollectionName, "vector", idx, false)
	}
	MilvusClient.LoadCollection(ctx, constant.CollectionName, false)
}

// 插入向量和数据
func InsertChunks(fileID int64, chunks []string, vectors [][]float32) error {
	ctx := context.Background()

	fileIds := make([]int64, len(chunks))
	for i := range fileIds {
		fileIds[i] = fileID
	}

	idCol := entity.NewColumnInt64("file_id", fileIds)
	contentCol := entity.NewColumnVarChar("content", chunks)
	vectorCol := entity.NewColumnFloatVector("vector", 1152, vectors)

	_, err := MilvusClient.Insert(ctx, constant.CollectionName, "", idCol, contentCol, vectorCol)
	MilvusClient.Flush(ctx, constant.CollectionName, false) // 强制落盘
	return err
}

// 相似度检索
func SearchKnowledge(queryVector []float32, topK int) ([]string, error) {
	ctx := context.Background()
	sp, _ := entity.NewIndexHNSWSearchParam(74) // 创建HNSW索引搜索参数(ef=74)

	searchResult, err := MilvusClient.Search(
		ctx, constant.CollectionName, // 1. collName: 集合名称
		[]string{},          // 2. partitions: 分区列表，传空数组代表全库检索，不做条件过滤
		"",                  // 3. expr表达式过滤，如果想要限定某个 file_id 可以在这里写 "file_id == 1"
		[]string{"content"}, // 4. outputFields: 需要一同返回的标量字段
		[]entity.Vector{entity.FloatVector(queryVector)}, // 5. vectors: 要查询的向量列表
		"vector",  // 6. vectorField: 数据库里存向量的字段名
		entity.L2, // 7. metricType: 计算距离的方式（L2 欧氏距离）
		topK,      // 8. topK: 返回的最相似结果数量
		sp,        // 9. sp: 特定算法的搜索参数
	)
	if err != nil {
		return nil, fmt.Errorf("Milvus 搜索失败: %v", err)
	}

	var results []string
	for _, res := range searchResult {
		// 安全检查：如果该向量没有找到任何匹配项，直接跳过
		if res.ResultCount == 0 {
			continue
		}

		// 获取列数据
		column := res.Fields.GetColumn("content")
		if column == nil {
			// 如果没拿到列，说明可能字段名写错了或者该结果集为空
			continue
		}

		// 取出刚才在 outputFields 里要求返回的 "content" 字段
		contentCol := res.Fields.GetColumn("content").(*entity.ColumnVarChar)
		for i := 0; i < contentCol.Len(); i++ {
			results = append(results, contentCol.Data()[i])
		}
	}
	return results, nil
}
