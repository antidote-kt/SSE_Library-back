package controllers

import (
	"log"
	"os"

	"github.com/antidote-kt/SSE_Library-back/utils"
)

func LearnDocument(fid int, cosUrl string) {
	// 开启异步 Goroutine
	go func(fid int64, url string) {
		// 1. 下载到本地临时文件
		tmpPath, err := utils.DownloadFromCOSToTemp(url)
		if err != nil {
			log.Printf("MilVus [错误]: 下载文档 %d 失败: %v\n", fid, err)
			return
		}
		log.Printf("MilVus: 下载文档 %d 成功，临时路径为: %s\n", fid, tmpPath)
		defer os.Remove(tmpPath)

		// 2. 提取、清洗并切片
		text, err := utils.ExtractTextFromPDF(tmpPath)
		if err != nil {
			log.Printf("MilVus [错误]: 提取文档 %d 的 PDF 文本失败: %v\n", fid, err)
			return
		}

		cleanedText := utils.CleanText(text)
		chunks := utils.ChunkText(cleanedText, 500, 50)

		if len(chunks) == 0 {
			log.Printf("MilVus [警告]: 文档 %d 未提取到任何文本内容(可能为扫描件)\n", fid)
			return
		}

		// 3. 批量向量化
		vectors, err := utils.GetEmbeddingsInBatches(chunks, 20)
		if err != nil {
			log.Printf("MilVus [错误]: 文档 %d 向量化失败: %v\n", fid, err)
			return
		}

		// 4. 存入 Milvus
		err = utils.InsertChunks(fid, chunks, vectors)
		if err != nil {
			log.Printf("MilVus [错误]: 文档 %d 存入 Milvus 失败: %v\n", fid, err)
			return
		}

		log.Printf("MilVus [成功]: 文档 %d 学习完成！共成功处理 %d 个切片\n", fid, len(chunks))
	}(int64(fid), cosUrl)
}
