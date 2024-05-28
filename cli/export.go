package cli

import (
	"encoding/json"
	"one-api/common/logger"
	"one-api/relay/relay_util"
	"os"
	"sort"
)

func ExportPrices() {
	prices := relay_util.GetPricesList("default")

	if len(prices) == 0 {
		logger.SysError("No prices found")
		return
	}

	// Sort prices by ChannelType
	sort.Slice(prices, func(i, j int) bool {
		if prices[i].ChannelType == prices[j].ChannelType {
			return prices[i].Model < prices[j].Model
		}
		return prices[i].ChannelType < prices[j].ChannelType
	})

	// 导出到当前目录下的 prices.json 文件
	file, err := os.Create("prices.json")
	if err != nil {
		logger.SysError("Failed to create file: " + err.Error())
		return
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(prices, "", "  ")
	if err != nil {
		logger.SysError("Failed to encode prices: " + err.Error())
		return
	}

	_, err = file.Write(jsonData)
	if err != nil {
		logger.SysError("Failed to write to file: " + err.Error())
		return
	}

	logger.SysLog("Prices exported to prices.json")
}
