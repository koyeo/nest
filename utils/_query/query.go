package _query

import (
	"fmt"
	"gorm.io/gorm"
	"strings"
)

func NewBaseQuery(page, limit uint64, conditions []*QueryCondition, orders []*QueryOrder) (*BaseQuery, error) {
	q := &BaseQuery{
		Page:          page,
		Limit:         limit,
		AndConditions: conditions,
		Orders:        orders,
	}
	for _, v := range q.AndConditions {
		if err := v.Valid(); err != nil {
			return nil, err
		}
	}

	for _, v := range q.Orders {
		if err := v.Valid(); err != nil {
			return nil, err
		}
	}

	return q, nil
}

type BaseQuery struct {
	Page          uint64            `json:"page"`
	Limit         uint64            `json:"limit"`
	AndConditions []*QueryCondition `json:"condition,omitempty"` // 只支持 and 条件拼接, or 条件需要客户端自行拼接
	Orders        []*QueryOrder     `json:"sort,omitempty"`
}

func (q BaseQuery) Offset() uint64 {
	// TODO 检查逻辑
	return q.Limit * q.Page
}

func (q BaseQuery) Valid() error {
	if q.Page == 0 {
		return fmt.Errorf("query page expect > 1, got: %d", q.Page)
	}
	if q.Limit == 0 {
		return fmt.Errorf("query limit expect > 1, got: %d", q.Limit)
	}

	for _, v := range q.AndConditions {
		if err := v.Valid(); err != nil {
			return err
		}
	}

	for _, v := range q.Orders {
		if err := v.Valid(); err != nil {
			return err
		}
	}

	return nil
}

func (q BaseQuery) PrepareGormWheres(db *gorm.DB) (*gorm.DB, error) {
	for _, v := range q.AndConditions {
		if err := v.Valid(); err != nil {
			return nil, err
		}
		db = db.Where(fmt.Sprintf("%s %s ?", v.Field, v.Operation), v.Value)
	}
	return db, nil
}

type QueryCondition struct {
	Operation QueryConditionOperation `json:"operation"` // 为空时，默认为 = 号
	Field     string                  `json:"field"`
	Value     string                  `json:"value"`
}

func (q QueryCondition) Valid() error {
	if err := q.Operation.Valid(); err != nil {
		return err
	}
	if q.Field == "" {
		return fmt.Errorf("query condition.field require no empty")
	}
	if q.Value == "" {
		return fmt.Errorf("query condition.value require no empty")
	}
	return nil
}

type QueryConditionOperation string

func (q QueryConditionOperation) String() string {
	s := strings.TrimSpace(string(q))
	if s == "" {
		return "="
	}
	return s
}

func (q QueryConditionOperation) Valid() error {
	s := q.String()
	switch s {
	case "=", ">", "<", ">=", "<=", "like", "<>":
		return nil
	}
	return fmt.Errorf("invalid query condition operation: %s", s)
}

type QueryOrder struct {
	Field string `json:"field"`
	Sort  string `json:"sort"`
}

func (q QueryOrder) Valid() error {
	if q.Field == "" {
		return fmt.Errorf("query order.field require no empty")
	}
	if q.Sort != "asc" && q.Sort != "desc" {
		return fmt.Errorf("invalid query order.sort, expect asc or desc")
	}
	return nil
}
