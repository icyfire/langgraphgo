# 使用 FalkorDB 的简单 RAG 知识图谱

本示例提供了一个简化方法来使用 FalkorDB 与 RAG，演示如何手动填充知识图谱并将其用于增强检索。

## 概述

本示例展示：

1. **手动知识图谱创建**：直接添加实体和关系
2. **简单实体管理**：知识图谱的基本 CRUD 操作
3. **直接图查询**：使用 Cypher 查询进行精确数据检索
4. **关系遍历**：探索连接的实体
5. **知识图谱统计**：监控图谱增长和结构

## 关键特性

### 🚀 高性能
- **快速设置**：手动实体定义（无需 LLM 调用）
- **微秒查询**：图查询 <1ms 内完成
- **高效存储**：直接的 Redis/FalkorDB 操作
- **可扩展性**：处理数千个实体和关系

### 📊 完整功能
- **实体管理**：创建、读取、更新、删除操作
- **关系管理**：定义和遍历关系
- **类型过滤**：按实体类型查询（PERSON、ORGANIZATION 等）
- **图谱统计**：跟踪节点和关系

## 前置条件

1. **FalkorDB 服务器**：运行 FalkorDB 实例
   ```bash
   docker run -p 6379:6379 falkordb/falkordb
   ```

2. **Go 依赖**：
   ```bash
   go mod tidy
   ```

## 运行示例

```bash
cd examples/rag_falkordb_simple_fixed
go run main.go
```

## 快速开始

### 1. 基本实体和关系创建

```go
// 创建实体
entities := []*rag.Entity{
    {
        ID:   "john_smith",
        Name: "张三",
        Type: "PERSON",
        Properties: map[string]any{
            "role":        "高级软件工程师",
            "company":     "Google",
        },
    },
    {
        ID:   "google",
        Name: "Google",
        Type: "ORGANIZATION",
        Properties: map[string]any{
            "industry": "科技",
            "location": "山景城，加利福尼亚",
        },
    },
}

// 创建关系
relationships := []*rag.Relationship{
    {
        ID:     "john_works_at_google",
        Source: "john_smith",
        Target: "google",
        Type:   "WORKS_AT",
    },
}
```

### 2. 直接图查询

```go
// 查询特定实体类型
cypherQuery := "MATCH (n:PERSON) RETURN n.id, n.name, n.role, n.company"
result, err := client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", cypherQuery).Result()
```

### 3. 关系遍历

```go
// 查找谁在 Google 工作
cypherQuery := "MATCH (p:PERSON)-[r:WORKS_AT]->(o:ORGANIZATION) WHERE o.name = 'Google' RETURN p.name, r, o.name"
result, err := client.Do(ctx, "GRAPH.QUERY", "simple_rag_graph", cypherQuery).Result()
```

## 示例数据

### 实体

示例创建这些示例实体：

**人员 (PERSON)：**
- 张三：Google 高级软件工程师，专长机器学习/人工智能
- 李四：TechStart Inc. CEO，专注区块链技术

**组织 (ORGANIZATION)：**
- Google：科技公司，位于山景城，加利福尼亚
- TechStart Inc.：区块链初创公司，位于旧金山

**技术 (TECHNOLOGY)：**
- Python：用于机器学习、Web 开发、数据科学的编程语言

**概念 (CONCEPT)：**
- 机器学习：人工智能的子集，使计算机能够从数据中学习

### 关系

**雇佣关系：**
- 张三 `WORKS_AT` Google
- 李四 `CEO_OF` TechStart Inc.

**专业关系：**
- 张三 `SPECIALIZES_IN` 机器学习
- Python `USED_FOR` 机器学习

## 查询示例

### 实体类型查询

```go
// 查找所有人员
"MATCH (n:PERSON) RETURN n.id, n.name, n.role, n.company"

// 查找所有组织
"MATCH (n:ORGANIZATION) RETURN n.id, n.name, n.industry"

// 查找所有技术
"MATCH (n:TECHNOLOGY) RETURN n.id, n.name, n.type, n.uses"

// 查找所有概念
"MATCH (n:CONCEPT) RETURN n.id, n.name, n.description"
```

### 关系查询

```go
// 所有关系
"MATCH (a)-[r]->(b) RETURN a.name, type(r), b.name"

// 谁在哪里工作
"MATCH (p:PERSON)-[r:WORKS_AT]->(o:ORGANIZATION) RETURN p.name, r, o.name"

// 张三专长什么
"MATCH (p {name: '张三'})-[r:SPECIALIZES_IN]->(c) RETURN p.name, type(r), c.name"
```

### 复杂查询

