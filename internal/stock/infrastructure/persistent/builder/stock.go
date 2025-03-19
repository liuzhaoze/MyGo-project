package builder

import (
	"github.com/liuzhaoze/MyGo-project/common/util"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Stock struct {
	Id        []int64  `json:"id,omitempty"`
	ProductID []string `json:"product_id,omitempty"`
	Quantity  []int32  `json:"quantity,omitempty"`
	Version   []int64  `json:"version,omitempty"`

	// extend fields
	OrderBy       string `json:"order_by,omitempty"`
	ForUpdateLock bool   `json:"for_update_lock,omitempty"`
}

func NewStock() *Stock {
	return &Stock{}
}

func (s *Stock) FormatArg() (string, error) {
	return util.MarshalString(s)
}

func (s *Stock) Fill(db *gorm.DB) *gorm.DB {
	db = s.FillWhere(db)
	if s.OrderBy != "" {
		db = db.Order(s.OrderBy)
	}
	return db
}

func (s *Stock) FillWhere(db *gorm.DB) *gorm.DB {
	if len(s.Id) > 0 {
		db = db.Where("Id in (?)", s.Id)
	}
	if len(s.ProductID) > 0 {
		db = db.Where("product_id in (?)", s.ProductID)
	}
	if len(s.Version) > 0 {
		db = db.Where("Version in (?)", s.Version)
	}
	if len(s.Quantity) > 0 {
		db = s.fillQuantityGE(db)
	}
	if s.ForUpdateLock {
		db = db.Clauses(clause.Locking{Strength: clause.LockingStrengthUpdate})
	}
	return db
}

func (s *Stock) fillQuantityGE(db *gorm.DB) *gorm.DB {
	if len(s.Quantity) > 0 {
		db = db.Where("Quantity >= ?", s.Quantity)
	}
	return db
}

func (s *Stock) IDs(v ...int64) *Stock {
	s.Id = v
	return s
}

func (s *Stock) QuantityGE(v ...int32) *Stock {
	s.Quantity = v
	return s
}

func (s *Stock) ProductIDs(v ...string) *Stock {
	s.ProductID = v
	return s
}

func (s *Stock) Versions(v ...int64) *Stock {
	s.Version = v
	return s
}

func (s *Stock) Order(v string) *Stock {
	s.OrderBy = v
	return s
}

func (s *Stock) ForUpdate() *Stock {
	s.ForUpdateLock = true
	return s
}
