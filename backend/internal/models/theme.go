package models

import "gorm.io/gorm"

type ThemeConfig struct {
	BaseModel
	Palette            string `json:"palette" gorm:"default:aozora"`
	BackgroundEnabled  bool   `json:"background_enabled" gorm:"default:false"`
	BackgroundImage    string `json:"background_image" gorm:"type:text"`
	Blur               int    `json:"blur" gorm:"default:0"`
	Brightness         int    `json:"brightness" gorm:"default:100"`
	MaskOpacity        int    `json:"mask_opacity" gorm:"default:34"`
	BackgroundSize     string `json:"background_size" gorm:"default:cover"`
	BackgroundPosition string `json:"background_position" gorm:"default:center"`
	CardOpacity        int    `json:"card_opacity" gorm:"default:72"`
	// DockImage 是侧边栏左下角时钟卡的背景长图（base64 或 URL），随主题配置保存。
	DockImage string `json:"dock_image" gorm:"type:text"`
}

func (t *ThemeConfig) AfterFind(tx *gorm.DB) (err error) {
	if t.Palette == "" {
		t.Palette = "aozora"
	}
	if t.BackgroundSize == "" {
		t.BackgroundSize = "cover"
	}
	if t.BackgroundPosition == "" {
		t.BackgroundPosition = "center"
	}
	return
}
