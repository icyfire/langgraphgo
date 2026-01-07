package main

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"time"
)

// HotelBookingParams 酒店预订参数
type HotelBookingParams struct {
	CheckIn         string   `json:"check_in"`
	CheckOut        string   `json:"check_out"`
	Guests          int      `json:"guests"`
	RoomType        string   `json:"room_type"`
	Breakfast       bool     `json:"breakfast"`
	Parking         bool     `json:"parking"`
	View            string   `json:"view"`
	MaxPrice        float64  `json:"max_price"`
	SpecialRequests []string `json:"special_requests"`
}

// HotelBookingTool 酒店预订工具
type HotelBookingTool struct{}

func (t HotelBookingTool) Name() string {
	return "hotel_booking"
}

func (t HotelBookingTool) Description() string {
	return `预订酒店房间，支持多种选项配置。

	参数说明：
	- check_in: 入住日期（格式：YYYY-MM-DD）
	- check_out: 退房日期（格式：YYYY-MM-DD）
	- guests: 客人数量（1-10人）
	- room_type: 房间类型（标准间standard、豪华间deluxe、套房suite、总统套房penthouse）
	- breakfast: 是否包含早餐（true/false）
	- parking: 是否需要停车位（true/false）
	- view: 房间景观偏好（无none、城市景观city、海景ocean、山景mountain）
	- max_price: 每晚最高价格（例如：200.00）
	- special_requests: 特殊要求数组（可选）`
}

func (t HotelBookingTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"check_in": map[string]any{
				"type":        "string",
				"description": "入住日期，格式为YYYY-MM-DD",
				"format":      "date",
			},
			"check_out": map[string]any{
				"type":        "string",
				"description": "退房日期，格式为YYYY-MM-DD",
				"format":      "date",
			},
			"guests": map[string]any{
				"type":        "integer",
				"description": "客人数量（1-10人）",
				"minimum":     1,
				"maximum":     10,
			},
			"room_type": map[string]any{
				"type":        "string",
				"description": "房间类型",
				"enum":        []string{"standard", "deluxe", "suite", "penthouse"},
			},
			"breakfast": map[string]any{
				"type":        "boolean",
				"description": "是否包含早餐",
			},
			"parking": map[string]any{
				"type":        "boolean",
				"description": "是否需要停车位",
			},
			"view": map[string]any{
				"type":        "string",
				"description": "房间景观偏好",
				"enum":        []string{"none", "city", "ocean", "mountain"},
			},
			"max_price": map[string]any{
				"type":        "number",
				"description": "每晚最高价格",
			},
			"special_requests": map[string]any{
				"type":        "array",
				"description": "特殊要求列表",
				"items": map[string]any{
					"type": "string",
				},
			},
		},
		"required": []string{"check_in", "check_out", "guests", "room_type"},
	}
}

