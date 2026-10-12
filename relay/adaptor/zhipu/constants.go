package zhipu

// https://open.bigmodel.cn/pricing

var ModelList = []string{
	// GLM-5 / 4.7 / 4.6 系列（v4 接口）
	"glm-5.3", "glm-5.3-flash", "glm-5.3-flashx",
	"glm-5.2", "glm-5.1", "glm-5",
	"glm-4.7", "glm-4.7-flashx", "glm-4.7-flash",
	"glm-4.6",
	// GLM-4.5 系列（v4 接口）
	"glm-4.5", "glm-4.5-air", "glm-4.5-x", "glm-4.5-airx", "glm-4.5-flash",
	// 视觉理解（多模态）
	"glm-4.6v", "glm-4.6v-flashx", "glm-4.6v-flash",
	"glm-4.5v",
	// GLM-4 经典系列（v4 接口）
	"glm-zero-preview", "glm-4-plus", "glm-4-0520", "glm-4-airx",
	"glm-4-air", "glm-4-long", "glm-4-flashx", "glm-4-flash",
	"glm-4", "glm-3-turbo",
	"glm-4v-plus", "glm-4v", "glm-4v-flash",
	// 图像生成
	"cogview-4", "cogview-3-plus", "cogview-3", "cogview-3-flash",
	"cogviewx", "cogviewx-flash",
	// 角色 / 代码 / 向量
	"charglm-4", "emohaa", "codegeex-4",
	"embedding-2", "embedding-3",
}