```go
// 查找在科技公司工作的人员
"MATCH (p:PERSON)-[r:WORKS_AT]->(o:ORGANIZATION) WHERE o.industry = '科技' RETURN p.name, o.name"

// 查找与机器学习的所有连接
"MATCH (n)-[*1..2]-(m {name: 'Machine Learning'}) RETURN DISTINCT n.name, type(n)"

// 实体和关系统计
"MATCH (n) RETURN labels(n) as types, count(n) as count ORDER BY types"
"MATCH ()-[r]->() RETURN type(r) as types, count(r) as count ORDER BY types"
```

## 性能特征

### 设置时间

- **实体创建**：6个实体和4个关系约 10ms
- **无 LLM 依赖**：快速且可预测的性能
- **直接数据库操作**：最小开销

### 查询性能

- **简单查询**：约 300-500 微秒
- **复杂查询**：约 1-2 毫秒
- **关系遍历**：约 1-3 毫秒

### 内存使用

- **高效存储**：Redis 中的紧凑表示
- **可扩展**：处理数千个实体，影响最小
- **缓存友好**：内置 Redis 缓存提高性能

## 使用场景

### 1. 快速知识库设置

非常适合使用已知信息创建知识库：

```go
// 预定义公司信息
company := &rag.Entity{
    ID:   "acme_corp",
    Name: "ACME 公司",
    Type: "ORGANIZATION",
    Properties: map[string]any{
        "founded": "1950",
        "employees": 5000,
        "industry": "制造业",
    },
}
```

### 2. 企业关系映射

映射组织关系：

```go
// 员工-公司关系
relationships := []*rag.Relationship{
    {ID: "emp001_works_at_acme", Source: "emp001", Target: "acme_corp", Type: "WORKS_AT"},
    {ID: "emp001_reports_to", Source: "emp001", Target: "mgr001", Type: "REPORTS_TO"},
    {ID: "mgr001_manages", Source: "mgr001", Target: "dept001", Type: "MANAGES"},
}
```

### 3. 产品知识图谱

创建产品层次结构和关系：

```go
// 产品类别和关系
product := &rag.Entity{
    ID:   "iphone_15",
    Name: "iPhone 15",
    Type: "PRODUCT",
    Properties: map[string]any{
        "category": "智能手机",
        "brand":     "苹果",
        "year":      "2023",
    },
}
```

### 4. 技能和专长跟踪

跟踪员工技能和专长：

```go
// 技能关系
skill := &rag.Entity{
    ID:   "python_programming",
    Name: "Python 编程",
    Type: "SKILL",
    Properties: map[string]any{
        "level":      "高级",
        "experience": "5年",
    },
}

relationship := &rag.Relationship{
    ID:     "john_has_python",
    Source: "john_smith",
    Target: "python_programming",
    Type:   "HAS_SKILL",
    Properties: map[string]any{
        "proficiency": "专家",
        "certified":   true,
    },
}
```

## 高级功能

### 1. 自定义 Cypher 查询

使用 Cypher 查询语言的全部功能：

```go
// 复杂多步查询
cypherQuery := `
    MATCH path = (start:PERSON {name: $personName})-[*1..3]-(end:PERSON)
    WHERE end.name <> start.name
    RETURN [node in path | node.name] as path,
           length(path) as distance
`

// 条件查询
cypherQuery := `
    MATCH (n:ORGANIZATION)
    WHERE n.founded >= $year
    RETURN n.name, n.founded
    ORDER BY n.founded
```

### 2. 图遍历模式

```go
// 多层级的同事
cypherQuery := `
    MATCH (p:PERSON {name: $personName})
    -[:REPORTS_TO*1..3]->(colleagues:PERSON)
    RETURN DISTINCT colleagues.name
`

// 拥有相似技能的人员
cypherQuery := `
    MATCH (p1:PERSON)-[:HAS_SKILL]->(s:SKILL)<-[:HAS_SKILL]-(p2:PERSON)
    WHERE p1.name <> p2.name
    RETURN p1.name, p2.name, s.name
```

### 3. 图统计和分析

```go
// 图密度分析
cypherQuery := `
    MATCH (n)
    RETURN count(n) as total_nodes,
           avg(size((n)-[])) as avg_degree
`

// 连通性分析
cypherQuery := `
    MATCH (a:PERSON), (b:PERSON)
    WHERE EXISTS((a)-[*]-(b))
    RETURN count(DISTINCT a) as connected_people
```

## 集成模式

### 1. 与传统 RAG 结合

与向量搜索结合进行混合检索：

```go
// 结合向量和图搜索
vectorResults := vectorStore.Search(query, 5)
graphResults := knowledgeGraph.Query(query)

