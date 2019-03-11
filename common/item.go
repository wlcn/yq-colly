package common

// Item stores information about a item
type Item struct {
	// 信息源URL
	SourceURL string
	// 商品标题
	Title string
	// // 信息源截图
	// SourcePic string
	// 品牌
	Brand string
	// // 型号
	// Model string
	// // 洗衣机类型
	// TypeOFWasher string
	// // 洗涤桶数
	// WashingLoads string
	// // 面板
	// Surface string
	// // 内桶材质
	// CylinderMaterial string
	// // 洗涤容量（公斤）
	// WashCapacity string
	// // 洗涤转数每分钟
	// WashRevoPerMin string
	// // 洗涤噪音（分贝）
	// WashNoise string
	// // 脱水甩干功能
	// Dehydaration string
	// // 烘干一体
	// IntegratedDryer string
	// // 烘干方式
	// DryingMethod string
	// // 衣物投放口径（毫米）
	// Caliber string
	// // 除菌
	// Degerming string
	// // 自动化
	// Automation string
	// // 脱水容量（公斤）
	// DryingCapacity string
	// // 甩干转数每分钟
	// SpinRevoPerMin string
	// // 甩干噪音（分贝）
	// SpinNoise string
	// // 变频
	// Inverter string
	// // 显示屏
	// Screen string
	// // 控制方式
	// Control string
	// // 剩余时间显示
	// TimeRemaining string
	// // 预约
	// Booking string
	// // 自动添加洗涤剂
	// AutoAddDetergent string
	// // 自清洁
	// SelfCleaning string
	// // 儿童锁
	// ChildLock string
	// // 温水洗
	// WarmWash string
	// // 转速调节
	// SpeedAdjustment string
	// // 排水方式
	// Drain string
	// // 智能控制
	// SmartControl string
	// // 能效等级
	// EEI string
	// // 高（毫米）
	// Height string
	// // 宽（毫米）
	// Width string
	// // 厚（毫米）
	// Depth string
	// // 人工智能系统
	// AISystem string
	// 描述字段
	Description map[string]string
}
