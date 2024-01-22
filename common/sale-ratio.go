package common

import "encoding/json"

var SaleRatio = map[string]float64{}

func SaleRatio2JSONString() string {
	jsonBytes, err := json.Marshal(SaleRatio)
	if err != nil {
		SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateSaleRatioByJSONString(jsonStr string) error {
	SaleRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &SaleRatio)
}

func GetSaleRatio(name string) float64 {
	ratio, ok := SaleRatio[name]
	if !ok {
		return 0
	}
	return ratio
}
