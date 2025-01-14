package ali

import "github.com/songquanpeng/one-api/relay/billing/ratio"

// https://help.aliyun.com/zh/model-studio/getting-started/models
// https://help.aliyun.com/zh/dashscope/developer-reference/tongyi-thousand-questions-metering-and-billing
var RatioMap = map[string]ratio.Ratio{
	"qwen-long":                   {Input: 0.0005 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"qwen-turbo":                  {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"qwen-turbo-latest":           {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"qwen-turbo-2024-09-19":       {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"qwen-turbo-0919":             {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"qwen-turbo-2024-06-24":       {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"qwen-turbo-0624":             {Input: 0.0003 * ratio.RMB, Output: 0.0006 * ratio.RMB},
	"qwen-turbo-2024-02-06":       {Input: 0.002 * ratio.RMB, Output: 0.006 * ratio.RMB},
	"qwen-turbo-0206":             {Input: 0.002 * ratio.RMB, Output: 0.006 * ratio.RMB},
	"qwen-plus":                   {Input: 0.0008 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"qwen-plus-latest":            {Input: 0.0008 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"qwen-plus-2024-09-19":        {Input: 0.0008 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"qwen-plus-0919":              {Input: 0.0008 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"qwen-plus-2024-08-06":        {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-plus-0806":              {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-plus-2024-07-23":        {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-plus-0723":              {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-plus-2024-06-24":        {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-plus-0624":              {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-plus-2024-02-06":        {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-plus-0206":              {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-max":                    {Input: 0.02 * ratio.RMB, Output: 0.06 * ratio.RMB},
	"qwen-max-latest":             {Input: 0.02 * ratio.RMB, Output: 0.06 * ratio.RMB},
	"qwen-max-2024-09-19":         {Input: 0.02 * ratio.RMB, Output: 0.06 * ratio.RMB},
	"qwen-max-0919":               {Input: 0.02 * ratio.RMB, Output: 0.06 * ratio.RMB},
	"qwen-max-2024-04-28":         {Input: 0.04 * ratio.RMB, Output: 0.12 * ratio.RMB},
	"qwen-max-0428":               {Input: 0.04 * ratio.RMB, Output: 0.12 * ratio.RMB},
	"qwen-max-2024-04-03":         {Input: 0.04 * ratio.RMB, Output: 0.12 * ratio.RMB},
	"qwen-max-0403":               {Input: 0.04 * ratio.RMB, Output: 0.12 * ratio.RMB},
	"qwen-max-2024-01-07":         {Input: 0.04 * ratio.RMB, Output: 0.12 * ratio.RMB},
	"qwen-max-0107":               {Input: 0.04 * ratio.RMB, Output: 0.12 * ratio.RMB},
	"qwen-vl-max":                 {Input: 0.003 * ratio.RMB, Output: 0.009 * ratio.RMB},
	"qwen-vl-max-latest":          {Input: 0.003 * ratio.RMB, Output: 0.009 * ratio.RMB},
	"qwen-vl-max-2024-12-30":      {Input: 0.003 * ratio.RMB, Output: 0.009 * ratio.RMB},
	"qwen-vl-max-1230":            {Input: 0.003 * ratio.RMB, Output: 0.009 * ratio.RMB},
	"qwen-vl-max-2024-11-19":      {Input: 0.003 * ratio.RMB, Output: 0.009 * ratio.RMB},
	"qwen-vl-max-1119":            {Input: 0.003 * ratio.RMB, Output: 0.009 * ratio.RMB},
	"qwen-vl-max-2024-10-30":      {Input: 0.02 * ratio.RMB, Output: 0.02 * ratio.RMB},
	"qwen-vl-max-1030":            {Input: 0.02 * ratio.RMB, Output: 0.02 * ratio.RMB},
	"qwen-vl-max-2024-08-09":      {Input: 0.02 * ratio.RMB, Output: 0.02 * ratio.RMB},
	"qwen-vl-max-0809":            {Input: 0.02 * ratio.RMB, Output: 0.02 * ratio.RMB},
	"qwen-vl-max-2024-02-01":      {Input: 0.02 * ratio.RMB, Output: 0.02 * ratio.RMB},
	"qwen-vl-max-0201":            {Input: 0.02 * ratio.RMB, Output: 0.02 * ratio.RMB},
	"qwen-vl-plus":                {Input: 0.0015 * ratio.RMB, Output: 0.0045 * ratio.RMB},
	"qwen-vl-plus-latest":         {Input: 0.0015 * ratio.RMB, Output: 0.0045 * ratio.RMB},
	"qwen-vl-plus-2024-08-09":     {Input: 0.0015 * ratio.RMB, Output: 0.0045 * ratio.RMB},
	"qwen-vl-plus-0809":           {Input: 0.0015 * ratio.RMB, Output: 0.0045 * ratio.RMB},
	"qwen-vl-plus-2023-12-01":     {Input: 0.008 * ratio.RMB, Output: 0.008 * ratio.RMB},
	"qwen-vl-ocr":                 {Input: 0.005 * ratio.RMB, Output: 0.005 * ratio.RMB},
	"qwen-vl-ocr-latest":          {Input: 0.005 * ratio.RMB, Output: 0.005 * ratio.RMB},
	"qwen-vl-ocr-2024-10-28":      {Input: 0.005 * ratio.RMB, Output: 0.005 * ratio.RMB},
	"qwen-audio-turbo":            {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen-audio-turbo-latest":     {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen-audio-turbo-2024-12-04": {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen-audio-turbo-1204":       {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen-audio-turbo-2024-08-07": {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen-audio-turbo-0807":       {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen-math-plus":              {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-math-plus-latest":       {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-math-plus-2024-09-19":   {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-math-plus-0919":         {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-math-plus-2024-08-16":   {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-math-plus-0816":         {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-math-turbo":             {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-math-turbo-latest":      {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-math-turbo-2024-09-19":  {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-math-turbo-0919":        {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen-coder-plus":             {Input: 0.0035 * ratio.RMB, Output: 0.007 * ratio.RMB},
	"qwen-coder-plus-latest":      {Input: 0.0035 * ratio.RMB, Output: 0.007 * ratio.RMB},
	"qwen-coder-plus-2024-11-06":  {Input: 0.0035 * ratio.RMB, Output: 0.007 * ratio.RMB},
	"qwen-coder-plus-1106":        {Input: 0.0035 * ratio.RMB, Output: 0.007 * ratio.RMB},
	"qwen-coder-turbo":            {Input: 0.002 * ratio.RMB, Output: 0.006 * ratio.RMB},
	"qwen-coder-turbo-latest":     {Input: 0.002 * ratio.RMB, Output: 0.006 * ratio.RMB},
	"qwen-coder-turbo-2024-09-19": {Input: 0.002 * ratio.RMB, Output: 0.006 * ratio.RMB},
	"qwen-coder-turbo-0919":       {Input: 0.002 * ratio.RMB, Output: 0.006 * ratio.RMB},
	"qwq-32b-preview":             {Input: 0.0035 * ratio.RMB, Output: 0.007 * ratio.RMB},
	"qwen2.5-72b-instruct":        {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen2.5-32b-instruct":        {Input: 0.0035 * ratio.RMB, Output: 0.007 * ratio.RMB},
	"qwen2.5-14b-instruct":        {Input: 0.002 * ratio.RMB, Output: 0.006 * ratio.RMB},
	"qwen2.5-7b-instruct":         {Input: 0.001 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"qwen2.5-3b-instruct":         {Input: 0.1, Output: 0.1}, // 限时免费
	"qwen2.5-1.5b-instruct":       {Input: 0.1, Output: 0.1}, // 限时免费
	"qwen2.5-0.5b-instruct":       {Input: 0.1, Output: 0.1}, // 限时免费
	"qwen2-72b-instruct":          {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen2-57b-a14b-instruct":     {Input: 0.0035 * ratio.RMB, Output: 0.007 * ratio.RMB},
	"qwen2-7b-instruct":           {Input: 0.001 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"qwen2-1.5b-instruct":         {Input: 0.1, Output: 0.1}, // 限时免费
	"qwen2-0.5b-instruct":         {Input: 0.1, Output: 0.1}, // 限时免费
	"qwen1.5-110b-chat":           {Input: 0.007 * ratio.RMB, Output: 0.014 * ratio.RMB},
	"qwen1.5-72b-chat":            {Input: 0.005 * ratio.RMB, Output: 0.01 * ratio.RMB},
	"qwen1.5-32b-chat":            {Input: 0.0035 * ratio.RMB, Output: 0.007 * ratio.RMB},
	"qwen1.5-14b-chat":            {Input: 0.002 * ratio.RMB, Output: 0.004 * ratio.RMB},
	"qwen1.5-7b-chat":             {Input: 0.001 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"qwen1.5-1.8b-chat":           {Input: 0.1, Output: 0.1}, // 限时免费
	"qwen1.5-0.5b-chat":           {Input: 0.1, Output: 0.1}, // 限时免费
	"qwen-72b-chat":               {Input: 0.02 * ratio.RMB, Output: 0.02 * ratio.RMB},
	"qwen-14b-chat":               {Input: 0.008 * ratio.RMB, Output: 0.008 * ratio.RMB},
	"qwen-7b-chat":                {Input: 0.006 * ratio.RMB, Output: 0.006 * ratio.RMB},
	"qwen-1.8b-chat":              {Input: 0.1, Output: 0.1}, // 限时免费
	"qwen-1.8b-longcontext-chat":  {Input: 0.1, Output: 0.1}, // 限时免费（需申请）
	"qwen2-vl-72b-instruct":       {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen2-vl-7b-instruct":        {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen2-vl-2b-instruct":        {Input: 0.1, Output: 0.1}, // 限时免费
	"qwen-vl-v1":                  {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen-vl-chat-v1":             {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen2-audio-instruct":        {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen-audio-chat":             {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen2.5-math-72b-instruct":   {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen2.5-math-7b-instruct":    {Input: 0.001 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"qwen2.5-math-1.5b-instruct":  {Input: 0.1, Output: 0.1}, // 限时免费
	"qwen2-math-72b-instruct":     {Input: 0.004 * ratio.RMB, Output: 0.012 * ratio.RMB},
	"qwen2-math-7b-instruct":      {Input: 0.001 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"qwen2-math-1.5b-instruct":    {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen2.5-coder-32b-instruct":  {Input: 0.0035 * ratio.RMB, Output: 0.007 * ratio.RMB},
	"qwen2.5-coder-14b-instruct":  {Input: 0.002 * ratio.RMB, Output: 0.006 * ratio.RMB},
	"qwen2.5-coder-7b-instruct":   {Input: 0.001 * ratio.RMB, Output: 0.002 * ratio.RMB},
	"qwen2.5-coder-3b-instruct":   {Input: 0.1, Output: 0.1}, // 限时免费
	"qwen2.5-coder-1.5b-instruct": {Input: 0.1, Output: 0.1}, // 目前仅供免费体验。免费额度用完后不可调用，敬请关注后续动态。
	"qwen2.5-coder-0.5b-instruct": {Input: 0.1, Output: 0.1}, // 限时免费
	"text-embedding-v3":           {Input: 0.0007 * ratio.RMB, Output: 0},
	"text-embedding-v2":           {Input: 0.0007 * ratio.RMB, Output: 0},
	"text-embedding-v1":           {Input: 0.0007 * ratio.RMB, Output: 0},
	"text-embedding-async-v2":     {Input: 0.0007 * ratio.RMB, Output: 0},
	"text-embedding-async-v1":     {Input: 0.0007 * ratio.RMB, Output: 0},
	"ali-stable-diffusion-xl":     {Input: 8.00 * ratio.RMB, Output: 0},
	"ali-stable-diffusion-v1.5":   {Input: 8.00 * ratio.RMB, Output: 0},
	"wanx-v1":                     {Input: 8.00 * ratio.RMB, Output: 0},
}
