package main

import (
	"context"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
	"github.com/smallnest/langgraphgo/rag"
	"github.com/smallnest/langgraphgo/rag/store"
)

func main() {
	ctx := context.Background()

	// 连接到FalkorDB
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	defer client.Close()

	fmt.Println("=== 调试FalkorDB查询问题 ===\n")

	// 1. 检查图中有多少节点
	fmt.Println("1. 检查所有节点:")
	allNodesQuery := "MATCH (n) RETURN count(n) as count"
	result, err := client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", allNodesQuery, "--compact").Result()
	if err != nil {
		log.Printf("查询节点数失败: %v", err)
	} else {
		fmt.Printf("总节点数: %v\n", result)
	}

	// 2. 检查所有标签
	fmt.Println("\n2. 检查所有标签:")
	allLabelsQuery := "CALL db.labels() YIELD label RETURN label"
	result, err = client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", allLabelsQuery, "--compact").Result()
	if err != nil {
		log.Printf("查询标签失败: %v", err)
	} else {
		fmt.Printf("所有标签: %v\n", result)
	}

	// 3. 检查特定类型的节点
	fmt.Println("\n3. 检查PERSON节点:")
	personQuery := "MATCH (n:PERSON) RETURN n"
	result, err = client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", personQuery, "--compact").Result()
	if err != nil {
		log.Printf("查询PERSON节点失败: %v", err)
	} else {
		fmt.Printf("PERSON节点: %v\n", result)
		if r, ok := result.([]any); ok && len(r) > 1 {
			if rows, ok := r[1].([]any); ok {
				fmt.Printf("找到 %d 个PERSON节点\n", len(rows))
				for i, row := range rows {
					fmt.Printf("  [%d] %v\n", i, row)
				}
			}
		}
	}

	// 4. 检查所有属性
	fmt.Println("\n4. 检查john_smith节点的详细信息:")
	detailQuery := "MATCH (n) WHERE n.id = 'john_smith' RETURN n, labels(n), properties(n)"
	result, err = client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", detailQuery, "--compact").Result()
	if err != nil {
		log.Printf("查询详细信息失败: %v", err)
	} else {
		fmt.Printf("详细信息: %v\n", result)
	}

	// 5. 测试不使用compact模式
	fmt.Println("\n5. 测试不使用compact模式:")
	normalQuery := "MATCH (n:PERSON) RETURN n.id, n.name"
	result, err = client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", normalQuery).Result()
	if err != nil {
		log.Printf("普通查询失败: %v", err)
	} else {
		fmt.Printf("普通查询结果: %v\n", result)
	}

	// 6. 创建一个新的测试节点
	fmt.Println("\n6. 创建新的测试节点:")
	createQuery := "CREATE (p:Person {id: 'test_person', name: 'Test Person', role: 'test'})"
	result, err = client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", createQuery, "--compact").Result()
	if err != nil {
		log.Printf("创建测试节点失败: %v", err)
	} else {
		fmt.Printf("创建结果: %v\n", result)
	}

	// 7. 再次查询Person
	fmt.Println("\n7. 再次查询Person节点:")
	result, err = client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", personQuery, "--compact").Result()
	if err != nil {
		log.Printf("查询PERSON节点失败: %v", err)
	} else {
		fmt.Printf("PERSON节点: %v\n", result)
	}

	// 8. 使用知识图接口测试
	fmt.Println("\n8. 使用知识图接口测试:")
	kg, err := store.NewFalkorDBGraph("falkordb://localhost:6379/simple_rag_graph")
	if err != nil {
		log.Printf("创建知识图失败: %v", err)
		return
	}
	defer func() {
		if falkorDB, ok := kg.(*store.FalkorDBGraph); ok {
			falkorDB.Close()
		}
	}()

	// 测试GetEntity
	fmt.Println("测试GetEntity:")
	testEntity, err := kg.GetEntity(ctx, "test_person")
	if err != nil {
		log.Printf("获取test_person失败: %v", err)
	} else {
		fmt.Printf("获取到test_person: ID=%s, Name=%s, Type=%s\n", testEntity.ID, testEntity.Name, testEntity.Type)
		fmt.Printf("Properties: %+v\n", testEntity.Properties)
	}

	// 调试：手动解析test_person
	fmt.Println("\n调试test_person的原始数据:")
	debugQuery := "MATCH (n) WHERE n.id = 'test_person' RETURN n"
	debugResult, err := client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", debugQuery, "--compact").Result()
	if err != nil {
		log.Printf("调试查询失败: %v", err)
	} else {
		fmt.Printf("test_person原始数据: %v\n", debugResult)
		if r, ok := debugResult.([]any); ok && len(r) > 1 {
			if rows, ok := r[1].([]any); ok && len(rows) > 0 {
				if row, ok := rows[0].([]any); ok && len(row) > 0 {
					if node, ok := row[0].([]any); ok {
						fmt.Printf("node结构长度: %d\n", len(node))
						for i, part := range node {
							fmt.Printf("  [%d] type: %T, value: %v\n", i, part, part)
							if arr, ok := part.([]any); ok {
								fmt.Printf("      子数组长度: %d\n", len(arr))
								for j, subpart := range arr {
									fmt.Printf("        [%d] type: %T, value: %v\n", j, subpart, subpart)
								}
							}
						}
					}
				}
			}
		}
	}

	// 测试Query
	fmt.Println("\n测试Query方法:")
	graphQuery := &rag.GraphQuery{
		EntityTypes: []string{"PERSON"},
		Limit:       10,
	}
	queryResult, err := kg.Query(ctx, graphQuery)
	if err != nil {
		log.Printf("Query失败: %v", err)
	} else {
		fmt.Printf("Query结果: 找到%d个实体, %d个关系\n", len(queryResult.Entities), len(queryResult.Relationships))
		for i, entity := range queryResult.Entities {
			fmt.Printf("  [%d] %s (%s)\n", i, entity.Name, entity.Type)
		}
	}
}
