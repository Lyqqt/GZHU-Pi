package models

import (
	"github.com/jmoiron/sqlx/types"
	"gopkg.in/guregu/null.v3"
	"time"
)

type TComment struct {
	ID       int64    `json:"id,omitempty" remark:"自增id" gorm:"primary_key"`
	ObjectID int64       `json:"object_id,omitempty" remark:"主题对象记录ID" gorm:"type:real;not null"`
	Content  null.String `json:"content,omitempty" remark:"主体内容" gorm:"type:varchar"`
	//Mention  null.String `json:"mention,omitempty" remark:"提及@用户" gorm:"type:varchar"`
	ReplyID null.Int `json:"reply_id,omitempty" remark:"回复id" gorm:"type:real"` // reference to t_comment(id)

	Image     types.JSONText `json:"image,omitempty" remark:"图片地址[string]" gorm:"type:varchar[]"`
	Anonymous null.Bool      `json:"anonymous,omitempty" remark:"是否匿名" gorm:"type:bool;default:false"`
	Anonymity null.String    `json:"anonymity,omitempty" remark:"匿名/化名" gorm:"type:varchar"`

	Type      null.String    `json:"type,omitempty" remark:"留言类型" gorm:"type:varchar"`
	Addi      types.JSONText `json:"addi,omitempty" remark:"附加信息" gorm:"type:jsonb"`
	Status    null.Int       `json:"status,omitempty" remark:"状态" gorm:"type:real;default:0"`
	CreatedBy null.Int       `json:"created_by,omitempty" remark:"创建者" gorm:"type:real"`
	CreatedAt time.Time      `json:"created_at,omitempty" remark:"创建时间" gorm:"default:current_timestamp"`
}