package baidu

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/hlrk4akp7
var RatioMap = map[string]ratio.Ratio{
	"ERNIE-4.0-Turbo-128K":                {Input: 0.02 * ratio.RMB, Output: 0.06 * ratio.RMB},
	"ERNIE-4.0-Turbo-8K":                  {Input: 0.02 * ratio.RMB, Output: 0.06 * ratio.RMB},
	"ERNIE-4.0-Turbo-8K-Preview":          {Input: 0.02 * ratio.RMB, Output: 0.06 * ratio.RMB},
	"ERNIE-4.0-Turbo-8K-0628":             {Input: 0.02 * ratio.RMB, Output: 0.06 * ratio.RMB},
	"ERNIE-4.0-8K":                        {Input: 0.03 * ratio.RMB, Output: 0.09 * ratio.RMB},
	"ERNIE-4.0-8K-0613":                   {Input: 0.03 * ratio.RMB, Output: 0.09 * ratio.RMB},
	"ERNIE-4.0-8K-Latest":                 {Input: 0.03 * ratio.RMB, Output: 0.09 * ratio.RMB},
	"ERNIE-4.0-8K-Preview":                {Input: 0.03 * ratio.RMB, Output: 0.09 * ratio.RMB},
	"ERNIE-3.5-128K":                      {Input: 0.0008 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"ERNIE-3.5-8K":                        {Input: 0.0008 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"ERNIE-3.5-8K-0701":                   {Input: 0.0008 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"ERNIE-3.5-8K-Preview":                {Input: 0.0008 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"ERNIE-3.5-8K-0613":                   {Input: 0.0008 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"ERNIE-Speed-Pro-128K":                {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"ERNIE-Novel-8K":                      {Input: 0.04 * ratio.RMB, Output: 0.12 * ratio.RMB},
	"ERNIE-Speed-128K":                    {Input: 0.1, Output: 0.1}, // 免费
	"ERNIE-Speed-8K":                      {Input: 0.1, Output: 0.1}, // 免费
	"ERNIE-Lite-8K":                       {Input: 0.1, Output: 0.1}, // 免费
	"ERNIE-Tiny-8K":                       {Input: 0.1, Output: 0.1}, // 免费
	"ERNIE-Functions-8K":                  {Input: 0.004 * ratio.RMB, Output: 0.008 * ratio.RMB},
	"ERNIE-Character-8K":                  {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"ERNIE-Character-Fiction-8K":          {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"ERNIE-Character-Fiction-8K-Preview	": {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"ERNIE-Lite-Pro-128K":                 {Input: 0.0002 * ratio.RMB, Output: 0.0004 * ratio.RMB},
	"Qianfan-Agent-Speed-8K":              {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"Qianfan-Agent-Lite-8K":               {Input: 0.0002 * ratio.RMB, Output: 0.0004 * ratio.RMB},
	"BLOOMZ-7B":                           {Input: 0.004 * ratio.RMB, Output: 0.004 * ratio.RMB},
	"Embedding-V1":                        {Input: 0.0005 * ratio.RMB, Output: 0},
	"bge-large-zh":                        {Input: 0.0005 * ratio.RMB, Output: 0},
	"bge-large-en":                        {Input: 0.0005 * ratio.RMB, Output: 0},
	"tao-8k":                              {Input: 0.0005 * ratio.RMB, Output: 0},
}
