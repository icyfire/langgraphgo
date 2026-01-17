# 基于 memU 的智能代理示例

本示例展示了如何将 [memU](https://memu.so) 与 LangGraphGo 集成，实现在 AI 代理中进行高级的持久化内存管理。

## 概述

一个简单的聊天代理，利用 memU 在多轮对话中记住用户信息和偏好。与基础的内存存储不同，memU 提供了一个基于云的持久化内存，能够自动提取和组织信息。

## 关键特性

- **持久化内存**：内存可以在不同的会话、甚至应用程序的不同运行之间持久存在。
- **AI 驱动的提取**：memU 自动从对话中识别并提取重要信息（如姓名和偏好）。
- **上下文感知检索**：代理根据当前用户输入从 memU 检索相关内存，提供个性化响应。
- **分层组织**：信息被组织成类别（Category）和项（Item），便于管理。

## 前提条件

1. 从 [memu.so](https://memu.so) 获取 API 密钥。
2. 设置 `MEMU_API_KEY` 环境变量。

```bash
export MEMU_API_KEY='your-api-key'
```

## 运行示例

```bash
cd examples/memu_agent
go run main.go
```

本示例将运行一个模拟对话：
1. Alice 介绍了自己和她的偏好。
2. 代理将其存储在 memU 中。
3. 代理在随后的对话中回忆起 Alice 的姓名和最喜欢的饮料。
4. 最后，显示内存统计信息。

## 工作原理

该示例遵循以下模式：

1. **初始化 memU 客户端**：
   ```go
   memClient, err := memu.NewClient(memu.Config{
       APIKey: os.Getenv("MEMU_API_KEY"),
       UserID: "demo-user",
       RetrieveMethod: "rag",
   })
   ```

2. **添加消息**：
   每条用户和助手消息都会添加到 memU。
   ```go
   msg := memory.NewMessage("user", userMsg)
   memClient.AddMessage(ctx, msg)
   ```

3. **检索上下文**：
   在生成响应之前，代理向 memU 请求相关上下文。
   ```go
   memories, err := memClient.GetContext(ctx, userMsg)
   ```

4. **生成个性化响应**：
   代理利用检索到的内存来构建了解用户的响应。

## 相关示例

- [memory_agent](../memory_agent/) - 使用各种策略的通用内存代理
- [memory_chatbot](../memory_chatbot/) - 带内存的基础聊天机器人
- [memory_basic](../memory_basic/) - 简单内存使用