func (t HotelBookingTool) Call(ctx context.Context, input string) (string, error) {
	var params HotelBookingParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("输入参数无效: %w", err)
	}

	// 验证必填字段
	if params.CheckIn == "" {
		return "", fmt.Errorf("入住日期是必填项")
	}
	if params.CheckOut == "" {
		return "", fmt.Errorf("退房日期是必填项")
	}
	if params.Guests < 1 || params.Guests > 10 {
		return "", fmt.Errorf("客人数量必须在1到10之间")
	}

	// 计算入住天数
	checkIn, _ := time.Parse("2006-01-02", params.CheckIn)
	checkOut, _ := time.Parse("2006-01-02", params.CheckOut)
	nights := int(checkOut.Sub(checkIn).Hours() / 24)

	if nights <= 0 {
		return "", fmt.Errorf("退房日期必须晚于入住日期")
	}

	// 计算基础价格
	basePrice := 100.0
	switch params.RoomType {
	case "deluxe":
		basePrice = 180.0
	case "suite":
		basePrice = 300.0
	case "penthouse":
		basePrice = 500.0
	}

	// 添加额外服务费用
	if params.Breakfast {
		basePrice += 25.0
	}
	if params.Parking {
		basePrice += 15.0
	}
	if params.View == "ocean" {
		basePrice += 50.0
	} else if params.View == "city" {
		basePrice += 30.0
	} else if params.View == "mountain" {
		basePrice += 35.0
	}

	// 检查最高价格限制
	if params.MaxPrice > 0 && basePrice > params.MaxPrice {
		return "", fmt.Errorf("房间价格（%.2f）超过了最高价格限制（%.2f）", basePrice, params.MaxPrice)
	}

	totalPrice := basePrice * float64(nights)

	result := map[string]any{
		"booking_id":       fmt.Sprintf("HTL-%d", time.Now().Unix()),
		"check_in":         params.CheckIn,
		"check_out":        params.CheckOut,
		"nights":           nights,
		"guests":           params.Guests,
		"room_type":        params.RoomType,
		"breakfast":        params.Breakfast,
		"parking":          params.Parking,
		"view":             params.View,
		"price_per_night":  basePrice,
		"total_price":      totalPrice,
		"special_requests": params.SpecialRequests,
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// MortgageCalculationParams 房贷计算参数
type MortgageCalculationParams struct {
	Principal    float64          `json:"principal"`
	InterestRate float64          `json:"interest_rate"`
	Years        int              `json:"years"`
	DownPayment  float64          `json:"down_payment"`
	PropertyTax  float64          `json:"property_tax"`
	Insurance    float64          `json:"insurance"`
	ExtraPayment ExtraPaymentInfo `json:"extra_payment"`
}

// ExtraPaymentInfo 额外还款信息（嵌套对象）
type ExtraPaymentInfo struct {
	Enabled   bool    `json:"enabled"`
	Amount    float64 `json:"amount"`
	Frequency string  `json:"frequency"` // monthly, yearly, onetime
	StartYear int     `json:"start_year"`
}

// MortgageCalculatorTool 房贷计算器工具
type MortgageCalculatorTool struct{}

func (t MortgageCalculatorTool) Name() string {
	return "mortgage_calculator"
}

func (t MortgageCalculatorTool) Description() string {
	return `计算房贷月供，支持额外还款、税费和保险。

	参数说明：
	- principal: 贷款本金（例如：300000）
	- interest_rate: 年利率百分比（例如：6.5表示6.5%）
	- years: 贷款年限（例如：30）
	- down_payment: 首付金额（例如：60000）
	- property_tax: 年房产税（例如：3000）
	- insurance: 年保险费（例如：1200）
	- extra_payment: 额外还款配置（嵌套对象）：
		- enabled: 是否启用额外还款
		- amount: 额外还款金额
		- frequency: 还款频率（monthly按月、yearly按年、onetime一次性）
		- start_year: 开始额外还款的年份`
}

func (t MortgageCalculatorTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"principal": map[string]any{
				"type":        "number",
				"description": "贷款本金金额",
			},
			"interest_rate": map[string]any{
				"type":        "number",
				"description": "年利率百分比（例如：6.5表示6.5%）",
			},
			"years": map[string]any{
				"type":        "integer",
				"description": "贷款年限",
				"minimum":     1,
				"maximum":     50,
			},
			"down_payment": map[string]any{
				"type":        "number",
				"description": "首付金额",
			},
			"property_tax": map[string]any{
				"type":        "number",
				"description": "年度房产税金额",
			},
			"insurance": map[string]any{
				"type":        "number",
				"description": "年度保险金额",
			},
			"extra_payment": map[string]any{
				"type":        "object",
				"description": "额外还款配置",
				"properties": map[string]any{
					"enabled": map[string]any{
						"type":        "boolean",
						"description": "是否启用额外还款",
					},
					"amount": map[string]any{
						"type":        "number",
						"description": "额外还款金额",
					},
					"frequency": map[string]any{
						"type":        "string",
						"description": "还款频率",
						"enum":        []string{"monthly", "yearly", "onetime"},
					},
					"start_year": map[string]any{
						"type":        "integer",
						"description": "开始额外还款的年份",
					},
				},
			},
		},
		"required": []string{"principal", "interest_rate", "years"},
	}
}