// 合并和排序结果
mergedResults := mergeSearchResults(vectorResults, graphResults)
```

### 2. 与 Web 应用集成

作为 REST API 暴露：

```go
// 图查询的 HTTP 处理器
func handleEntityQuery(w http.ResponseWriter, r *http.Request) {
    entityTypes := r.URL.Query()["types"]
    query := buildCypherQuery(entityTypes)
    result := executeGraphQuery(query)
    json.NewEncoder(w).Encode(result)
}
```

### 3. 与聊天机器人集成

使用图知识增强聊天机器人响应：

```go
func chatbotResponse(query string) string {
    // 检查查询是否包含已知实体
    entities := extractEntities(query)

    // 使用图知识丰富响应
    context := getGraphContext(entities)

    // 生成增强上下文的响应
    return generateAnswer(query, context)
}
```

## 最佳实践

### 1. 数据建模

**良好实践：**
- 使用一致的实体 ID（小写，无空格）
- 标准化关系类型（使用大写）
- 包含搜索的基本属性
- 规划实体层次结构

```go
// 良好：一致的 ID 和类型
entity := &rag.Entity{
    ID:   "apple_inc",
    Name: "苹果公司",
    Type: "ORGANIZATION",
    Properties: map[string]any{
        "industry": "科技",
        "founded": "1976",
        "ticker": "AAPL",
    },
}

// 良好：清晰的关系类型
relationship := &rag.Relationship{
    ID:     "apple_founded_by_jobs",
    Source: "steve_jobs",
    Target: "apple_inc",
    Type:   "FOUNDED_BY",
}
```

### 2. 查询优化

**性能提示：**
- 在 WHERE 子句中使用特定过滤器
- 适当时限制结果（LIMIT）
- 为频繁查询的属性添加索引
- 缓存复杂查询结果

```go
// 优化的查询
cypherQuery := `
    MATCH (n:PERSON {company: 'Google'})
    RETURN n.name, n.role
    LIMIT 50
`

// 避免：返回所有节点
cypherQuery := "MATCH (n) RETURN n"  // 对于大图很慢
```

### 3. 错误处理

```go
result, err := client.Do(ctx, "GRAPH.QUERY", graphName, query).Result()
if err != nil {
    log.Printf("查询失败: %v", err)
    return nil
}

// 验证结果
if r, ok := result.([]interface{}); ok && len(r) > 1 {
    // 处理结果
}
```

## 故障排除

### 常见问题

1. **连接错误**：
   ```bash
   # 测试 FalkorDB 连接
   redis-cli -p 6379 GRAPH.QUERY test "RETURN 1"
   ```

2. **查询语法错误**：
   ```go
   // 先测试简单查询
   simpleQuery := "MATCH (n) RETURN count(n)"
   ```

3. **找不到数据**：
   ```go
   // 检查存在哪些实体
   countQuery := "MATCH (n) RETURN labels(n), count(n)"
   ```

### 调试模式

启用详细日志：

```go
// 记录所有查询
fmt.Printf("执行查询: %s\n", cypherQuery)

// 记录查询结果
fmt.Printf("查询结果: %+v\n", result)

// 记录处理时间
startTime := time.Now()
// ... 执行查询 ...
duration := time.Since(startTime)
fmt.Printf("查询完成于: %v\n", duration)
```

## 扩展考虑

### 对于大型知识图谱

1. **批处理操作**：批量添加实体和关系
2. **连接池化**：使用 Redis 连接池
3. **查询优化**：添加索引并优化 Cypher 查询
4. **内存管理**：监控 Redis 内存使用

### 高可用性

1. **复制**：使用 Redis 复制实现容错
2. **备份**：定期备份图数据
3. **监控**：跟踪查询性能和错误率
4. **负载均衡**：在多个实例间分配查询

## 扩展

### 1. 时间关系

跟踪随时间变化的关系：

```go
relationship := &rag.Relationship{
    ID:     "john_worked_at_google_2023",
    Source: "john_smith",
    Target: "google",
    Type:   "WORKED_AT",
    Properties: map[string]any{
        "start_date": "2023-01-01",
        "end_date":   "2023-12-31",
        "position":   "高级工程师",
    },
}
```

### 2. 加权关系

为关系添加权重以进行排序：

```go
relationship := &rag.Relationship{
    Type: "PARTNER",
    Weight: 0.8, // 关系强度
    Properties: map[string]any{
        "collaboration_count": 15,
        "project_success_rate": 0.9,
    },
}
```

### 3. 基于事件的更新

基于事件自动更新知识图谱：

```go
func handleEmployeeEvent(event EmployeeEvent) {
    switch event.Type {
    case "HIRE":
        addEmployeeRelationship(event.Employee, event.Company, "WORKS_AT")
    case "PROMOTION":
        updateEmployeeRole(event.Employee, event.NewRole)
    case "TRANSFER":
        updateEmployeeCompany(event.Employee, event.NewCompany)
    }
}
```

## 贡献

欢迎为改进简单的 FalkorDB 示例做出贡献！请：

1. Fork 仓库
2. 创建功能分支
3. 添加您的改进
4. 包含测试
5. 提交 Pull Request

## 许可证

此示例是 LangGraphGo 项目的一部分。有关许可证信息，请参阅主仓库。