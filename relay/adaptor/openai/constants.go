package openai

import "github.com/songquanpeng/one-api/relay/billing/ratio"

var RatioMap = map[string]ratio.Ratio{
	"gpt-4":                      {Input: 15, Output: 30},
	"gpt-4-0314":                 {Input: 15, Output: 30},
	"gpt-4-0613":                 {Input: 15, Output: 30},
	"gpt-4-32k":                  {Input: 30, Output: 60},
	"gpt-4-32k-0314":             {Input: 30, Output: 60},
	"gpt-4-32k-0613":             {Input: 30, Output: 60},
	"gpt-4-1106-preview":         {Input: 5, Output: 15},
	"gpt-4-0125-preview":         {Input: 5, Output: 15},
	"gpt-4-turbo-preview":        {Input: 5, Output: 15},               // $0.01 / 1K tokens
	"gpt-4-turbo":                {Input: 5, Output: 15},               // $0.01 / 1K tokens
	"gpt-4-turbo-2024-04-09":     {Input: 5, Output: 15},               // $0.01 / 1K tokens
	"gpt-4o":                     {Input: 1.25, Output: 5},             // $0.005 / 1K tokens
	"chatgpt-4o-latest":          {Input: 2.5, Output: 7.5},            // $0.005 / 1K tokens
	"gpt-4o-2024-05-13":          {Input: 2.5, Output: 7.5},            // $0.005 / 1K tokens
	"gpt-4o-2024-08-06":          {Input: 1.25, Output: 5},             // $0.0025 / 1K tokens
	"gpt-4o-2024-11-20":          {Input: 1.25, Output: 5},             // $0.0025 / 1K tokens
	"gpt-4o-mini":                {Input: 0.075, Output: 0.3},          // $0.00015 / 1K tokens
	"gpt-4o-mini-2024-07-18":     {Input: 0.075, Output: 0.3},          // $0.00015 / 1K tokens
	"gpt-4-vision-preview":       {Input: 5, Output: 15},               // $0.01 / 1K tokens
	"gpt-3.5-turbo":              {Input: 0.25, Output: 0.75},          // $0.0005 / 1K tokens
	"gpt-3.5-turbo-0301":         {Input: 0.75, Output: 1},             // $0.0015 / 1K tokens
	"gpt-3.5-turbo-0613":         {Input: 0.75, Output: 1},             // $0.0015 / 1K tokens
	"gpt-3.5-turbo-16k":          {Input: 1.5, Output: 2},              // $0.003 / 1K tokens
	"gpt-3.5-turbo-16k-0613":     {Input: 1.5, Output: 2},              // $0.003 / 1K tokens
	"gpt-3.5-turbo-instruct":     {Input: 0.75, Output: 1},             // $0.0015 / 1K tokens
	"gpt-3.5-turbo-1106":         {Input: 0.5, Output: 1},              // $0.001 / 1K tokens
	"gpt-3.5-turbo-0125":         {Input: 0.25, Output: 0.75},          // $0.0005 / 1K tokens
	"davinci-002":                {Input: 1, Output: 1},                // $0.002 / 1K tokens
	"babbage-002":                {Input: 0.2, Output: 0.2},            // $0.0004 / 1K tokens
	"text-ada-001":               {Input: 0.2, Output: 0.2},            // $0.0004 / 1K tokens
	"text-babbage-001":           {Input: 0.25, Output: 0.25},          // $0.0005 / 1K tokens
	"text-curie-001":             {Input: 1, Output: 1},                // $0.002 / 1K tokens
	"text-davinci-002":           {Input: 10, Output: 10},              // $0.02 / 1K tokens
	"text-davinci-003":           {Input: 10, Output: 10},              // $0.02 / 1K tokens
	"text-davinci-edit-001":      {Input: 10, Output: 10},              // $0.02 / 1K tokens
	"code-davinci-edit-001":      {Input: 10, Output: 10},              // $0.02 / 1K tokens
	"whisper-1":                  {Input: 1, Output: 1},                // $0.006 / minute -> $0.002 / 20 seconds -> $0.002 / 1K tokens -> 20 seconds / 1K tokens
	"tts-1":                      {Input: 7.5, Output: 7.5},            // $0.015 / 1K characters
	"tts-1-1106":                 {Input: 7.5, Output: 7.5},            // $0.015 / 1K characters
	"tts-1-hd":                   {Input: 15, Output: 15},              // $0.030 / 1K characters
	"tts-1-hd-1106":              {Input: 15, Output: 15},              // $0.030 / 1K characters
	"davinci":                    {Input: 10, Output: 10},              // $0.02 / 1K tokens
	"curie":                      {Input: 10, Output: 10},              // $0.02 / 1K tokens
	"babbage":                    {Input: 10, Output: 10},              // $0.02 / 1K tokens
	"ada":                        {Input: 10, Output: 10},              // $0.02 / 1K tokens
	"text-embedding-ada-002":     {Input: 0.05, Output: 0},             // $0.001 / 1K tokens
	"text-embedding-3-small":     {Input: 0.01, Output: 0},             // $0.0002 / 1K tokens
	"text-embedding-3-large":     {Input: 0.065, Output: 0},            // $0.0013 / 1K tokens
	"text-search-ada-doc-001":    {Input: 10, Output: 0},               // $0.02 / 1K tokens
	"text-moderation-stable":     {Input: 0.1, Output: 0},              // currently free to use
	"text-moderation-latest":     {Input: 0.1, Output: 0},              // currently free to use
	"omni-moderation-latest":     {Input: 0.1, Output: 0},              // currently free to use
	"omni-moderation-2024-09-26": {Input: 0.1, Output: 0},              // currently free to use
	"dall-e-2":                   {Input: 0.02 * ratio.USD, Output: 0}, // $0.016 - $0.020 / image
	"dall-e-3":                   {Input: 0.04 * ratio.USD, Output: 0}, // $0.040 - $0.120 / image
}