func (t MortgageCalculatorTool) Call(ctx context.Context, input string) (string, error) {
	var params MortgageCalculationParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("输入参数无效: %w", err)
	}

	// 验证必填字段
	if params.Principal <= 0 {
		return "", fmt.Errorf("贷款本金必须大于0")
	}
	if params.InterestRate <= 0 {
		return "", fmt.Errorf("利率必须大于0")
	}
	if params.Years <= 0 {
		return "", fmt.Errorf("贷款年限必须大于0")
	}

	// 计算实际贷款金额（扣除首付）
	loanAmount := params.Principal - params.DownPayment
	if loanAmount <= 0 {
		return "", fmt.Errorf("首付金额不能超过贷款本金")
	}

	// 计算月供
	monthlyRate := params.InterestRate / 100 / 12
	numPayments := params.Years * 12

	monthlyPayment := loanAmount * (monthlyRate * math.Pow(1+monthlyRate, float64(numPayments))) /
		(math.Pow(1+monthlyRate, float64(numPayments)) - 1)

	// 添加月度房产税和保险
	monthlyTax := params.PropertyTax / 12
	monthlyInsurance := params.Insurance / 12
	totalMonthlyPayment := monthlyPayment + monthlyTax + monthlyInsurance

	// 计算总还款额和利息
	totalPayment := monthlyPayment * float64(numPayments)
	totalInterest := totalPayment - loanAmount

	// 计算额外还款的影响
	extraPaymentResult := map[string]any{}
	if params.ExtraPayment.Enabled && params.ExtraPayment.Amount > 0 {
		savedInterest := params.ExtraPayment.Amount * 100 // 简化计算
		paidOffMonths := numPayments - int(params.ExtraPayment.Amount/2)

		extraPaymentResult = map[string]any{
			"extra_payment_enabled":    true,
			"extra_payment_amount":     params.ExtraPayment.Amount,
			"extra_payment_frequency":  params.ExtraPayment.Frequency,
			"estimated_interest_saved": savedInterest,
			"paid_off_months_early":    numPayments - paidOffMonths,
			"new_payoff_date":          time.Now().AddDate(0, paidOffMonths, 0).Format("2006-01-02"),
		}
	} else {
		extraPaymentResult = map[string]any{
			"extra_payment_enabled": false,
		}
	}

	result := map[string]any{
		"loan_amount":                loanAmount,
		"down_payment":               params.DownPayment,
		"interest_rate":              params.InterestRate,
		"loan_term_years":            params.Years,
		"monthly_principal_interest": monthlyPayment,
		"monthly_property_tax":       monthlyTax,
		"monthly_insurance":          monthlyInsurance,
		"total_monthly_payment":      totalMonthlyPayment,
		"total_payment":              totalPayment + (monthlyTax+monthlyInsurance)*float64(numPayments),
		"total_interest":             totalInterest,
		"payoff_date":                time.Now().AddDate(params.Years, 0, 0).Format("2006-01-02"),
		"extra_payment_info":         extraPaymentResult,
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}

// BatchOperationItem 批量操作中的单个项目
type BatchOperationItem struct {
	ID       string                 `json:"id"`
	Action   string                 `json:"action"` // create, update, delete
	Quantity int                    `json:"quantity"`
	Price    float64                `json:"price"`
	Metadata map[string]interface{} `json:"metadata"`
}

// BatchOperationParams 批量操作参数
type BatchOperationParams struct {
	Operation string               `json:"operation"`
	Items     []BatchOperationItem `json:"items"`
	DryRun    bool                 `json:"dry_run"`
	Priority  string               `json:"priority"`
}

// BatchOperationTool 批量操作工具
type BatchOperationTool struct{}

func (t BatchOperationTool) Name() string {
	return "batch_operation"
}

func (t BatchOperationTool) Description() string {
	return `对多个项目执行批量操作。

	参数说明：
	- operation: 操作类型（process处理、validate验证、export导出）
	- items: 项目数组，每个项目包含：
		- id: 唯一标识符
		- action: 要执行的操作（create创建、update更新、delete删除）
		- quantity: 数量
		- price: 单价
		- metadata: 额外的键值对（嵌套对象）
	- dry_run: 如果为true，仅验证不执行（true/false）
	- priority: 操作优先级（low低、normal普通、high高、urgent紧急）`
}

func (t BatchOperationTool) Schema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"operation": map[string]any{
				"type":        "string",
				"description": "批量操作类型",
				"enum":        []string{"process", "validate", "export"},
			},
			"items": map[string]any{
				"type":        "array",
				"description": "要处理的项目数组",
				"items": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"id": map[string]any{
							"type":        "string",
							"description": "项目唯一标识符",
						},
						"action": map[string]any{
							"type":        "string",
							"description": "要执行的操作",
							"enum":        []string{"create", "update", "delete"},
						},
						"quantity": map[string]any{
							"type":        "integer",
							"description": "项目数量",
							"minimum":     0,
						},
						"price": map[string]any{
							"type":        "number",
							"description": "单价",
							"minimum":     0,
						},
						"metadata": map[string]any{
							"type":        "object",
							"description": "额外的元数据（键值对）",
						},
					},
					"required": []string{"id", "action"},
				},
			},
			"dry_run": map[string]any{
				"type":        "boolean",
				"description": "如果为true，仅验证而不执行",
			},
			"priority": map[string]any{
				"type":        "string",
				"description": "操作优先级",
				"enum":        []string{"low", "normal", "high", "urgent"},
			},
		},
		"required": []string{"operation", "items"},
	}
}

func (t BatchOperationTool) Call(ctx context.Context, input string) (string, error) {
	var params BatchOperationParams
	if err := json.Unmarshal([]byte(input), &params); err != nil {
		return "", fmt.Errorf("输入参数无效: %w", err)
	}

	// 验证必填字段
	if params.Operation == "" {
		return "", fmt.Errorf("操作类型是必填项")
	}
	if len(params.Items) == 0 {
		return "", fmt.Errorf("至少需要一个项目")
	}

	// 处理每个项目
	results := make([]map[string]any, 0, len(params.Items))
	totalValue := 0.0
	summary := map[string]int{
		"create": 0,
		"update": 0,
		"delete": 0,
		"failed": 0,
	}

	for _, item := range params.Items {
		itemResult := map[string]any{
			"id":     item.ID,
			"action": item.Action,
			"status": "pending",
		}

		if params.DryRun {
			itemResult["status"] = "validated (dry run)"
			itemResult["message"] = "将会处理此项目"
		} else {
			// 模拟处理
			if item.Action == "create" || item.Action == "update" {
				itemValue := float64(item.Quantity) * item.Price
				totalValue += itemValue
				itemResult["value"] = itemValue
				itemResult["status"] = "success"
				itemResult["message"] = fmt.Sprintf("已处理 %d 件，单价 %.2f", item.Quantity, item.Price)
			}
			summary[item.Action]++
		}

		if item.Metadata != nil {
			itemResult["metadata_processed"] = len(item.Metadata)
		}

		results = append(results, itemResult)
	}

	result := map[string]any{
		"operation":   params.Operation,
		"dry_run":     params.DryRun,
		"priority":    params.Priority,
		"total_items": len(params.Items),
		"summary":     summary,
		"total_value": totalValue,
		"timestamp":   time.Now().Format(time.RFC3339),
		"results":     results,
	}

	jsonResult, _ := json.MarshalIndent(result, "", "  ")
	return string(jsonResult), nil
}
